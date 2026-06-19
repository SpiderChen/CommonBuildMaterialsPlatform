package main

import (
	"context"
	"os"
)

type DesktopApp struct {
	ctx    context.Context
	apiURL string
}

func NewDesktopApp() *DesktopApp {
	return &DesktopApp{apiURL: env("CBMP_API_URL", "http://127.0.0.1:8088")}
}

func (a *DesktopApp) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *DesktopApp) Shutdown(ctx context.Context) {
}

func (a *DesktopApp) APIBaseURL() string {
	return a.apiURL
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
