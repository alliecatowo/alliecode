package keybindings

import "testing"

func TestWave5NormalizeComboCanonicalizesPagingAliases(t *testing.T) {
	if got := normalizeCombo("PgDn"); got != "pagedown" {
		t.Fatalf("expected pgdn alias to normalize, got %q", got)
	}
	if got := normalizeCombo("ctrl+pgup"); got != "ctrl+pageup" {
		t.Fatalf("expected ctrl+pgup alias to normalize, got %q", got)
	}
	if got := normalizeCombo("Prior"); got != "pageup" {
		t.Fatalf("expected prior alias to normalize, got %q", got)
	}
	if got := normalizeCombo("Next"); got != "pagedown" {
		t.Fatalf("expected next alias to normalize, got %q", got)
	}
}

func TestWave5NormalizeComboCanonicalizesShiftTabAliases(t *testing.T) {
	for _, input := range []string{"S-Tab", "backtab", "iso-left-tab"} {
		if got := normalizeCombo(input); got != "shift+tab" {
			t.Fatalf("expected %q alias to normalize to shift+tab, got %q", input, got)
		}
	}
}

func TestWave5NormalizeComboCanonicalizesShiftTabTerminalSequences(t *testing.T) {
	for _, input := range []string{"\x1b[Z", "\x1b[1;2Z", "\x1b[9;2u", "\x1b[27;2;9~", "[Z", "1;2Z", "9;2u", "27;2;9~"} {
		if got := normalizeCombo(input); got != "shift+tab" {
			t.Fatalf("expected %q sequence to normalize to shift+tab, got %q", input, got)
		}
	}
}

func TestWave5NormalizeComboCanonicalizesShiftEnterTerminalSequences(t *testing.T) {
	for _, input := range []string{"s-enter", "\x1b[13;2u", "\x1b[27;2;13~", "13;2u", "27;2;13~"} {
		if got := normalizeCombo(input); got != "shift+enter" {
			t.Fatalf("expected %q sequence to normalize to shift+enter, got %q", input, got)
		}
	}
}
