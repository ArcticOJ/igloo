package http

import (
	"context"
	"fmt"
	"github.com/uptrace/bunrouter"
	"igloo/igloo/config"
	"igloo/igloo/judge/runtimes"
	"igloo/igloo/logger"
	"net"
	"net/http"
	"time"
)

func StartServer(ctx context.Context) {
	r := bunrouter.New()
	r.GET("/version", func(w http.ResponseWriter, req bunrouter.Request) error {
		return bunrouter.JSON(w, bunrouter.H{
			"version": "0.1.0-prealpha",
		})
	})
	r.GET("/status", func(w http.ResponseWriter, req bunrouter.Request) error {
		return nil
	})
	r.GET("/runtimes", func(w http.ResponseWriter, req bunrouter.Request) error {
		return bunrouter.JSON(w, runtimes.Runtimes)
	})
	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      r,
	}
	addr := fmt.Sprintf("%s:%d", config.Config.Host, config.Config.Port)
	l, e := net.Listen("tcp", addr)
	logger.Panic(e, "error while listening on port %s", addr)
	go func() {
		<-ctx.Done()
		logger.Panic(srv.Shutdown(context.Background()), "could not gracefully shutdown http server")
	}()
	logger.Logger.Info().Msgf("http listening on %s", addr)
	logger.Logger.Fatal().Err(srv.Serve(l))
}
