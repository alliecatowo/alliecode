package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTeaProviderPanelQuickOpenChangeModelShowsPanelAndRuntimeContext(t *testing.T) {
	app := readySizedApp(t, 180, 30)

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	app = typeTestText(t, app, "model")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")

	plain := stripANSIForTest(app.View())
	for _, want := range []string{"model picker: /model", "runtime: idle", "context: model picker /model", "provider:"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("expected model panel view to contain %q, got:\n%s", want, plain)
		}
	}
}

func TestTeaProviderPanelModelPickerShowsProviderStateAndFooter(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	app = typeTestText(t, app, "/model ")

	view := stripANSIForTest(app.renderModelPicker())
	for _, want := range []string{"model picker: /model", "selected: /model ", "provider:", "enter", "caps:"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected model picker to contain %q, got %q", want, view)
		}
	}
	if !strings.Contains(view, "[ready]") && !strings.Contains(view, "[auth]") {
		t.Fatalf("expected model picker readiness markers, got %q", view)
	}
}
