//go:build linux

package runner

import (
	"context"
	"fmt"
	"github.com/ArcticOJ/igloo/v0/config"
	"github.com/ArcticOJ/igloo/v0/logger"
	"github.com/ArcticOJ/igloo/v0/runner/shared"
	"github.com/ArcticOJ/igloo/v0/runtimes"
	"github.com/ArcticOJ/igloo/v0/utils"
	"github.com/criyle/go-sandbox/container"
	"github.com/criyle/go-sandbox/pkg/cgroup"
	"github.com/criyle/go-sandbox/pkg/forkexec"
	"github.com/criyle/go-sandbox/pkg/memfd"
	"github.com/criyle/go-sandbox/pkg/mount"
	"github.com/criyle/go-sandbox/pkg/rlimit"
	"github.com/criyle/go-sandbox/runner"
	"io"
	"math"
	"os"
	"path"
	"time"
)

var (
	mounts              []mount.Mount
	symlinks            []container.SymbolicLink
	requiredControllers = &cgroup.Controllers{
		CPU:    true,
		CPUSet: true,
		Memory: true,
		Pids:   true,
	}
	apexCgroup cgroup.Cgroup
)

type LinuxRunner struct {
	cpu uint16
	container.Environment
}

func init() {
	var e error
	mounts, symlinks, e = Config.Build()
	logger.Panic(e, "could not build config for containers")
	ct, e := cgroup.GetAvailableController()
	if !ct.Contains(requiredControllers) {
		logger.Logger.Fatal().Strs("expected", requiredControllers.Names()).Strs("got", ct.Names()).Msgf("missing required cgroup controller(s)")
	}
	logger.Panic(cgroup.EnableV2Nesting(), "failed query for available controllers")
	apexCgroup, e = cgroup.New("igloo.slice", ct)
	logger.Panic(e, "could not initialize cgroup")
}

func New(cpu uint16) (r *LinuxRunner, e error) {
	uid := os.Getuid()
	if uid == 0 {
		// fallback to 1536 on root
		uid = 1536
	}
	cb := container.Builder{
		Root:          "/tmp",
		TmpRoot:       "igloo-container-*",
		Mounts:        mounts,
		SymbolicLinks: symlinks,
		MaskPaths:     Config.MaskPaths,
		WorkDir:       "/home/igloo",
		CredGenerator: newCredGen(uint32(uid)),
		CloneFlags:    uintptr(forkexec.UnshareFlags),
		HostName:      "igloo",
		DomainName:    "arctic",
		ContainerUID:  Config.UID,
		ContainerGID:  Config.GID,
	}
	if config.Config.Debug {
		cb.Stderr = os.Stderr
	}
	r = &LinuxRunner{cpu: cpu}
	if e != nil {
		return nil, e
	}
	r.Environment, e = cb.Build()
	if e != nil {
		return nil, e
	}
	return r, nil
}

func (r *LinuxRunner) Compile(rt runtimes.Runtime, sourceCode string, ctx context.Context) (outPath, compOut string, e error) {
	ext := path.Ext(sourceCode)
	rand := utils.NextRand()
	outPath = fmt.Sprintf("/tmp/igloo_%s", rand)
	srcPath := fmt.Sprintf("/tmp/%s%s", rand, ext)
	f, e := r.Open([]container.OpenCmd{{
		Path: srcPath,
		Flag: os.O_WRONLY | os.O_CREATE | os.O_TRUNC,
		Perm: 0644,
	}})
	if e != nil {
		return
	}
	_f, e := os.Open(sourceCode)
	if e != nil {
		return
	}
	_, e = io.Copy(f[0], _f)
	f[0].Close()
	_f.Close()
	if e != nil {
		return
	}
	cmd, args := rt.BuildCompileCommand(srcPath, outPath)
	mf, e := memfd.New("compiler_output_" + utils.NextRand())
	if e != nil {
		return
	}
	defer mf.Close()
	fd := mf.Fd()
	fds := []uintptr{0, fd, fd}
	if e != nil {
		return
	}
	rl := rlimit.RLimits{
		CPU: 5,
		// 256MB
		Data:        256 << 20,
		DisableCore: true,
	}
	// limit at 5 seconds
	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cg, e := apexCgroup.Random("compiler-*")
	if e != nil {
		return
	}
	params := container.ExecveParam{
		Args:    append([]string{cmd}, args...),
		Env:     Config.Env,
		Files:   fds,
		RLimits: rl.PrepareRLimit(),
		SyncFunc: func(pid int) error {
			return cg.AddProc(pid)
		},
	}
	if e != nil {
		return
	}
	defer cg.Destroy()
	if e = cg.SetCPUSet([]byte(fmt.Sprintf("%d", r.cpu))); e != nil {
		return
	}
	res := r.Execve(c, params)
	mf.Seek(0, 0)
	out, e := io.ReadAll(mf)
	compOut = string(out)
	if res.Status != runner.StatusNormal {
		e = fmt.Errorf("failed to compile, got %v", res.Status)
	}
	return
}

func (r *LinuxRunner) Judge(args []string, config *shared.Config, ctx context.Context) (func(string) (*runner.Result, error), error) {
	tl := uint64(math.Floor(float64(config.TimeLimit)))
	tlHard := uint64(math.Ceil(float64(config.TimeLimit)))
	rl := rlimit.RLimits{
		// round to nearest second
		CPU:         tl,
		CPUHard:     tlHard,
		FileSize:    uint64(config.OutputLimit),
		Stack:       uint64(config.StackLimit),
		Data:        uint64(config.MemoryLimit),
		DisableCore: true,
	}
	return func(input string) (*runner.Result, error) {
		files, err := prepareFiles(input, config.OutputFile)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare files: %v", err)
		}
		defer closeFiles(files)
		cg, err := apexCgroup.Random("runner-*")
		if err != nil {
			return nil, err
		}
		defer cg.Destroy()
		if err = cg.SetMemoryLimit(uint64(config.MemoryLimit)); err != nil {
			return nil, err
		}
		if err = cg.SetCPUSet([]byte(fmt.Sprintf("%d", r.cpu))); err != nil {
			return nil, err
		}
		if err = cg.SetProcLimit(1); err != nil {
			return nil, err
		}
		fds := make([]uintptr, len(files))
		for i, f := range files {
			if f != nil {
				fds[i] = f.Fd()
			} else {
				fds[i] = uintptr(i)
			}
		}
		params := container.ExecveParam{
			Args:    args,
			Env:     Config.Env,
			Files:   fds,
			RLimits: rl.PrepareRLimit(),
			SyncFunc: func(pid int) error {
				return cg.AddProc(pid)
			},
		}
		c, cancel := context.WithTimeout(ctx, time.Duration(config.TimeLimit*float32(time.Second)))
		defer cancel()
		res := r.Execve(c, params)
		memory, err := cg.MemoryMaxUsage()
		if err != nil && os.IsNotExist(err) {
			return nil, err
		}
		if memory > 0 {
			res.Memory = runner.Size(memory)
		}
		cpu, err := cg.CPUUsage()
		// else fallback to measure time exceeded using time package
		if err == nil {
			res.Time = time.Duration(cpu)
		}
		if res.Memory.MiB() >= uint64(config.MemoryLimit>>20) {
			res.Status = runner.StatusMemoryLimitExceeded
		}
		return &res, nil
	}, nil
}

func (r *LinuxRunner) Cleanup() error {
	return r.Reset()
}

func (r *LinuxRunner) Destroy() error {
	return r.Reset()
}

func DestroyAll() error {
	return apexCgroup.Destroy()
}
