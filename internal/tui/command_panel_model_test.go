package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCommandPanelModelCanSelectConcreteModelAndApply(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	app = typeTestText(t, app, "/model ")
	if !app.modelPickerActive() {
		t.Fatalf("expected model picker after staging /model input")
	}
	for i, item := range app.slashAutocomplete.modelPicker.items {
		if strings.HasPrefix(item.provider, "anthropic") {
			app.slashAutocomplete.modelPicker.selected = i
			break
		}
	}
	app = sendTestKeyAndRun(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	if len(app.timeline) == 0 {
		t.Fatalf("expected timeline row after model panel apply")
	}
	if !strings.Contains(app.timeline[len(app.timeline)-1].text, "Model set to") {
		t.Fatalf("expected model confirmation, got %q", app.timeline[len(app.timeline)-1].text)
	}
}

func TestCommandPanelModelPickerUsesActualProviderReadiness(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	app.cmdState.ProviderName = "anthropic"
	app.cmdState.Model = "claude-opus-4-20250514"
	app.cmdState.ModelRef = "anthropic/claude-opus-4-20250514"
	app.cmdState.LoggedIn = true
	app.cmdState.AuthProvider = "openai"

	items := app.buildModelPickerItems()
	for _, item := range items {
		if item.provider == "anthropic" {
			if item.ready {
				t.Fatalf("expected anthropic picker rows to be not ready when auth provider mismatches")
			}
			return
		}
	}
	t.Fatalf("expected anthropic item in picker")
}

func TestCommandPanelRightArrowAppliesSelection(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	if !app.openInteractiveCommandPanel("doctor") {
		t.Fatalf("expected doctor panel to open")
	}
	app = sendTestKeyAndRun(t, app, tea.KeyMsg{Type: tea.KeyRight}, "right")
	if len(app.timeline) == 0 {
		t.Fatalf("expected right arrow to apply selected panel action")
	}
	if !strings.Contains(app.timeline[len(app.timeline)-1].text, "DOCTOR_REPORT") {
		t.Fatalf("expected doctor output after right arrow apply, got %q", app.timeline[len(app.timeline)-1].text)
	}
}
