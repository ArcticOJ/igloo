package models

type (
	CaseVerdict  int8
	FinalVerdict int8
	CaseResult   struct {
		Duration float32
		Memory   uint32
		Message  string
		Verdict  CaseVerdict
	}
	FinalResult struct {
		CompilerOutput string
		Verdict        FinalVerdict
	}
)

const (
	Accepted CaseVerdict = iota
	WrongAnswer
	InternalError
	TimeLimitExceeded
	MemoryLimitExceeded
	OutputLimitExceeded
	RuntimeError
)

const (
	Normal FinalVerdict = iota
	ShortCircuit
	Rejected
	Cancelled
	CompileError
	InitError
)
