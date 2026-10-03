package tui

import "testing"

func TestActivityPanelsPreferSearchAndPermissionAsFirstClassSurfaces(t *testing.T) {
	app := readySizedApp(t, 120, 28)
	app.startQuickOpen()
	panels := app.activityPanels()
	if len(panels) != 1 || panels[0].key != "search" {
		t.Fatalf("expected search panel stack entry, got %#v", panels)
	}
	app.state = statePermissionPrompt
	app.syncInputMode()
	panels = app.activityPanels()
	if len(panels) != 1 || panels[0].key != "permission" {
		t.Fatalf("expected permission panel stack entry, got %#v", panels)
	}
}

func TestInputPanelsIncludeHintModeAndInputInSearch(t *testing.T) {
	app := readySizedApp(t, 120, 28)
	app.startQuickOpen()
	panels := app.inputPanels()
	if len(panels) != 2 {
		t.Fatalf("expected input mode + input panels when searching, got %#v", panels)
	}
	if panels[0].key != "input-mode" {
		t.Fatalf("expected first input panel to be mode indicator, got %q", panels[0].key)
	}
	if panels[1].key != "input" {
		t.Fatalf("expected second input panel to be composer, got %q", panels[1].key)
	}
}

func TestOverlayPanelsMergeActivityAndDrawerNearComposer(t *testing.T) {
	app := readySizedApp(t, 120, 28)
	app.startQuickOpen()
	overlays := app.overlayPanels()
	if len(overlays) != 1 {
		t.Fatalf("expected one overlay panel in quick-open mode, got %#v", overlays)
	}
	if overlays[0].key != "search" {
		t.Fatalf("expected search overlay key, got %q", overlays[0].key)
	}
}
