package main

import (
	"embed"
	"net/url"
	"os"

	"github.com/wailsapp/wails/v2"
	wailsassetserver "github.com/wailsapp/wails/v2/pkg/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options"
	assetserveroptions "github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:dist
var assets embed.FS

func main() {
	app := NewDesktopApp()
	err := wails.Run(&options.App{
		Title:            "Common Build Materials Platform",
		Width:            1440,
		Height:           900,
		MinWidth:         1180,
		MinHeight:        760,
		AssetServer:      desktopAssetServer(),
		BackgroundColour: &options.RGBA{R: 245, G: 247, B: 250, A: 1},
		OnStartup:        app.Startup,
		OnShutdown:       app.Shutdown,
		Bind:             []interface{}{app},
		Windows: &windows.Options{
			IsZoomControlEnabled: false,
			ZoomFactor:           1,
			DisablePinchZoom:     true,
		},
		Mac: &mac.Options{
			DisableZoom: true,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}

func desktopAssetServer() *assetserveroptions.Options {
	options := assetserveroptions.Options{Assets: assets}
	devURL := os.Getenv("CBMP_FRONTEND_DEV_URL")
	if devURL == "" {
		return &options
	}

	parsedURL, err := url.Parse(devURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		println("Invalid CBMP_FRONTEND_DEV_URL:", devURL)
		return &options
	}

	return &assetserveroptions.Options{
		Handler: wailsassetserver.NewExternalAssetsHandler(nil, options, parsedURL),
	}
}
