package runner

import (
	"context"
	"github.com/ArcticOJ/igloo/v0/runner/shared"
	"github.com/ArcticOJ/igloo/v0/runtimes"
	"github.com/ArcticOJ/polar/v0/types"
	"github.com/criyle/go-sandbox/runner"
)

type (
	Runner interface {
		Judge(args []string, config *shared.Config, ctx context.Context) (
			// judge func
			func(input string) (*runner.Result, error),
			// initialization error
			error,
		)
		Compile(rt runtimes.Runtime, sourceCode string, ctx context.Context) (string, string, error)
		Cleanup() error
		Destroy() error
	}
)

func Convert(s runner.Status) types.CaseVerdict {
	switch s {
	case runner.StatusTimeLimitExceeded:
		return types.CaseVerdictTimeLimitExceeded
	case runner.StatusMemoryLimitExceeded:
		return types.CaseVerdictTimeLimitExceeded
	case runner.StatusOutputLimitExceeded:
		return types.CaseVerdictOutputLimitExceeded
	case runner.StatusNonzeroExitStatus, runner.StatusSignalled, runner.StatusRunnerError:
		return types.CaseVerdictRuntimeError
	default:
		return -1
	}
}
