package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestStageJDoctorSingleEnterRendersIntentWithoutLegacyHeader(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app = typeTestText(t, app, "/doctor")
	app = sendTestKeyAndRun(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")

	plain := stripANSIForTest(app.View())
	if !strings.Contains(plain, "Doctor") {
		t.Fatalf("expected doctor intent content in view, got:\n%s", plain)
	}
	if strings.Contains(plain, "DOCTOR_REPORT") {
		t.Fatalf("expected no legacy doctor header in intent-backed view, got:\n%s", plain)
	}
}
