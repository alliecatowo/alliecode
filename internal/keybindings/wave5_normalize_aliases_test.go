package keybindings

import "testing"

func TestWave5NormalizeComboCanonicalizesPagingAliases(t *testing.T) {
	if got := normalizeCombo("PgDn"); got != "pagedown" {
		t.Fatalf("expected pgdn alias to normalize, got %q", got)
	}
	if got := normalizeCombo("ctrl+pgup"); got != "ctrl+pageup" {
		t.Fatalf("expected ctrl+pgup alias to normalize, got %q", got)
	}
}
