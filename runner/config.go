package runner

import (
	_ "embed"
	"github.com/ArcticOJ/igloo/v0/logger"
	"github.com/criyle/go-sandbox/container"
	"github.com/criyle/go-sandbox/pkg/mount"
	"gopkg.in/yaml.v2"
	"os"
	"path"
)

//go:embed container.yml
var conf []byte

var Config = new(ContainerConfig)

func init() {
	logger.Panic(yaml.Unmarshal(conf, Config), "failed to parse container config")
}

type (
	MountType string

	SubmissionConfig struct {
		Target      string
		TimeLimit   float32
		MemoryLimit uint32
		OutputLimit uint32
		StackLimit  uint32
		OutputFile  string
		Verbose     bool
	}

	Mount struct {
		Type     MountType
		Source   string
		Target   string
		Readonly bool
		Options  string
	}

	Link struct {
		LinkPath string
		Target   string
	}
	ContainerConfig struct {
		Env        []string
		Mounts     []Mount
		Symlinks   []Link
		MaskPaths  []string
		WorkDir    string
		HostName   string
		DomainName string
		UID        int
		GID        int
		Proc       bool
	}
)

const (
	Bind MountType = "bind"
	Temp           = "tmpfs"
)

func (c *ContainerConfig) Build() ([]mount.Mount, []container.SymbolicLink, error) {
	mb := mount.NewBuilder()
	var _symlinks []container.SymbolicLink
	if c.Proc {
		mb.WithProc()
	}
	cwd, e := os.Getwd()
	if e != nil {
		return nil, nil, e
	}
	for _, mnt := range c.Mounts {
		dest := mnt.Target
		if path.IsAbs(dest) {
			dest = path.Clean(dest[1:])
		}
		src := mnt.Source
		if !path.IsAbs(src) {
			src = path.Join(cwd, src)
		}
		switch mnt.Type {
		case Bind:
			mb.WithBind(src, dest, mnt.Readonly)
		case Temp:
			mb.WithTmpfs(dest, mnt.Options)
		}
	}
	if c.Proc {
		mb.WithProc()
	}
	for _, l := range c.Symlinks {
		_symlinks = append(_symlinks, container.SymbolicLink{LinkPath: l.LinkPath, Target: l.Target})
	}
	return mb.FilterNotExist().Mounts, nil, nil
}
