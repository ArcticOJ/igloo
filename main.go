package main

import (
	"context"
	"github.com/ArcticOJ/igloo/v0/logger"
	"github.com/ArcticOJ/igloo/v0/worker"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	w := worker.New(ctx)
	go func() {
		<-ctx.Done()
		logger.Logger.Info().Msg("received signal, exiting...")
		w.Destroy()
		os.Exit(0)
	}()
	w.Work()
}
