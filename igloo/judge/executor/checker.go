package checker

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
)

type (
	Checker struct {
		proc    *exec.Cmd
		Context context.Context
		stdin   *json.Encoder
		stdout  *json.Decoder
		stderr  *json.Decoder
		Compile func() (string, []string)
		dispose func()
	}
	CheckerInput struct {
		Actual   string `json:"actual"`
		Expected string `json:"expected"`
	}

	CheckerOutput struct {
		Status  int8   `json:"status"`
		Message string `json:"message"`
	}
)

func (c *Checker) Init() error {
	if c.proc != nil {
		_ = c.proc.Cancel()
		c.proc = nil
	}
	cmd, args := c.Compile()
	c.proc = exec.CommandContext(c.Context, cmd, args...)
	stdin, e := c.proc.StdinPipe()
	if e != nil {
		return e
	}
	stdout, e := c.proc.StderrPipe()
	if e != nil {
		return e
	}
	stderr, e := c.proc.StderrPipe()
	if e != nil {
		return e
	}
	c.stdin = json.NewEncoder(stdin)
	c.stdout = json.NewDecoder(stdout)
	c.stderr = json.NewDecoder(stderr)
	c.dispose = func() {
		stdin.Close()
		stdout.Close()
		stderr.Close()
	}
	return nil
}

func (c *Checker) Dispose() {
	c.dispose()
	c.proc.Cancel()
	c.proc = nil
	c.stdin = nil
	c.stdout = nil
	c.stderr = nil
}

func (c *Checker) Check(ctx context.Context, output, expected []byte) (error, bool) {
	if c.proc == nil {
		return errors.New("uninitialized"), false
	}
	//c.stdin.Encode(&Input{Type: ProcOutput, Data: output})
	return nil, false
}
