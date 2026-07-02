package clientupdater

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func LoadConfig(path string) (Config, error) {
	cfg, err := LoadConfigFile(path)
	if err != nil {
		return cfg, err
	}
	return NormalizeConfig(cfg)
}

func LoadConfigFile(path string) (Config, error) {
	cfg := defaultConfig()
	if path != "" {
		raw, err := os.ReadFile(path)
		if err != nil {
			return cfg, err
		}
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return cfg, err
		}
	}
	applyEnv(&cfg)
	return cfg, nil
}

func NormalizeConfig(cfg Config) (Config, error) {
	if cfg.BaseURL == "" {
		return cfg, fmt.Errorf("baseUrl is required")
	}
	if cfg.UpdaterToken == "" {
		return cfg, fmt.Errorf("updaterToken is required")
	}
	if cfg.RootDir == "" {
		return cfg, fmt.Errorf("rootDir is required")
	}
	if cfg.PollIntervalSeconds <= 0 {
		cfg.PollIntervalSeconds = 30
	}
	if cfg.Components == nil {
		cfg.Components = map[string]ComponentConfig{}
	}
	abs, err := filepath.Abs(cfg.RootDir)
	if err != nil {
		return cfg, err
	}
	cfg.RootDir = abs
	return cfg, nil
}

func defaultConfig() Config {
	return Config{
		BaseURL:               "http://127.0.0.1:8095",
		RootDir:               "data/client-updater",
		TargetComponent:       "client",
		PollIntervalSeconds:   30,
		ResumeDownloads:       true,
		AutoRollbackOnFailure: true,
		Components:            map[string]ComponentConfig{},
	}
}

func applyEnv(cfg *Config) {
	if value := strings.TrimSpace(os.Getenv("CBMP_CLIENT_UPDATER_BASE_URL")); value != "" {
		cfg.BaseURL = value
	}
	if value := strings.TrimSpace(os.Getenv("CBMP_CLIENT_UPDATER_TOKEN")); value != "" {
		cfg.UpdaterToken = value
	}
	if value := strings.TrimSpace(os.Getenv("CBMP_CLIENT_UPDATER_WATERMARK")); value != "" {
		cfg.Watermark = value
	}
	if value := strings.TrimSpace(os.Getenv("CBMP_CLIENT_UPDATER_ROOT")); value != "" {
		cfg.RootDir = value
	}
	if value := strings.TrimSpace(os.Getenv("CBMP_CLIENT_UPDATER_TARGET_COMPONENT")); value != "" {
		cfg.TargetComponent = value
	}
	if value := strings.TrimSpace(os.Getenv("CBMP_CLIENT_UPDATER_INTERVAL_SECONDS")); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			cfg.PollIntervalSeconds = parsed
		}
	}
	if value := strings.TrimSpace(os.Getenv("CBMP_CLIENT_UPDATER_STRICT_CHECKSUM")); value != "" {
		cfg.StrictChecksum = value == "1" || strings.EqualFold(value, "true") || strings.EqualFold(value, "yes")
	}
	if value := strings.TrimSpace(os.Getenv("CBMP_CLIENT_UPDATER_RESUME_DOWNLOADS")); value != "" {
		cfg.ResumeDownloads = value == "1" || strings.EqualFold(value, "true") || strings.EqualFold(value, "yes")
	}
	if value := strings.TrimSpace(os.Getenv("CBMP_CLIENT_UPDATER_AUTO_ROLLBACK")); value != "" {
		cfg.AutoRollbackOnFailure = value == "1" || strings.EqualFold(value, "true") || strings.EqualFold(value, "yes")
	}
}

func SampleConfig() Config {
	return Config{
		BaseURL:               "http://127.0.0.1:8095",
		UpdaterToken:          "updater-token-from-operations-platform",
		Watermark:             "CBMP-CUSTOMER",
		RootDir:               "/opt/cbmp/client-updater",
		TargetComponent:       "client",
		PollIntervalSeconds:   30,
		StrictChecksum:        false,
		ResumeDownloads:       true,
		AutoRollbackOnFailure: true,
		Components: map[string]ComponentConfig{
			"client": {StopCommand: []string{"systemctl", "stop", "cbmp-client"}, StartCommand: []string{"systemctl", "start", "cbmp-client"}, HealthCommand: []string{"systemctl", "is-active", "cbmp-client"}, TimeoutSeconds: 60},
		},
	}
}

func WriteSampleConfig() error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(SampleConfig())
}
