package tui

import "testing"

func TestWaveMLane3_ResolveScrollAnchorNormalizesNegativeOffset(t *testing.T) {
	offset, bottom := resolveScrollAnchor(-10, false, false, 120, 150, 24, 24)
	if bottom {
		t.Fatalf("expected fixed offset mode")
	}
	if offset != 0 {
		t.Fatalf("expected normalized offset 0, got %d", offset)
	}
}
