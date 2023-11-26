package worker

import (
	"context"
	"fmt"
	"github.com/ArcticOJ/igloo/v0/config"
	"github.com/ArcticOJ/igloo/v0/logger"
	"github.com/ArcticOJ/igloo/v0/runner"
	"github.com/ArcticOJ/igloo/v0/sys"
	_ "github.com/ArcticOJ/igloo/v0/sys"
	polar "github.com/ArcticOJ/polar/v0/client"
	"github.com/ArcticOJ/polar/v0/types"
	"math"
	"runtime"
	"slices"
	"strings"
	"sync/atomic"
)

type (
	JudgeWorker struct {
		pool []*_runner
		// TODO: implement auto reconnecting
		p   *polar.Polar
		ctx context.Context
	}

	_runner struct {
		*JudgeRunner
		currentSubmission atomic.Uint32
		ctx               context.Context
		cancel            func()
	}
)

func New(ctx context.Context) (w *JudgeWorker) {
	w = &JudgeWorker{
		ctx: ctx,
	}
	maxCpus := uint16(runtime.NumCPU())
	// remove invalid cpus
	config.Config.CPUs = slices.DeleteFunc(config.Config.CPUs, func(cpu uint16) bool {
		return cpu >= maxCpus
	})
	slices.Sort(config.Config.CPUs)
	config.Config.Parallelism = uint16(len(config.Config.CPUs))
	if maxCpus/2 < config.Config.Parallelism {
		logger.Logger.Warn().Msg("running with more than 50% logical cores is not recommended.")
	}

	w.Connect()

	w.pool = make([]*_runner, config.Config.Parallelism)
	for i := range w.pool {
		_ctx, cancel := context.WithCancel(ctx)
		w.pool[i] = &_runner{
			JudgeRunner: NewJudge(config.Config.CPUs[i]),
			ctx:         _ctx,
			cancel:      cancel,
		}
		w.pool[i].currentSubmission.Store(math.MaxUint32)
	}
	logger.Logger.Info().Uints16("cpus", config.Config.CPUs).Uint16("parallelism", config.Config.Parallelism).Msgf("initializing igloo")
	return
}

func (w *JudgeWorker) WaitForSignal() {
	for {
		select {
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
	j := types.Judge{
		Name:        config.Config.ID,
		BootedSince: sys.BootTimestamp.Unix(),
		OS:          sys.OS,
		Memory:      sys.Memory,
		Parallelism: config.Config.Parallelism,
		Version:     "0.0.1-prealpha",
	}
	for name, rt := range Runtimes {
		j.Runtimes = append(j.Runtimes, types.Runtime{
			ID:        name,
			Compiler:  rt.Program,
			Arguments: rt.Arguments,
			Version:   rt.Version,
		})
	}
	slices.SortStableFunc(j.Runtimes, func(a, b types.Runtime) int {
		if a.ID == b.ID {
			return 0
		}
		if a.ID > b.ID {
			return 1
		}
		return -1
	})
	w.p, e = polar.New(w.ctx, config.Config.Polar.Host, config.Config.Polar.Port, j)
	logger.Panic(e, "error creating new polar instance")
	logger.Logger.Info().Msg("successfully connected to polar")
}

func (w *JudgeWorker) Consume(r *_runner) {
	c, e := w.p.NewConsumer()
	logger.Panic(e, "error creating consumer for runner #%d", r.boundCpu)
	go func() {
		logger.Panic(c.Consume(), "error creating consumer for runner #%d", r.boundCpu)
	}()
	for {
		select {
		case <-w.ctx.Done():
			return
		case s := <-c.MessageChan:
			logger.Logger.Debug().Interface("submission", s).Msg("received submission")
			w.Judge(r, s)
		}
	}
}

func (w *JudgeWorker) Judge(r *_runner, sub types.Submission) {
	// TODO: ensure that RAM is sufficient to handle submissions
	prod, e := w.p.NewProducer(sub.ID)
	defer prod.Close()
	if e != nil {
		//w.consumer.Reject(sub.ID)
		return
	}
	r.currentSubmission.Store(sub.ID)
	// TODO: immediately halt runner when getting error
	judge := r.Judge(sub, r.ctx, func(caseId uint16) bool {
		return prod.Report(types.ResultAnnouncement, caseId) == nil
	}, func(r types.CaseResult) bool {
		return prod.Report(types.ResultCase, r) == nil
	})
	finalResult := judge()
	if finalResult != nil {
		// replace actual new line characters with \\n to avoid shattered payloads when serializing response with msgpack
		finalResult.CompilerOutput = strings.ReplaceAll(finalResult.CompilerOutput, "\n", "\\n")
	}
	prod.Report(types.ResultFinal, finalResult)
}

func (w *JudgeWorker) Work() {
	for i := range w.pool {
		go w.Consume(w.pool[i])
	}
	w.WaitForSignal()
}

func (w *JudgeWorker) Destroy() {
	w.p.Close()
	for i, j := range w.pool {
		if j != nil {
			fmt.Printf("destroying %d\n", i)
			_ = j.Destroy()
		}
	}
	runner.Destroy()
}
