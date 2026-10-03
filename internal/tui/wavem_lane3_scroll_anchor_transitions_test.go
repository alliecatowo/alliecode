package tui

import "testing"

func TestWaveMLane3_ResolveScrollAnchorRetainsBottomGapOnViewportChanges(t *testing.T) {
	offset, bottom := resolveScrollAnchor(70, false, false, 120, 160, 30, 20)
	if bottom {
		t.Fatalf("expected fixed offset mode")
	}
	if offset != 120 {
		t.Fatalf("expected bottom-gap retention offset 120, got %d", offset)
	}
}
