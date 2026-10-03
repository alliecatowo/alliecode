package snapshots_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSnapshot_LoginStatusCommandRendersCleanRows(t *testing.T) {
	app := readyApp(t, 160, 28)
	app = submitText(t, app, "/login status")

	plain := stripANSI(app.View())
	got := strings.Join([]string{
		strings.TrimSpace(mustFindLineContaining(t, plain, "Login status")),
		strings.TrimSpace(mustFindLineContaining(t, plain, "Provider: -")),
		strings.TrimSpace(mustFindLineContaining(t, plain, "Logged in: no")),
	}, "\n")
	const want = "Login status\nProvider: -\nLogged in: no"
	if got != want {
		t.Fatalf("login status snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestSnapshot_StatusCommandRendersCleanRows(t *testing.T) {
	app := readyApp(t, 160, 28)
	app = submitText(t, app, "/status")

	plain := stripANSI(app.View())
	if !strings.Contains(plain, "- Diagnostics: /status diagnostics") {
		t.Fatalf("expected compact status card output, got:\n%s", plain)
	}
}

func TestSnapshot_StatusCommandPanelizedCardLines(t *testing.T) {
	app := readyApp(t, 160, 28)
	app = submitText(t, app, "/status")

	plain := stripANSI(app.View())
	got := strings.Join([]string{
		strings.TrimSpace(mustFindLineContaining(t, plain, "Provider:")),
		strings.TrimSpace(mustFindLineContaining(t, plain, "Logged in:")),
		strings.TrimSpace(mustFindLineContaining(t, plain, "Tasks:")),
	}, "\n")
	const want = "Provider: -\nLogged in: no\nTasks: 0 total / 0 running / 0 done"
	if got != want {
		t.Fatalf("status command snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestSnapshot_ModelListAllCommandPanelizedCardLines(t *testing.T) {
	app := readyApp(t, 180, 30)
	app = submitText(t, app, "/model list all")

	plain := stripANSI(app.View())
	got := strings.Join([]string{
		strings.TrimSpace(mustFindLineContaining(t, plain, "Provider  | Model")),
		strings.TrimSpace(mustFindLineContaining(t, plain, "openai")),
	}, "\n")
	const want = "Provider  | Model                     | Context\nopenai    | gpt-4o                    | 128000"
	if got != want {
		t.Fatalf("model list all snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestSnapshot_QuickOpenModelPreviewPanel(t *testing.T) {
	app := readyApp(t, 180, 28)
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("model")}, "type:model")

	plain := stripANSI(app.View())
	got := strings.Join([]string{
		trimRight(mustFindLineContaining(t, plain, "quick-open preview:")),
		trimRight(mustFindLineContaining(t, plain, "action: command.model")),
		trimRight(mustFindLineContaining(t, plain, "selected: Change model")),
	}, "\n")
	const want = "quick-open preview:\n  action: command.model\n  selected: Change model"
	if got != want {
		t.Fatalf("quick-open preview snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestSnapshot_QuickOpenModelEnterShowsPanelizedPicker(t *testing.T) {
	app := readyApp(t, 180, 30)
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("model")}, "type:model")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")

	plain := stripANSI(app.View())
	if !strings.Contains(plain, "model picker: /model") {
		t.Fatalf("expected model picker header, got:\n%s", plain)
	}
	if !strings.Contains(plain, "context: model picker /model anthropic/") {
		t.Fatalf("expected anthropic model picker context, got:\n%s", plain)
	}
	if !strings.Contains(plain, "provider: anthropic (requires login)") {
		t.Fatalf("expected anthropic provider state, got:\n%s", plain)
	}
}

func TestSnapshot_SlashTasksOpensPanelizedDrawerRows(t *testing.T) {
	app := readyApp(t, 180, 30)
	app = submitText(t, app, "/tasks add stabilize panel rows")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/tasks")}, "type:/tasks")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")

	plain := stripANSI(app.View())
	for _, needle := range []string{"drawer: tasks panel: /tasks", "## Overview", "## Open tasks", "Complete task 1", "[open]      stabilize panel rows"} {
		if !strings.Contains(plain, needle) {
			t.Fatalf("expected tasks drawer snapshot line %q, got:\n%s", needle, plain)
		}
	}
}

func TestSnapshot_SlashIssueOpensPanelizedDrawerRows(t *testing.T) {
	app := readyApp(t, 180, 30)
	app = submitText(t, app, "/issue create panel parity drift")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/issue")}, "type:/issue")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")

	plain := stripANSI(app.View())
	for _, needle := range []string{"drawer: issue panel: /issue", "## Overview", "## Issues", "Close ISSUE-1", "[open]      panel parity drift"} {
		if !strings.Contains(plain, needle) {
			t.Fatalf("expected issue drawer snapshot line %q, got:\n%s", needle, plain)
		}
	}
}
