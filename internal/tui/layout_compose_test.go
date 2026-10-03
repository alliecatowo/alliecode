package tui

import "testing"

func TestComposeMeasuredLayoutIncludesPanelGroups(t *testing.T) {
	app := readySizedApp(t, 120, 28)
	layout := app.composeMeasuredLayout(120)
	if len(layout.header.panels) == 0 {
		t.Fatalf("expected header panels in measured layout")
	}
	if len(layout.status.panels) == 0 {
		t.Fatalf("expected status panels in measured layout")
	}
	if len(layout.composer.panels) == 0 {
		t.Fatalf("expected composer panels in measured layout")
	}
}

func TestComposeMeasuredLayoutSeparatesOverlaysFromComposer(t *testing.T) {
	app := readySizedApp(t, 120, 28)
	app.startQuickOpen()
	layout := app.composeMeasuredLayout(120)
	if len(layout.overlays.panels) == 0 {
		t.Fatalf("expected overlay panels in measured layout when quick open is active")
	}
	if len(layout.composer.panels) == 0 {
		t.Fatalf("expected composer panels in measured layout when quick open is active")
	}
	if layout.overlays.panels[0].key != "search" {
		t.Fatalf("expected search overlay first, got %q", layout.overlays.panels[0].key)
	}
	lastComposer := layout.composer.panels[len(layout.composer.panels)-1]
	if lastComposer.key != "input" {
		t.Fatalf("expected composer to end with input row, got %q", lastComposer.key)
	}
}
