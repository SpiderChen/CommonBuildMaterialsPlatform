package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"common-build-materials-operations/backend/internal/ops"
)

func main() {
	addr := flag.String("addr", getenv("CBM_OPS_ADDR", ":8095"), "HTTP listen address")
	dataPath := flag.String("data", getenv("CBM_OPS_DATA", "data/ops.json"), "JSON data file")
	frontendDir := flag.String("frontend", getenv("CBM_OPS_FRONTEND_DIR", ""), "static frontend directory")
	flag.Parse()

	store := ops.NewStore(filepath.Clean(*dataPath))
	cleanFrontendDir := ""
	if *frontendDir != "" {
		cleanFrontendDir = filepath.Clean(*frontendDir)
	}
	app := ops.NewApp(store, cleanFrontendDir)
	log.Printf("CommonBuildMaterialsOperationsPlatform listening on %s", *addr)
	if err := http.ListenAndServe(*addr, app.Routes()); err != nil {
		log.Fatal(err)
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
