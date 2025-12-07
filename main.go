package main

import (
	"embed"

    "github.com/wailsapp/wails/v2"
    "github.com/wailsapp/wails/v2/pkg/options"
    "github.com/wailsapp/wails/v2/pkg/options/assetserver"
	// "github.com/wailsapp/wails/v2/pkg/menu"
	// "context"
	// "github.com/getlantern/systray"
	// "github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)


//go:embed favicon.ico
// var appIcon []byte

//go:embed all:frontend/dist
var assets embed.FS


func main() {
	// Create an instance of the app structure
	app := NewApp()
	// go systray.Run(func() { onReady(app) }, onExit)

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "YzuAutologin",
		Width:  300,
		Height: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		OnStartup:        app.startup,
        Frameless: true,
        DisableResize: true,
		Bind: []interface{}{
			app,
		},
        Windows: &windows.Options{
			WebviewIsTransparent:              true,
			WindowIsTranslucent:               true,
			BackdropType:                      windows.Acrylic,
			DisableWindowIcon:                 false,
			DisableFramelessWindowDecorations: false,
			WebviewUserDataPath:               "",
			Theme:                             windows.SystemDefault,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
	

}

// func onReady(a *App) {
// 	// systray.SetIcon(appIcon)
//     // systray.SetTitle("My Wails App")
//     // systray.SetTooltip("My Wails App Tooltip")

//     // mShow := systray.AddMenuItem("Show", "Show the application")
//     // mHide := systray.AddMenuItem("Hide", "Hide the application")
//     // mQuit := systray.AddMenuItem("Quit", "Quit the application")

//     // go func() {
//     //     for {
//     //         select {
//     //         case <-mShow.ClickedCh:
//     //             runtime.Show(a.ctx)
//     //         case <-mHide.ClickedCh:
//     //             runtime.Hide(a.ctx)
//     //         case <-mQuit.ClickedCh:
//     //             runtime.Quit(a.ctx)
//     //         }
//     //     }
//     // }()
// }

// func onExit() {
//     // 清理工作
// }