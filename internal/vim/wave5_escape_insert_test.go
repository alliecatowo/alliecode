package vim

import "testing"

func TestWave5EscFromInsertReturnsToNormalAndStepsBack(t *testing.T) {
	e := NewEngine([]string{"abc"})
	e.HandleKeys("A")
	e.HandleKeys("z")
	state := e.HandleKeys("<esc>")
	if state.Mode != ModeNormal {
		t.Fatalf("expected normal mode after esc, got %d", state.Mode)
	}
	if state.Cursor.Col != 3 {
		t.Fatalf("expected cursor to step back one cell on esc, got %+v", state.Cursor)
	}
}
