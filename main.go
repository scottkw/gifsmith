package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"gifsmith/internal/library"
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
	"github.com/wailsapp/wails/v3/pkg/events"
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
	ConfigDir             string // settings.json and the library files live here
	StagingDir            string
	DownloadsDir          string                             // Download saves here; never created by the app
	Allow                 func(provider.ID, *url.URL) bool   // download allowlist for staging
	Platform              platform.Platform                  // native in the app; platform.Fake in the UAT build
	Autostart             services.Autostarter               // nil in the app (app.Autostart); in-memory in the UAT build
	Registrar             services.Registrar                 // nil in the app (app.GlobalShortcut); a fake in the UAT build
	PickFolder            func(start string) (string, error) // nil in the app (the native picker); FAKE_PICK_FOLDER in the UAT build
}

func main() {
	d := wire()
	plat := d.Platform
	// Downloads get their own client: Fetch's 45 s deadline bounds them, and the
	// API client's 15 s timeout would cut off a multi-MB original on a slow link.
	stage := &staging.Store{Dir: d.StagingDir, HTTP: &http.Client{}, Allow: d.Allow}
	actions := &services.ActionService{Platform: plat, Staging: stage}
	// Favorites, Recents and recent searches live next to settings.json; metadata only (D-09).
	lib := &library.Store{Dir: d.ConfigDir, Allow: d.Allow}
	actions.Recents = lib
	panel := &services.PanelService{}
	keys := settings.NewKeys(d.Secrets)
	keys.Load()
	var settingsSvc *services.SettingsService // set below, before any request; clients read Params per request
	params := func() provider.Params { return settingsSvc.Params() }
	clipsClient := &giphy.Client{BaseURL: d.GiphyBase, Key: keys.Func(provider.GIPHY), HTTP: d.HTTP, Clips: true, Params: params}
	svc := search.New(map[provider.ID]provider.Provider{
		provider.GIPHY:     &giphy.Client{BaseURL: d.GiphyBase, Key: keys.Func(provider.GIPHY), HTTP: d.HTTP, Params: params},
		provider.Clips:     clipsClient,
		provider.Klipy:     &klipy.Client{BaseURL: d.KlipyBase, Key: keys.Func(provider.Klipy), HTTP: d.HTTP, Params: params},
		provider.FinerGifs: &finergifs.Client{BaseURL: d.FinerBase, MediaBase: d.FinerMedia, HTTP: d.HTTP, Params: params},
	}, nil)
	clips := &search.Clips{
		// Same key, same 429 budget as GIPHY GIFs: a probe 429 blocks both.
		Probe: func(ctx context.Context) error {
			return svc.Guard(provider.Clips, func() error { return clipsClient.ProbeClips(ctx) })
		},
		HasKey:  func() bool { _, ok := keys.Get(provider.GIPHY); return ok },
		Blocked: func() error { return svc.BlockedErr(provider.Clips) },
	}
	store := settings.OpenStore(d.ConfigDir) // the only settings.json owner
	gifs := &services.GifService{Svc: svc, Keys: keys, Clips: clips}
	settingsSvc = &services.SettingsService{Keys: keys, Store: store, Platform: plat, Executable: os.Executable, Svc: svc, Clips: clips, NewClient: func(p provider.ID, key string) provider.Provider {
		candidate := func() (string, bool) { return key, true }
		if p == provider.Klipy {
			return &klipy.Client{BaseURL: d.KlipyBase, Key: candidate, HTTP: d.HTTP, Params: params}
		}
		return &giphy.Client{BaseURL: d.GiphyBase, Key: candidate, HTTP: d.HTTP, Params: params}
	}, PickFolder: d.PickFolder, DefaultDownloads: d.DownloadsDir}
	actions.DownloadsDir = settingsSvc.DownloadDir // read at call time: a new folder applies to the next download

	var sh *shell.Shell // set below, before app.Run; a second launch arrives only after Run starts
	// No ApplicationShouldTerminateAfterLastWindowClosed: the hidden panel counts as closed.
	app := application.New(application.Options{
		Name: "Gifsmith",
		// One process owns the library, staging and the hotkey (WR-01): a second launch, including
		// a wails3 dev build while the installed app runs, exits and shows this panel instead.
		// No-op in -tags server builds.
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID:               "com.kenscott.gifsmith",
			OnSecondInstanceLaunch: func(application.SecondInstanceData) { sh.Show() },
		},
		Services: []application.Service{
			application.NewService(actions),
			application.NewService(gifs),
			application.NewService(panel),
			application.NewService(settingsSvc),
			application.NewService(&services.LibraryService{Store: lib}),
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
	settingsSvc.Emit = func(n string, data ...any) { app.Event.Emit(n, data...) }
	settingsSvc.Autostart = d.Autostart
	if d.Autostart == nil {
		settingsSvc.Autostart = app.Autostart // SMAppService on macOS 13+ for the bundled app
	}
	clips.OnChange = func() {
		log.Printf("clips: %s", clips.State()) // state only: records the native probe outcome
		app.Event.Emit("providers:changed")
	}
	sh = shell.New(app, plat, trayIcon)
	if settingsSvc.PickFolder == nil {
		settingsSvc.PickFolder = sh.PickFolder // the native NSOpenPanel sheet on Settings
	}
	reg := d.Registrar
	if reg == nil {
		reg = app.GlobalShortcut // Carbon hotkeys: no Accessibility needed
	}
	hk := &services.Hotkeys{Reg: reg, Actions: map[string]func(){
		"open":      sh.Toggle,
		"favorites": func() { sh.ShowTab("favorites") },
		"recents":   func() { sh.ShowTab("recents") },
	}, Emit: settingsSvc.Emit, Now: time.Now}
	settingsSvc.Hotkeys = hk
	settingsSvc.OnOpenHotkey = sh.SetOpenAccelerator
	sh.OnSettingsInactive = hk.Resume
	actions.Hide = sh.Hide
	actions.Previous = sh.Previous
	panel.HideFunc = sh.Hide
	panel.OpenSettingsFunc = sh.OpenSettings
	settingsSvc.ApplyTheme = sh.ApplyTheme
	// The stored theme drives native appearance and window backgrounds before any window shows.
	// Hotkeys register once the app runs, so an OS rejection comes back to Start instead of the error handler.
	app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*application.ApplicationEvent) {
		sh.ApplyTheme(store.Get().Theme)
		settingsSvc.ApplyStartup()
		hk.Start(store.Get().Hotkeys)
		sh.SetOpenAccelerator(hk.Bindings().Open)
	})
	if d.Registrar != nil {
		hk.Start(store.Get().Hotkeys) // the server build's fake: server mode never fires ApplicationStarted (and has no tray)
	}

	// Staging holds only the clipboard's file; nothing from an earlier run is kept (D-17).
	// ActionService.ServiceShutdown clears it again at quit.
	if err := stage.Clear(); err != nil {
		log.Printf("staging clear: %v", err)
	}
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
