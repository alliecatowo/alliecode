package tui

import "testing"

func TestResolveScrollAnchorFollowTail(t *testing.T) {
	_, bottom := resolveScrollAnchor(4, false, true, 28, 30, 10, 10)
	if !bottom {
		t.Fatalf("expected follow tail to keep bottom anchor")
	}
}

func TestResolveScrollAnchorClamp(t *testing.T) {
	offset, bottom := resolveScrollAnchor(50, false, false, 28, 30, 10, 10)
	if bottom {
		t.Fatalf("expected fixed offset mode")
	}
	if offset != 20 {
		t.Fatalf("expected clamped offset 20, got %d", offset)
	}
}

func TestResolveScrollAnchorHandlesNegativeOffset(t *testing.T) {
	offset, bottom := resolveScrollAnchor(-3, false, false, 20, 30, 10, 10)
	if bottom {
		t.Fatalf("expected fixed offset mode")
	}
	if offset != 0 {
		t.Fatalf("expected clamped offset 0, got %d", offset)
	}
}

func TestResolveScrollAnchorShortContent(t *testing.T) {
	offset, bottom := resolveScrollAnchor(4, false, false, 4, 5, 10, 10)
	if bottom {
		t.Fatalf("expected no bottom anchor for short content")
	}
	if offset != 0 {
		t.Fatalf("expected offset 0 for short content, got %d", offset)
	}
}

func TestResolveScrollAnchorAtBottomWins(t *testing.T) {
	_, bottom := resolveScrollAnchor(4, true, false, 29, 30, 10, 10)
	if !bottom {
		t.Fatalf("expected previous bottom state to keep bottom anchor")
	}
}

func TestResolveScrollAnchorBurstAfterShortContentFollowsTail(t *testing.T) {
	offset, bottom := resolveScrollAnchor(0, false, false, 8, 40, 10, 10)
	if !bottom {
		t.Fatalf("expected anchor to follow bottom after short-to-long burst")
	}
	if offset != 0 {
		t.Fatalf("expected zero offset in bottom-follow mode, got %d", offset)
	}
}

func TestResolveScrollAnchorBurstWhenUserScrolledUpStaysPinned(t *testing.T) {
	offset, bottom := resolveScrollAnchor(5, false, false, 25, 60, 10, 10)
	if bottom {
		t.Fatalf("expected fixed offset mode when user intentionally scrolled up")
	}
	if offset != 5 {
		t.Fatalf("expected preserved offset 5, got %d", offset)
	}
}

func TestResolveScrollAnchorKeepsOffsetWhenFollowTailDisabledAndNotAtBottom(t *testing.T) {
	offset, bottom := resolveScrollAnchor(7, false, false, 80, 92, 20, 20)
	if bottom {
		t.Fatalf("expected fixed offset mode while not following tail")
	}
	if offset != 7 {
		t.Fatalf("expected stable offset 7, got %d", offset)
	}
}

func TestResolveScrollAnchorDetectsBottomUsingPreviousViewportHeight(t *testing.T) {
	_, bottom := resolveScrollAnchor(90, false, false, 100, 120, 10, 20)
	if !bottom {
		t.Fatalf("expected bottom anchor when previous offset matched previous viewport max")
	}
}

func TestResolveScrollAnchorClampsWhenViewportGrows(t *testing.T) {
	offset, bottom := resolveScrollAnchor(85, false, false, 100, 100, 10, 20)
	if bottom {
		t.Fatalf("expected fixed offset mode when not at bottom")
	}
	if offset != 85 {
		t.Fatalf("expected preserved offset 85 when only viewport changed, got %d", offset)
	}
}

func TestResolveScrollAnchorKeepsOffsetAcrossOverlayToggle(t *testing.T) {
	offset, bottom := resolveScrollAnchor(25, false, false, 180, 180, 20, 18)
	if bottom {
		t.Fatalf("expected fixed offset mode when overlay toggles")
	}
	if offset != 25 {
		t.Fatalf("expected preserved offset 25 across overlay toggle, got %d", offset)
	}
}
