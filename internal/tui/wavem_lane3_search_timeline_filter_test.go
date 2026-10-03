package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestWaveMLane3_TimelineSearchFiltersWhileActiveOnly(t *testing.T) {
	app := readySizedApp(t, 120, 28)
	app.addTimeline(timelineEntry{kind: timelineAssistant, text: "keep this row", turn: 1})
	app.addTimeline(timelineEntry{kind: timelineAssistant, text: "needle row", turn: 2})
	app.addTimeline(timelineEntry{kind: timelineAssistant, text: "keep too", turn: 3})

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlF}, "ctrl+f")
	app = typeTestText(t, app, "needle")
	if len(app.visibleRows) != 1 || app.visibleRows[0] != 1 {
		t.Fatalf("expected only needle row visible during timeline search, got visible=%v", app.visibleRows)
	}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")
	if len(app.visibleRows) != 3 {
		t.Fatalf("expected full timeline visible after search exit, got visible=%v", app.visibleRows)
	}
}
