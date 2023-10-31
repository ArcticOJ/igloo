package runner

import (
	runner "github.com/ArcticOJ/igloo/v0/runner/linux"
	"github.com/criyle/go-sandbox/container"
)

func init() {
	_ = container.Init()
}

func New(cpu uint16) (Runner, error) {
	return runner.New(cpu)
}
