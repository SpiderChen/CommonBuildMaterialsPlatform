package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"common-build-materials-platform/industrial-control-gateway/internal/gateway"
)

func main() {
	cfg := gateway.LoadConfigFromEnv()
	logger := log.New(os.Stdout, "industrial-control-gateway ", log.LstdFlags|log.Lmicroseconds)

	forwarder, err := gateway.NewForwarder(cfg, logger)
	if err != nil {
		logger.Fatalf("init forwarder: %v", err)
	}
	service := gateway.NewService(cfg, forwarder, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.TCPAddr != "" {
		go func() {
			if err := service.ListenTCP(ctx, cfg.TCPAddr); err != nil && !errors.Is(err, context.Canceled) {
				logger.Printf("tcp listener stopped: %v", err)
				stop()
			}
		}()
	}
	if cfg.InputFile != "" {
		go service.WatchFile(ctx, cfg.InputFile)
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           service.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		logger.Printf("http listening on %s", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Printf("http server stopped: %v", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Printf("http shutdown: %v", err)
	}
}
