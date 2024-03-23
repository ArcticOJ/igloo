package worker

import (
	"context"
	"github.com/ArcticOJ/igloo/v0/logger"
	"github.com/ArcticOJ/igloo/v0/runner"
	"github.com/ArcticOJ/polar/v0/pb"
	"sync/atomic"
)

type JudgeRunner struct {
	boundCpu uint16
	isBusy   atomic.Bool
	runner   *runner.Runner
}

func NewJudge(boundCpu uint16) (r *JudgeRunner) {
	r = &JudgeRunner{boundCpu: boundCpu}
	_r, e := runner.New(boundCpu)
	logger.Panic(e, "could not spawn runner for cpu %d", boundCpu)
	r.runner = _r
	return
}

func (j *JudgeRunner) Busy() bool {
	return j.isBusy.Load()
}

func (j *JudgeRunner) Judge(sub *pb.Submission, ctx context.Context, notifyAck func() bool, callback func(*pb.CaseResult) bool) func() *pb.FinalResult {
	j.isBusy.Store(true)
	return func() *pb.FinalResult {
		defer func() {
			_ = j.runner.Cleanup()
			j.isBusy.Store(false)
		}()
		rt := Runtimes[sub.Runtime]
		if !notifyAck() {
			return &pb.FinalResult{Verdict: pb.FinalVerdict_CANCELLED}
		}
		outPath, compOut, e := j.Compile(rt, sub, ctx)
		if e != nil {
			logger.Logger.Debug().Err(e).Interface("submission", sub).Msg("judging error")
			return &pb.FinalResult{Verdict: pb.FinalVerdict_COMPILATION_ERROR, CompilerOutput: compOut}
		}
		fv, e := j.Run(rt, sub, outPath, callback, ctx)
		if e != nil {
			logger.Logger.Error().Interface("submission", sub).Err(e).Msg("error judging submission")
		}
		return &pb.FinalResult{
			Verdict:        fv,
			CompilerOutput: compOut,
		}
	}
}

func (j *JudgeRunner) Destroy() error {
	return j.runner.Destroy()
}
