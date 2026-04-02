package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAppClearsBuddyReactionOnScrollKey(t *testing.T) {
	app := New(Config{})
	base := time.Unix(500, 0)
	app.now = func() time.Time { return base }

	_, _ = app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	app.buddy.bubble = "hello"
	app.buddy.bubbleUntil = base.Add(10 * time.Second)
	app.buddy.bubbleFadeAt = base.Add(7 * time.Second)

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyDown})
	if app.buddy.bubble != "" {
		t.Fatalf("expected scroll key to clear buddy reaction, got %q", app.buddy.bubble)
	}
}

func TestIsScrollActivityMsg(t *testing.T) {
	if !isScrollActivityMsg(tea.KeyMsg{Type: tea.KeyPgDown}) {
		t.Fatalf("expected pgdown as scroll activity")
	}
	if isScrollActivityMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}) {
		t.Fatalf("expected regular rune key not to count as scroll")
	}
}
