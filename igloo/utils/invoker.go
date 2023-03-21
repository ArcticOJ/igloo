package utils

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

func Invoke(name string, args ...string) (string, error) {
	fmt.Println(name, args)
	cmd := exec.CommandContext(context.Background(), name, args...)

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Start(); err != nil {
		return buf.String(), err
	}

	if err := cmd.Wait(); err != nil {
		return buf.String(), err
	}

	return buf.String(), nil
}
