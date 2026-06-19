package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"common-build-materials-platform/backend/internal/appliance"
)

func main() {
	addr := flag.String("addr", env("CBMP_ADDR", "127.0.0.1:8088"), "listen address")
	dataPath := flag.String("data", env("CBMP_DATA", filepath.Join("data", "app.vault")), "encrypted data vault path")
	frontendDir := flag.String("frontend", env("CBMP_FRONTEND_DIR", ""), "optional static frontend directory")
	flag.Parse()

	var dataStore appliance.DataStore
	if dsn := os.Getenv("CBMP_POSTGRES_DSN"); dsn != "" {
		dataStore = appliance.NewPostgresStore(dsn, os.Getenv("CBMP_DATA_KEY"))
		log.Printf("storage=postgres dsn=%s", redactDSN(dsn))
	} else {
		dataStore = appliance.NewStore(*dataPath, os.Getenv("CBMP_DATA_KEY"))
		log.Printf("storage=vault data=%s", *dataPath)
	}
	if err := dataStore.Load(); err != nil {
		log.Fatalf("load data vault: %v", err)
	}

	app := appliance.NewApp(dataStore, *frontendDir)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := app.StartDeviceGateways(ctx); err != nil {
		log.Fatalf("start device gateways: %v", err)
	}
	server := &http.Server{Addr: *addr, Handler: app.Routes()}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	log.Printf("CBMP Go appliance listening on http://%s", *addr)
	if *frontendDir != "" {
		log.Printf("frontend=%s", *frontendDir)
	} else {
		log.Printf("frontend=disabled")
	}
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func redactDSN(value string) string {
	if value == "" {
		return ""
	}
	return "<configured>"
}
