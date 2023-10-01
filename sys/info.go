package sys

import (
	"github.com/elastic/go-sysinfo"
	"igloo/logger"
	"time"
)

// Memory as KB
var Memory uint32 = 0

var OS = "Unknown"

var BootTimestamp time.Time

func init() {
	host, e := sysinfo.Host()
	logger.Panic(e, "could not get host info")
	inf := host.Info()
	OS = inf.OS.Name
	BootTimestamp = inf.BootTime
	mem, e := host.Memory()
	logger.Panic(e, "could not get host memory info")
	Memory = uint32(mem.Total >> 10)
}
