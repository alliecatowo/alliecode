package snapshots_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSnapshot_StageJStatusViewHidesLegacyReportHeader(t *testing.T) {
	app := readyApp(t, 160, 28)
	app = submitText(t, app, "/status")
	plain := stripANSI(app.View())

	if !strings.Contains(plain, "Provider:") || !strings.Contains(plain, "Tasks:") {
		t.Fatalf("expected status card details in snapshot view, got:\n%s", plain)
	}
	if strings.Contains(plain, "STATUS_REPORT") {
		t.Fatalf("expected no STATUS_REPORT in panelized status view, got:\n%s", plain)
	}
}

func TestSnapshot_StageJDoctorViewHidesLegacyReportHeader(t *testing.T) {
	app := readyApp(t, 160, 28)
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/doctor")}, "type:/doctor")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	plain := stripANSI(app.View())

	if !strings.Contains(plain, "Doctor") {
		t.Fatalf("expected doctor heading in snapshot view, got:\n%s", plain)
	}
	if strings.Contains(plain, "DOCTOR_REPORT") {
		t.Fatalf("expected no DOCTOR_REPORT in panelized doctor view, got:\n%s", plain)
	}
}

func TestSnapshot_StageJTasksPanelOpenHidesLegacyTaskPayloads(t *testing.T) {
	app := readyApp(t, 180, 30)
	app = submitText(t, app, "/tasks add stagej snapshot")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/tasks")}, "type:/tasks")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	plain := stripANSI(app.View())

	if !strings.Contains(plain, "drawer: tasks panel: /tasks") {
		t.Fatalf("expected tasks panel open in snapshot view, got:\n%s", plain)
	}
	for _, token := range []string{"TASKS_LIST", "TASKS_ADD", "TASKS_DONE"} {
		if strings.Contains(plain, token) {
			t.Fatalf("expected no legacy token %q while panelized tasks view is active:\n%s", token, plain)
		}
	}
}

func TestSnapshot_StageJIssuePanelOpenHidesLegacyIssuePayloads(t *testing.T) {
	app := readyApp(t, 180, 30)
	app = submitText(t, app, "/issue create stagej snapshot")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/issue")}, "type:/issue")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	plain := stripANSI(app.View())

	if !strings.Contains(plain, "drawer: issue panel: /issue") {
		t.Fatalf("expected issue panel open in snapshot view, got:\n%s", plain)
	}
	for _, token := range []string{"ISSUE_STATUS", "ISSUE_LIST", "ISSUE_CREATE"} {
		if strings.Contains(plain, token) {
			t.Fatalf("expected no legacy token %q while panelized issue view is active:\n%s", token, plain)
		}
	}
}
