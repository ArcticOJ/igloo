// TODO: Add support for interactive problems, non-standard I/O (e.g to files)
// TODO: Compress and cache test cases to save disk space
// TODO: Enforce I/O throttling to ensure timing consistency among runners

package worker

import (
	"cmp"
	"context"
	"errors"
	"github.com/ArcticOJ/igloo/v0/build"
	"github.com/ArcticOJ/igloo/v0/config"
	"github.com/ArcticOJ/igloo/v0/logger"
	"github.com/ArcticOJ/igloo/v0/runner"
	"github.com/ArcticOJ/igloo/v0/sys"
	_ "github.com/ArcticOJ/igloo/v0/sys"
	polar "github.com/ArcticOJ/polar/v0/client"
	"github.com/ArcticOJ/polar/v0/pb"
	runner2 "github.com/criyle/go-sandbox/runner"
	"math"
	"runtime"
	"slices"
	"sync/atomic"
	"time"
)

type (
	JudgeWorker struct {
		pool        []*_runner
		p           *polar.Polar
		ctx         context.Context
		runnerQueue chan uint16
	}

	_runner struct {
		*JudgeRunner
		id                uint16
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
	// Remove invalid CPUs
	config.Config.CPUs = slices.DeleteFunc(config.Config.CPUs, func(cpu uint16) bool {
		return cpu >= maxCpus
	})
	slices.Sort(config.Config.CPUs)
	config.Config.Parallelism = uint16(len(config.Config.CPUs))
	if maxCpus/2 < config.Config.Parallelism {
		logger.Logger.Warn().Msg("running with more than 50% logical cores is not recommended.")
	}
	w.Connect()
	w.runnerQueue = make(chan uint16, config.Config.Parallelism)
	w.pool = make([]*_runner, config.Config.Parallelism)
	for i := range w.pool {
		_ctx, cancel := context.WithCancel(ctx)
		w.pool[i] = &_runner{
			JudgeRunner: NewJudge(config.Config.CPUs[i]),
			ctx:         _ctx,
			cancel:      cancel,
			id:          uint16(i),
		}
		w.pool[i].currentSubmission.Store(math.MaxUint32)
		w.runnerQueue <- uint16(i)
	}
	logger.Logger.Info().Uints16("cpus", config.Config.CPUs).Uint16("parallelism", config.Config.Parallelism).Msgf("initializing igloo")
	return
}

func (w *JudgeWorker) Connect() {
	var e error
	j := &pb.Judge{
		Name:        config.Config.ID,
		BootedSince: sys.BootTimestamp.UTC().Unix(),
		Os:          sys.OS,
		Memory:      sys.Memory,
		Parallelism: uint32(config.Config.Parallelism),
		Version:     build.Version,
	}
	for name, rt := range Runtimes {
		j.Runtimes = append(j.Runtimes, &pb.Judge_Runtime{
			Id:        name,
			Compiler:  rt.Program,
			Arguments: rt.Arguments,
			Version:   rt.Version,
		})
	}
	slices.SortStableFunc(j.Runtimes, func(a, b *pb.Judge_Runtime) int {
		return cmp.Compare(a.Id, b.Id)
	})
	w.p, e = polar.New(w.ctx, j)
	logger.Panic(e, "error creating new polar instance")
	logger.Logger.Info().Msg("successfully connected to polar")
}

func (w *JudgeWorker) Judge(r *_runner, sub *pb.Submission) {
	// TODO: Avoid system crash by checking whether current submission may exceed current available memory?
	prod, e := w.p.NewProducer(sub.Id, r.ctx)
	defer prod.Close()
	if e != nil {
		// prod.Close will be invoked and this submission will be marked as rejected
		return
	}
	// bind this submission to this runner
	r.currentSubmission.Store(sub.Id)
	judge := r.Judge(sub, r.ctx, func() bool {
		logger.Logger.Debug().Uint32("id", sub.Id).Uint16("runner_id", r.id).Msg("submission acknowledged")
		return prod.Report(&pb.Result{
			Data: &pb.Result_None{},
		}) == nil
	}, func(res *pb.CaseResult) bool {
		logger.Logger.Debug().
			Uint32("case_id", res.CaseId).
			Stringer("execution_time", time.Duration(res.ExecutionTime)*time.Millisecond).
			// Since `res.Memory` is represented in KB, we have to cast it into byte
			Stringer("memory", runner2.Size(res.Memory*1024)).
			Str("message", res.Message).
			Stringer("verdict", res.Verdict).
			Uint32("id", sub.Id).
			Uint16("runner_id", r.id).
			Msg("case result")
		return prod.Report(&pb.Result{
			Data: &pb.Result_Case{
				Case: res,
			},
		}) == nil
	})
	finalResult := judge()
	if finalResult != nil {
		logger.Logger.Debug().
			Str("compiler_output", finalResult.CompilerOutput).
			Stringer("verdict", finalResult.Verdict).
			Uint16("runner_id", r.id).
			Uint32("id", sub.Id).Msg("final result")
	}
	prod.Report(&pb.Result{
		Data: &pb.Result_Final{
			Final: finalResult,
		},
	})
}

func (w *JudgeWorker) Work() {
	for {
		select {
		case <-w.ctx.Done():
			for i := range w.pool {
				w.pool[i].cancel()
			}
			return
		case id := <-w.runnerQueue:
			logger.Logger.Debug().Uint16("id", id).Msg("runner available")
			sub, e := w.p.Consume()
			if errors.Is(e, context.Canceled) {
				logger.Logger.Fatal().Msg("remote peer closed connection")
				return
			}
			if sub == nil {
				logger.Logger.Panic().Msg("could not consume submission")
				return
			}
			logger.Logger.Debug().Interface("submission", sub).Msg("received submission")
			go func() {
				w.Judge(w.pool[id], sub)
				// Release runner back to queue
				w.runnerQueue <- id
			}()
		}
	}
}

func (w *JudgeWorker) Destroy() {
	w.p.Close()
	for i, j := range w.pool {
		if j != nil {
			logger.Logger.Info().Msgf("destroying runner #%d", i)
			_ = j.Destroy()
		}
	}
	runner.Destroy()
}
