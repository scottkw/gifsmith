package shell

import "gifsmith/internal/platform"

// panelOrigin centers a w×h panel horizontally in wa with its top edge one sixth down.
// shortcut: no clamping, a display narrower than the panel overhangs; clamp if that ever matters.
func panelOrigin(wa platform.Rect, w, h int) (x, y int) {
	return wa.X + (wa.Width-w)/2, wa.Y + wa.Height/6
}

func settingsOrigin(wa platform.Rect, w, h int) (x, y int) { return 0, 0 } // RED skeleton
