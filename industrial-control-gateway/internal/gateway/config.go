package gateway

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr        string
	TCPAddr         string
	InputFile       string
	OutputFile      string
	HTTPTargets     []string
	HTTPBearer      string
	SharedSecret    string
	DefaultProtocol string
	DefaultSource   string
	ForwardMode     string
	IncludeRaw      bool
	PollInterval    time.Duration
	HTTPTimeout     time.Duration
	MaxBodyBytes    int64
}

func LoadConfigFromEnv() Config {
	return Config{
		HTTPAddr:        env("ICG_ADDR", "0.0.0.0:19101"),
		TCPAddr:         strings.TrimSpace(os.Getenv("ICG_TCP_ADDR")),
		InputFile:       strings.TrimSpace(os.Getenv("ICG_FILE")),
		OutputFile:      strings.TrimSpace(os.Getenv("ICG_OUT_FILE")),
		HTTPTargets:     splitList(os.Getenv("ICG_HTTP_TARGETS")),
		HTTPBearer:      strings.TrimSpace(os.Getenv("ICG_HTTP_BEARER_TOKEN")),
		SharedSecret:    strings.TrimSpace(os.Getenv("ICG_SHARED_SECRET")),
		DefaultProtocol: env("ICG_PROTOCOL", "industrial-json"),
		DefaultSource:   env("ICG_SOURCE", "industrial-control-gateway"),
		ForwardMode:     env("ICG_FORWARD_MODE", "normalized"),
		IncludeRaw:      envBool("ICG_INCLUDE_RAW", false),
		PollInterval:    envDuration("ICG_POLL_INTERVAL", 2*time.Second),
		HTTPTimeout:     envDuration("ICG_HTTP_TIMEOUT", 5*time.Second),
		MaxBodyBytes:    envInt64("ICG_MAX_BODY_BYTES", 256*1024),
	}
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func splitList(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func envBool(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	switch strings.ToLower(raw) {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}

func envDuration(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	if d, err := time.ParseDuration(raw); err == nil {
		return d
	}
	if ms, err := strconv.Atoi(raw); err == nil && ms > 0 {
		return time.Duration(ms) * time.Millisecond
	}
	return fallback
}

func envInt64(key string, fallback int64) int64 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	if n, err := strconv.ParseInt(raw, 10, 64); err == nil && n > 0 {
		return n
	}
	return fallback
}
