package tui

import "testing"

func TestStage7ModalOverlayPrecedenceSearchOverCommandPanel(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	if !app.openInteractiveCommandPanel("permissions") {
		t.Fatalf("expected command panel open")
	}
	if got := app.activeModalSurface(); got != modalSurfaceCommandPanel {
		t.Fatalf("expected command panel active before search, got %q", got)
	}
	app.startQuickOpen()
	if got := app.activeModalSurface(); got != modalSurfaceSearch {
		t.Fatalf("expected search to override command panel, got %q", got)
	}
}
