package keybindings

import "testing"

func TestNormalizeComboPageAliases(t *testing.T) {
	if got := normalizeCombo("PgDn"); got != "pagedown" {
		t.Fatalf("expected pgdn alias to normalize, got %q", got)
	}
	if got := normalizeCombo("PgUp"); got != "pageup" {
		t.Fatalf("expected pgup alias to normalize, got %q", got)
	}
}
