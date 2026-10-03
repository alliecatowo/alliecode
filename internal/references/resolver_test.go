package references

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseReferences(t *testing.T) {
	input := "check @README.md and @internal/agent/loop.go:42, email me at dev@example.com"
	refs := ParseReferences(input)
	if len(refs) != 2 {
		t.Fatalf("len(refs) = %d, want 2", len(refs))
	}
	if refs[0].Raw != "@README.md" || refs[0].Path != "README.md" || refs[0].HasLine {
		t.Fatalf("first ref = %+v", refs[0])
	}
	if refs[1].Raw != "@internal/agent/loop.go:42" {
		t.Fatalf("second raw = %q", refs[1].Raw)
	}
	if refs[1].Path != "internal/agent/loop.go" || !refs[1].HasLine || refs[1].Line != 42 {
		t.Fatalf("second ref = %+v", refs[1])
	}
}

func TestParseReferencesInvalidLineSuffix(t *testing.T) {
	input := "look at @docs/spec.md:abc"
	refs := ParseReferences(input)
	if len(refs) != 1 {
		t.Fatalf("len(refs) = %d, want 1", len(refs))
	}
	if refs[0].Path != "docs/spec.md:abc" || refs[0].HasLine {
		t.Fatalf("ref = %+v", refs[0])
	}
}

func TestResolveReferencesFileDirectoryAndMissing(t *testing.T) {
	root := t.TempDir()
	dirPath := filepath.Join(root, "docs")
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	filePath := filepath.Join(root, "docs", "a.txt")
	if err := os.WriteFile(filePath, []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	resolver := NewResolver(root)
	refs := resolver.Resolve("@docs/a.txt:1 @docs @missing.txt")
	if len(refs) != 3 {
		t.Fatalf("len(refs) = %d, want 3", len(refs))
	}

	if !refs[0].Exists || refs[0].ResourceType != "file" || !refs[0].HasLine || refs[0].Line != 1 {
		t.Fatalf("file ref = %+v", refs[0])
	}
	if !refs[1].Exists || refs[1].ResourceType != "directory" {
		t.Fatalf("dir ref = %+v", refs[1])
	}
	if refs[2].Exists || refs[2].ResourceType != "missing" {
		t.Fatalf("missing ref = %+v", refs[2])
	}
}

func TestSuggestIncludesWorkspaceAndRecentWithFuzzyMatch(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "internal", "references"))
	mustWriteFile(t, filepath.Join(root, "internal", "references", "resolver.go"), "package references\n")
	mustWriteFile(t, filepath.Join(root, "README.md"), "# demo\n")

	resolver := NewResolver(root)
	suggestions := resolver.Suggest("rslvr", []string{"README.md"}, nil, nil, 8)
	if len(suggestions) == 0 {
		t.Fatalf("expected fuzzy suggestions, got none")
	}
	if suggestions[0].Path != "internal/references/resolver.go" {
		t.Fatalf("expected resolver.go to rank first for fuzzy query, got %q", suggestions[0].Path)
	}
	if suggestions[0].MatchReason == "" {
		t.Fatalf("expected ranked suggestion to include a match reason")
	}
	if suggestions[0].Preview == "" {
		t.Fatalf("expected ranked suggestion to include preview text")
	}
	if suggestions[0].Section == "" {
		t.Fatalf("expected ranked suggestion to include section")
	}

	recent := resolver.Suggest("read", []string{"README.md"}, nil, nil, 8)
	if len(recent) == 0 || recent[0].Path != "README.md" {
		t.Fatalf("expected recent README.md to rank first, got %#v", recent)
	}
	if recent[0].Section == "" || !strings.Contains(strings.ToLower(recent[0].Section), "recent") {
		t.Fatalf("expected recent suggestion section marker, got %#v", recent[0])
	}
}

func TestSuggestProvidesReasonAndPreviewForDirectoryAndFile(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "internal", "tui"))
	mustWriteFile(t, filepath.Join(root, "internal", "tui", "app.go"), "package tui\n")

	resolver := NewResolver(root)
	suggestions := resolver.Suggest("internal/t", nil, nil, nil, 8)
	if len(suggestions) == 0 {
		t.Fatalf("expected suggestions for internal/t query")
	}
	foundReason := false
	foundPreview := false
	for _, s := range suggestions {
		if s.MatchReason != "" {
			foundReason = true
		}
		if s.Preview != "" {
			foundPreview = true
		}
	}
	if !foundReason || !foundPreview {
		t.Fatalf("expected suggestions to include reasons and previews, got %#v", suggestions)
	}
	if !strings.Contains(suggestions[0].Preview, "depth:") {
		t.Fatalf("expected preview to include path depth hint, got %q", suggestions[0].Preview)
	}
}

func TestSuggestPreservesLineSuffixScaffold(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "internal", "tui"))
	mustWriteFile(t, filepath.Join(root, "internal", "tui", "app.go"), "package tui\n")

	resolver := NewResolver(root)
	suggestions := resolver.Suggest("internal/tui/app.go:42", nil, nil, nil, 5)
	if len(suggestions) == 0 {
		t.Fatalf("expected suggestions for line scaffold query")
	}
	if suggestions[0].Path != "internal/tui/app.go:42" {
		t.Fatalf("expected line suffix to be preserved, got %q", suggestions[0].Path)
	}
}

func TestSuggestionSectionGroupsBySourceSets(t *testing.T) {
	recent := map[string]struct{}{"docs/readme.md": {}}
	open := map[string]struct{}{"docs/open.md": {}}
	context := map[string]struct{}{"docs/context.md": {}}

	if got := suggestionSectionFor("docs/readme.md", false, recent, open, context); got != "Recent Files" {
		t.Fatalf("expected recent file section, got %q", got)
	}
	if got := suggestionSectionFor("docs/open.md", false, recent, open, context); got != "Open Files" {
		t.Fatalf("expected open file section, got %q", got)
	}
	if got := suggestionSectionFor("docs/context.md", false, recent, open, context); got != "Context Files" {
		t.Fatalf("expected context file section, got %q", got)
	}
	if got := suggestionSectionFor("docs", true, recent, open, context); got != "Workspace Folders" {
		t.Fatalf("expected workspace folder section, got %q", got)
	}
}

func TestSuggestSectionPrecedenceRecentOverOpenOverContext(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "docs.md"), "x\n")

	resolver := NewResolver(root)
	suggestions := resolver.Suggest("docs", []string{"docs.md"}, []string{"docs.md"}, []string{"docs.md"}, 4)
	if len(suggestions) == 0 {
		t.Fatalf("expected suggestions")
	}
	if suggestions[0].Section != "Recent Files" {
		t.Fatalf("expected recent section precedence, got %q", suggestions[0].Section)
	}
}

func TestSplitSuggestionQuerySupportsHashLineMarker(t *testing.T) {
	path, suffix := splitSuggestionQuery("internal/tui/app.go#L12-18")
	if path != "internal/tui/app.go" || suffix != "#L12-18" {
		t.Fatalf("unexpected split result path=%q suffix=%q", path, suffix)
	}
}

func TestParseReferencesSupportsHashLineMarker(t *testing.T) {
	refs := ParseReferences("inspect @internal/tui/app.go#L77")
	if len(refs) != 1 {
		t.Fatalf("expected one ref, got %d", len(refs))
	}
	if refs[0].Path != "internal/tui/app.go" || !refs[0].HasLine || refs[0].Line != 77 {
		t.Fatalf("unexpected parsed ref: %+v", refs[0])
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
