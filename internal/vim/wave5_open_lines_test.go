package vim

import "testing"

func TestWave5OpenLineCommandsEnterInsertMode(t *testing.T) {
	e := NewEngine([]string{"first", "second"})
	state := e.HandleKeys("o")
	if state.Mode != ModeInsert {
		t.Fatalf("expected o to switch to insert mode, got %d", state.Mode)
	}
	if len(state.Lines) != 3 {
		t.Fatalf("expected new line inserted below, got %#v", state.Lines)
	}

	e = NewEngine([]string{"first", "second"})
	state = e.HandleKeys("O")
	if state.Mode != ModeInsert {
		t.Fatalf("expected O to switch to insert mode, got %d", state.Mode)
	}
	if len(state.Lines) != 3 {
		t.Fatalf("expected new line inserted above, got %#v", state.Lines)
	}
}
