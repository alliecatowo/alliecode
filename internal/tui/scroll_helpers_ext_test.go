package tui

import "testing"

func TestResolveScrollAnchorClampsOverflowOffset(t *testing.T) {
	offset, bottom := resolveScrollAnchor(50, false, false, 100, 30, 10, 10)
	if bottom {
		t.Fatalf("expected clamped offset instead of bottom anchor")
	}
	if offset != 20 {
		t.Fatalf("expected clamped max offset 20, got %d", offset)
	}
}
