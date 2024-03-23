package worker

import (
	"context"
	"github.com/ArcticOJ/igloo/v0/checker"
	"github.com/ArcticOJ/igloo/v0/config"
	"github.com/ArcticOJ/igloo/v0/logger"
	"github.com/ArcticOJ/igloo/v0/runner"
	"github.com/ArcticOJ/igloo/v0/runtimes"
	"github.com/ArcticOJ/igloo/v0/utils"
	"github.com/ArcticOJ/polar/v0/pb"
	r "github.com/criyle/go-sandbox/runner"
	"os"
	"path"
	"strconv"
)

func (j *JudgeRunner) Compile(rt runtimes.Runtime, sub *pb.Submission, ctx context.Context) (out, compOut string, e error) {
	srcCode := path.Join(config.Config.Storage.Submissions, sub.SourcePath)
	return j.runner.Compile(rt, srcCode, ctx)
}

func (j *JudgeRunner) Run(rt runtimes.Runtime, sub *pb.Submission, prog string, callback func(*pb.CaseResult) bool, ctx context.Context) (pb.FinalVerdict, error) {
	cmd, args := rt.BuildExecCommand(prog)
	c := sub.Constraints
	// TODO: Support per-language time limit
	out, e := utils.CreateRandomFile("output-")

	isShortCircuit := pb.Submission_Constraints_SHORT_CIRCUIT.In(c.Flags)

	if e != nil {
		return pb.FinalVerdict_INITIALIZATION_ERROR, e
	}
	defer utils.Clean(out)
	judge, e := j.runner.Judge(append([]string{cmd}, args...), &runner.SubmissionConfig{
		MemoryLimit: c.MemoryLimit << 20,
		TimeLimit:   c.TimeLimit,
		StackLimit:  c.MemoryLimit << 20,
		OutputLimit: c.OutputLimit << 20,
		Verbose:     true,
		OutputFile:  out,
	}, ctx)
	if e != nil {
		return pb.FinalVerdict_INITIALIZATION_ERROR, e
	}
	for i := uint32(1); i <= sub.TestCount; i++ {
		casePath := path.Join(config.Config.Storage.Problems, sub.ProblemId, strconv.FormatUint(uint64(i), 10))
		res, e := judge(path.Join(casePath, "input.inp"))
		if res == nil || e != nil {
			callback(&pb.CaseResult{CaseId: i, Verdict: pb.CaseVerdict_INTERNAL_ERROR})
			continue
		}
		result := &pb.CaseResult{
			CaseId:        i,
			ExecutionTime: float32(res.Time.Microseconds() / 1000),
			Memory:        uint32(res.Memory.KiB()),
			Verdict:       runner.ConvertStatus(res.Status),
		}
		if e != nil {
			result.Verdict = pb.CaseVerdict_RUNTIME_ERROR
			if isShortCircuit {
				return pb.FinalVerdict_SHORT_CIRCUIT, nil
			}
			continue
		}
		if res.Status != r.StatusNormal {
			callback(result)
			if isShortCircuit {
				return pb.FinalVerdict_SHORT_CIRCUIT, nil
			}
		} else {
			//expOut, _err := cache.Get(ctx, sub.ProblemID, i, "out")
			expOut, e := os.Open(path.Join(casePath, "output.out"))
			if e != nil {
				logger.Logger.Err(e).Interface("submission", sub).Msg("error judging submission")
				result.Verdict = pb.CaseVerdict_INTERNAL_ERROR
				callback(result)
				continue
			}
			ok, msg := checker.Check(out, expOut)
			result.Message = msg
			if ok {
				result.Verdict = pb.CaseVerdict_ACCEPTED
				callback(result)
			} else {
				result.Verdict = pb.CaseVerdict_WRONG_ANSWER
				callback(result)
				if isShortCircuit {
					return pb.FinalVerdict_SHORT_CIRCUIT, nil
				}
			}
		}
	}
	return pb.FinalVerdict_NORMAL, nil
}
