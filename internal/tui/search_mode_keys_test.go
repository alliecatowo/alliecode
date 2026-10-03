package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSearchModeLeftArrowExitsQuickOpen(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	if app.inputMode != inputModeQuickOpen {
		t.Fatalf("expected quick-open mode, got %s", app.inputMode)
	}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyLeft}, "left")
	if app.state != stateIdle {
		t.Fatalf("expected left arrow to exit search, got state %d", app.state)
	}
	if app.inputMode != inputModeChat {
		t.Fatalf("expected chat mode after left arrow, got %s", app.inputMode)
	}
}

func TestSearchModeRightArrowStagesHistorySelection(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.input.history = []string{"deploy release", "open logs"}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlR}, "ctrl+r")
	app = typeTestText(t, app, "log")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyRight}, "right")
	if got := app.input.Value(); got != "open logs" {
		t.Fatalf("expected right arrow to stage history selection, got %q", got)
	}
	if app.state != stateIdle {
		t.Fatalf("expected history search to close after right arrow, got %d", app.state)
	}
}

func TestSearchModeEnterExitsTimelineSearch(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.addTimeline(timelineEntry{kind: timelineAssistant, text: "status ready"})
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlF}, "ctrl+f")
	app = typeTestText(t, app, "status")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	if app.state != stateIdle {
		t.Fatalf("expected enter to close timeline search, got %d", app.state)
	}
}

func TestSearchModeRightArrowExitsTimelineSearch(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.addTimeline(timelineEntry{kind: timelineAssistant, text: "status ready"})
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlF}, "ctrl+f")
	app = typeTestText(t, app, "status")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyRight}, "right")
	if app.state != stateIdle {
		t.Fatalf("expected right arrow to close timeline search, got %d", app.state)
	}
}

func TestSearchModeShiftTabSequenceVariantReversesQuickOpenSelection(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyTab}, "tab")
	selected := app.quickOpen.selected
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'', '[', 'Z'}}, "shift-tab-csi-z")
	if app.quickOpen.selected == selected {
		t.Fatalf("expected shift+tab sequence alias to reverse quick-open selection")
	}
	if app.inputMode != inputModeQuickOpen {
		t.Fatalf("expected quick-open mode after reverse, got %s", app.inputMode)
	}
}
