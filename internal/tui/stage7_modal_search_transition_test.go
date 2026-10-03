package tui

import "testing"

func TestStage7ModalSearchTransitionBackToChat(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.startQuickOpen()
	if app.inputMode != inputModeQuickOpen {
		t.Fatalf("expected quick-open mode before exit")
	}
	app.exitSearch()
	if app.inputMode != inputModeChat {
		t.Fatalf("expected chat mode after search exit, got %s", app.inputMode)
	}
	if got := app.activeModalSurface(); got != modalSurfaceNone {
		t.Fatalf("expected no active modal after search exit, got %q", got)
	}
}
