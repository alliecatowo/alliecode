package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestStage5PanelStackContextQuickOpenIncludesSelectionSummary(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	plain := stripANSIForTest(app.renderStatusRuntimePanes())
	if !strings.Contains(plain, "context: search mode quick-open selection=") {
		t.Fatalf("expected quick-open context summary in runtime pane, got %q", plain)
	}
}

func TestStage5PanelStackContextCommandPanelIncludesSelectionSummary(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	if !app.openInteractiveCommandPanel("permissions") {
		t.Fatalf("expected permissions command panel to open")
	}
	plain := stripANSIForTest(app.renderStatusRuntimePanes())
	if !strings.Contains(plain, "context: command panel /permissions") {
		t.Fatalf("expected command panel context summary in runtime pane, got %q", plain)
	}
}


func TestStage5PanelStackContextHistoryIncludesSelectionSummary(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	app.input.history = []string{"deploy release", "open logs"}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlR}, "ctrl+r")

	plain := stripANSIForTest(app.renderStatusRuntimePanes())
	if !strings.Contains(plain, "context: search mode history-search selection=") {
		t.Fatalf("expected history context summary in runtime pane, got %q", plain)
	}
}
