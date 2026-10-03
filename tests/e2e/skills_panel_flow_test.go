package e2e_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestE2ESkillsPanelOpenThenApply(t *testing.T) {
	app := e2eReadyApp(t, 180, 30)
	app = e2eTypeText(t, app, "/skills")
	app = e2eSendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter})

	openView := e2ePlainView(app)
	if !strings.Contains(openView, "drawer: skills panel: /skills") {
		t.Fatalf("expected /skills one-enter to open skills panel")
	}
	if strings.Contains(openView, "SKILLS_LIST") {
		t.Fatalf("expected no immediate skills output while panel open, got:\n%s", openView)
	}

	app = e2eSendKeyAndRun(t, app, tea.KeyMsg{Type: tea.KeyEnter})
	afterView := e2ePlainView(app)
	if strings.Contains(afterView, "drawer: skills panel: /skills") {
		t.Fatalf("expected panel to close after apply")
	}
	if !strings.Contains(afterView, "Skill inventory with source and state diagnostics.") {
		t.Fatalf("expected skills list intent output after apply, got:\n%s", afterView)
	}
}
