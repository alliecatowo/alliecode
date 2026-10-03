package tui

import "testing"

func TestStage7StatusPanelsIncludePrimaryAndSecondaryRows(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	app.turns = 2
	app.input.SetValue("/")
	app.syncSlashAutocomplete()
	panels := app.statusPanels()
	if len(panels) < 2 {
		t.Fatalf("expected primary and secondary status rows, got %#v", panels)
	}
	if panels[0].key != "status-primary" {
		t.Fatalf("expected first status panel key status-primary, got %q", panels[0].key)
	}
	if panels[1].key != "status-secondary" {
		t.Fatalf("expected second status panel key status-secondary, got %q", panels[1].key)
	}
}
