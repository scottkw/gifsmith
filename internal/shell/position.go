package shell

import "gifsmith/internal/platform"

// panelOrigin centers a w×h panel horizontally in wa with its top edge one sixth down.
// shortcut: no clamping, a display narrower than the panel overhangs; clamp if that ever matters.
func panelOrigin(wa platform.Rect, w, h int) (x, y int) {
	return wa.X + (wa.Width-w)/2, wa.Y + wa.Height/6
}

// settingsOrigin places the Settings window the way the panel is placed.
func settingsOrigin(wa platform.Rect, w, h int) (x, y int) { return panelOrigin(wa, w, h) }
