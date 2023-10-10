package worker

import (
	"context"
	"fmt"
	"github.com/ArcticOJ/igloo/v0/config"
	"github.com/ArcticOJ/igloo/v0/logger"
	"github.com/ArcticOJ/igloo/v0/models"
	"github.com/ArcticOJ/igloo/v0/runtimes"
	"github.com/ArcticOJ/igloo/v0/sys"
	_ "github.com/ArcticOJ/igloo/v0/sys"
	amqp "github.com/rabbitmq/amqp091-go"
	amqp2 "github.com/rabbitmq/rabbitmq-stream-go-client/pkg/amqp"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/stream"
	"github.com/vmihailenco/msgpack/v5"
	"math"
	"net"
	"runtime"
	"sync/atomic"
	"time"
)

type (
	JudgeWorker struct {
		pool   []*_runner
		rc     <-chan amqp.Return
		ec     <-chan *amqp.Error
		mqConn *amqp.Connection
		mqChan *amqp.Channel
		mq     amqp.Queue
		env    *stream.Environment
		ctx    context.Context
	}

	_runner struct {
		*JudgeRunner
		currentSubmission atomic.Uint32
		c                 <-chan amqp.Delivery
		ctx               context.Context
		cancel            func()
	}
)

func New(ctx context.Context) *JudgeWorker {
	maxCpus := runtime.NumCPU()
	if config.Config.Parallelism == -1 || int(config.Config.Parallelism) > maxCpus {
		config.Config.Parallelism = int16(maxCpus / 2)
		logger.Logger.Info().Msgf("automatically allocating %d cores for workers", config.Config.Parallelism)
	} else if int16(maxCpus/2) < config.Config.Parallelism {
		logger.Logger.Warn().Msg("running with more than 50% logical cores is not recommended.")
	}
	offset := maxCpus - int(config.Config.Parallelism)
	runners := make([]*_runner, config.Config.Parallelism)
	for i := range runners {
		_ctx, cancel := context.WithCancel(ctx)
		runners[i] = &_runner{
			JudgeRunner: NewJudge(offset + i),
			c:           make(<-chan amqp.Delivery, 1),
			ctx:         _ctx,
			cancel:      cancel,
		}
		runners[i].currentSubmission.Store(math.MaxUint32)
	}
	logger.Logger.Info().Msgf("initializing igloo with %d concurrent runner(s)", config.Config.Parallelism)
	return &JudgeWorker{
		ctx:  ctx,
		pool: runners,
	}
}

func (w *JudgeWorker) WaitForSignal() {
	for {
		select {
		case e := <-w.ec:
			logger.Logger.Debug().Err(e).Msg("reconnect")
			w.Connect()
		case d := <-w.rc:
			fmt.Println(d)
		case <-w.ctx.Done():
			for i := range w.pool {
				w.pool[i].cancel()
			}
			return
		}
	}
}

func (w *JudgeWorker) Connect() {
	var e error
	if w.mqConn != nil {
		w.mqConn.Close()
		w.mqChan.Close()
	}
	conf := config.Config.RabbitMQ
	w.mqConn, e = amqp.DialConfig(fmt.Sprintf("amqp://%s:%s@%s", conf.Username, conf.Password, net.JoinHostPort(conf.Host, fmt.Sprint(conf.Port))), amqp.Config{
		Heartbeat: time.Second,
		Vhost:     conf.VHost,
	})
	logger.Panic(e, "failed to establish a connection to rabbitmq")
	w.mqChan, e = w.mqConn.Channel()
	logger.Panic(e, "failed to open a channel for queue")
	logger.Panic(w.mqChan.ExchangeDeclare("submissions", "direct", true, false, false, false, amqp.Table{
		"x-consumer-timeout": 3600000,
	}), "failed to declare exchange for submissions")
	w.mq, e = w.mqChan.QueueDeclare(fmt.Sprintf("judge-worker-%s-%d", config.Config.ID, time.Now().UTC().UnixMilli()), true, false, true, false, amqp.Table{
		"Name":        config.Config.ID,
		"BootedSince": sys.BootTimestamp,
		"OS":          sys.OS,
		"Memory":      int64(sys.Memory),
		"Parallelism": config.Config.Parallelism,
		"Version":     "0.0.1-prealpha",
	})
	logger.Panic(e, "failed to open queue for submissions")
	for name, _rt := range runtimes.Runtimes {
		rt := _rt
		logger.Panic(w.mqChan.QueueBind(w.mq.Name, name, "submissions", false, amqp.Table{
			"Version":   rt.Version,
			"Compiler":  rt.Program,
			"Arguments": rt.Arguments,
		}), "could not register runtime '%s' to queue", name)
		logger.Logger.Debug().Msgf("binding '%s' to queue", name)
	}
	logger.Panic(w.mqChan.Qos(int(config.Config.Parallelism), 0, true), "error whilst setting QoS")
	for i := range w.pool {
		w.pool[i].c, e = w.mqChan.Consume(w.mq.Name, fmt.Sprintf("%s#%d", w.mq.Name, i), false, false, false, false, nil)
		logger.Panic(e, "could not create a consumer for runner #%d", i)
	}
	w.rc = w.mqChan.NotifyReturn(make(chan amqp.Return, 1))
	w.ec = w.mqConn.NotifyClose(make(chan *amqp.Error, 1))
}

func (w *JudgeWorker) CreateStream() {
	var e error
	conf := config.Config.RabbitMQ
	w.env, e = stream.NewEnvironment(
		stream.NewEnvironmentOptions().
			SetHost(conf.Host).
			SetPort(int(conf.StreamPort)).
			SetUser(conf.Username).
			SetPassword(conf.Password).
			SetVHost(conf.VHost))
	if e != nil {
		logger.Panic(e, "failed to start a stream connection")
	}
}

func (w *JudgeWorker) PublishResult(prod *stream.Producer, headers map[string]interface{}, body interface{}) error {
	if b, e := msgpack.Marshal(map[string]interface{}{
		"Headers": headers,
		"Body":    body,
	}); e == nil {
		return prod.Send(amqp2.NewMessage(b))
	} else {
		return e
	}
}

func (w *JudgeWorker) Consume(r *_runner) {
	for {
		select {
		case <-w.ctx.Done():
			return
		case d := <-r.c:
			if d.CorrelationId != "" {
				logger.Logger.Debug().Str("id", d.CorrelationId).Msg("received submission")
				w.Judge(r, d)
			}
		}
	}
}

func (w *JudgeWorker) Judge(r *_runner, d amqp.Delivery) {
	// TODO: ensure that ram is adequate to handle submissions
	if r.Busy() {
		d.Reject(true)
		return
	}
	var sub models.Submission
	if msgpack.Unmarshal(d.Body, &sub) != nil {
		d.Reject(false)
		return
	}
	prod, e := w.env.NewProducer(d.ReplyTo, stream.NewProducerOptions().SetProducerName(fmt.Sprintf("judge-%s#%d", config.Config.ID, r.boundCpu)))
	if e != nil {
		d.Reject(true)
		return
	}
	defer prod.Close()
	r.currentSubmission.Store(sub.ID)
	judge := r.Judge(&sub, r.ctx, func(caseId uint16, r models.CaseResult) bool {
		return w.PublishResult(prod, map[string]interface{}{
			"from":    config.Config.ID,
			"case-id": int32(caseId),
			"ttl":     int(sub.Constraints.TimeLimit + 15),
		}, r) != nil
	})
	w.PublishResult(prod, map[string]interface{}{
		"from": config.Config.ID,
		"type": "final",
	}, judge())
	d.Ack(false)
}

func (w *JudgeWorker) Work() {
	w.Connect()
	w.CreateStream()
	for i := range w.pool {
		// TODO: improve this a bit
		go w.Consume(w.pool[i])
	}
	w.WaitForSignal()
}

func (w *JudgeWorker) Destroy() {
	if w.mqConn != nil {
		w.mqConn.Close()
	}
	for i, j := range w.pool {
		if j != nil {
			fmt.Printf("destroying %d\n", i)
			_ = j.Destroy()
		}
	}
}
