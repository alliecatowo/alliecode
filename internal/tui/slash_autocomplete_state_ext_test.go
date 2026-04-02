package tui

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
)

func TestSlashSuggestionSectionBuckets(t *testing.T) {
	if got := slashSuggestionSection("name-prefix"); got != "Prefix Matches" {
		t.Fatalf("unexpected section bucket: %q", got)
	}
	if got := slashSuggestionSection("fuzzy"); got != "Fuzzy" {
		t.Fatalf("unexpected fuzzy section: %q", got)
	}
}

func TestSlashAutocompleteEnsureVisibleOffset(t *testing.T) {
	state := slashAutocompleteState{selected: 9}
	for i := 0; i < 12; i++ {
		state.items = append(state.items, commands.Suggestion{Name: "x"})
	}
	state.ensureVisible(6)
	if state.offset == 0 {
		t.Fatalf("expected offset advance for deep selection")
	}
}

func TestSlashAutocompleteSelectionMemoryByQuery(t *testing.T) {
	state := slashAutocompleteState{selected: -1}
	state.setItems("mo", []commands.Suggestion{{Name: "model"}, {Name: "mode"}})
	state.moveSelection(1)
	state.setItems("mo", []commands.Suggestion{{Name: "model"}, {Name: "mode"}})
	if state.selected != 1 {
		t.Fatalf("expected memory to restore selected command, got %d", state.selected)
	}
}
