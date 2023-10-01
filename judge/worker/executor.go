package worker

import (
	"context"
	"fmt"
	r "github.com/criyle/go-sandbox/runner"
	"igloo/config"
	"igloo/judge/checker"
	"igloo/judge/runner"
	"igloo/judge/runner/shared"
	"igloo/judge/runtimes"
	"igloo/models"
	"igloo/utils"
	"path"
	"strconv"
)

func (jc *JudgeRunner) Compile(rt *runtimes.Runtime, sub *models.Submission, ctx context.Context) (out, compOut string, e error) {
	srcCode := path.Join(config.Config.Storage.Submissions, sub.SourcePath)
	return jc.runner.Compile(rt, srcCode, ctx)
}

func (jc *JudgeRunner) Run(rt *runtimes.Runtime, sub *models.Submission, prog string, callback func(uint16, models.CaseResult) bool, ctx context.Context) (models.FinalVerdict, error) {
	cmd, args := rt.BuildExecCommand(prog)
	c := sub.Constraints
	// TODO: implement per language time limit
	// TODO: store output file inside container to run checker there
	out, e := utils.CreateRandomFile("output-")
	if e != nil {
		return models.InitError, e
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
		return models.InitError, e
	}
	// TODO: stop judging once context is done.
	for i := uint16(0); i < sub.TestCount; i++ {
		casePath := path.Join(config.Config.Storage.Problems, sub.ProblemID)
		caseId := strconv.Itoa(int(i + 1))
		inp, expectedOut := path.Join(casePath, caseId, fmt.Sprintf("%s.inp", sub.ProblemID)), path.Join(casePath, caseId, fmt.Sprintf("%s.out", sub.ProblemID))
		res, _e := judge(inp)
		result := models.CaseResult{Duration: float32(res.Time.Microseconds() / 1000), Memory: uint32(res.Memory.KiB()), Verdict: runner.Convert(res.Status)}
		if _e != nil {
			result.Verdict = models.RuntimeError
			callback(i, result)
			if c.ShortCircuit {
				return models.ShortCircuit, _e
			}
			continue
		}
		if res == nil {
			callback(i, models.CaseResult{Verdict: models.InternalError})
			continue
		}
		if res.Status != r.StatusNormal {
			callback(i, result)
			if c.ShortCircuit {
				return models.ShortCircuit, _e
			}
		} else {
			ok, msg := checker.Check(out, expectedOut)
			result.Message = msg
			if ok {
				result.Verdict = models.Accepted
				callback(i, result)
			} else {
				result.Verdict = models.WrongAnswer
				callback(i, result)
				if c.ShortCircuit {
					return models.ShortCircuit, _e
				}
			}
		}
	}
	return models.Normal, nil
}
