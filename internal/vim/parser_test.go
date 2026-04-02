package vim

import "testing"

func TestParseKeysTokenEnter(t *testing.T) {
	keys := ParseKeys("<enter>")
	if len(keys) != 1 || keys[0].Special != SpecialEnter {
		t.Fatalf("expected enter token parse, got %#v", keys)
	}
}
