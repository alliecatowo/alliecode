package tui

type measuredPanelGroup struct {
	panels []panelSurface
	height int
}

type measuredLayout struct {
	header   measuredPanelGroup
	status   measuredPanelGroup
	overlays measuredPanelGroup
	composer measuredPanelGroup
	buddy    measuredPanelGroup
	reserved int
}

func (a *App) composeMeasuredLayout(width int) measuredLayout {
	headerPanels := a.headerPanels()
	statusPanels := a.statusPanels()
	overlayPanels := a.overlayPanels()
	composerPanels := a.composerPanels()
	buddyPanels := a.buddyPanels()
	headerHeight := a.panelStackHeight(headerPanels, width)
	statusHeight := a.panelStackHeight(statusPanels, width)
	overlayHeight := a.panelStackHeight(overlayPanels, width)
	composerHeight := a.panelStackHeight(composerPanels, width)
	buddyHeight := a.panelStackHeight(buddyPanels, width)
	return measuredLayout{
		header:   measuredPanelGroup{panels: headerPanels, height: headerHeight},
		status:   measuredPanelGroup{panels: statusPanels, height: statusHeight},
		overlays: measuredPanelGroup{panels: overlayPanels, height: overlayHeight},
		composer: measuredPanelGroup{panels: composerPanels, height: composerHeight},
		buddy:    measuredPanelGroup{panels: buddyPanels, height: buddyHeight},
		reserved: headerHeight + statusHeight + overlayHeight + composerHeight + buddyHeight,
	}
}
