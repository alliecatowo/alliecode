package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestWave5SearchCtrlWDeletesLastWord(t *testing.T) {
	app := New(Config{})
	app.startHistorySearch()
	app.searchQuery = "deploy release now"
	app.syncSearchHelpers()

	app.updateSearchInput(tea.KeyMsg{Type: tea.KeyCtrlW})
	if app.searchQuery != "deploy release" {
		t.Fatalf("expected ctrl+w to delete trailing word, got %q", app.searchQuery)
	}
}
