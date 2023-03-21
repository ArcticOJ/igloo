package judge

import (
	"context"
	"github.com/criyle/go-sandbox/container"
	"github.com/criyle/go-sandbox/pkg/rlimit"
)

type Judge struct {
	env container.Environment
}

func InitEnv() *Judge {
	builder := &container.Builder{
		ContainerGID: 0,
		ContainerUID: 1000,
		HostName:     "igloo-sandbox",
		DomainName:   "igloo-sandbox",
	}
	if env, e := builder.Build(); e != nil {
		return nil
	} else {
		return &Judge{env}
	}
}

func (judge *Judge) Exec(ctx context.Context) {
	limits := &rlimit.RLimits{}
	judge.env.Execve(
		ctx,
		container.ExecveParam{RLimits: limits.PrepareRLimit()},
	)
}
