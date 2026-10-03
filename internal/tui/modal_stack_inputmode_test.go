package tui

import "testing"

func TestModalStackDrivesInputModeQuickOpen(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.startQuickOpen()
	if app.inputMode != inputModeQuickOpen {
		t.Fatalf("expected quick-open input mode, got %s", app.inputMode)
	}
}
