package ops

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type App struct {
	store       *Store
	frontendDir string
}

func NewApp(store *Store, frontendDir string) *App {
	return &App{store: store, frontendDir: frontendDir}
}

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/", a.api)
	if a.frontendDir != "" {
		mux.HandleFunc("/", a.frontend)
	}
	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *App) frontend(w http.ResponseWriter, r *http.Request) {
	path := filepath.Clean(filepath.Join(a.frontendDir, r.URL.Path))
	if rel, err := filepath.Rel(a.frontendDir, path); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			http.ServeFile(w, r, path)
			return
		}
	}
	http.ServeFile(w, r, filepath.Join(a.frontendDir, "index.html"))
}

func (a *App) api(w http.ResponseWriter, r *http.Request) {
	parts := splitPath(strings.TrimPrefix(r.URL.Path, "/api"))
	if len(parts) == 0 {
		writeJSON(w, http.StatusOK, map[string]string{"name": "Common Build Materials Operations API"})
		return
	}
	if parts[0] == "health" {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	}

	switch {
	case r.Method == http.MethodGet && parts[0] == "summary":
		a.snapshotJSON(w, func(data AppData) interface{} { return Summary(data) })
	case r.Method == http.MethodGet && parts[0] == "customers":
		a.snapshotJSON(w, func(data AppData) interface{} { return SortedCustomers(data) })
	case r.Method == http.MethodPost && len(parts) == 3 && parts[0] == "customers" && parts[2] == "renewals":
		id, ok := parseID(w, parts[1])
		if ok {
			var req RenewLicenseRequest
			if decodeBody(w, r, &req) {
				a.updateJSON(w, func(data *AppData) (interface{}, error) { return RenewCustomer(data, id, req) })
			}
		}
	case r.Method == http.MethodGet && parts[0] == "renewals":
		a.snapshotJSON(w, func(data AppData) interface{} { return data.Renewals })
	case r.Method == http.MethodGet && parts[0] == "alerts":
		a.snapshotJSON(w, func(data AppData) interface{} { return data.Alerts })
	case r.Method == http.MethodPost && len(parts) == 3 && parts[0] == "alerts" && parts[2] == "ack":
		id, ok := parseID(w, parts[1])
		if ok {
			a.updateJSON(w, func(data *AppData) (interface{}, error) { return AcknowledgeAlert(data, id, "ops") })
		}
	case r.Method == http.MethodPost && len(parts) == 3 && parts[0] == "alerts" && parts[2] == "resolve":
		id, ok := parseID(w, parts[1])
		if ok {
			var req ResolveAlertRequest
			if decodeBody(w, r, &req) {
				a.updateJSON(w, func(data *AppData) (interface{}, error) { return ResolveAlert(data, id, req) })
			}
		}
	case r.Method == http.MethodGet && parts[0] == "update-packages":
		a.snapshotJSON(w, func(data AppData) interface{} { return data.UpdatePackages })
	case r.Method == http.MethodPost && len(parts) == 1 && parts[0] == "update-packages":
		var req CreateUpdatePackageRequest
		if decodeBody(w, r, &req) {
			a.updateJSON(w, func(data *AppData) (interface{}, error) { return CreatePackage(data, req) })
		}
	case r.Method == http.MethodPost && len(parts) == 3 && parts[0] == "update-packages" && parts[2] == "publish":
		id, ok := parseID(w, parts[1])
		if ok {
			a.updateJSON(w, func(data *AppData) (interface{}, error) { return PublishPackage(data, id) })
		}
	case r.Method == http.MethodPost && len(parts) == 3 && parts[0] == "update-packages" && parts[2] == "assign":
		id, ok := parseID(w, parts[1])
		if ok {
			var req AssignUpdatePackageRequest
			if decodeBody(w, r, &req) {
				a.updateJSON(w, func(data *AppData) (interface{}, error) { return AssignPackage(data, id, req) })
			}
		}
	case r.Method == http.MethodGet && parts[0] == "assignments":
		a.snapshotJSON(w, func(data AppData) interface{} { return data.Assignments })
	case r.Method == http.MethodGet && parts[0] == "audit-logs":
		a.snapshotJSON(w, func(data AppData) interface{} { return data.AuditLogs })
	default:
		writeError(w, http.StatusNotFound, "unknown endpoint")
	}
}

func (a *App) snapshotJSON(w http.ResponseWriter, view func(AppData) interface{}) {
	data, err := a.store.Snapshot()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, view(data))
}

func (a *App) updateJSON(w http.ResponseWriter, fn func(*AppData) (interface{}, error)) {
	var result interface{}
	_, err := a.store.Update(func(data *AppData) error {
		var err error
		result, err = fn(data)
		return err
	})
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrNotFound) {
			status = http.StatusNotFound
		}
		if errors.Is(err, ErrBadRequest) {
			status = http.StatusBadRequest
		}
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}

func parseID(w http.ResponseWriter, value string) (int64, bool) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func decodeBody(w http.ResponseWriter, r *http.Request, target interface{}) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
