package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSlashPaletteSelectionMemoryAcrossReopen(t *testing.T) {
	app := New(Config{})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyDown})
	selected := app.slashAutocomplete.selected
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEsc})

	app.input.SetValue("/mo")
	app.syncSlashAutocomplete()
	if app.slashAutocomplete.selected != selected {
		t.Fatalf("expected slash selection memory restore, got %d want %d", app.slashAutocomplete.selected, selected)
	}
}

func TestReferencePaletteSelectionMemoryAcrossReopen(t *testing.T) {
	app := New(Config{})
	app.input.SetValue("review @int")
	app.syncReferenceAutocomplete()
	if !app.refAuto.active || len(app.refAuto.suggestions) < 2 {
		t.Skip("workspace does not produce enough reference suggestions")
	}
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyDown})
	selected := app.refAuto.selected
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEsc})

	app.input.SetValue("review @int")
	app.syncReferenceAutocomplete()
	if app.refAuto.selected != selected {
		t.Fatalf("expected reference selection memory restore, got %d want %d", app.refAuto.selected, selected)
	}
}
