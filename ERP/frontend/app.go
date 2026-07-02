package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type DesktopApp struct {
	ctx          context.Context
	apiURL       string
	initialRoute string
}

func NewDesktopApp() *DesktopApp {
	return &DesktopApp{
		apiURL:       env("CBMP_API_URL", "http://127.0.0.1:8088"),
		initialRoute: sanitizeStandaloneRoute(os.Getenv("CBMP_STANDALONE_ROUTE")),
	}
}

func (a *DesktopApp) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *DesktopApp) Shutdown(ctx context.Context) {
}

func (a *DesktopApp) APIBaseURL() string {
	return a.apiURL
}

func (a *DesktopApp) InitialRoute() string {
	return a.initialRoute
}

func (a *DesktopApp) OpenStandaloneWindow(route string) error {
	route = sanitizeStandaloneRoute(route)
	if route == "" {
		return fmt.Errorf("invalid standalone route")
	}

	executable, err := os.Executable()
	if err != nil {
		return err
	}

	cmd := exec.Command(executable)
	cmd.Dir = filepath.Dir(executable)
	cmd.Env = withEnv(os.Environ(),
		"CBMP_STANDALONE_ROUTE="+route,
		"CBMP_API_URL="+a.apiURL,
		"CBMP_FRONTEND_DEV_URL="+os.Getenv("CBMP_FRONTEND_DEV_URL"),
	)
	return cmd.Start()
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func sanitizeStandaloneRoute(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || !strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") {
		return ""
	}
	if strings.Contains(value, "://") || strings.ContainsAny(value, "\r\n\t") {
		return ""
	}
	switch strings.TrimRight(strings.SplitN(value, "?", 2)[0], "/") {
	case "/fulfillment/dispatch", "/fulfillment/map-center":
		return value
	default:
		return ""
	}
}

func withEnv(base []string, values ...string) []string {
	result := append([]string{}, base...)
	for _, value := range values {
		key := strings.SplitN(value, "=", 2)[0]
		replaced := false
		for index, existing := range result {
			if strings.SplitN(existing, "=", 2)[0] == key {
				result[index] = value
				replaced = true
				break
			}
		}
		if !replaced {
			result = append(result, value)
		}
	}
	return result
}
