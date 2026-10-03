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

func TestInputEscRestoresDraftWhenBrowsingHistory(t *testing.T) {
	model := NewInput()
	model.history = []string{"first", "second"}
	model.SetValue("draft")
	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	next, _ = next.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if next.Value() != "draft" {
		t.Fatalf("expected esc to restore draft, got %q", next.Value())
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

func TestInferInputHintPromptContexts(t *testing.T) {
	cases := []struct {
		name string
		text string
		want string
	}{
		{name: "empty", text: "", want: "chat (enter sends)"},
		{name: "slash root", text: "/", want: "slash command picker"},
		{name: "slash command", text: "/status", want: "slash command mode (tab/down to pick)"},
		{name: "slash args", text: "/model sonnet", want: "slash arguments (enter sends)"},
		{name: "reference", text: "check @internal/tui/app.go", want: "reference mode (@ opens picker)"},
		{name: "message", text: "hello there", want: "message (ctrl+j newline)"},
	}

	for _, tc := range cases {
		if got := inferInputHint(tc.text); got != tc.want {
			t.Fatalf("%s: inferInputHint(%q)=%q want %q", tc.name, tc.text, got, tc.want)
		}
	}
}

func TestInputShiftEnterVariantsInsertNewline(t *testing.T) {
	model := NewInput()
	model.SetValue("hello")
	for _, msg := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'\x1b', '[', '1', '3', ';', '2', 'u'}},
		{Type: tea.KeyRunes, Runes: []rune{'\x1b', '[', '2', '7', ';', '2', ';', '1', '3', '~'}},
	} {
		next, cmd := model.Update(msg)
		if cmd != nil {
			t.Fatalf("expected shift-enter variant %q to avoid submit command", msg.String())
		}
		if next.Value() != "hello\n" {
			t.Fatalf("expected shift-enter variant %q to insert newline, got %q", msg.String(), next.Value())
		}
	}
}
