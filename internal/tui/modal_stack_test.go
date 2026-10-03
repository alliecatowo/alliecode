package tui

import "testing"

func TestModalStackEmptyDefaultsToNone(t *testing.T) {
	app := New(Config{})
	if got := app.activeModalSurface(); got != modalSurfaceNone {
		t.Fatalf("expected no active modal surface, got %q", got)
	}
}
