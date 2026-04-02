package keybindings

import "testing"

func TestWave5ModeStackWithNoPrimaryAddsGlobalWhenRequested(t *testing.T) {
	stack := ModeStack(nil, true)
	if len(stack) != 1 || stack[0] != ModeGlobal {
		t.Fatalf("expected only global mode in empty stack, got %#v", stack)
	}
}
