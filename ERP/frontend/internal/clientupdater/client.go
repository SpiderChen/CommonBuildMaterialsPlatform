package clientupdater

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	apiBaseURL string
	token      string
	watermark  string
	rootDir    string
	resume     bool
	httpClient *http.Client
}

func NewClient(cfg Config) (*Client, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("baseUrl is required")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("baseUrl must include scheme and host")
	}
	apiBase := baseURL
	if !strings.HasSuffix(apiBase, "/api") {
		apiBase += "/api"
	}
	origin := parsed.Scheme + "://" + parsed.Host
	return &Client{
		baseURL:    origin,
		apiBaseURL: apiBase,
		token:      cfg.UpdaterToken,
		watermark:  cfg.Watermark,
		rootDir:    cfg.RootDir,
		resume:     cfg.ResumeDownloads,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func (c *Client) Poll(ctx context.Context) (PollResponse, error) {
	payload := map[string]string{"updaterToken": c.token, "watermark": c.watermark}
	var out PollResponse
	err := c.doJSON(ctx, http.MethodPost, c.apiBaseURL+"/product-ops/system-updates/tasks", payload, &out)
	return out, err
}

func (c *Client) DownloadTask(ctx context.Context, task ProductSystemUpdateTask) (UpdatePackageDownload, []byte, error) {
	if !c.resume || strings.TrimSpace(c.rootDir) == "" {
		return c.Download(ctx, task.DownloadURL)
	}
	return c.downloadWithResume(ctx, task)
}

func (c *Client) Download(ctx context.Context, downloadURL string) (UpdatePackageDownload, []byte, error) {
	resolved, err := c.resolveURL(downloadURL)
	if err != nil {
		return UpdatePackageDownload{}, nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resolved, nil)
	if err != nil {
		return UpdatePackageDownload{}, nil, err
	}
	req.Header.Set("X-CBMP-Updater-Token", c.token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return UpdatePackageDownload{}, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return UpdatePackageDownload{}, nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return UpdatePackageDownload{}, body, fmt.Errorf("download update package failed: %s %s", resp.Status, string(body))
	}
	var out UpdatePackageDownload
	if err := json.Unmarshal(body, &out); err != nil {
		return UpdatePackageDownload{}, body, err
	}
	return out, body, nil
}

func (c *Client) downloadWithResume(ctx context.Context, task ProductSystemUpdateTask) (UpdatePackageDownload, []byte, error) {
	resolved, err := c.resolveURL(task.DownloadURL)
	if err != nil {
		return UpdatePackageDownload{}, nil, err
	}
	cacheDir := filepath.Join(c.rootDir, "downloads", safeName(task.TaskNo))
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return UpdatePackageDownload{}, nil, err
	}
	finalPath := filepath.Join(cacheDir, "package.json")
	partPath := finalPath + ".part"
	if raw, err := os.ReadFile(finalPath); err == nil {
		var out UpdatePackageDownload
		if err := json.Unmarshal(raw, &out); err == nil {
			return out, raw, nil
		}
		_ = os.Remove(finalPath)
	}
	offset := int64(0)
	if info, err := os.Stat(partPath); err == nil {
		offset = info.Size()
	}
	appendMode := offset > 0
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resolved, nil)
	if err != nil {
		return UpdatePackageDownload{}, nil, err
	}
	req.Header.Set("X-CBMP-Updater-Token", c.token)
	if appendMode {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", offset))
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return UpdatePackageDownload{}, nil, err
	}
	defer resp.Body.Close()
	if appendMode && resp.StatusCode == http.StatusOK {
		appendMode = false
		offset = 0
		_ = os.Remove(partPath)
	}
	if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
		_ = os.Remove(partPath)
		return c.DownloadTask(ctx, task)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return UpdatePackageDownload{}, body, fmt.Errorf("download update package failed: %s %s", resp.Status, string(body))
	}
	if appendMode && resp.StatusCode != http.StatusPartialContent {
		return UpdatePackageDownload{}, nil, fmt.Errorf("resume download expected 206 partial content, got %s", resp.Status)
	}
	flags := os.O_CREATE | os.O_WRONLY
	if appendMode {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}
	file, err := os.OpenFile(partPath, flags, 0o644)
	if err != nil {
		return UpdatePackageDownload{}, nil, err
	}
	if _, err := io.Copy(file, resp.Body); err != nil {
		_ = file.Close()
		return UpdatePackageDownload{}, nil, err
	}
	if err := file.Close(); err != nil {
		return UpdatePackageDownload{}, nil, err
	}
	raw, err := os.ReadFile(partPath)
	if err != nil {
		return UpdatePackageDownload{}, nil, err
	}
	var out UpdatePackageDownload
	if err := json.Unmarshal(raw, &out); err != nil {
		return UpdatePackageDownload{}, raw, fmt.Errorf("download incomplete; kept resumable partial file %s: %w", partPath, err)
	}
	if err := os.Rename(partPath, finalPath); err != nil {
		return UpdatePackageDownload{}, raw, err
	}
	return out, raw, nil
}

func (c *Client) Report(ctx context.Context, taskNo string, report ReportRequest) error {
	report.UpdaterToken = c.token
	report.UpdaterVersion = fallbackText(report.UpdaterVersion, Version)
	path := c.apiBaseURL + "/product-ops/system-updates/tasks/" + url.PathEscape(taskNo) + "/report"
	return c.doJSON(ctx, http.MethodPost, path, report, nil)
}

func (c *Client) doJSON(ctx context.Context, method, endpoint string, in any, out any) error {
	var body io.Reader
	if in != nil {
		raw, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CBMP-Updater-Token", c.token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%s %s failed: %s %s", method, endpoint, resp.Status, string(raw))
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

func (c *Client) resolveURL(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", fmt.Errorf("downloadUrl is empty")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	if parsed.IsAbs() {
		return parsed.String(), nil
	}
	if strings.HasPrefix(raw, "/") {
		return c.baseURL + raw, nil
	}
	return strings.TrimRight(c.apiBaseURL, "/") + "/" + strings.TrimLeft(raw, "/"), nil
}

func fallbackText(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}
