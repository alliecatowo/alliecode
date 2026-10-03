package tui

import (
	"strings"
	"testing"
)

func TestStage7SecondaryStatusIncludesSearchFacts(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	app.startQuickOpen()
	line := stripANSIForTest(app.renderSecondaryStatusLine())
	if !strings.Contains(line, "search:quick-open") {
		t.Fatalf("expected secondary line to include quick-open search mode, got %q", line)
	}
	if !strings.Contains(line, "matches:") {
		t.Fatalf("expected secondary line to include match count, got %q", line)
	}
}
