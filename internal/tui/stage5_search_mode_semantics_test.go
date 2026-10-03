package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestStage5SearchModeSemanticsRightAppliesSelectedHistoryPrompt(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	app.input.history = []string{"deploy release", "open logs"}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlR}, "ctrl+r")
	app = typeTestText(t, app, "log")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyRight}, "right")
	if got := app.input.Value(); got != "open logs" {
		t.Fatalf("expected right to stage selected history prompt, got %q", got)
	}
	if app.state != stateIdle {
		t.Fatalf("expected history panel closed after right apply, state=%d", app.state)
	}
}

func TestStage5SearchModeSemanticsEnterExitsTimelineAfterSelection(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	app.addTimeline(timelineEntry{kind: timelineAssistant, text: "searchable status row"})
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlF}, "ctrl+f")
	app = typeTestText(t, app, "status")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	if app.state != stateIdle {
		t.Fatalf("expected timeline search to close on enter, got state=%d", app.state)
	}
}
