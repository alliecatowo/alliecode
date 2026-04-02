package tui

import "testing"

func TestBuildToolPreviewHonorsLimit(t *testing.T) {
	preview := buildToolPreview("abc", "defghi", 6)
	if preview != "abc..." {
		t.Fatalf("expected truncated preview, got %q", preview)
	}
}
