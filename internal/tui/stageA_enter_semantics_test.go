package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestStageAEnterSemanticsSlashModelOpensPickerWithOneEnter(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app = typeTestText(t, app, "/model")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")

	if got := app.input.Value(); got != "/model " {
		t.Fatalf("expected one-enter /model stage with trailing space, got %q", got)
	}
	if !app.modelPickerActive() {
		t.Fatalf("expected one-enter /model stage to open model picker")
	}
	if len(app.timeline) != 0 {
		t.Fatalf("expected staged-only /model flow, timeline=%d", len(app.timeline))
	}
}

func TestStageAEnterSemanticsSlashDoctorRunsDefaultPanelAction(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app = typeTestText(t, app, "/doctor")
	app = sendTestKeyAndRun(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")

	if app.commandPanel.active {
		t.Fatalf("expected doctor panel action applied immediately")
	}
	if len(app.timeline) == 0 || !strings.Contains(app.timeline[len(app.timeline)-1].text, "DOCTOR_REPORT") {
		t.Fatalf("expected doctor output after one enter, got %#v", app.timeline)
	}
}

func TestStageAEnterSemanticsSlashTasksOpensPanelOnOneEnter(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app = typeTestText(t, app, "/tasks")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")

	if !app.commandPanel.active {
		t.Fatalf("expected one-enter /tasks to open command panel")
	}
	if app.commandPanel.panel.Command != "tasks" {
		t.Fatalf("expected tasks panel, got %q", app.commandPanel.panel.Command)
	}
	if len(app.timeline) != 0 {
		t.Fatalf("expected no timeline rows for one-enter panel-open flow, got %d", len(app.timeline))
	}
}

func TestStageAEnterSemanticsSlashIssueOpensPanelOnOneEnter(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app = typeTestText(t, app, "/issue")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")

	if !app.commandPanel.active {
		t.Fatalf("expected one-enter /issue to open command panel")
	}
	if app.commandPanel.panel.Command != "issue" {
		t.Fatalf("expected issue panel, got %q", app.commandPanel.panel.Command)
	}
	if len(app.timeline) != 0 {
		t.Fatalf("expected no timeline rows for one-enter panel-open flow, got %d", len(app.timeline))
	}
}

func TestStageAEnterSemanticsSlashProviderOpensPanelOnOneEnter(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app = typeTestText(t, app, "/provider")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")

	if !app.commandPanel.active {
		t.Fatalf("expected one-enter /provider to open command panel")
	}
	if app.commandPanel.panel.Command != "provider" {
		t.Fatalf("expected provider panel, got %q", app.commandPanel.panel.Command)
	}
	if len(app.timeline) != 0 {
		t.Fatalf("expected no timeline rows for one-enter panel-open flow, got %d", len(app.timeline))
	}
}
