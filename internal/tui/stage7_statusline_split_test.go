package tui

import (
	"strings"
	"testing"
)

func TestStage7StatuslinePrimaryAndSecondarySplit(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	app.turns = 4
	app.input.SetValue("hello")
	app.syncInputMode()
	primary := stripANSIForTest(app.renderStatusBar())
	secondary := stripANSIForTest(app.renderSecondaryStatusLine())
	if !strings.Contains(primary, "input chat") {
		t.Fatalf("expected primary status to include input mode, got %q", primary)
	}
	if !strings.Contains(secondary, "turns:4") {
		t.Fatalf("expected secondary status to include turns, got %q", secondary)
	}
}
