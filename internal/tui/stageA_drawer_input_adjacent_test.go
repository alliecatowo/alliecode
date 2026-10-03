package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestStageADrawerInputAdjacentSlashBetweenStatusAndComposer(t *testing.T) {
	app := readySizedApp(t, 120, 28)
	app = typeTestText(t, app, "/")
	plain := stripANSIForTest(app.View())
	drawerIdx := strings.Index(plain, "commands: /")
	inputIdx := strings.Index(plain, "┃")
	buddyIdx := buddyIndex(plain)
	statusIdx := strings.Index(plain, "hints:")
	if statusIdx < 0 || drawerIdx < 0 || inputIdx < 0 || buddyIdx < 0 {
		t.Fatalf("expected slash drawer, input, buddy, and status sections in view")
	}
	if !(drawerIdx < inputIdx && inputIdx < buddyIdx && buddyIdx < statusIdx) {
		t.Fatalf("expected slash drawer -> input -> buddy -> status ordering, got:\n%s", plain)
	}
}

func TestStageADrawerInputAdjacentSearchStaysAboveStatus(t *testing.T) {
	app := readySizedApp(t, 120, 28)
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	plain := stripANSIForTest(app.View())
	searchIdx := strings.Index(plain, "search:")
	statusIdx := strings.Index(plain, "hints:")
	if searchIdx < 0 || statusIdx < 0 {
		t.Fatalf("expected search drawer and status sections in view")
	}
	if !(searchIdx < statusIdx) {
		t.Fatalf("expected search drawer above status stack, got:\n%s", plain)
	}
}
