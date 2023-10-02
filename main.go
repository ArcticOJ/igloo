package main

import (
	"context"
	"igloo/global"
	"igloo/judge/worker"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	global.Worker = worker.New(ctx)
	go func() {
		<-ctx.Done()
		global.Worker.Destroy()
		os.Exit(0)
	}()
	global.Worker.Work()
}
