package worker

import (
	"context"
	"github.com/ArcticOJ/igloo/v0/checker"
	"github.com/ArcticOJ/igloo/v0/config"
	"github.com/ArcticOJ/igloo/v0/models"
	"github.com/ArcticOJ/igloo/v0/runner"
	"github.com/ArcticOJ/igloo/v0/runner/shared"
	"github.com/ArcticOJ/igloo/v0/runtimes"
	"github.com/ArcticOJ/igloo/v0/utils"
	r "github.com/criyle/go-sandbox/runner"
	"os"
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
	for i := uint16(1); i <= sub.TestCount; i++ {
		casePath := path.Join(config.Config.Storage.Problems, sub.ProblemID, strconv.FormatUint(uint64(i), 10))
		res, e := judge(path.Join(casePath, "input.inp"))
		if res == nil {
			callback(i, models.CaseResult{Verdict: models.InternalError})
			continue
		}
		result := models.CaseResult{Duration: float32(res.Time.Microseconds() / 1000), Memory: uint32(res.Memory.KiB()), Verdict: runner.Convert(res.Status)}
		if e != nil {
			result.Verdict = models.RuntimeError
			callback(i, result)
			if c.ShortCircuit {
				return models.ShortCircuit, e
			}
			continue
		}
		if res.Status != r.StatusNormal {
			callback(i, result)
			if c.ShortCircuit {
				return models.ShortCircuit, nil
			}
		} else {
			//expOut, _err := cache.Get(ctx, sub.ProblemID, i, "out")
			expOut, e := os.Open(path.Join(casePath, "output.out"))
			if e != nil {
				result.Verdict = models.InternalError
				callback(i, result)
				continue
			}
			//_x, _ := io.ReadAll(expOut)
			//fmt.Println(string(_x))
			//fmt.Println()
			ok, msg := checker.Check(out, expOut)
			result.Message = msg
			if ok {
				result.Verdict = models.Accepted
				callback(i, result)
			} else {
				result.Verdict = models.WrongAnswer
				callback(i, result)
				if c.ShortCircuit {
					return models.ShortCircuit, nil
				}
			}
		}
	}
	return models.Normal, nil
}
