package tui

import (
	"strings"
	"testing"
)

func TestCommandPanelStatusSlashSubmitRunsOnSingleEnter(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	updated, _ := app.Update(submitMsg{text: "/status"})
	app = updated.(*App)

	if app.commandPanel.active {
		t.Fatalf("expected status command panel to be skipped for single-enter submit")
	}
	if len(app.timeline) != 1 || !strings.Contains(app.timeline[0].text, "STATUS_REPORT") {
		t.Fatalf("expected status report after single-enter submit, timeline=%#v", app.timeline)
	}
}
