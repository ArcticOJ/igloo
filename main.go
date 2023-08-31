package main

import (
	"context"
	"igloo/igloo/global"
	"igloo/igloo/http"
	"igloo/igloo/judge/worker"
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
	}()
	go http.StartServer(ctx)
	global.Worker.Work()
}
