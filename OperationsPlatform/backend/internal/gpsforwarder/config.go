package forwarder

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
	DeviceKey       string
	SharedSecret    string
	DefaultProtocol string
	DefaultSource   string
	ForwardMode     string
	IncludeRaw      bool
	DedupeWindow    time.Duration
	PollInterval    time.Duration
	HTTPTimeout     time.Duration
	MaxBodyBytes    int64
}

func LoadConfigFromEnv() Config {
	return Config{
		HTTPAddr:        env("GPSF_ADDR", "0.0.0.0:19102"),
		TCPAddr:         strings.TrimSpace(os.Getenv("GPSF_TCP_ADDR")),
		InputFile:       strings.TrimSpace(os.Getenv("GPSF_FILE")),
		OutputFile:      strings.TrimSpace(os.Getenv("GPSF_OUT_FILE")),
		HTTPTargets:     splitList(os.Getenv("GPSF_HTTP_TARGETS")),
		HTTPBearer:      strings.TrimSpace(os.Getenv("GPSF_HTTP_BEARER_TOKEN")),
		DeviceKey:       strings.TrimSpace(os.Getenv("GPSF_DEVICE_KEY")),
		SharedSecret:    strings.TrimSpace(os.Getenv("GPSF_SHARED_SECRET")),
		DefaultProtocol: env("GPSF_PROTOCOL", "gps-json"),
		DefaultSource:   env("GPSF_SOURCE", "gps-forwarder"),
		ForwardMode:     env("GPSF_FORWARD_MODE", "protocol-frame"),
		IncludeRaw:      envBool("GPSF_INCLUDE_RAW", false),
		DedupeWindow:    envDuration("GPSF_DEDUPE_WINDOW", 10*time.Minute),
		PollInterval:    envDuration("GPSF_POLL_INTERVAL", 2*time.Second),
		HTTPTimeout:     envDuration("GPSF_HTTP_TIMEOUT", 5*time.Second),
		MaxBodyBytes:    envInt64("GPSF_MAX_BODY_BYTES", 128*1024),
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
