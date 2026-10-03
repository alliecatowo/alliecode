package snapshots_test

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func TestSnapshot_QuickOpenDrawerAndStatusContext(t *testing.T) {
	app := tui.New(tui.Config{Version: "test"})
	updated, _ := app.Update(tea.WindowSizeMsg{Width: 100, Height: 26})
	ready := updated.(*tui.App)
	updated, _ = ready.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	ready = updated.(*tui.App)
	plain := stripANSI(ready.View())
	if want := "search:"; !containsSnapshot(plain, want) {
		t.Fatalf("snapshot missing %q\n%s", want, plain)
	}
	if want := "context: search mode quick-open"; !containsSnapshot(plain, want) {
		t.Fatalf("snapshot missing %q\n%s", want, plain)
	}
	if want := "> "; !containsSnapshot(plain, want) {
		t.Fatalf("snapshot missing input prompt %q\n%s", want, plain)
	}
}

func containsSnapshot(haystack, needle string) bool { return strings.Contains(haystack, needle) }

func TestSnapshot_InputPanelsComposeWithoutOverlap(t *testing.T) {
	app := readyApp(t, 90, 20)
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/permissions")}, "type:/permissions")

	plain := stripANSI(app.View())
	if !containsSnapshot(plain, "hint: usage: /permissions") {
		t.Fatalf("snapshot missing command usage hint\n%s", plain)
	}
	if !containsSnapshot(plain, "input mode: slash") {
		t.Fatalf("snapshot missing input mode row\n%s", plain)
	}
	if !containsSnapshot(plain, "┃ /permissions") {
		t.Fatalf("snapshot missing staged input row\n%s", plain)
	}
}

func TestSnapshot_DrawerAnchorsAboveComposerNearBottom(t *testing.T) {
	app := readyApp(t, 120, 26)
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")

	plain := stripANSI(app.View())
	searchIdx := mustFindLineIndexContaining(t, plain, "search:")
	composerIdx := mustFindLineIndexContaining(t, plain, "┃ Type a message")
	statusIdx := mustFindLineIndexContaining(t, plain, "hints:")
	if !(searchIdx < composerIdx && composerIdx < statusIdx) {
		t.Fatalf("expected bottom-anchored ordering search->composer->status, got:\n%s", plain)
	}
}
