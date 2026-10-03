package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTeaStatuslineContextTracksQuickOpenAndTimelinePivot(t *testing.T) {
	app := readySizedApp(t, 180, 30)

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	quickOpen := stripANSIForTest(app.renderStatusRuntimePanes())
	for _, want := range []string{"runtime: search", "context: search mode quick-open selection=Search timeline"} {
		if !strings.Contains(quickOpen, want) {
			t.Fatalf("expected quick-open runtime pane to contain %q, got %q", want, quickOpen)
		}
	}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlF}, "ctrl+f")
	timeline := stripANSIForTest(app.renderStatusRuntimePanes())
	if !strings.Contains(timeline, "context: search mode timeline selection=") {
		t.Fatalf("expected timeline runtime pane context after ctrl+f pivot, got %q", timeline)
	}
}

func TestTeaStatuslineContextTracksCommandPanelAndModelPicker(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	if !app.openInteractiveCommandPanel("permissions") {
		t.Fatalf("expected permissions panel to open")
	}
	panel := stripANSIForTest(app.renderStatusRuntimePanes())
	if !strings.Contains(panel, "context: command panel /permissions") {
		t.Fatalf("expected command panel context, got %q", panel)
	}

	app = readySizedApp(t, 180, 30)
	app = typeTestText(t, app, "/model ")
	picker := stripANSIForTest(app.renderStatusRuntimePanes())
	if !strings.Contains(picker, "context: model picker /model") {
		t.Fatalf("expected model picker context, got %q", picker)
	}
}

func TestTeaStatuslineContextTracksReferencePalette(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	app = typeTestText(t, app, "review @READ")

	pane := stripANSIForTest(app.renderStatusRuntimePanes())
	if !strings.Contains(pane, "context: reference @") {
		t.Fatalf("expected reference context, got %q", pane)
	}
	if !strings.Contains(stripANSIForTest(app.renderStatusHints()), "reference @") {
		t.Fatalf("expected reference hint marker, got %q", app.renderStatusHints())
	}
}

func TestTeaStatuslineContextHiddenForIdleChat(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	if got := stripANSIForTest(app.renderStatusRuntimePanes()); strings.TrimSpace(got) != "" {
		t.Fatalf("expected idle chat runtime pane hidden, got %q", got)
	}
}


func TestTeaStatuslineContextTracksHistoryAndSlashPalette(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	app.input.history = []string{"deploy release", "open logs"}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlR}, "ctrl+r")
	history := stripANSIForTest(app.renderStatusRuntimePanes())
	if !strings.Contains(history, "context: search mode history-search selection=") {
		t.Fatalf("expected history search context, got %q", history)
	}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")
	app = typeTestText(t, app, "/")
	slash := stripANSIForTest(app.renderStatusRuntimePanes())
	if !strings.Contains(slash, "context: command /") {
		t.Fatalf("expected slash command context, got %q", slash)
	}
}
