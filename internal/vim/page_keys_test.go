package vim

import "testing"

func TestParseKeysPageTokens(t *testing.T) {
	keys := ParseKeys("<pgdown>")
	if len(keys) != 1 || keys[0].Special != SpecialPageDown {
		t.Fatalf("expected pgdown token parse, got %#v", keys)
	}
	keys = ParseKeys("<pgup>")
	if len(keys) != 1 || keys[0].Special != SpecialPageUp {
		t.Fatalf("expected pgup token parse, got %#v", keys)
	}
}

func TestHandleKeyWithTransitionSupportsPageKeys(t *testing.T) {
	e := NewEngine([]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"})
	step := e.HandleKeyWithTransition(Key{Special: SpecialPageDown})
	if step.After.Cursor.Row != 10 {
		t.Fatalf("expected pagedown to move cursor by 10 rows, got %+v", step.After.Cursor)
	}
	step = e.HandleKeyWithTransition(Key{Special: SpecialPageUp})
	if step.After.Cursor.Row != 0 {
		t.Fatalf("expected pageup to move cursor back toward top, got %+v", step.After.Cursor)
	}
}
