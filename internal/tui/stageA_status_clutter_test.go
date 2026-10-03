package tui

import "testing"

func TestStageAStatusClutterDefaultShowsPrimaryOnly(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	panels := app.statusPanels()
	if len(panels) != 1 {
		t.Fatalf("expected default status to render single primary row, got %#v", panels)
	}
	if panels[0].key != "status-primary" {
		t.Fatalf("expected primary status row key, got %q", panels[0].key)
	}
}

func TestStageAStatusClutterContextCanShowSecondaryRows(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.startQuickOpen()
	panels := app.statusPanels()
	if len(panels) < 2 {
		t.Fatalf("expected contextual status rows when quick-open active, got %#v", panels)
	}
}
