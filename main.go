package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options.
	//
	// The floating bar is a fixed-size window that keeps the native
	// close/minimize traffic lights (users can dismiss the app or send it
	// to the Dock as usual). We enforce the exact bar dimensions with
	// matching Min/Max size so the user cannot grab a corner and resize.
	// Programmatic resize (used when entering Capture Region selection) is
	// still permitted because App.tsx relaxes these bounds via
	// WindowSetMinSize / WindowSetMaxSize before calling WindowSetSize,
	// then re-locks them on restore.
	err := wails.Run(&options.App{
		Title:       "pasha-go",
		Width:       960,
		Height:      96,
		MinWidth:    960,
		MinHeight:   96,
		MaxWidth:    960,
		MaxHeight:   96,
		AlwaysOnTop: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		Mac: &mac.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			// Keep the traffic-light buttons (close/minimize) visible in
			// the top-left corner while hiding the title bar background,
			// so the window still looks like a compact floating bar.
			TitleBar: mac.TitleBarHiddenInset(),
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
