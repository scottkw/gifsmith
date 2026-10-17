// Package shell owns the Wails panel window and the global hotkey. No cgo here.
package shell

import (
	"log"

	"gifsmith/internal/platform"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

const (
	PanelWidth  = 720
	PanelHeight = 480
	Hotkey      = "Ctrl+Option+Cmd+G"
)

type Shell struct {
	app   *application.App
	plat  platform.Platform
	panel *application.WebviewWindow
}

// New creates the panel once (hidden; never closed) and registers the hotkey after start.
func New(app *application.App, plat platform.Platform) *Shell {
	s := &Shell{app: app, plat: plat}
	s.panel = app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:            "panel",
		Width:           PanelWidth,
		Height:          PanelHeight,
		URL:             "/",
		Frameless:       true,
		Hidden:          true,
		DisableResize:   true,
		HideOnEscape:    true,
		HideOnFocusLost: true,
		Mac: application.MacWindow{
			WindowClass:      application.MacWindowClassPanel,
			PanelPreferences: application.MacPanelPreferences{NonActivating: true},
			CollectionBehavior: application.MacWindowCollectionBehaviorCanJoinAllSpaces |
				application.MacWindowCollectionBehaviorFullScreenAuxiliary,
			WindowLevel: application.MacWindowLevelFloating,
		},
	})

	// Registered once the app runs so an OS rejection comes back here instead of the error handler.
	app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*application.ApplicationEvent) {
		if err := app.GlobalShortcut.Register(Hotkey, s.Toggle); err != nil {
			log.Printf("hotkey register error: %v", err)
			return
		}
		log.Printf("hotkey registered: %s", Hotkey)
	})
	return s
}

func (s *Shell) Toggle() {
	if s.panel.IsVisible() {
		s.Hide()
		return
	}
	s.Show()
}

func (s *Shell) Show() {
	s.panel.Center() // plan 01-02: cursor-screen placement
	s.panel.Show()
	s.panel.Focus() // Show() alone never makes a non-activating panel key
	s.plat.FocusWebView(s.panel.NativeWindow())
	s.app.Event.Emit("panel:show")
}

func (s *Shell) Hide() { s.panel.Hide() }
