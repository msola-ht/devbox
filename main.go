package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"devbox/internal/devtools"
	"devbox/internal/proxy"
	"devbox/internal/runtimeenv"
	"devbox/internal/systeminfo"

	"github.com/egoist/mygo"
)

func main() {
	if handled, err := proxy.HandleSystemRouteCommand(os.Args); handled {
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	proxyService := &proxy.Service{}
	toolsService, stopTools := devtools.NewService(func(ctx context.Context) ([]string, error) {
		return proxy.HomebrewEnvironment(ctx, proxyService)
	}, func(ctx context.Context) (*http.Client, error) {
		return proxy.DownloadClient(ctx, proxyService)
	})
	mygo.App.OnQuit(stopTools)

	mygo.BindAs("SystemService", &systeminfo.Service{})
	mygo.BindAs("EnvironmentService", &runtimeenv.Service{})
	mygo.BindAs("ProxyService", proxyService)
	mygo.BindAs("DevToolsService", toolsService)
	mygo.BindAs("ProjectDialog", projectDialog{})

	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{
			Title:           "devbox",
			Width:           1280,
			Height:          720,
			UseContentSize:  true,
			MinWidth:        480,
			MinHeight:       360,
			BackgroundColor: "light-dark(#f6f7f9, #0f1115)",
			// The frontend: devUrl during `mygo dev`, frontendDist
			// embedded by `mygo build` otherwise (see mygo.config.ts).
			URL: "/",
		})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
