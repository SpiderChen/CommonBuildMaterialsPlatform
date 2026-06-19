package gateway

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

type Metrics struct {
	Received      atomic.Uint64 `json:"-"`
	Accepted      atomic.Uint64 `json:"-"`
	Rejected      atomic.Uint64 `json:"-"`
	Forwarded     atomic.Uint64 `json:"-"`
	ForwardErrors atomic.Uint64 `json:"-"`
}

type Service struct {
	cfg       Config
	forwarder Forwarder
	logger    *log.Logger
	metrics   Metrics
}

func NewService(cfg Config, forwarder Forwarder, logger *log.Logger) *Service {
	return &Service{cfg: cfg, forwarder: forwarder, logger: logger}
}

func (s *Service) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.healthz)
	mux.HandleFunc("/metrics", s.metricsJSON)
	mux.HandleFunc("/ingest", s.ingestHTTP)
	return mux
}

func (s *Service) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "industrial-control-gateway"})
}

func (s *Service) metricsJSON(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]uint64{
		"received":      s.metrics.Received.Load(),
		"accepted":      s.metrics.Accepted.Load(),
		"rejected":      s.metrics.Rejected.Load(),
		"forwarded":     s.metrics.Forwarded.Load(),
		"forwardErrors": s.metrics.ForwardErrors.Load(),
	})
}

func (s *Service) ingestHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, s.cfg.MaxBodyBytes))
	if err != nil {
		http.Error(w, "read request body", http.StatusBadRequest)
		return
	}
	raw, protocol, source := unwrapBody(body, s.cfg.DefaultProtocol, s.cfg.DefaultSource)
	msg, err := s.Ingest(r.Context(), raw, "http", protocol, source)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusAccepted, msg.withoutRaw())
}

func (s *Service) Ingest(ctx context.Context, raw, channel, protocol, source string) (Message, error) {
	s.metrics.Received.Add(1)
	msg, err := ParseFrame(raw, source, channel, protocol, time.Now())
	if err != nil {
		s.metrics.Rejected.Add(1)
		return Message{}, err
	}
	s.metrics.Accepted.Add(1)
	if err := s.forwarder.Forward(ctx, msg); err != nil {
		s.metrics.ForwardErrors.Add(1)
		return Message{}, err
	}
	s.metrics.Forwarded.Add(1)
	return msg, nil
}

func (s *Service) ListenTCP(ctx context.Context, addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()
	s.logger.Printf("tcp listening on %s", addr)
	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}
		go s.handleConn(ctx, conn)
	}
}

func (s *Service) handleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 0, 64*1024), int(s.cfg.MaxBodyBytes))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if _, err := s.Ingest(ctx, line, "tcp", s.cfg.DefaultProtocol, s.cfg.DefaultSource); err != nil {
			_, _ = fmt.Fprintf(conn, "ERR %s\n", err.Error())
			continue
		}
		_, _ = io.WriteString(conn, "OK\n")
	}
	if err := scanner.Err(); err != nil && !errors.Is(err, net.ErrClosed) {
		s.logger.Printf("tcp scan: %v", err)
	}
}

func (s *Service) WatchFile(ctx context.Context, path string) {
	ticker := time.NewTicker(s.cfg.PollInterval)
	defer ticker.Stop()
	var offset int64
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			next, err := s.readNewLines(ctx, path, offset)
			if err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					s.logger.Printf("file watch %s: %v", path, err)
				}
				continue
			}
			offset = next
		}
	}
}

func (s *Service) readNewLines(ctx context.Context, path string, offset int64) (int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return offset, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return offset, err
	}
	if info.Size() < offset {
		offset = 0
	}
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return offset, err
	}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), int(s.cfg.MaxBodyBytes))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			if _, err := s.Ingest(ctx, line, "file", s.cfg.DefaultProtocol, s.cfg.DefaultSource); err != nil {
				s.logger.Printf("reject line: %v", err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return offset, err
	}
	pos, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return offset, err
	}
	return pos, nil
}

func unwrapBody(body []byte, defaultProtocol, defaultSource string) (raw, protocol, source string) {
	protocol, source = defaultProtocol, defaultSource
	raw = strings.TrimSpace(string(body))
	var envelope struct {
		Raw      string `json:"raw"`
		Protocol string `json:"protocol"`
		Source   string `json:"source"`
	}
	if json.Unmarshal(body, &envelope) == nil && strings.TrimSpace(envelope.Raw) != "" {
		raw = strings.TrimSpace(envelope.Raw)
		if strings.TrimSpace(envelope.Protocol) != "" {
			protocol = strings.TrimSpace(envelope.Protocol)
		}
		if strings.TrimSpace(envelope.Source) != "" {
			source = strings.TrimSpace(envelope.Source)
		}
	}
	return raw, protocol, source
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
