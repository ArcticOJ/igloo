package worker

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/rueidis"
	"github.com/vmihailenco/msgpack/v5"
	"igloo/igloo/config"
	"igloo/igloo/judge/runtimes"
	"igloo/igloo/logger"
	"igloo/igloo/models"
	"runtime"
)

type (
	JudgeWorker struct {
		pool   []*JudgeRunner
		redis  rueidis.Client
		mqConn *amqp.Connection
		mqChan *amqp.Channel
		rmq    amqp.Queue
		mq     amqp.Queue
		ctx    context.Context
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
	runners := make([]*JudgeRunner, config.Config.Parallelism)
	for i := range runners {
		runners[i] = NewJudge(offset + i)
	}
	conf := config.Config.RabbitMQ
	conn, e := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d", conf.Username, conf.Password, conf.Host, conf.Port))
	logger.Panic(e, "failed to establish a connection to rabbitmq")
	ch, e := conn.Channel()
	logger.Panic(e, "failed to open a channel for queue")
	logger.Panic(ch.ExchangeDeclare("submissions", "direct", true, false, false, false, amqp.Table{
		"x-queue-type":       "quorum",
		"x-consumer-timeout": 3600000,
	}), "failed to declare exchange for submissions")
	q, e := ch.QueueDeclare(fmt.Sprintf("judge-worker-%s", config.Config.ID), true, false, false, false, amqp.Table{
		"x-consumer-timeout": 3600000,
	})
	logger.Panic(e, "failed to open queue for submissions")
	for rt := range runtimes.Runtimes {
		logger.Panic(ch.QueueBind(q.Name, rt, "submissions", false, nil), "could not register runtime '%s' to queue", rt)
		logger.Logger.Info().Msgf("binding '%s' to queue", rt)
	}
	rq, e := ch.QueueDeclare("results", false, false, false, false, amqp.Table{
		"x-single-active-consumer": true,
	})
	logger.Panic(e, "failed to open queue for results")
	logger.Panic(ch.Qos(1, 0, false), "error whilst setting QoS")
	logger.Logger.Info().Msgf("initializing igloo with %d concurrent runners", config.Config.Parallelism)
	return &JudgeWorker{
		ctx:    ctx,
		pool:   runners,
		mqConn: conn,
		mqChan: ch,
		mq:     q,
		rmq:    rq,
	}
}

func (w *JudgeWorker) PublishResult(ctx context.Context, headers map[string]interface{}, body interface{}) error {
	if b, e := msgpack.Marshal(body); e == nil {
		return w.mqChan.PublishWithContext(ctx,
			"",
			w.rmq.Name,
			true,
			false,
			amqp.Publishing{
				DeliveryMode: amqp.Transient,
				ContentType:  "application/msgpack",
				Headers:      headers,
				Body:         b,
			})
	} else {
		return e
	}
}

func (w *JudgeWorker) Consume(c <-chan amqp.Delivery, runner *JudgeRunner) {
	for s := range c {
		if runner.Busy() {
			s.Reject(true)
			continue
		}
		var sub models.Submission
		if msgpack.Unmarshal(s.Body, &sub) != nil {
			s.Reject(false)
			continue
		}
		judge := runner.Judge(&sub, w.ctx, func(caseId uint16, r models.CaseResult) bool {
			return w.PublishResult(context.Background(), map[string]interface{}{
				"submission-id": int64(sub.ID),
				"case-id":       int32(caseId),
				"ttl":           int(sub.Constraints.TimeLimit + 15),
			}, r) != nil
		})
		w.PublishResult(context.Background(), map[string]interface{}{
			"submission-id": int64(sub.ID),
			"type":          "final",
		}, judge())
		s.Ack(false)
	}
}

func (w *JudgeWorker) Work() {
	for i := range w.pool {
		// TODO: improve this, using a controllable pool
		runner := w.pool[i]
		c, e := w.mqChan.Consume(w.mq.Name, fmt.Sprintf("judge-worker-%s#%d", config.Config.ID, runner.boundCpu), false, false, false, false, nil)
		logger.Panic(e, "failed to create a consumer for submissions on cpu %d", runner.boundCpu)
		go w.Consume(c, runner)
	}
	<-w.ctx.Done()
}

func (w *JudgeWorker) StopAllContainers() {

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
