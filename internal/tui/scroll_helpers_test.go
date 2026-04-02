package tui

import "testing"

func TestResolveScrollAnchorFollowTail(t *testing.T) {
	_, bottom := resolveScrollAnchor(4, false, true, 28, 30, 10)
	if !bottom {
		t.Fatalf("expected follow tail to keep bottom anchor")
	}
}

func TestResolveScrollAnchorClamp(t *testing.T) {
	offset, bottom := resolveScrollAnchor(50, false, false, 28, 30, 10)
	if bottom {
		t.Fatalf("expected fixed offset mode")
	}
	if offset != 20 {
		t.Fatalf("expected clamped offset 20, got %d", offset)
	}
}

func TestResolveScrollAnchorHandlesNegativeOffset(t *testing.T) {
	offset, bottom := resolveScrollAnchor(-3, false, false, 20, 30, 10)
	if bottom {
		t.Fatalf("expected fixed offset mode")
	}
	if offset != 0 {
		t.Fatalf("expected clamped offset 0, got %d", offset)
	}
}

func TestResolveScrollAnchorShortContent(t *testing.T) {
	offset, bottom := resolveScrollAnchor(4, false, false, 4, 5, 10)
	if bottom {
		t.Fatalf("expected no bottom anchor for short content")
	}
	if offset != 0 {
		t.Fatalf("expected offset 0 for short content, got %d", offset)
	}
}

func TestResolveScrollAnchorAtBottomWins(t *testing.T) {
	_, bottom := resolveScrollAnchor(4, true, false, 29, 30, 10)
	if !bottom {
		t.Fatalf("expected previous bottom state to keep bottom anchor")
	}
}

func TestResolveScrollAnchorBurstAfterShortContentFollowsTail(t *testing.T) {
	offset, bottom := resolveScrollAnchor(0, false, false, 8, 40, 10)
	if !bottom {
		t.Fatalf("expected anchor to follow bottom after short-to-long burst")
	}
	if offset != 0 {
		t.Fatalf("expected zero offset in bottom-follow mode, got %d", offset)
	}
}

func TestResolveScrollAnchorBurstWhenUserScrolledUpStaysPinned(t *testing.T) {
	offset, bottom := resolveScrollAnchor(5, false, false, 25, 60, 10)
	if bottom {
		t.Fatalf("expected fixed offset mode when user intentionally scrolled up")
	}
	if offset != 5 {
		t.Fatalf("expected preserved offset 5, got %d", offset)
	}
}
