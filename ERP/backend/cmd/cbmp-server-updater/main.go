package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"common-build-materials-platform/backend/internal/serverupdater"
)

func main() {
	configPath := flag.String("config", "", "path to server updater JSON config")
	once := flag.Bool("once", false, "poll and execute once, then exit")
	printSample := flag.Bool("print-sample-config", false, "print a sample JSON config")
	printServiceFiles := flag.Bool("print-service-files", false, "print systemd/launchd/Windows service install templates")
	baseURL := flag.String("base-url", "", "platform base URL, for example http://127.0.0.1:8088")
	token := flag.String("token", "", "customer product instance updater/probe token")
	watermark := flag.String("watermark", "", "optional customer product instance watermark")
	rootDir := flag.String("root", "", "local serverupdater work root")
	binaryPath := flag.String("binary", "", "serverupdater binary path used by -print-service-files")
	flag.Parse()

	if *printSample {
		if err := serverupdater.WriteSampleConfig(); err != nil {
			log.Fatal(err)
		}
		return
	}
	if *printServiceFiles {
		if err := serverupdater.WriteServiceFiles(os.Stdout, *binaryPath, *configPath); err != nil {
			log.Fatal(err)
		}
		return
	}

	cfg, err := serverupdater.LoadConfigFile(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	if *baseURL != "" {
		cfg.BaseURL = *baseURL
	}
	if *token != "" {
		cfg.UpdaterToken = *token
	}
	if *watermark != "" {
		cfg.Watermark = *watermark
	}
	if *rootDir != "" {
		cfg.RootDir = *rootDir
	}
	cfg, err = serverupdater.NormalizeConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	logger := log.New(os.Stdout, "cbmp-server-updater ", log.LstdFlags)
	runner, err := serverupdater.NewRunner(cfg, logger)
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if *once {
		if err := runner.RunOnce(ctx); err != nil {
			log.Fatal(err)
		}
		return
	}
	if err := runner.Run(ctx); err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}
