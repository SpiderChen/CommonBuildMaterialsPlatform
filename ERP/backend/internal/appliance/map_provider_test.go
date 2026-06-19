package appliance

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestMapProviderConfigDefaultsToOSM(t *testing.T) {
	clearMapEnv(t)
	cfg := NewMapProviderConfigFromEnv()
	if cfg.Provider != "osm" || cfg.TileStatus() != "default" {
		t.Fatalf("expected default osm provider, got %+v", cfg)
	}
	if cfg.MaxZoom != 19 || cfg.CoordinateSystem != "wgs84" {
		t.Fatalf("unexpected default zoom/coordinate config: %+v", cfg)
	}
}

func TestMapProviderConfigSupportsPrivateTiles(t *testing.T) {
	clearMapEnv(t)
	t.Setenv("CBMP_MAP_TILE_URL", "https://tiles.internal/{z}/{x}/{y}.png")
	t.Setenv("CBMP_MAP_ATTRIBUTION", "客户内网地图")
	t.Setenv("CBMP_MAP_SUBDOMAINS", "a,b")
	t.Setenv("CBMP_MAP_MIN_ZOOM", "4")
	t.Setenv("CBMP_MAP_MAX_ZOOM", "17")

	cfg := NewMapProviderConfigFromEnv()
	if cfg.Provider != "private" || cfg.TileURL != "https://tiles.internal/{z}/{x}/{y}.png" {
		t.Fatalf("expected private tile config, got %+v", cfg)
	}
	if cfg.Attribution != "客户内网地图" || len(cfg.Subdomains) != 2 || cfg.Subdomains[1] != "b" {
		t.Fatalf("expected private attribution/subdomains, got %+v", cfg)
	}
	if cfg.MinZoom != 4 || cfg.MaxZoom != 17 || cfg.TileStatus() != "configured" {
		t.Fatalf("unexpected private zoom/status: %+v", cfg)
	}
}

func TestMapProviderConfigSupportsDomesticProviderKey(t *testing.T) {
	clearMapEnv(t)
	t.Setenv("CBMP_MAP_PROVIDER", "tianditu")
	t.Setenv("CBMP_MAP_API_KEY", "test-key")

	cfg := NewMapProviderConfigFromEnv()
	if cfg.Provider != "tianditu" || cfg.CoordinateSystem != "wgs84" {
		t.Fatalf("expected tianditu config, got %+v", cfg)
	}
	if !cfg.APIKeyConfigured || strings.Contains(cfg.TileURL, "{key}") || !strings.Contains(cfg.TileURL, "test-key") {
		t.Fatalf("expected configured key substitution, got %+v", cfg)
	}
	if got := cfg.PublicRuntimeTileURL(); got != "<configured>" {
		t.Fatalf("runtime URL should be masked, got %q", got)
	}
}

func TestMapProviderConfigSupportsBaiduProvider(t *testing.T) {
	clearMapEnv(t)
	t.Setenv("CBMP_MAP_PROVIDER", "baidu")
	t.Setenv("CBMP_MAP_API_KEY", "baidu-key")

	cfg := NewMapProviderConfigFromEnv()
	if cfg.Provider != "baidu" || cfg.CoordinateSystem != "bd09" || cfg.Name != "百度地图" {
		t.Fatalf("expected baidu bd09 config, got %+v", cfg)
	}
	if !cfg.APIKeyConfigured || strings.Contains(cfg.TileURL, "{key}") || !strings.Contains(cfg.TileURL, "baidu-key") {
		t.Fatalf("expected configured baidu key substitution, got %+v", cfg)
	}
}

func TestSystemMapConfigEndpoint(t *testing.T) {
	clearMapEnv(t)
	t.Setenv("CBMP_MAP_PROVIDER", "amap")
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodGet, "/api/system/map-config", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("map config status %d: %s", rec.Code, rec.Body.String())
	}
	var cfg MapProviderConfig
	if err := json.Unmarshal(rec.Body.Bytes(), &cfg); err != nil {
		t.Fatalf("decode map config: %v", err)
	}
	if cfg.Provider != "amap" || cfg.CoordinateSystem != "gcj02" || cfg.TileURL == "" {
		t.Fatalf("unexpected endpoint map config: %+v", cfg)
	}
}

func clearMapEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"CBMP_MAP_PROVIDER",
		"CBMP_MAP_TILE_URL",
		"CBMP_MAP_ATTRIBUTION",
		"CBMP_MAP_SUBDOMAINS",
		"CBMP_MAP_MIN_ZOOM",
		"CBMP_MAP_MAX_ZOOM",
		"CBMP_MAP_COORDINATE_SYSTEM",
		"CBMP_MAP_API_KEY",
	} {
		t.Setenv(key, "")
	}
}
