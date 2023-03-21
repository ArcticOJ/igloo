package config

// ProgramConfig defines the extra config apply to program type
type ProgramConfig struct {
	Syscall    SyscallConfig
	FileAccess FileAccessConfig
}

// SyscallConfig defines extra syscallConfig apply to program type
type SyscallConfig struct {
	ExtraAllow, ExtraBan []string
	ExtraCount           map[string]int
}

// FileAccessConfig defines extra file access permission for the program type
type FileAccessConfig struct {
	ExtraRead, ExtraWrite, ExtraStat, ExtraBan []string
}
