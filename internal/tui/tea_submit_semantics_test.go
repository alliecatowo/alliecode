package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTeaSubmitSemanticsQuickOpenChangeModelNeedsOnlyOneEnter(t *testing.T) {
	app := readySizedApp(t, 160, 28)

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	app = typeTestText(t, app, "model")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")

	if app.state != stateIdle {
		t.Fatalf("expected quick-open apply to return idle, got %d", app.state)
	}
	if got := app.input.Value(); got != "/model " {
		t.Fatalf("expected quick-open apply to stage /model input, got %q", got)
	}
	if !app.modelPickerActive() {
		t.Fatalf("expected staged /model input to open model picker")
	}
	if len(app.timeline) != 0 {
		t.Fatalf("expected change-model action to stage, not submit, timeline=%d", len(app.timeline))
	}
}

func TestTeaSubmitSemanticsSlashStatusRunsOnSingleEnter(t *testing.T) {
	app := readySizedApp(t, 160, 28)

	updated, _ := app.Update(submitMsg{text: "/status"})
	app = updated.(*App)

	if app.slashAutocomplete.isVisible() {
		t.Fatalf("expected slash drawer closed after submit")
	}
	if app.commandPanel.active {
		t.Fatalf("expected status command panel skipped for single-enter submit")
	}
	if got := app.input.Value(); got != "" {
		t.Fatalf("expected input reset after immediate submit, got %q", got)
	}
	if len(app.timeline) != 1 {
		t.Fatalf("expected one timeline row after immediate submit, got %d", len(app.timeline))
	}
	if !strings.Contains(app.timeline[0].text, "STATUS_REPORT") {
		t.Fatalf("expected status command output in timeline, got %q", app.timeline[0].text)
	}
}

func TestTeaSubmitSemanticsModelPickerRunsSelectionOnSingleEnter(t *testing.T) {
	app := readySizedApp(t, 160, 28)

	app = typeTestText(t, app, "/model ")
	selected, ok := app.selectedModelPickerItem()
	if !ok {
		t.Fatalf("expected initial model picker selection")
	}

	app = sendTestKeyAndRun(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")

	if app.modelPickerActive() {
		t.Fatalf("expected model picker closed after single-enter apply")
	}
	if got := app.input.Value(); got != "" {
		t.Fatalf("expected input reset after model picker apply, got %q", got)
	}
	if len(app.timeline) != 1 {
		t.Fatalf("expected one model confirmation row, got %d", len(app.timeline))
	}
	if last := app.timeline[0].text; !strings.Contains(last, "Model set to "+selected.model.Model) {
		t.Fatalf("expected model confirmation for %q, got %q", selected.model.Model, last)
	}
}

func TestTeaSubmitSemanticsHistoryStagesSelectionOnSingleEnter(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.input.history = []string{"deploy release", "open logs"}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlR}, "ctrl+r")
	app = typeTestText(t, app, "log")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")

	if app.state != stateIdle {
		t.Fatalf("expected history apply to return idle, got %d", app.state)
	}
	if got := app.input.Value(); got != "open logs" {
		t.Fatalf("expected selected history entry staged into input, got %q", got)
	}
	if len(app.timeline) != 0 {
		t.Fatalf("expected history apply to stage only, timeline=%d", len(app.timeline))
	}
}

func TestTeaSubmitSemanticsTimelineRightExitsWithoutSubmission(t *testing.T) {
	app := readySizedApp(t, 160, 28)

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlF}, "ctrl+f")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyRight}, "right")

	if app.state != stateIdle || app.inputMode != inputModeChat {
		t.Fatalf("expected right in timeline search to exit to chat, state=%d mode=%s", app.state, app.inputMode)
	}
	if len(app.timeline) != 0 {
		t.Fatalf("expected no timeline output from timeline search exit, got %d", len(app.timeline))
	}
}

func TestTeaSubmitSemanticsQuickOpenRightStagesModelCommand(t *testing.T) {
	app := readySizedApp(t, 160, 28)

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	app = typeTestText(t, app, "model")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyRight}, "right")

	if app.inputMode != inputModeModelPicker {
		t.Fatalf("expected quick-open right to apply model action and open picker, got %s", app.inputMode)
	}
	if got := app.input.Value(); got != "/model " {
		t.Fatalf("expected /model staged by quick-open right, got %q", got)
	}
	if len(app.timeline) != 0 {
		t.Fatalf("expected staging-only transition, timeline=%d", len(app.timeline))
	}
}
