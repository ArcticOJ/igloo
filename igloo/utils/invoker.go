package utils

import (
	"context"
	"io"
	"os/exec"
)

func InvokeStream(name string, args ...string) (io.WriteCloser, io.ReadCloser, io.ReadCloser, func() error, func() error) {
	cmd := exec.CommandContext(context.Background(), name, args...)
	{
		stdin, e := cmd.StdinPipe()
		if e != nil {
			goto quit
		}
		stdout, e := cmd.StdoutPipe()
		if e != nil {
			goto quit
		}
		stderr, e := cmd.StderrPipe()
		if e != nil {
			goto quit
		}
		defer cmd.Start()
		return stdin, stdout, stderr, cmd.Start, cmd.Cancel
	}
quit:
	cmd.Cancel()
	return nil, nil, nil, nil, nil
}

func Invoke(name string, args ...string) (string, error) {
	cmd := exec.CommandContext(context.Background(), name, args...)
	out, e := cmd.CombinedOutput()
	return string(out), e
}
