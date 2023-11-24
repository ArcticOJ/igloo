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

func (jc *JudgeRunner) Judge(sub types.Submission, ctx context.Context, announce func(caseId uint16) bool, callback func(types.CaseResult) bool) func() *types.FinalResult {
	jc.isBusy.Store(true)
	return func() *types.FinalResult {
		defer func() {
			_ = jc.runner.Cleanup()
			jc.isBusy.Store(false)
		}()
		rt := Runtimes[sub.Runtime]
		if !announce(0) {
			return nil
		}
		outPath, compOut, e := jc.Compile(rt, sub, ctx)
		if e != nil {
			return &types.FinalResult{Verdict: types.FinalCompileError, CompilerOutput: compOut}
		}
		fv, p, e := jc.Run(rt, sub, announce, outPath, callback, ctx)
		if e != nil {

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
