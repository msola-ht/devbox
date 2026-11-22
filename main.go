package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync/atomic"

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
		return proxy.CommandEnvironment(ctx, proxyService)
	}, func(ctx context.Context) (*http.Client, error) {
		return proxy.DownloadClient(ctx, proxyService)
	})
	mygo.App.OnQuit(stopTools)
	var installNotice atomic.Bool
	sessionEnding := false
	guardClose := func(sessionEnding bool, preventClose func(), parent *mygo.Window) {
		if sessionEnding || devtools.PrepareQuit(toolsService) {
			return
		}
		preventClose()
		if installNotice.CompareAndSwap(false, true) {
			go func() {
				defer installNotice.Store(false)
				_, _ = mygo.Dialog.Message(mygo.MessageOptions{
					Parent:  parent,
					Title:   "正在安装开发工具",
					Message: "请等待安装结束后再退出。如果系统正在请求授权，请完成或关闭授权窗口。",
				})
			}()
		}
	}
	mygo.App.OnBeforeQuit(func(event *mygo.QuitEvent) {
		sessionEnding = event.SessionEnding
		guardClose(event.SessionEnding, event.PreventDefault, nil)
	})
	mygo.App.OnQuitCanceled(func() { devtools.ResumeAfterCanceledQuit(toolsService) })

	mygo.BindAs("SystemService", &systeminfo.Service{})
	mygo.BindAs("EnvironmentService", &runtimeenv.Service{})
	mygo.BindAs("ProxyService", proxyService)
	mygo.BindAs("DevToolsService", toolsService)
	mygo.BindAs("ProjectDialog", projectDialog{})

	mygo.App.WhenReady(func() {
		window := mygo.NewWindow(mygo.WindowOptions{
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
		window.OnClose(func(event *mygo.CloseEvent) {
			guardClose(sessionEnding, event.PreventDefault, window)
		})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
