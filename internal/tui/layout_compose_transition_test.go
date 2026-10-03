package tui

import "testing"

func TestComposeMeasuredLayoutChangesAcrossSearchTransition(t *testing.T) {
	app := readySizedApp(t, 120, 28)
	before := app.composeMeasuredLayout(120)
	app.startQuickOpen()
	after := app.composeMeasuredLayout(120)
	if after.overlays.height <= before.overlays.height {
		t.Fatalf("expected search transition to increase overlay height, before=%d after=%d", before.overlays.height, after.overlays.height)
	}
}
