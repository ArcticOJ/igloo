package runner

import (
	"context"
	"github.com/ArcticOJ/igloo/v0/models"
	"github.com/ArcticOJ/igloo/v0/runner/shared"
	"github.com/ArcticOJ/igloo/v0/runtimes"
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
		Compile(rt *runtimes.Runtime, sourceCode string, ctx context.Context) (string, string, error)
		Cleanup() error
		Destroy() error
	}
)

func Convert(s runner.Status) models.CaseVerdict {
	switch s {
	case runner.StatusTimeLimitExceeded:
		return models.TimeLimitExceeded
	case runner.StatusMemoryLimitExceeded:
		return models.TimeLimitExceeded
	case runner.StatusOutputLimitExceeded:
		return models.OutputLimitExceeded
	case runner.StatusNonzeroExitStatus, runner.StatusSignalled, runner.StatusRunnerError:
		return models.RuntimeError
	default:
		return -1
	}
}
