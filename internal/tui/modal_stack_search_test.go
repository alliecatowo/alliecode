package tui

import "testing"

func TestModalStackSearchSurfaces(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.startTimelineSearch()
	if got := app.activeModalSurface(); got != modalSurfaceSearch {
		t.Fatalf("expected timeline search modal, got %q", got)
	}
	app.startHistorySearch()
	if got := app.activeModalSurface(); got != modalSurfaceSearch {
		t.Fatalf("expected history search modal, got %q", got)
	}
}
