package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"common-build-materials-platform/frontend/internal/clientupdater"
)

func main() {
	configPath := flag.String("config", "", "path to client updater JSON config")
	once := flag.Bool("once", false, "poll and execute once, then exit")
	printSample := flag.Bool("print-sample-config", false, "print a sample JSON config")
	printServiceFiles := flag.Bool("print-service-files", false, "print systemd/launchd/Windows service install templates")
	baseURL := flag.String("base-url", "", "operations platform base URL, for example http://127.0.0.1:8095")
	token := flag.String("token", "", "updater token issued by OperationsPlatform")
	watermark := flag.String("watermark", "", "optional customer instance watermark")
	rootDir := flag.String("root", "", "local clientupdater work root")
	binaryPath := flag.String("binary", "", "clientupdater binary path used by -print-service-files")
	flag.Parse()

	if *printSample {
		if err := clientupdater.WriteSampleConfig(); err != nil {
			log.Fatal(err)
		}
		return
	}
	if *printServiceFiles {
		if err := clientupdater.WriteServiceFiles(os.Stdout, *binaryPath, *configPath); err != nil {
			log.Fatal(err)
		}
		return
	}

	cfg, err := clientupdater.LoadConfigFile(*configPath)
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
	cfg, err = clientupdater.NormalizeConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	logger := log.New(os.Stdout, "cbmp-client-updater ", log.LstdFlags)
	runner, err := clientupdater.NewRunner(cfg, logger)
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
