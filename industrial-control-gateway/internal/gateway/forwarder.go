package gateway

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

type Forwarder interface {
	Forward(context.Context, Message) error
}

func NewForwarder(cfg Config, logger *log.Logger) (Forwarder, error) {
	forwarders := []Forwarder{}
	for _, target := range cfg.HTTPTargets {
		forwarders = append(forwarders, &HTTPForwarder{
			target:     target,
			token:      cfg.HTTPBearer,
			secret:     cfg.SharedSecret,
			mode:       strings.ToLower(cfg.ForwardMode),
			includeRaw: cfg.IncludeRaw,
			client:     &http.Client{Timeout: cfg.HTTPTimeout},
		})
	}
	if cfg.OutputFile != "" {
		fileForwarder, err := NewFileForwarder(cfg.OutputFile, cfg.IncludeRaw)
		if err != nil {
			return nil, err
		}
		forwarders = append(forwarders, fileForwarder)
	}
	if len(forwarders) == 0 {
		forwarders = append(forwarders, &StdoutForwarder{logger: logger, includeRaw: cfg.IncludeRaw})
	}
	return MultiForwarder(forwarders), nil
}

type MultiForwarder []Forwarder

func (m MultiForwarder) Forward(ctx context.Context, msg Message) error {
	var errs []string
	for _, forwarder := range m {
		if err := forwarder.Forward(ctx, msg); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

type HTTPForwarder struct {
	target     string
	token      string
	secret     string
	mode       string
	includeRaw bool
	client     *http.Client
}

func (h *HTTPForwarder) Forward(ctx context.Context, msg Message) error {
	body, err := h.body(msg)
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

func (h *HTTPForwarder) body(msg Message) ([]byte, error) {
	if h.mode == "protocol-frame" {
		return json.Marshal(ProtocolFrameEnvelope{
			Channel:  msg.Channel,
			Protocol: msg.Protocol,
			Raw:      msg.Raw,
		})
	}
	if !h.includeRaw {
		msg = msg.withoutRaw()
	}
	return json.Marshal(msg)
}

type FileForwarder struct {
	path       string
	includeRaw bool
	mu         sync.Mutex
}

func NewFileForwarder(path string, includeRaw bool) (*FileForwarder, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	return &FileForwarder{path: path, includeRaw: includeRaw}, nil
}

func (f *FileForwarder) Forward(_ context.Context, msg Message) error {
	if !f.includeRaw {
		msg = msg.withoutRaw()
	}
	body, err := json.Marshal(msg)
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

type StdoutForwarder struct {
	logger     *log.Logger
	includeRaw bool
}

func (s *StdoutForwarder) Forward(_ context.Context, msg Message) error {
	if !s.includeRaw {
		msg = msg.withoutRaw()
	}
	body, err := json.Marshal(msg)
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
