package tui

import (
	"strings"
	"testing"
)

func TestWave5SearchSurfaceRendersQuickOpenPreviewMeta(t *testing.T) {
	app := New(Config{})
	app.width = 120
	app.startQuickOpen()
	line := app.renderSearchDetailsPane()
	if !strings.Contains(line, "quick-open preview") {
		t.Fatalf("expected quick-open preview heading, got %q", line)
	}
	if !strings.Contains(line, "action:") {
		t.Fatalf("expected quick-open action detail, got %q", line)
	}
}
