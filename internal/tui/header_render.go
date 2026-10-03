package tui

import (
	"fmt"
	"strings"
)

func (a *App) headerPanels() []panelSurface {
	header := strings.TrimSpace(a.renderHeaderBar())
	if header == "" {
		return nil
	}
	return []panelSurface{{key: "header", content: header, minLines: 1, maxLines: 1}}
}

func (a *App) renderHeaderBar() string {
	left := fmt.Sprintf("ALLIECODE  %s", strings.ToUpper(a.stateLabel()))
	right := fmt.Sprintf("%s/%s", a.statusProviderLabel(), a.statusModelLabel())
	return renderStatusLineWithStyle(left, right, a.width, headerBarStyle)
}
