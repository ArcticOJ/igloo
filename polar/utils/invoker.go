package utils

import (
	"bytes"
	"context"
	"os/exec"
)

func Invoke(name string, arg ...string) ([]byte, error) {
	cmd := exec.CommandContext(context.Background(), name, arg...)

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Start(); err != nil {
		return buf.Bytes(), err
	}

	if err := cmd.Wait(); err != nil {
		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}
