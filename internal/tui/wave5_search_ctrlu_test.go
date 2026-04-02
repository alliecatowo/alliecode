package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestWave5SearchCtrlUClearsQuery(t *testing.T) {
	app := New(Config{})
	app.startQuickOpen()
	app.searchQuery = "model picker"
	app.syncSearchHelpers()

	app.updateSearchInput(tea.KeyMsg{Type: tea.KeyCtrlU})
	if app.searchQuery != "" {
		t.Fatalf("expected ctrl+u to clear search query, got %q", app.searchQuery)
	}
}
