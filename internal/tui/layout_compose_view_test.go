package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestComposeMeasuredLayoutKeepsActivityBeforeStatusAndInput(t *testing.T) {
	app := readySizedApp(t, 120, 28)
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	plain := stripANSIForTest(app.View())
	searchIdx := strings.Index(plain, "search:")
	statusIdx := strings.Index(plain, "hints:")
	inputIdx := strings.Index(plain, "┃")
	buddyIdx := buddyIndex(plain)
	if searchIdx < 0 || statusIdx < 0 || inputIdx < 0 || buddyIdx < 0 {
		t.Fatalf("expected search/input/buddy/status sections present")
	}
	if !(searchIdx < inputIdx && inputIdx < buddyIdx && buddyIdx < statusIdx) {
		t.Fatalf("expected measured layout ordering preserved")
	}
}

func TestComposeMeasuredLayoutDrawerAnchorsNearComposer(t *testing.T) {
	app := readySizedApp(t, 120, 28)
	app = typeTestText(t, app, "/")
	plain := stripANSIForTest(app.View())
	drawerIdx := strings.Index(plain, "commands: /")
	inputIdx := strings.Index(plain, "┃ /")
	statusIdx := strings.Index(plain, "hints:")
	if drawerIdx < 0 || inputIdx < 0 || statusIdx < 0 {
		t.Fatalf("expected slash drawer, composer, and status rows, got:\n%s", plain)
	}
	if !(drawerIdx < inputIdx && inputIdx < statusIdx) {
		t.Fatalf("expected drawer directly above composer and below transcript, got:\n%s", plain)
	}
}
