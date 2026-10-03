package tui

import (
	"strings"
	"testing"
)

func TestStageJStatusViewUsesIntentRenderingWithoutLegacyHeader(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	updated, _ := app.Update(submitMsg{text: "/status"})
	app = updated.(*App)

	plain := stripANSIForTest(app.View())
	if !strings.Contains(plain, "Provider:") || !strings.Contains(plain, "Tasks:") {
		t.Fatalf("expected status intent content in view, got:\n%s", plain)
	}
	if strings.Contains(plain, "STATUS_REPORT") {
		t.Fatalf("expected no legacy status header in intent-backed view, got:\n%s", plain)
	}
}
