package tui

import "testing"

func TestWaveMLane3_ResolveScrollAnchorShortToLongFollowsTailByDefault(t *testing.T) {
	offset, bottom := resolveScrollAnchor(0, false, false, 6, 50, 10, 10)
	if !bottom || offset != 0 {
		t.Fatalf("expected short-to-long expansion to follow tail, got offset=%d bottom=%t", offset, bottom)
	}
}
