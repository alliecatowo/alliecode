package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestWaveMLane3_SearchTransitionPreservesAnchorWhenLeavingQuickOpen(t *testing.T) {
	app := readySizedApp(t, 120, 30)
	for i := 0; i < 140; i++ {
		app.addTimeline(timelineEntry{kind: timelineAssistant, text: strings.Repeat("payload ", 12), turn: i + 1})
	}
	app.followTail = false
	app.viewport.SetYOffset(36)
	base := app.viewport.YOffset
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")
	if diff := absInt(app.viewport.YOffset - base); diff > 2 {
		t.Fatalf("expected stable offset after quick-open enter/exit, base=%d now=%d", base, app.viewport.YOffset)
	}
}
