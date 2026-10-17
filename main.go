package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"gifsmith/internal/platform"
	"gifsmith/internal/provider"
	"gifsmith/internal/provider/finergifs"
	"gifsmith/internal/services"
	"gifsmith/internal/shell"

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
}

// stageSample writes the embedded sample GIF to the cache dir. The file must persist:
// the pasteboard holds a URL to it, not its bytes.
func stageSample() (string, error) {
	b, err := fs.ReadFile(assets, "frontend/dist/sample.gif")
	if err != nil {
		return "", err
	}
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	p := filepath.Join(dir, "com.kenscott.gifsmith", "staging", "sample.gif")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return "", err
	}
	return p, os.WriteFile(p, b, 0o644)
}

func main() {
	sample, err := stageSample()
	if err != nil {
		log.Fatal(err)
	}
	d := wire()
	plat := platform.New()
	actions := &services.ActionService{Platform: plat, SamplePath: sample}
	gifs := &services.GifService{Providers: map[provider.ID]provider.Provider{
		provider.FinerGifs: &finergifs.Client{BaseURL: d.FinerBase, MediaBase: d.FinerMedia, HTTP: d.HTTP},
	}}

	// No ApplicationShouldTerminateAfterLastWindowClosed: the hidden panel counts as closed.
	app := application.New(application.Options{
		Name: "Gifsmith",
		Services: []application.Service{
			application.NewService(actions),
			application.NewService(gifs),
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

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
