package services

// PanelService lets the page hide the panel (after a copy confirmation; Esc in plan 02-02)
// and open Settings (⌘,; the shell hides the panel first).
type PanelService struct {
	HideFunc         func()
	OpenSettingsFunc func()
}

func (p *PanelService) Hide() {
	if p.HideFunc != nil {
		p.HideFunc()
	}
}

func (p *PanelService) OpenSettings() {
	if p.OpenSettingsFunc != nil {
		p.OpenSettingsFunc()
	}
}
