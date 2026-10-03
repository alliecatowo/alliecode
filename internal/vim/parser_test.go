package vim

import "testing"

func TestParseKeysTokenEnter(t *testing.T) {
	keys := ParseKeys("<enter>")
	if len(keys) != 1 || keys[0].Special != SpecialEnter {
		t.Fatalf("expected enter token parse, got %#v", keys)
	}
}

func TestParseKeysTokenShiftTabAliases(t *testing.T) {
	for _, token := range []string{"<s-tab>", "<shift-tab>"} {
		keys := ParseKeys(token)
		if len(keys) != 1 || keys[0].Special != SpecialShiftTab {
			t.Fatalf("expected shift-tab parse for %q, got %#v", token, keys)
		}
	}
}
