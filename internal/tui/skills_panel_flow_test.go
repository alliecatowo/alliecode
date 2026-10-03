package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSkillsPanelOpensFromSlashAndShowsActionRows(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.cmdState.Skills = []string{"openclaw-status"}
	app.cmdState.SkillsSources = map[string]string{"openclaw-status": "plugin"}
	app.cmdState.SkillsOrigins = map[string]string{"openclaw-status": "plugin:openclaw"}
	app.cmdState.SkillsEnabled = map[string]bool{"openclaw-status": true}
	app = typeTestText(t, app, "/skills")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	if !app.commandPanel.active {
		t.Fatalf("expected /skills to open command panel")
	}
	if app.commandPanel.panel.Command != "skills" {
		t.Fatalf("expected skills panel, got %q", app.commandPanel.panel.Command)
	}
	if len(app.timeline) != 0 {
		t.Fatalf("expected no immediate execution while skills panel opens")
	}
	plain := stripANSIForTest(app.renderCommandPanel())
	for _, token := range []string{"skills panel: /skills", "Sync skill sources", "Repair via sync", "Repair duplicates"} {
		if !strings.Contains(plain, token) {
			t.Fatalf("expected %q in skills panel, got:\n%s", token, plain)
		}
	}
}

func TestSkillsPanelApplyRunsSkillsListOutput(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.cmdState.Skills = []string{"openclaw-status"}
	app = typeTestText(t, app, "/skills")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	if !app.commandPanel.active {
		t.Fatalf("expected skills panel open")
	}
	app = sendTestKeyAndRun(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	if app.commandPanel.active {
		t.Fatalf("expected skills panel closed after apply")
	}
	if len(app.timeline) == 0 {
		t.Fatalf("expected timeline output after skills apply")
	}
	if !strings.Contains(app.timeline[len(app.timeline)-1].text, "SKILLS_LIST") || !strings.Contains(app.timeline[len(app.timeline)-1].text, "skill.1.source=") {
		t.Fatalf("expected rich skills contract output, got %q", app.timeline[len(app.timeline)-1].text)
	}
}
