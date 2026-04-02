package tui

import "testing"

func TestTruncateDisplayWidthEllipsisFallback(t *testing.T) {
	if got := truncateDisplayWidth("abcdef", 3, "....."); got != "abc" {
		t.Fatalf("expected hard fit when ellipsis too wide, got %q", got)
	}
}
