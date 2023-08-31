package runner

import (
	"github.com/criyle/go-sandbox/container"
	runner "igloo/igloo/judge/runner/linux"
)

func init() {
	_ = container.Init()
}

func New(cpu uint8) (Runner, error) {
	return runner.New(cpu)
}
