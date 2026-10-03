package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestViewRendersDrawerBeforeStatusAndInput(t *testing.T) {
	app := readySizedApp(t, 120, 28)
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	view := stripANSIForTest(app.View())
	searchIdx := strings.Index(view, "search:")
	statusIdx := strings.Index(view, "hints:")
	inputIdx := strings.Index(view, "┃ Type a message")
	buddyIdx := buddyIndex(view)
	if searchIdx < 0 || statusIdx < 0 || inputIdx < 0 || buddyIdx < 0 {
		t.Fatalf("expected search, input, buddy, and status in view, got %q", view)
	}
	if !(searchIdx < inputIdx && inputIdx < buddyIdx && buddyIdx < statusIdx) {
		t.Fatalf("expected drawer -> input -> buddy -> status ordering, got %q", view)
	}
}

func TestViewSearchPanelClampsBeforeStatusAndInput(t *testing.T) {
	app := readySizedApp(t, 100, 24)
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlR}, "ctrl+r")
	app.input.history = []string{"one\ntwo\nthree\nfour\nfive\nsix\nseven\neight\nnine\nten"}
	app.history = newHistorySearchState(historyEntriesFromInput(app.input.history))
	app.searchQuery = "one"
	app.syncSearchHelpers()
	plain := stripANSIForTest(app.View())
	if strings.Count(plain, "history details:") != 1 {
		t.Fatalf("expected single visible history details block, got:\n%s", plain)
	}
	searchIdx := strings.Index(plain, "search:")
	statusIdx := strings.Index(plain, "hints:")
	inputIdx := strings.Index(plain, "┃")
	buddyIdx := buddyIndex(plain)
	if searchIdx < 0 || statusIdx < 0 || inputIdx < 0 || buddyIdx < 0 {
		t.Fatalf("expected search/input/buddy/status present, got:\n%s", plain)
	}
	if !(searchIdx < inputIdx && inputIdx < buddyIdx && buddyIdx < statusIdx) {
		t.Fatalf("expected clamped search block above input/buddy/status, got:\n%s", plain)
	}
}

func TestViewInputPanelStackOrderOnSlashHint(t *testing.T) {
	app := readySizedApp(t, 100, 24)
	app = typeTestText(t, app, "/permissions")
	plain := stripANSIForTest(app.View())
	hintIdx := strings.Index(plain, "hint: usage: /permissions")
	modeIdx := strings.Index(plain, "input mode:")
	inputIdx := strings.Index(plain, "┃ /permissions")
	if hintIdx < 0 || modeIdx < 0 || inputIdx < 0 {
		t.Fatalf("expected hint/mode/input rows present, got:\n%s", plain)
	}
	if !(hintIdx < modeIdx && modeIdx < inputIdx) {
		t.Fatalf("expected input panel stack ordering hint->mode->input, got:\n%s", plain)
	}
}

func buddyIndex(view string) int {
	if idx := strings.Index(view, "`-vvvv-`"); idx >= 0 {
		return idx
	}
	return strings.Index(view, "<°~°> buddy")
}
