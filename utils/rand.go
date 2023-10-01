package utils

import (
	"strconv"
	_ "unsafe"
)

//go:linkname fastrand runtime.fastrand
func fastrand() uint32

func NextRand() string {
	return strconv.FormatUint(uint64(fastrand()), 10)
}
