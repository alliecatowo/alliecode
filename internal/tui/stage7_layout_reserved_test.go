package tui

import "testing"

func TestStage7LayoutReservedHeightFeedsViewportSizing(t *testing.T) {
	app := readySizedApp(t, 120, 28)
	layout := app.composeMeasuredLayout(120)
	want := app.height - layout.reserved
	if want < 1 {
		want = 1
	}
	if app.viewport.Height != want {
		t.Fatalf("expected viewport height %d from measured reserved layout, got %d", want, app.viewport.Height)
	}
}
