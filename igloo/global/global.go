package global

import (
	"igloo/igloo/judge/worker"
	"time"
)

var (
	BootTimestamp = time.Now()
	Worker        *worker.JudgeWorker
)
