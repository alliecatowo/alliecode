package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestWaveMLane3_OverlayChurnPreservesViewportAnchor(t *testing.T) {
	app := readySizedApp(t, 120, 30)
	for i := 0; i < 180; i++ {
		app.addTimeline(timelineEntry{kind: timelineAssistant, text: strings.Repeat("segment ", 16), turn: i + 1})
	}
	app.followTail = false
	app.viewport.SetYOffset(44)
	base := app.viewport.YOffset
	for i := 0; i < 12; i++ {
		app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
		app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")
	}
	if diff := absInt(app.viewport.YOffset - base); diff > 2 {
		t.Fatalf("expected stable anchor after overlay churn, base=%d now=%d", base, app.viewport.YOffset)
	}
}
