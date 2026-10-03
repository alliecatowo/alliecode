package vim

import "testing"

func TestHandleKeyWithTransitionCapturesModeChange(t *testing.T) {
	e := NewEngine([]string{"alpha"})
	step := e.HandleKeyWithTransition(Key{Rune: 'i', Special: SpecialNone})

	if step.Before.Mode != ModeNormal {
		t.Fatalf("expected before mode normal, got %v", step.Before.Mode)
	}
	if step.After.Mode != ModeInsert {
		t.Fatalf("expected after mode insert, got %v", step.After.Mode)
	}
}

func TestHandleKeysWithTransitionsTracksCursorAndLines(t *testing.T) {
	e := NewEngine([]string{"cat"})
	steps := e.HandleKeysWithTransitions("iX")
	if len(steps) != 2 {
		t.Fatalf("expected 2 transitions, got %d", len(steps))
	}

	if steps[1].After.Lines[0] != "Xcat" {
		t.Fatalf("expected insert transition to mutate line, got %q", steps[1].After.Lines[0])
	}
	if steps[1].After.Cursor.Col != 1 {
		t.Fatalf("expected cursor advance after insert, got %+v", steps[1].After.Cursor)
	}
}

func TestHandleKeysWithTransitionsIncludesUppercaseCommands(t *testing.T) {
	e := NewEngine([]string{"one"})
	steps := e.HandleKeysWithTransitions("A!")
	if len(steps) != 2 {
		t.Fatalf("expected 2 transitions, got %d", len(steps))
	}
	if steps[0].After.Mode != ModeInsert {
		t.Fatalf("expected A to enter insert mode, got %v", steps[0].After.Mode)
	}
}

func TestHandleKeyWithTransitionEscapesInsertMode(t *testing.T) {
	e := NewEngine([]string{"one"})
	e.HandleKey(Key{Rune: 'i', Special: SpecialNone})
	step := e.HandleKeyWithTransition(Key{Special: SpecialEsc})

	if step.Before.Mode != ModeInsert {
		t.Fatalf("expected before mode insert, got %v", step.Before.Mode)
	}
	if step.After.Mode != ModeNormal {
		t.Fatalf("expected after mode normal, got %v", step.After.Mode)
	}
}
