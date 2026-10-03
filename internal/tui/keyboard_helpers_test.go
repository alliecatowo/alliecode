package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNormalizeComboAliases(t *testing.T) {
	got := normalizeCombo("Command + Shift + K")
	if got != "ctrl+shift+k" {
		t.Fatalf("expected alias normalization, got %q", got)
	}
}

func TestNormalizeComboSynonyms(t *testing.T) {
	cases := map[string]string{
		"Esc":              "escape",
		"RETURN":           "enter",
		"PgUp":             "pageup",
		"PgDn":             "pagedown",
		"S-Tab":            "shift+tab",
		"Ctrl + Alt + Del": "ctrl+alt+delete",
	}
	for in, want := range cases {
		if got := normalizeCombo(in); got != want {
			t.Fatalf("normalizeCombo(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestKeyMatchesShiftTabVariants(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyShiftTab}
	for _, combo := range []string{"shift+tab", "s-tab", "iso-left-tab", "backtab"} {
		if !keyMatches(msg, combo) {
			t.Fatalf("expected shift-tab message to match combo %q", combo)
		}
	}
}

func TestKeyMatchesCtrlCombo(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyCtrlF}
	if !keyMatches(msg, "ctrl+f") {
		t.Fatalf("expected ctrl+f to match")
	}
}

func TestKeyMatchesAltEnter(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyEnter, Alt: true}
	if !keyMatches(msg, "alt+enter") {
		t.Fatalf("expected alt+enter to match")
	}
}

func TestKeyMatchesSequenceUsesLastStep(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyCtrlF}
	if !keyMatches(msg, "ctrl+k>ctrl+f") {
		t.Fatalf("expected sequence tail to match current key")
	}
}

func TestKeyMatchesAlternativeVariants(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyCtrlF}
	if !keyMatches(msg, "ctrl+g | ctrl+f") {
		t.Fatalf("expected fallback variant to match")
	}
}

func TestKeyMatchesDeleteAsBackspace(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyDelete}
	if !keyMatches(msg, "backspace") {
		t.Fatalf("expected delete to satisfy backspace matcher")
	}
}

func TestNormalizeComboPreservesCtrlJ(t *testing.T) {
	if got := normalizeCombo("Ctrl + J"); got != "ctrl+j" {
		t.Fatalf("expected ctrl+j normalization, got %q", got)
	}
}

func TestNormalizeComboTerminalSequences(t *testing.T) {
	cases := map[string]string{
		"\x1b[Z":       "shift+tab",
		"\x1b[1;2Z":    "shift+tab",
		"\x1b[9;2u":    "shift+tab",
		"\x1b[27;2;9~": "shift+tab",
		"\x1b[13;2u":   "shift+enter",
	}
	for in, want := range cases {
		if got := normalizeCombo(in); got != want {
			t.Fatalf("normalizeCombo(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestKeyMatchesShiftEnterVariants(t *testing.T) {
	for _, msg := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'\x1b', '[', '1', '3', ';', '2', 'u'}},
		{Type: tea.KeyRunes, Runes: []rune{'\x1b', '[', '2', '7', ';', '2', ';', '1', '3', '~'}},
	} {
		if !keyMatches(msg, "shift+enter") {
			t.Fatalf("expected shift+enter variant %q to match", msg.String())
		}
	}
}

func TestKeyMatchesShiftTabSequenceVariants(t *testing.T) {
	for _, msg := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'\x1b', '[', 'Z'}},
		{Type: tea.KeyRunes, Runes: []rune{'\x1b', '[', '1', ';', '2', 'Z'}},
		{Type: tea.KeyRunes, Runes: []rune{'\x1b', '[', '9', ';', '2', 'u'}},
		{Type: tea.KeyRunes, Runes: []rune{'\x1b', '[', '2', '7', ';', '2', ';', '9', '~'}},
	} {
		if !keyMatches(msg, "shift+tab") {
			t.Fatalf("expected shift+tab variant %q to match", msg.String())
		}
	}
}
