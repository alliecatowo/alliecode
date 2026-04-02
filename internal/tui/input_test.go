package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInputCtrlJInsertsNewline(t *testing.T) {
	model := NewInput()
	model.SetValue("hello")
	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlJ})
	if next.Value() != "hello\n" {
		t.Fatalf("expected ctrl+j newline insertion, got %q", next.Value())
	}
}

func TestInputCtrlPNavigatesHistory(t *testing.T) {
	model := NewInput()
	model.history = []string{"first", "second"}
	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	if next.Value() != "second" {
		t.Fatalf("expected ctrl+p to open latest history entry, got %q", next.Value())
	}
	next, _ = next.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	if next.Value() != "" {
		t.Fatalf("expected ctrl+n to restore draft, got %q", next.Value())
	}
}

func TestInputCtrlWDeletesTrailingWord(t *testing.T) {
	model := NewInput()
	model.SetValue("hello world")
	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlW})
	if next.Value() != "hello" {
		t.Fatalf("expected ctrl+w to delete trailing word, got %q", next.Value())
	}
}

func TestInputCtrlUClearsInput(t *testing.T) {
	model := NewInput()
	model.SetValue("hello world")
	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	if next.Value() != "" {
		t.Fatalf("expected ctrl+u to clear input, got %q", next.Value())
	}
}
