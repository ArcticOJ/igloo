package runner

import (
	"sync/atomic"
	"syscall"
)

type credGen struct {
	current uint32
}

func newCredGen(uid uint32) (c *credGen) {
	c = &credGen{current: uid}
	return
}

func (c *credGen) Get() syscall.Credential {
	n := atomic.AddUint32(&c.current, 1)
	return syscall.Credential{
		Uid: n,
		Gid: n,
	}
}
