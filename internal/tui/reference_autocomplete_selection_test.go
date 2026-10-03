package tui

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/references"
)

func TestReferenceSelectionMemoryFallsBackToLastSelection(t *testing.T) {
	app := New(Config{})
	app.refAuto.active = true
	app.refAuto.token = referenceToken{Query: "int"}
	app.refAuto.suggestions = referencesSuggestionsForTest{{Path: "internal/tui/app.go"}, {Path: "internal/tui/input.go"}}.toSuggestions()
	app.refAuto.selected = 1
	app.rememberReferenceSelection()
	app.clearReferenceAutocomplete()
	if got := app.recallReferenceSelection("internal/tui"); got != "internal/tui/input.go" {
		t.Fatalf("expected stable selected path restore, got %q", got)
	}
}

type referencesSuggestionForTest struct{ Path string }

type referencesSuggestionsForTest []referencesSuggestionForTest

func (items referencesSuggestionsForTest) toSuggestions() []references.Suggestion {
	out := make([]references.Suggestion, 0, len(items))
	for _, item := range items {
		out = append(out, references.Suggestion{Path: item.Path, Exists: true})
	}
	return out
}

func TestReferenceAutocompleteClearInitializesMemoryMap(t *testing.T) {
	app := New(Config{})
	app.refAuto.memory = nil
	app.clearReferenceAutocomplete()
	if app.refAuto.memory == nil {
		t.Fatalf("expected reference memory map initialized after clear")
	}
}
