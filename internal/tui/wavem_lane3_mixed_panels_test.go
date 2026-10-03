package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestWaveMLane3_MixedPanelTransitionsKeepViewportValid(t *testing.T) {
	app := readySizedApp(t, 110, 24)
	for i := 0; i < 90; i++ {
		app.addTimeline(timelineEntry{kind: timelineAssistant, text: "output output output", turn: i + 1})
	}
	app.followTail = false
	app.viewport.SetYOffset(20)
	steps := []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune("/")},
		{Type: tea.KeyEsc},
		{Type: tea.KeyCtrlO},
		{Type: tea.KeyEsc},
		{Type: tea.KeyCtrlR},
		{Type: tea.KeyEsc},
		{Type: tea.KeyCtrlF},
		{Type: tea.KeyEsc},
	}
	for _, step := range steps {
		app = sendTestKey(t, app, step, "step")
		if app.viewport.YOffset < 0 {
			t.Fatalf("expected non-negative viewport offset, got %d", app.viewport.YOffset)
		}
	}
}
