package utils

import (
	"bytes"
	"os/exec"
)

func InvokeStdout(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var b bytes.Buffer
	cmd.Stdout = &b
	e := cmd.Run()
	return b.String(), e
}
