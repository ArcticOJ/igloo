package judge

import (
	"context"
	"fmt"
	"github.com/criyle/go-sandbox/pkg/memfd"
	"igloo/igloo/judge/config"
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
	StatusNormal  Status = iota // 0
	StatusInvalid               // 1
	StatusRE                    // 2
	StatusMLE                   // 3
	StatusTLE                   // 4
	StatusOLE                   // 5
	StatusBan                   // 6
	StatusFatal                 // 7
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

func Start(conf *Config, argv []string) (*runner.Result, error) {
	var rt runner.Result
	args, allow, trace, h := config.GetConf(conf.Type, conf.WorkDir, argv, []string{}, []string{}, false)
	inpFile, outFile, errFile := conf.getIO()
	files, err := prepareFiles(inpFile, outFile, errFile)
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
		CPU:         conf.TimeLimit,
		CPUHard:     conf.TimeLimitHard,
		FileSize:    conf.OutputLimit,
		Stack:       conf.StackLimit,
		Data:        conf.MemoryLimit,
		OpenFile:    256,
		DisableCore: true,
	}
	actionDefault := libseccomp.ActionKill
	if conf.Verbose {
		// TODO: change to |=
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
		TimeLimit:   time.Duration(conf.TimeLimit) * time.Second,
		MemoryLimit: runner.Size(conf.MemoryLimit),
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
		WorkDir:     conf.WorkDir,
		ShowDetails: conf.Verbose,
		Unsafe:      false,
		Handler:     h,
		SyncFunc:    syncFunc,
	}
	// gracefully shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	// Run tracer
	sTime := time.Now()
	c, cancel := context.WithTimeout(context.Background(), time.Duration(int64(conf.TimeLimit+2)*int64(time.Second)))
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
