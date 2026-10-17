package main

import (
	"embed"
	"log"
	"net/http"
	"net/url"

	"gifsmith/internal/platform"
	"gifsmith/internal/provider"
	"gifsmith/internal/provider/finergifs"
	"gifsmith/internal/services"
	"gifsmith/internal/shell"
	"gifsmith/internal/staging"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

// deps is what differs between the desktop build (wire_desktop.go) and the
// -tags server UAT build (wire_server.go). Services are still registered here,
// untagged: the binding generator finds them by static analysis without tags.
type deps struct {
	HTTP                  *http.Client
	FinerBase, FinerMedia string
	StagingDir            string
	Allow                 func(provider.ID, *url.URL) bool // download allowlist for staging
}

func main() {
	d := wire()
	plat := platform.New()
	// Downloads get their own client: Fetch's 45 s deadline bounds them, and the
	// API client's 15 s timeout would cut off a multi-MB original on a slow link.
	stage := &staging.Store{Dir: d.StagingDir, HTTP: &http.Client{}, Allow: d.Allow}
	actions := &services.ActionService{Platform: plat, Staging: stage}
	panel := &services.PanelService{}
	gifs := &services.GifService{Providers: map[provider.ID]provider.Provider{
		provider.FinerGifs: &finergifs.Client{BaseURL: d.FinerBase, MediaBase: d.FinerMedia, HTTP: d.HTTP},
	}}

	// No ApplicationShouldTerminateAfterLastWindowClosed: the hidden panel counts as closed.
	app := application.New(application.Options{
		Name: "Gifsmith",
		Services: []application.Service{
			application.NewService(actions),
			application.NewService(gifs),
			application.NewService(panel),
		},
		MarshalError: provider.MarshalError,
		Assets:       application.AssetOptions{Handler: application.AssetFileServerFS(assets)},
		Mac:          application.MacOptions{ActivationPolicy: application.ActivationPolicyAccessory},
		// Used only by -tags server builds (browser UAT); WAILS_SERVER_PORT overrides.
		Server: application.ServerOptions{Host: "localhost", Port: 8089},
	})
	sh := shell.New(app, plat)
	actions.Hide = sh.Hide
	actions.Previous = sh.Previous
	panel.HideFunc = sh.Hide

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
