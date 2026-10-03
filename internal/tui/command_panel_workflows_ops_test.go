package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alliecatowo/alliecode/internal/commands"
)

func TestCommandPanelTasksFlowFromSlashAndApply(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	app.cmdState.Tasks = []string{"stabilize output rows"}
	app = typeTestText(t, app, "/tasks")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	if !app.commandPanel.active {
		t.Fatalf("expected tasks panel active")
	}
	plain := stripANSIForTest(app.renderCommandPanel())
	if !strings.Contains(plain, "tasks panel: /tasks") || !strings.Contains(plain, "## Overview") {
		t.Fatalf("expected polished tasks panel rows, got:\n%s", plain)
	}
	app = sendTestKeyAndRun(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	if len(app.timeline) == 0 || !strings.Contains(app.timeline[len(app.timeline)-1].text, "TASKS_LIST") {
		t.Fatalf("expected tasks list command output after panel apply")
	}
}

func TestCommandPanelIssueOpensFromSlashSelection(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	app.cmdState.Issues = []commands.IssueRecord{{ID: "ISSUE-1", Title: "Panel mismatch", Status: "open", Provider: "github"}}
	app = typeTestText(t, app, "/iss")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	if !app.commandPanel.active || app.commandPanel.panel.Command != "issue" {
		t.Fatalf("expected issue panel opened from slash drawer")
	}
	view := stripANSIForTest(app.renderCommandPanel())
	if !strings.Contains(view, "## Issues") {
		t.Fatalf("expected issue section header, got:\n%s", view)
	}
}
