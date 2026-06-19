package appliance

import (
	"os"
	"strconv"
	"strings"
)

type MapProviderConfig struct {
	Provider          string   `json:"provider"`
	Name              string   `json:"name"`
	TileURL           string   `json:"tileUrl"`
	Attribution       string   `json:"attribution"`
	Subdomains        []string `json:"subdomains"`
	MinZoom           int      `json:"minZoom"`
	MaxZoom           int      `json:"maxZoom"`
	CoordinateSystem  string   `json:"coordinateSystem"`
	Offline           bool     `json:"offline"`
	APIKeyConfigured  bool     `json:"apiKeyConfigured"`
	RequiresClientKey bool     `json:"requiresClientKey"`
}

func NewMapProviderConfigFromEnv() MapProviderConfig {
	providerEnv := strings.TrimSpace(strings.ToLower(os.Getenv("CBMP_MAP_PROVIDER")))
	provider := fallback(providerEnv, "osm")
	if providerEnv == "" && strings.TrimSpace(os.Getenv("CBMP_MAP_TILE_URL")) != "" {
		provider = "private"
	}

	cfg := mapProviderPreset(provider)
	if customURL := strings.TrimSpace(os.Getenv("CBMP_MAP_TILE_URL")); customURL != "" {
		cfg.TileURL = customURL
		cfg.Offline = false
	}
	if attribution := strings.TrimSpace(os.Getenv("CBMP_MAP_ATTRIBUTION")); attribution != "" {
		cfg.Attribution = attribution
	}
	if subdomains := splitCSV(os.Getenv("CBMP_MAP_SUBDOMAINS")); len(subdomains) > 0 {
		cfg.Subdomains = subdomains
	}
	if minZoom, ok := parseZoomEnv("CBMP_MAP_MIN_ZOOM"); ok {
		cfg.MinZoom = minZoom
	}
	if maxZoom, ok := parseZoomEnv("CBMP_MAP_MAX_ZOOM"); ok {
		cfg.MaxZoom = maxZoom
	}
	if coordinateSystem := strings.TrimSpace(strings.ToLower(os.Getenv("CBMP_MAP_COORDINATE_SYSTEM"))); coordinateSystem != "" {
		cfg.CoordinateSystem = coordinateSystem
	}
	if apiKey := strings.TrimSpace(os.Getenv("CBMP_MAP_API_KEY")); apiKey != "" {
		cfg.APIKeyConfigured = true
		cfg.TileURL = strings.ReplaceAll(cfg.TileURL, "{key}", apiKey)
	}
	cfg.RequiresClientKey = strings.Contains(cfg.TileURL, "{key}")
	cfg.Offline = cfg.Offline || strings.TrimSpace(cfg.TileURL) == ""
	if cfg.Subdomains == nil {
		cfg.Subdomains = []string{}
	}
	return cfg
}

func (c MapProviderConfig) TileStatus() string {
	if c.Offline {
		return "offline"
	}
	if c.RequiresClientKey {
		return "missing-key"
	}
	if c.Provider == "osm" {
		return "default"
	}
	return "configured"
}

func (c MapProviderConfig) PublicRuntimeTileURL() string {
	if c.TileURL == "" {
		return ""
	}
	if c.APIKeyConfigured || c.RequiresClientKey {
		return "<configured>"
	}
	return c.TileURL
}

func mapProviderPreset(provider string) MapProviderConfig {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "baidu", "bmap":
		return MapProviderConfig{
			Provider:         "baidu",
			Name:             "百度地图",
			TileURL:          "https://maponline{s}.bdimg.com/tile/?qt=tile&x={x}&y={y}&z={z}&styles=pl&scaler=1&udt=20260618&ak={key}",
			Attribution:      "百度地图",
			Subdomains:       []string{"0", "1", "2", "3"},
			MinZoom:          3,
			MaxZoom:          18,
			CoordinateSystem: "bd09",
		}
	case "amap", "gaode":
		return MapProviderConfig{
			Provider:         "amap",
			Name:             "高德地图",
			TileURL:          "https://webrd0{s}.is.autonavi.com/appmaptile?lang=zh_cn&size=1&scale=1&style=8&x={x}&y={y}&z={z}",
			Attribution:      "高德地图",
			Subdomains:       []string{"1", "2", "3", "4"},
			MinZoom:          3,
			MaxZoom:          18,
			CoordinateSystem: "gcj02",
		}
	case "tianditu":
		return MapProviderConfig{
			Provider:         "tianditu",
			Name:             "天地图",
			TileURL:          "https://t{s}.tianditu.gov.cn/DataServer?T=vec_w&x={x}&y={y}&l={z}&tk={key}",
			Attribution:      "天地图",
			Subdomains:       []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			MinZoom:          3,
			MaxZoom:          18,
			CoordinateSystem: "wgs84",
		}
	case "tencent":
		return MapProviderConfig{
			Provider:         "tencent",
			Name:             "腾讯地图",
			TileURL:          "https://rt{s}.map.gtimg.com/tile?z={z}&x={x}&y={y}&type=vector&styleid=1",
			Attribution:      "腾讯地图",
			Subdomains:       []string{"0", "1", "2", "3"},
			MinZoom:          3,
			MaxZoom:          18,
			CoordinateSystem: "gcj02",
		}
	case "private":
		return MapProviderConfig{
			Provider:         "private",
			Name:             "客户私有瓦片",
			Attribution:      "Private tiles",
			MinZoom:          0,
			MaxZoom:          19,
			CoordinateSystem: "wgs84",
		}
	case "offline":
		return MapProviderConfig{
			Provider:         "offline",
			Name:             "离线地图",
			Attribution:      "Offline",
			MinZoom:          0,
			MaxZoom:          19,
			CoordinateSystem: "wgs84",
			Offline:          true,
		}
	default:
		return MapProviderConfig{
			Provider:         "osm",
			Name:             "OpenStreetMap",
			TileURL:          "https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png",
			Attribution:      "&copy; OpenStreetMap contributors",
			Subdomains:       []string{"a", "b", "c"},
			MinZoom:          0,
			MaxZoom:          19,
			CoordinateSystem: "wgs84",
		}
	}
}

func parseZoomEnv(key string) (int, bool) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return 0, false
	}
	zoom, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}
	if zoom < 0 {
		zoom = 0
	}
	if zoom > 22 {
		zoom = 22
	}
	return zoom, true
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
