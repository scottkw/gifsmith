// Package shell owns the Wails panel window and the global hotkey. No cgo here.
package shell

import (
	"log"
	"sync"

	"gifsmith/internal/platform"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

const (
	PanelWidth     = 720
	PanelHeight    = 480
	SettingsWidth  = 640
	SettingsHeight = 520
	Hotkey         = "Ctrl+Option+Cmd+G"
)

type Shell struct {
	app   *application.App
	plat  platform.Platform
	panel *application.WebviewWindow
	// settings is a normal titled window, created hidden once and never closed.
	settings *application.WebviewWindow
	// Hotkey, menu and bound calls each run on their own goroutine; mu serialises
	// Toggle/Show/Hide/OpenSettings and guards prev and settingsPlaced.
	mu             sync.Mutex
	prev           platform.App // frontmost app when the panel was last shown
	settingsPlaced bool         // the first Settings show of a launch is centred; later ones keep the user's spot
}

// New creates the panel once (hidden; never closed) and registers the hotkey after start.
// trayIcon is the menu bar template PNG, supplied by the caller (embedded in main).
func New(app *application.App, plat platform.Platform, trayIcon []byte) *Shell {
	s := &Shell{app: app, plat: plat}
	s.panel = app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:            "panel",
		Width:           PanelWidth,
		Height:          PanelHeight,
		URL:             "/",
		Frameless:       true,
		Hidden:          true,
		DisableResize:   true,
		HideOnEscape:    false, // the page owns Esc: clear the query first, then PanelService.Hide
		HideOnFocusLost: true,
		Mac: application.MacWindow{
			WindowClass:      application.MacWindowClassPanel,
			PanelPreferences: application.MacPanelPreferences{NonActivating: true},
			CollectionBehavior: application.MacWindowCollectionBehaviorCanJoinAllSpaces |
				application.MacWindowCollectionBehaviorFullScreenAuxiliary,
			WindowLevel: application.MacWindowLevelFloating,
		},
	})
	// NSPanel cancelOperation or ⌘W must never destroy the only panel. s.panel.Hide directly,
	// not s.Hide: this runs inside a window callback and must not take s.mu.
	s.panel.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		e.Cancel()
		s.panel.Hide()
	})

	s.settings = app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:                "settings",
		Title:               "Gifsmith Settings",
		Width:               SettingsWidth,
		Height:              SettingsHeight,
		URL:                 "/settings.html",
		Hidden:              true,
		DisableResize:       true,
		MinimiseButtonState: application.ButtonDisabled,
		MaximiseButtonState: application.ButtonHidden,
	})
	// The red button hides Settings; it is never destroyed (window callback: no s.mu).
	s.settings.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		e.Cancel()
		s.settings.Hide()
	})

	// Replaces Wails' default menu: no File (⌘W), View (⌘R reload, zoom) or Window (⌘M).
	// App keeps ⌘Q/⌘H; Edit keeps ⌘X/⌘C/⌘V/⌘A/⌘Z for text fields.
	appMenu := app.NewMenu()
	appMenu.AddRole(application.AppMenu)
	appMenu.AddRole(application.EditMenu)
	app.Menu.SetApplicationMenu(appMenu)

	// Tray is built here, not in the hotkey handler, so Open works even if the hotkey fails.
	// No window is attached: with only a menu, a click opens the menu.
	tray := app.SystemTray.New()
	tray.SetTemplateIcon(trayIcon)
	menu := app.NewMenu()
	menu.Add("Open Gifsmith").OnClick(func(*application.Context) { s.Show() })
	menu.Add("Settings…").OnClick(func(*application.Context) { s.OpenSettings() })
	menu.AddSeparator()
	menu.Add("Quit Gifsmith").OnClick(func(*application.Context) { app.Quit() })
	tray.SetMenu(menu)
	log.Printf("tray menu ready")

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
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.panel.IsVisible() {
		s.panel.Hide()
		return
	}
	s.show()
}

func (s *Shell) Show() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.show()
}

func (s *Shell) show() {
	app, err := s.plat.FrontmostApp()
	if err != nil {
		log.Printf("frontmost app: %v", err)
		app = platform.App{} // never paste into a target left over from an earlier show
	}
	s.prev = app
	if wa := s.plat.CursorScreenWorkArea(); wa.Width > 0 && wa.Height > 0 { // zero: no screen, keep position
		s.panel.SetPosition(panelOrigin(wa, PanelWidth, PanelHeight))
	}
	s.panel.Show()
	s.panel.Focus() // Show() alone never makes a non-activating panel key
	s.plat.FocusWebView(s.panel.NativeWindow())
	s.app.Event.Emit("panel:show")
}

func (s *Shell) Hide() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.panel.Hide()
}

// Previous is the app that was frontmost when the panel was last shown.
func (s *Shell) Previous() platform.App {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.prev
}

// OpenSettings hides the panel, then shows and focuses Settings: Focus activates the accessory
// app so the window is key and takes typing. Re-opening brings it to front; the page keeps its pane.
func (s *Shell) OpenSettings() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.panel.Hide()
	if !s.settingsPlaced {
		if wa := s.plat.CursorScreenWorkArea(); wa.Width > 0 && wa.Height > 0 {
			s.settings.SetPosition(panelOrigin(wa, SettingsWidth, SettingsHeight))
		}
		s.settingsPlaced = true
	}
	s.settings.Show()
	s.settings.Focus()
	s.app.Event.Emit("settings:show")
}

// ApplyTheme sets the native appearance ("" follows macOS) and both windows' backgrounds to --bg.
func (s *Shell) ApplyTheme(theme string) {
	if err := s.plat.SetAppearance(theme); err != nil {
		log.Printf("appearance: %v", err)
	}
	bg := application.NewRGB(0xFF, 0xFF, 0xFF)
	if theme == "dark" || (theme == "" && s.app.Env.IsDarkMode()) {
		bg = application.NewRGB(0x1E, 0x1E, 0x1E)
	}
	s.panel.SetBackgroundColour(bg)
	s.settings.SetBackgroundColour(bg)
}
