package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestWave5PermissionEscDeniesRequest(t *testing.T) {
	model := NewPermission("bash", "execute shell command")
	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if next.decision != PermissionNo {
		t.Fatalf("expected esc to deny permission, got %d", next.decision)
	}
}
