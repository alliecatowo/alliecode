package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/references"
)

func TestCurrentReferenceToken(t *testing.T) {
	token, ok := currentReferenceToken("please check @internal/tui/app.go:12")
	if !ok {
		t.Fatalf("expected token to be detected")
	}
	if token.Query != "internal/tui/app.go:12" {
		t.Fatalf("unexpected token query: %q", token.Query)
	}

	if _, found := currentReferenceToken("email dev@example.com"); found {
		t.Fatalf("did not expect email address to trigger @reference token")
	}
}

func TestBuildReferenceInsertionAddsSpaceAndReplacesToken(t *testing.T) {
	input := "open @int"
	token, ok := currentReferenceToken(input)
	if !ok {
		t.Fatalf("expected token in %q", input)
	}
	next := buildReferenceInsertion(input, token, references.Suggestion{Path: "internal/tui/app.go"})
	if next != "open @internal/tui/app.go " {
		t.Fatalf("unexpected insertion result: %q", next)
	}

	input2 := "open @int now"
	token2 := referenceToken{Start: 5, End: 9, Query: "int"}
	next2 := buildReferenceInsertion(input2, token2, references.Suggestion{Path: "internal/tui/app.go"})
	if next2 != "open @internal/tui/app.go now" {
		t.Fatalf("unexpected insertion with existing space: %q", next2)
	}
}

func TestCaptureRecentReferencesDedupAndOrder(t *testing.T) {
	app := New(Config{})
	app.captureRecentReferences("look @README.md and @internal/tui/app.go:12")
	app.captureRecentReferences("again @README.md and @internal/references/resolver.go")

	if len(app.recentRefs) < 3 {
		t.Fatalf("expected at least 3 recent refs, got %d", len(app.recentRefs))
	}
	if app.recentRefs[0] != "README.md" {
		t.Fatalf("expected README first, got %q", app.recentRefs[0])
	}
	if app.recentRefs[1] != "internal/references/resolver.go" {
		t.Fatalf("expected newest second ref after dedupe, got %q", app.recentRefs[1])
	}
	if app.recentRefs[2] != "internal/tui/app.go" {
		t.Fatalf("expected line suffix trimmed path, got %q", app.recentRefs[2])
	}
}

func TestReferenceAutocompleteRenderIncludesPreviewAndFooter(t *testing.T) {
	app := New(Config{})
	app.width = 120
	app.refAuto.active = true
	app.refAuto.selected = 0
	app.refAuto.suggestions = []references.Suggestion{
		{Path: "internal/tui/app.go", IsDir: false, Source: "workspace+recent", Section: "Workspace + Recent", Preview: "go file app.go", MatchReason: "name-prefix"},
		{Path: "internal/tui", IsDir: true, Source: "workspace", Section: "Workspace", Preview: "directory tui", MatchReason: "path-prefix"},
	}

	view := app.renderReferenceAutocomplete()
	if !strings.Contains(view, "preview: @internal/tui/app.go") {
		t.Fatalf("expected reference preview line, got %q", view)
	}
	if !strings.Contains(view, "[file recent name-prefix]") {
		t.Fatalf("expected normalized source tag in view, got %q", view)
	}
	if !strings.Contains(view, "Workspace + Recent") {
		t.Fatalf("expected section headers in reference palette, got %q", view)
	}
	if !strings.Contains(view, "meta: score:") {
		t.Fatalf("expected reference metadata line, got %q", view)
	}
	if !strings.Contains(view, "about:") {
		t.Fatalf("expected reference preview metadata line, got %q", view)
	}
	if !strings.Contains(view, "tab/enter apply") {
		t.Fatalf("expected keyboard footer hints, got %q", view)
	}
}

func TestReferenceAutocompletePagingAndJumpMaintainsOffset(t *testing.T) {
	app := New(Config{})
	app.refAuto.active = true
	app.refAuto.selected = 0
	for i := 0; i < 12; i++ {
		app.refAuto.suggestions = append(app.refAuto.suggestions, references.Suggestion{Path: "internal/tui/file.go", Section: "Workspace"})
	}

	app.pageReferenceSelection(1)
	if app.refAuto.selected != 5 {
		t.Fatalf("expected page down to move by 5, got %d", app.refAuto.selected)
	}
	app.jumpReferenceSelection(true)
	if app.refAuto.selected != len(app.refAuto.suggestions)-1 {
		t.Fatalf("expected jump end selection, got %d", app.refAuto.selected)
	}
	if app.refAuto.offset == 0 {
		t.Fatalf("expected offset to advance near end")
	}
}
