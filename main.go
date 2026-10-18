package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"net/url"

	"gifsmith/internal/platform"
	"gifsmith/internal/provider"
	"gifsmith/internal/provider/finergifs"
	"gifsmith/internal/provider/giphy"
	"gifsmith/internal/provider/klipy"
	"gifsmith/internal/search"
	"gifsmith/internal/services"
	"gifsmith/internal/settings"
	"gifsmith/internal/shell"
	"gifsmith/internal/staging"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed icons/menubar/GifsmithTemplate-tray-2x.png
var trayIcon []byte

// deps is what differs between the desktop build (wire_desktop.go) and the
// -tags server UAT build (wire_server.go). Services are still registered here,
// untagged: the binding generator finds them by static analysis without tags.
type deps struct {
	HTTP                  *http.Client
	FinerBase, FinerMedia string
	GiphyBase, KlipyBase  string
	Secrets               settings.SecretStore
	ConfigDir             string // settings.json lives here
	StagingDir            string
	DownloadsDir          string                           // Download saves here; never created by the app
	Allow                 func(provider.ID, *url.URL) bool // download allowlist for staging
	Platform              platform.Platform                // native in the app; platform.Fake in the UAT build
}

func main() {
	d := wire()
	plat := d.Platform
	// Downloads get their own client: Fetch's 45 s deadline bounds them, and the
	// API client's 15 s timeout would cut off a multi-MB original on a slow link.
	stage := &staging.Store{Dir: d.StagingDir, HTTP: &http.Client{}, Allow: d.Allow}
	actions := &services.ActionService{Platform: plat, Staging: stage, DownloadsDir: func() string { return d.DownloadsDir }}
	panel := &services.PanelService{}
	keys := settings.NewKeys(d.Secrets)
	keys.Load()
	clipsClient := &giphy.Client{BaseURL: d.GiphyBase, Key: keys.Func(provider.GIPHY), HTTP: d.HTTP, Clips: true}
	svc := search.New(map[provider.ID]provider.Provider{
		provider.GIPHY:     &giphy.Client{BaseURL: d.GiphyBase, Key: keys.Func(provider.GIPHY), HTTP: d.HTTP},
		provider.Clips:     clipsClient,
		provider.Klipy:     &klipy.Client{BaseURL: d.KlipyBase, Key: keys.Func(provider.Klipy), HTTP: d.HTTP},
		provider.FinerGifs: &finergifs.Client{BaseURL: d.FinerBase, MediaBase: d.FinerMedia, HTTP: d.HTTP},
	}, nil)
	clips := &search.Clips{
		// Same key, same 429 budget as GIPHY GIFs: a probe 429 blocks both.
		Probe: func(ctx context.Context) error {
			return svc.Guard(provider.Clips, func() error { return clipsClient.ProbeClips(ctx) })
		},
		HasKey:  func() bool { _, ok := keys.Get(provider.GIPHY); return ok },
		Blocked: func() error { return svc.BlockedErr(provider.Clips) },
	}
	gifs := &services.GifService{Svc: svc, Keys: keys, Clips: clips}
	settingsSvc := &services.SettingsService{Keys: keys, Dir: d.ConfigDir, Svc: svc, Clips: clips, NewClient: func(p provider.ID, key string) provider.Provider {
		candidate := func() (string, bool) { return key, true }
		if p == provider.Klipy {
			return &klipy.Client{BaseURL: d.KlipyBase, Key: candidate, HTTP: d.HTTP}
		}
		return &giphy.Client{BaseURL: d.GiphyBase, Key: candidate, HTTP: d.HTTP}
	}}

	// No ApplicationShouldTerminateAfterLastWindowClosed: the hidden panel counts as closed.
	app := application.New(application.Options{
		Name: "Gifsmith",
		Services: []application.Service{
			application.NewService(actions),
			application.NewService(gifs),
			application.NewService(panel),
			application.NewService(settingsSvc),
		},
		MarshalError: func(err error) []byte {
			if b := provider.MarshalError(err); b != nil {
				return b
			}
			return services.MarshalError(err)
		},
		Assets: application.AssetOptions{Handler: application.AssetFileServerFS(assets)},
		Mac:    application.MacOptions{ActivationPolicy: application.ActivationPolicyAccessory},
		// Used only by -tags server builds (browser UAT); WAILS_SERVER_PORT overrides.
		Server: application.ServerOptions{Host: "localhost", Port: 8089},
	})
	settingsSvc.Emit = func(n string) { app.Event.Emit(n) }
	clips.OnChange = func() {
		log.Printf("clips: %s", clips.State()) // state only: records the native probe outcome
		app.Event.Emit("providers:changed")
	}
	sh := shell.New(app, plat, trayIcon)
	actions.Hide = sh.Hide
	actions.Previous = sh.Previous
	panel.HideFunc = sh.Hide

	// Staging holds only the clipboard's file; nothing from an earlier run is kept (D-17).
	// ActionService.ServiceShutdown clears it again at quit.
	if err := stage.Clear(); err != nil {
		log.Printf("staging clear: %v", err)
	}
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
