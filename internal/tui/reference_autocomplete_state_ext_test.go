package tui

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/references"
)

func TestEnsureReferenceSelectionVisibleAdjustsOffset(t *testing.T) {
	app := New(Config{})
	app.refAuto.active = true
	app.refAuto.selected = 8
	for i := 0; i < 12; i++ {
		app.refAuto.suggestions = append(app.refAuto.suggestions, references.Suggestion{Path: "f"})
	}
	app.ensureReferenceSelectionVisible(6)
	if app.refAuto.offset == 0 {
		t.Fatalf("expected non-zero offset for deep selection")
	}
}

func TestReferenceAutocompleteSelectionMemoryByQuery(t *testing.T) {
	app := New(Config{})
	app.refAuto.active = true
	app.refAuto.token = referenceToken{Query: "internal"}
	app.refAuto.suggestions = []references.Suggestion{{Path: "internal/tui/app.go"}, {Path: "internal/references/resolver.go"}}
	app.refAuto.selected = 1
	app.rememberReferenceSelection()

	if got := app.recallReferenceSelection("internal"); got != "internal/references/resolver.go" {
		t.Fatalf("expected remembered selection path, got %q", got)
	}
}
