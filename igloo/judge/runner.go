package judge

import (
	"context"
	"fmt"
	"github.com/criyle/go-sandbox/pkg/memfd"
	"igloo/igloo/judge/config"
	"igloo/igloo/utils"
	"os"
	"os/signal"
	"time"

	"github.com/criyle/go-sandbox/pkg/rlimit"
	"github.com/criyle/go-sandbox/pkg/seccomp/libseccomp"
	"github.com/criyle/go-sandbox/runner"
	"github.com/criyle/go-sandbox/runner/ptrace"
)

const pathEnv, userEnv, hostnameEnv = "PATH=/usr/local/bin:/usr/bin:/bin", "USER=igloo", "hostname=igloo-sandbox"

type Status int

// UOJ run_program constants
const (
	StatusNormal Status = iota
	StatusInvalid
	StatusRE
	StatusMLE
	StatusTLE
	StatusOLE
	StatusBan
	StatusFatal
	StatusRunnerError
)

func getStatus(s runner.Status) int {
	switch s {
	case runner.StatusNormal:
		return int(StatusNormal)
	case runner.StatusInvalid:
		return int(StatusInvalid)
	case runner.StatusTimeLimitExceeded:
		return int(StatusTLE)
	case runner.StatusMemoryLimitExceeded:
		return int(StatusMLE)
	case runner.StatusOutputLimitExceeded:
		return int(StatusOLE)
	case runner.StatusDisallowedSyscall:
		return int(StatusBan)
	case runner.StatusSignalled, runner.StatusNonzeroExitStatus:
		return int(StatusRE)
	default:
		return int(StatusFatal)
	}
}

type Instance struct {
	config *Config
}

func New(conf *Config) (*Instance, error) {
	return &Instance{
		config: conf,
	}, nil
}

func (instance *Instance) Judge(input []byte, argv []string) (*runner.Result, error) {
	args, allow, trace, h := config.GetConf(instance.config.Type, instance.config.WorkDir, argv, []string{}, []string{}, false)
	inp, outFile, errFile := instance.config.getIO(input)
	files, err := prepareFiles(inp, outFile, errFile)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare files: %v", err)
	}
	defer closeFiles(files)

	fds := make([]uintptr, len(files))
	for i, f := range files {
		if f != nil {
			fds[i] = f.Fd()
		} else {
			fds[i] = uintptr(i)
		}
	}

	rlims := rlimit.RLimits{
		CPU:         uint64(instance.config.TimeLimit),
		CPUHard:     uint64(instance.config.TimeLimitHard),
		FileSize:    instance.config.OutputLimit,
		Stack:       instance.config.StackLimit,
		Data:        instance.config.MemoryLimit,
		OpenFile:    256,
		DisableCore: true,
	}
	actionDefault := libseccomp.ActionKill
	if instance.config.Verbose {
		actionDefault = libseccomp.ActionTrace
	}
	builder := libseccomp.Builder{
		Allow:   allow,
		Trace:   trace,
		Default: actionDefault,
	}
	filter, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create seccomp filter %v", err)
	}

	limit := runner.Limit{
		TimeLimit:   time.Duration(instance.config.TimeLimit) * time.Second,
		MemoryLimit: runner.Size(instance.config.MemoryLimit),
	}

	syncFunc := func(pid int) error {
		return nil
	}
	fin, err := os.Open(args[0])
	if err != nil {
		return nil, fmt.Errorf("filed to open args[0]: %v", err)
	}
	execf, err := memfd.DupToMemfd("run_program", fin)
	if err != nil {
		return nil, fmt.Errorf("dup to memfd failed: %v", err)
	}
	fin.Close()
	defer execf.Close()
	execFile := execf.Fd()
	r := &ptrace.Runner{
		Args:        args,
		Env:         []string{pathEnv, userEnv, hostnameEnv},
		ExecFile:    execFile,
		RLimits:     rlims.PrepareRLimit(),
		Limit:       limit,
		Files:       fds,
		Seccomp:     filter,
		WorkDir:     instance.config.WorkDir,
		ShowDetails: instance.config.Verbose,
		Unsafe:      false,
		Handler:     h,
		SyncFunc:    syncFunc,
	}
	var rt runner.Result

	// gracefully shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	// Run tracer
	sTime := time.Now()
	// TODO: ensure that hard time limit is working properly
	c, cancel := context.WithTimeout(context.Background(), time.Duration(int64(instance.config.TimeLimit+2)*int64(time.Second)))
	defer cancel()

	s := make(chan runner.Result, 1)
	go func() {
		s <- r.Run(c)
	}()
	rTime := time.Now()

	select {
	case <-sig:
		cancel()
		rt = <-s
		rt.Status = runner.StatusRunnerError
	case rt = <-s:
	}
	eTime := time.Now()
	if rt.SetUpTime == 0 {
		rt.SetUpTime = rTime.Sub(sTime)
		rt.RunningTime = eTime.Sub(rTime)
	}
	return &rt, nil
}

func (instance *Instance) Collect() (string, string, error) {
	_, fout, ferr := instance.config.getIO(nil)
	out, e := os.ReadFile(fout)
	if e != nil {
		return "", "", e
	}
	err, e := os.ReadFile(ferr)
	if e != nil {
		return "", "", e
	}
	defer utils.Clean(fout, ferr)
	return string(out), string(err), nil
}
