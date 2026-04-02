package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAppSlashAutocompleteOpenAndFilter(t *testing.T) {
	app := New(Config{})

	for _, r := range []rune("/mo") {
		_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	if !app.slashAutocomplete.isVisible() {
		t.Fatalf("expected slash autocomplete to open for slash input")
	}
	if app.slashAutocomplete.query != "mo" {
		t.Fatalf("expected query mo, got %q", app.slashAutocomplete.query)
	}
	if len(app.slashAutocomplete.items) == 0 {
		t.Fatalf("expected filtered command suggestions")
	}
	for _, item := range app.slashAutocomplete.items {
		if strings.TrimSpace(item.MatchReason) == "" {
			t.Fatalf("expected suggestions to include match reason for %q", item.Name)
		}
		if item.MatchReason == "browse" {
			t.Fatalf("expected filtered query to produce non-browse match reason, got %q for %q", item.MatchReason, item.Name)
		}
		if strings.TrimSpace(item.Description) == "" {
			t.Fatalf("expected suggestions to include description for %q", item.Name)
		}
	}
}

func TestAppSlashAutocompleteKeyboardSelectAndApply(t *testing.T) {
	app := New(Config{})

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if !app.slashAutocomplete.isVisible() {
		t.Fatalf("expected slash autocomplete to be visible after '/'")
	}

	if len(app.slashAutocomplete.items) < 2 {
		t.Fatalf("expected at least two commands for selection test")
	}

	first := app.slashAutocomplete.items[0].Name
	second := app.slashAutocomplete.items[1].Name

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyDown})
	if app.slashAutocomplete.selected != 1 {
		t.Fatalf("expected down key to move selected index to 1, got %d", app.slashAutocomplete.selected)
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.slashAutocomplete.isVisible() {
		t.Fatalf("expected slash autocomplete to close after applying selection")
	}
	if got := app.input.Value(); got != "/"+second+" " {
		t.Fatalf("expected selected command %q to be applied, got %q (first was %q)", second, got, first)
	}
}

func TestAppSlashAutocompleteEscAndPaging(t *testing.T) {
	app := New(Config{})

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if !app.slashAutocomplete.isVisible() {
		t.Fatalf("expected slash autocomplete to be visible after '/'")
	}
	if len(app.slashAutocomplete.items) < 6 {
		t.Fatalf("expected enough slash suggestions for paging")
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if app.slashAutocomplete.selected != 5 {
		t.Fatalf("expected pgdown to advance 5 rows, got %d", app.slashAutocomplete.selected)
	}
	if app.slashAutocomplete.offset != 0 {
		t.Fatalf("expected first page to keep offset at 0, got %d", app.slashAutocomplete.offset)
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyHome})
	if app.slashAutocomplete.selected != 0 {
		t.Fatalf("expected home to jump to first suggestion, got %d", app.slashAutocomplete.selected)
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnd})
	if app.slashAutocomplete.selected != len(app.slashAutocomplete.items)-1 {
		t.Fatalf("expected end to jump to last suggestion, got %d", app.slashAutocomplete.selected)
	}
	if app.slashAutocomplete.offset <= 0 {
		t.Fatalf("expected end jump to advance viewport offset, got %d", app.slashAutocomplete.offset)
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if app.slashAutocomplete.isVisible() {
		t.Fatalf("expected esc to close slash autocomplete")
	}
}

func TestAppCommandContextHintShownAfterSlashSelection(t *testing.T) {
	app := New(Config{})
	app.width = 120
	app.height = 30
	app.recalcLayout()

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	hint := app.renderCommandContextHint()
	if !strings.Contains(hint, "hint: /model [provider/model|model]") {
		t.Fatalf("expected model command usage hint, got %q", hint)
	}
	if !strings.Contains(hint, "Get or set active model") {
		t.Fatalf("expected model command description in hint, got %q", hint)
	}
}

func TestSlashAutocompleteRendersSectionHeaders(t *testing.T) {
	app := New(Config{})
	app.width = 120
	for _, r := range []rune("/mo") {
		_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	view := app.renderSlashAutocomplete()
	if !strings.Contains(view, "Prefix Matches") && !strings.Contains(view, "Contains") && !strings.Contains(view, "Best Match") && !strings.Contains(view, "Browse") {
		t.Fatalf("expected section header in slash palette, got %q", view)
	}
	if !strings.Contains(view, "preview: /") {
		t.Fatalf("expected slash preview line, got %q", view)
	}
}

func TestAppSlashAutocompleteHiddenWhenCommandHasArgs(t *testing.T) {
	app := New(Config{})
	for _, r := range []rune("/model gpt") {
		_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	if app.slashAutocomplete.isVisible() {
		t.Fatalf("expected slash autocomplete hidden once arguments begin")
	}
}
