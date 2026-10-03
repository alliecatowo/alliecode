package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTeaStage3ShiftTabReverseTraversalAcrossModes(t *testing.T) {
	app := readySizedApp(t, 160, 28)

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyDown}, "down")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyShiftTab}, "shift+tab")
	if app.quickOpen.selected != 0 {
		t.Fatalf("expected quick-open shift+tab reverse to first row, got %d", app.quickOpen.selected)
	}

	if !app.openInteractiveCommandPanel("permissions") {
		t.Fatalf("expected permissions panel to open")
	}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyDown}, "down")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyShiftTab}, "shift+tab")
	if app.commandPanel.selected != 0 {
		t.Fatalf("expected command-panel shift+tab reverse to previous row, got %d", app.commandPanel.selected)
	}

	app = readySizedApp(t, 160, 28)
	app.addTimeline(timelineEntry{kind: timelineAssistant, text: "status alpha"})
	app.addTimeline(timelineEntry{kind: timelineAssistant, text: "status beta"})
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlF}, "ctrl+f")
	app = typeTestText(t, app, "status")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlN}, "ctrl+n")
	if app.matchPos < 1 {
		t.Fatalf("expected second timeline match before reverse cycle, got %d", app.matchPos)
	}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyShiftTab}, "shift+tab")
	if app.matchPos != 0 {
		t.Fatalf("expected timeline shift+tab reverse to previous match, got %d", app.matchPos)
	}
}

func TestTeaStage3SingleEnterSemanticsPreserveStagingVsImmediate(t *testing.T) {
	app := readySizedApp(t, 160, 28)

	app = typeTestText(t, app, "/status")
	app = sendTestKeyAndRun(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	if len(app.timeline) != 1 {
		t.Fatalf("expected immediate slash command to run once, got %d rows", len(app.timeline))
	}
	if !strings.Contains(app.timeline[0].text, "STATUS_REPORT") {
		t.Fatalf("expected status output row, got %q", app.timeline[0].text)
	}

	app = readySizedApp(t, 160, 28)
	app = typeTestText(t, app, "/model")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	if got := app.input.Value(); got != "/model " {
		t.Fatalf("expected arg-required slash command to stage on first enter, got %q", got)
	}
	if !app.modelPickerActive() {
		t.Fatalf("expected model picker to open for staged /model command")
	}
	if len(app.timeline) != 0 {
		t.Fatalf("expected no submission for staged arg-required command, got %d rows", len(app.timeline))
	}
}

func TestTeaStage3ModeHintsExposeEnterTabEscSemantics(t *testing.T) {
	app := readySizedApp(t, 160, 28)

	chat := stripANSIForTest(app.renderInputModeIndicator())
	if !strings.Contains(chat, "enter send") || !strings.Contains(chat, "shift+tab history") {
		t.Fatalf("expected chat mode hint semantics, got %q", chat)
	}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	quickOpen := stripANSIForTest(app.renderInputModeIndicator())
	for _, want := range []string{"enter/right apply", "shift+tab reverse", "esc close"} {
		if !strings.Contains(quickOpen, want) {
			t.Fatalf("expected quick-open hint %q, got %q", want, quickOpen)
		}
	}

	app = readySizedApp(t, 160, 28)
	app = typeTestText(t, app, "/")
	slash := stripANSIForTest(app.renderInputModeIndicator())
	for _, want := range []string{"enter/right apply", "shift+tab reverse", "esc close"} {
		if !strings.Contains(slash, want) {
			t.Fatalf("expected slash hint %q, got %q", want, slash)
		}
	}
}

func TestTeaStage3ShiftTabSequenceReverseQuickOpen(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyDown}, "down")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'', '[', 'Z'}}, "shift-tab-csi-z")
	if app.quickOpen.selected != 0 {
		t.Fatalf("expected quick-open shift+tab sequence alias reverse to first row, got %d", app.quickOpen.selected)
	}
}
