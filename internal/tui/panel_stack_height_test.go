package tui

import "testing"

func TestPanelStackHeightUsesRenderedContentInsteadOfFixedConstants(t *testing.T) {
	app := New(Config{})
	panels := []panelSurface{{key: "test", content: "one\ntwo\nthree", minLines: 1, maxLines: 10}}
	if got := app.panelStackHeight(panels, 80); got != 3 {
		t.Fatalf("expected rendered height 3, got %d", got)
	}
}

func TestPanelStackHeightClampsToMaxLines(t *testing.T) {
	app := New(Config{})
	panels := []panelSurface{{key: "test", content: "one\ntwo\nthree\nfour\nfive", minLines: 1, maxLines: 3}}
	if got := app.panelStackHeight(panels, 80); got != 3 {
		t.Fatalf("expected clamped rendered height 3, got %d", got)
	}
}

func TestPanelStackHeightRespectsInputPanelComposition(t *testing.T) {
	app := readySizedApp(t, 120, 28)
	app.input.SetValue("/permissions")
	panels := app.inputPanels()
	if got := app.panelStackHeight(panels, 118); got < 3 {
		t.Fatalf("expected at least three lines for hint/mode/input stack, got %d", got)
	}
}

func TestPanelStackHeightClampsSearchOverlayNearComposer(t *testing.T) {
	app := readySizedApp(t, 120, 28)
	app.startQuickOpen()
	overlays := app.overlayPanels()
	if len(overlays) == 0 || overlays[0].key != "search" {
		t.Fatalf("expected search overlay panel, got %#v", overlays)
	}
	if got := app.panelStackHeight([]panelSurface{overlays[0]}, 118); got > searchPanelMaxLines {
		t.Fatalf("expected search overlay height <= %d, got %d", searchPanelMaxLines, got)
	}
}

func TestPanelStackHeightClampsSlashDrawerNearComposer(t *testing.T) {
	app := readySizedApp(t, 120, 28)
	app.input.SetValue("/")
	app.syncSlashAutocomplete()
	drawers := app.drawerPanels()
	if len(drawers) == 0 || drawers[0].key != "slash" {
		t.Fatalf("expected slash drawer panel, got %#v", drawers)
	}
	if got := app.panelStackHeight([]panelSurface{drawers[0]}, 118); got > slashPanelMaxLines {
		t.Fatalf("expected slash drawer height <= %d, got %d", slashPanelMaxLines, got)
	}
}
