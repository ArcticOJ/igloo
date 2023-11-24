package worker

import (
	"context"
	"github.com/ArcticOJ/igloo/v0/checker"
	"github.com/ArcticOJ/igloo/v0/config"
	"github.com/ArcticOJ/igloo/v0/logger"
	"github.com/ArcticOJ/igloo/v0/runner"
	"github.com/ArcticOJ/igloo/v0/runner/shared"
	"github.com/ArcticOJ/igloo/v0/runtimes"
	"github.com/ArcticOJ/igloo/v0/utils"
	"github.com/ArcticOJ/polar/v0/types"
	r "github.com/criyle/go-sandbox/runner"
	"os"
	"path"
	"strconv"
)

func (jc *JudgeRunner) Compile(rt runtimes.Runtime, sub types.Submission, ctx context.Context) (out, compOut string, e error) {
	srcCode := path.Join(config.Config.Storage.Submissions, sub.SourcePath)
	return jc.runner.Compile(rt, srcCode, ctx)
}

func (jc *JudgeRunner) Run(rt runtimes.Runtime, sub types.Submission, announce func(caseId uint16) bool, prog string, callback func(types.CaseResult) bool, ctx context.Context) (types.FinalVerdict, float64, error) {
	cmd, args := rt.BuildExecCommand(prog)
	c := sub.Constraints
	// TODO: implement per language time limit
	// TODO: store output file inside container to run checker there
	out, e := utils.CreateRandomFile("output-")
	var p float64 = 0
	if e != nil {
		return types.FinalVerdictInitializationError, p, e
	}
	defer utils.Clean(out)
	judge, e := jc.runner.Judge(append([]string{cmd}, args...), &shared.Config{
		MemoryLimit: c.MemoryLimit << 20,
		TimeLimit:   c.TimeLimit,
		StackLimit:  c.MemoryLimit << 20,
		OutputLimit: c.OutputLimit << 20,
		Verbose:     true,
		OutputFile:  out,
	}, ctx)
	if e != nil {
		return types.FinalVerdictInitializationError, p, e
	}
	// TODO: stop judging once context is done.
	for i := uint16(1); i <= sub.TestCount; i++ {
		announce(i)
		casePath := path.Join(config.Config.Storage.Problems, sub.ProblemID, strconv.FormatUint(uint64(i), 10))
		res, e := judge(path.Join(casePath, "input.inp"))
		if res == nil {
			callback(types.CaseResult{CaseID: i, Verdict: types.CaseVerdictInternalError})
			continue
		}
		result := types.CaseResult{CaseID: i, Duration: float32(res.Time.Microseconds() / 1000), Memory: uint32(res.Memory.KiB()), Verdict: runner.Convert(res.Status)}
		if e != nil {
			result.Verdict = types.CaseVerdictRuntimeError
			callback(result)
			if c.ShortCircuit {
				return types.FinalVerdictShortCircuit, p, e
			}
			continue
		}
		if res.Status != r.StatusNormal {
			callback(result)
			if c.ShortCircuit {
				return types.FinalVerdictShortCircuit, p, nil
			}
		} else {
			//expOut, _err := cache.Get(ctx, sub.ProblemID, i, "out")
			expOut, e := os.Open(path.Join(casePath, "output.out"))
			if e != nil {
				logger.Logger.Err(e).Interface("submission", sub).Msg("error judging submission")
				result.Verdict = types.CaseVerdictInternalError
				callback(result)
				continue
			}
			ok, msg := checker.Check(out, expOut)
			result.Message = msg
			if ok {
				result.Verdict = types.CaseVerdictAccepted
				if sub.Constraints.AllowPartial {
					p += sub.PointsPerTest
				}
				callback(result)
			} else {
				result.Verdict = types.CaseVerdictWrongAnswer
				callback(result)
				if c.ShortCircuit {
					return types.FinalVerdictShortCircuit, p, nil
				}
			}
		}
	}
	return types.FinalVerdictNormal, p, nil
}
