package tui

import "testing"

func TestStatusPanelsIncludeContextWhenAvailable(t *testing.T) {
	app := New(Config{})
	app.width = 120
	app.cmdState.ProviderName = "anthropic"
	app.cmdState.Model = "claude-sonnet"
	app.input.SetValue("/")
	app.syncSlashAutocomplete()
	panels := app.statusPanels()
	if len(panels) < 2 {
		t.Fatalf("expected status stack with contextual detail, got %#v", panels)
	}
}

func TestInputPanelsIncludeCommandHintWhenSlashCommandStaged(t *testing.T) {
	app := readySizedApp(t, 120, 28)
	app.input.SetValue("/permissions")
	panels := app.inputPanels()
	if len(panels) < 3 {
		t.Fatalf("expected hint + mode + input panels, got %#v", panels)
	}
	if panels[0].key != "input-hint" {
		t.Fatalf("expected first panel to be input hint, got %q", panels[0].key)
	}
}
