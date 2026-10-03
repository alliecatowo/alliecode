package tui

import (
	"encoding/json"
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

func TestMergeRefHistoryDedupAndLimit(t *testing.T) {
	got := mergeRefHistory([]string{"a.go", "a.go", "b.go"}, []string{"c.go", "d.go"}, 3)
	if len(got) != 3 {
		t.Fatalf("expected capped history length, got %d", len(got))
	}
	if got[0] != "a.go" || got[1] != "b.go" || got[2] != "c.go" {
		t.Fatalf("unexpected merged order: %#v", got)
	}
}

func TestCaptureToolReferencesTracksOpenAndContext(t *testing.T) {
	app := New(Config{})
	payload := json.RawMessage(`{"filePath":"internal/tui/app.go","cwd":"."}`)
	app.captureToolReferences("read", payload)
	if len(app.openRefs) == 0 {
		t.Fatalf("expected open refs captured from tool payload")
	}
	if len(app.contextRefs) == 0 {
		t.Fatalf("expected context refs captured from read payload")
	}
}

func TestSplitReferenceDisplayPathSupportsColonAndHash(t *testing.T) {
	path, marker := splitReferenceDisplayPath("internal/tui/app.go:9")
	if path != "internal/tui/app.go" || marker != "#L9" {
		t.Fatalf("unexpected colon split path=%q marker=%q", path, marker)
	}
	hPath, hMarker := splitReferenceDisplayPath("internal/tui/app.go#L10-14")
	if hPath != "internal/tui/app.go" || hMarker != "#L10-14" {
		t.Fatalf("unexpected hash split path=%q marker=%q", hPath, hMarker)
	}
}
