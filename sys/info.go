package sys

import (
	"github.com/ArcticOJ/igloo/v0/logger"
	"github.com/elastic/go-sysinfo"
	"time"
)

var Memory uint64 = 0

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
	Memory = mem.Total
}
