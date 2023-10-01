package worker

import (
	"context"
	"igloo/judge/runner"
	"igloo/judge/runtimes"
	"igloo/logger"
	"igloo/models"
	"sync/atomic"
)

type JudgeRunner struct {
	boundCpu uint8
	isBusy   atomic.Bool
	runner   runner.Runner
}

func NewJudge(boundCpu int) (r *JudgeRunner) {
	r = &JudgeRunner{boundCpu: uint8(boundCpu)}
	_r, e := runner.New(uint8(boundCpu))
	logger.Panic(e, "could not spawn runner for cpu %d", boundCpu)
	r.runner = _r
	return
}

func (jc *JudgeRunner) Busy() bool {
	return jc.isBusy.Load()
}

func (jc *JudgeRunner) Judge(sub *models.Submission, ctx context.Context, callback func(uint16, models.CaseResult) bool) func() models.FinalResult {
	jc.isBusy.Store(true)
	return func() models.FinalResult {
		defer func() {
			_ = jc.runner.Cleanup()
			jc.isBusy.Store(false)
		}()
		rt := runtimes.Runtimes[sub.Language]
		outPath, compOut, e := jc.Compile(rt, sub, ctx)
		if e != nil {
			return models.FinalResult{Verdict: models.CompileError, CompilerOutput: compOut}
		}
		fv, e := jc.Run(rt, sub, outPath, callback, ctx)
		return models.FinalResult{Verdict: fv, CompilerOutput: compOut}
	}
}

func (jc *JudgeRunner) Destroy() error {
	return jc.runner.Destroy()
}
