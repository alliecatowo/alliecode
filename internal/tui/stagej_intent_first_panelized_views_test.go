package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestStageJTasksPanelOneEnterStaysPanelizedWithoutLegacyDump(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	app.cmdState.Tasks = []string{"stabilize intent contracts"}
	app = typeTestText(t, app, "/tasks")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")

	if !app.commandPanel.active {
		t.Fatalf("expected /tasks to open panelized view")
	}
	plain := stripANSIForTest(app.View())
	if !strings.Contains(plain, "tasks panel: /tasks") {
		t.Fatalf("expected tasks panel chrome, got:\n%s", plain)
	}
	if strings.Contains(plain, "TASKS_LIST") {
		t.Fatalf("expected no legacy tasks dump while panel is open, got:\n%s", plain)
	}
}

func TestStageJIssuePanelOneEnterStaysPanelizedWithoutLegacyDump(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	app = typeTestText(t, app, "/issue")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")

	if !app.commandPanel.active {
		t.Fatalf("expected /issue to open panelized view")
	}
	plain := stripANSIForTest(app.View())
	if !strings.Contains(plain, "issue panel: /issue") {
		t.Fatalf("expected issue panel chrome, got:\n%s", plain)
	}
	if strings.Contains(plain, "ISSUE_STATUS") || strings.Contains(plain, "ISSUE_LIST") {
		t.Fatalf("expected no legacy issue dump while panel is open, got:\n%s", plain)
	}
}
