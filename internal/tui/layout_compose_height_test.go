package tui

import "testing"

func TestComposeMeasuredLayoutReservedMatchesGroupHeights(t *testing.T) {
	app := readySizedApp(t, 120, 28)
	layout := app.composeMeasuredLayout(120)
	if want := layout.header.height + layout.status.height + layout.overlays.height + layout.composer.height + layout.buddy.height; layout.reserved != want {
		t.Fatalf("expected reserved=%d, got %d", want, layout.reserved)
	}
}
