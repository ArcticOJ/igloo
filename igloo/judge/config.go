package judge

import (
	"bytes"
	"io"
)

type ProcessType = string

const (
	Compiler ProcessType = "runtimes"
	Normal               = "normal"
	Python3              = "python3"
	Cpp                  = "cpp"
)

type Config struct {
	Target        string
	TimeLimit     uint32
	TimeLimitHard uint32
	MemoryLimit   uint64
	OutputLimit   uint64
	StackLimit    uint64
	OutErrPath    string
	Type          ProcessType
	Verbose       bool
	WorkDir       string
}

func (config *Config) getIO(input []byte) (inp io.Reader, out, err string) {
	out, err = config.OutErrPath+".OUT", config.OutErrPath+".ERR"
	if input != nil {
		inp = bytes.NewReader(input)
	}
	return
}
