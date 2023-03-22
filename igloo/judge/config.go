package judge

import (
	"fmt"
)

type ProcessType = string

const (
	Compiler ProcessType = "compiler"
	Normal               = "normal"
	Python3              = "python3"
	Cpp                  = "cpp"
)

type Config struct {
	TimeLimit     uint64
	TimeLimitHard uint64
	MemoryLimit   uint64
	OutputLimit   uint64
	StackLimit    uint64
	IOFileName    string
	Type          ProcessType
	Verbose       bool
	WorkDir       string
}

func (config *Config) getIO() (input, output, err string) {
	if config.IOFileName == "" {
		return "", "", ""
	}
	file := config.IOFileName
	input = fmt.Sprintf("%s.INP", file)
	output = fmt.Sprintf("%s.OUT", file)
	err = fmt.Sprintf("%s.ERR", file)
	return
}
