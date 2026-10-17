package shell

import (
	"testing"

	"gifsmith/internal/platform"
)

func TestPanelOrigin(t *testing.T) {
	cases := []struct {
		name   string
		wa     platform.Rect
		wx, wy int
	}{
		{"primary", platform.Rect{X: 0, Y: 38, Width: 1710, Height: 1074}, 495, 217},
		{"secondary", platform.Rect{X: 1710, Y: -200, Width: 2560, Height: 1415}, 2630, 35},
		{"narrower than panel", platform.Rect{X: 0, Y: 0, Width: 600, Height: 900}, -60, 150},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			x, y := panelOrigin(c.wa, PanelWidth, PanelHeight)
			if x != c.wx || y != c.wy {
				t.Errorf("panelOrigin(%+v) = (%d, %d), want (%d, %d)", c.wa, x, y, c.wx, c.wy)
			}
		})
	}
}
