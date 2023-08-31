package shared

type Config struct {
	Target      string
	TimeLimit   float32
	MemoryLimit uint32
	OutputLimit uint32
	StackLimit  uint32
	OutputFile  string
	Verbose     bool
}
