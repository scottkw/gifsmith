package services

// PanelService lets the page hide the panel (after a copy confirmation; Esc in plan 02-02).
type PanelService struct {
	HideFunc func()
}

func (p *PanelService) Hide() {
	if p.HideFunc != nil {
		p.HideFunc()
	}
}
