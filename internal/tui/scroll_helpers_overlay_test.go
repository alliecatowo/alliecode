package tui

import "testing"

func TestResolveScrollAnchorKeepsTopAnchorWhenOverlayResizesViewport(t *testing.T) {
	offset, bottom := resolveScrollAnchor(12, false, false, 120, 120, 24, 16)
	if bottom {
		t.Fatalf("expected fixed anchor while overlay resizes viewport")
	}
	if offset != 12 {
		t.Fatalf("expected preserved top anchor 12, got %d", offset)
	}
}
