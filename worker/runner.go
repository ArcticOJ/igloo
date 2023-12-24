package worker

import (
	"context"
	"github.com/ArcticOJ/igloo/v0/logger"
	"github.com/ArcticOJ/igloo/v0/runner"
	"github.com/ArcticOJ/polar/v0/types"
	"sync/atomic"
)

type JudgeRunner struct {
	boundCpu uint16
	isBusy   atomic.Bool
	runner   runner.Runner
}

func NewJudge(boundCpu uint16) (r *JudgeRunner) {
	r = &JudgeRunner{boundCpu: boundCpu}
	_r, e := runner.New(boundCpu)
	logger.Panic(e, "could not spawn runner for cpu %d", boundCpu)
	r.runner = _r
	return
}

func (jc *JudgeRunner) Busy() bool {
	return jc.isBusy.Load()
}

func (jc *JudgeRunner) Judge(sub types.Submission, ctx context.Context, notifyAck func() bool, callback func(types.CaseResult) bool) func() *types.FinalResult {
	jc.isBusy.Store(true)
	return func() *types.FinalResult {
		defer func() {
			_ = jc.runner.Cleanup()
			jc.isBusy.Store(false)
		}()
		rt := Runtimes[sub.Runtime]
		if !notifyAck() {
			return &types.FinalResult{Verdict: types.FinalVerdictCancelled}
		}
		outPath, compOut, e := jc.Compile(rt, sub, ctx)
		if e != nil {
			logger.Logger.Debug().Err(e).Interface("submission", sub).Msg("judging error")
			return &types.FinalResult{Verdict: types.FinalCompileError, CompilerOutput: compOut}
		}
		fv, p, e := jc.Run(rt, sub, outPath, callback, ctx)
		if e != nil {
			logger.Logger.Error().Interface("submission", sub).Err(e).Msg("error judging submission")
		}
		return &types.FinalResult{
			Verdict:        fv,
			CompilerOutput: compOut,
			Points:         p,
			MaxPoints:      float64(sub.TestCount) * sub.PointsPerTest,
		}
	}
}

func (jc *JudgeRunner) Destroy() error {
	return jc.runner.Destroy()
}
