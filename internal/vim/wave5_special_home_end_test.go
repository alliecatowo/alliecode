package vim

import "testing"

func TestWave5SpecialHomeEndInNormalMode(t *testing.T) {
	e := NewEngine([]string{"alpha beta"})
	e.HandleKey(Key{Special: SpecialEnd})
	if e.State().Cursor.Col != len("alpha beta")-1 {
		t.Fatalf("expected end to move to last rune in normal mode, got %+v", e.State().Cursor)
	}
	e.HandleKey(Key{Special: SpecialHome})
	if e.State().Cursor.Col != 0 {
		t.Fatalf("expected home to move to column zero, got %+v", e.State().Cursor)
	}
}
