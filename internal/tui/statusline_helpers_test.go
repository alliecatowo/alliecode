package tui

import (
	"strings"
	"testing"
)

func TestRenderStatusBarCompactUsesCleanerPrimarySummary(t *testing.T) {
	app := New(Config{})
	app.width = 180
	app.cmdState.ProviderName = "anthropic"
	app.cmdState.Model = "claude-sonnet"
	app.turns = 3
	view := stripANSIForTest(app.renderStatusBarCompact())
	if !strings.Contains(view, "anthropic/claude-sonnet-4-20250514") {
		t.Fatalf("expected provider/model summary, got %q", view)
	}
	secondary := stripANSIForTest(app.renderSecondaryStatusLine())
	if !strings.Contains(secondary, "turns:3") {
		t.Fatalf("expected turns secondary detail, got %q", secondary)
	}
}
