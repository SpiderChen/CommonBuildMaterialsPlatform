package forwarder

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Sink interface {
	Write(context.Context, Location) error
}

func NewSink(cfg Config, logger *log.Logger) (Sink, error) {
	sinks := []Sink{}
	for _, target := range cfg.HTTPTargets {
		sinks = append(sinks, &HTTPSink{
			target:     target,
			token:      cfg.HTTPBearer,
			deviceKey:  cfg.DeviceKey,
			secret:     cfg.SharedSecret,
			mode:       strings.ToLower(cfg.ForwardMode),
			includeRaw: cfg.IncludeRaw,
			client:     &http.Client{Timeout: cfg.HTTPTimeout},
		})
	}
	if cfg.OutputFile != "" {
		fileSink, err := NewFileSink(cfg.OutputFile, cfg.IncludeRaw)
		if err != nil {
			return nil, err
		}
		sinks = append(sinks, fileSink)
	}
	if len(sinks) == 0 {
		sinks = append(sinks, &StdoutSink{logger: logger, includeRaw: cfg.IncludeRaw})
	}
	return MultiSink(sinks), nil
}

type MultiSink []Sink

func (m MultiSink) Write(ctx context.Context, loc Location) error {
	var errs []string
	for _, sink := range m {
		if err := sink.Write(ctx, loc); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

type HTTPSink struct {
	target     string
	token      string
	deviceKey  string
	secret     string
	mode       string
	includeRaw bool
	client     *http.Client
}

func (h *HTTPSink) Write(ctx context.Context, loc Location) error {
	body, err := h.body(loc)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.target, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if h.token != "" {
		req.Header.Set("Authorization", "Bearer "+h.token)
	}
	if h.deviceKey != "" {
		req.Header.Set("X-Device-Key", h.deviceKey)
	}
	if h.secret != "" {
		ts := time.Now().UTC().Format(time.RFC3339)
		req.Header.Set("X-CBMP-Timestamp", ts)
		req.Header.Set("X-CBMP-Signature", signBody(h.secret, ts, body))
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("http target %s returned %s", h.target, resp.Status)
	}
	return nil
}

func (h *HTTPSink) body(loc Location) ([]byte, error) {
	switch h.mode {
	case "protocol-frame":
		return json.Marshal(ProtocolFrameEnvelope{Channel: loc.Channel, Protocol: loc.Protocol, Raw: loc.Raw})
	case "location":
		if !h.includeRaw {
			loc = loc.withoutRaw()
		}
		return json.Marshal(loc)
	default:
		if !h.includeRaw {
			loc = loc.withoutRaw()
		}
		return json.Marshal(loc)
	}
}

type FileSink struct {
	path       string
	includeRaw bool
	mu         sync.Mutex
}

func NewFileSink(path string, includeRaw bool) (*FileSink, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	return &FileSink{path: path, includeRaw: includeRaw}, nil
}

func (f *FileSink) Write(_ context.Context, loc Location) error {
	if !f.includeRaw {
		loc = loc.withoutRaw()
	}
	body, err := json.Marshal(loc)
	if err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	file, err := os.OpenFile(f.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(append(body, '\n'))
	return err
}

type StdoutSink struct {
	logger     *log.Logger
	includeRaw bool
}

func (s *StdoutSink) Write(_ context.Context, loc Location) error {
	if !s.includeRaw {
		loc = loc.withoutRaw()
	}
	body, err := json.Marshal(loc)
	if err != nil {
		return err
	}
	s.logger.Print(string(body))
	return nil
}

func signBody(secret, timestamp string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("\n"))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
