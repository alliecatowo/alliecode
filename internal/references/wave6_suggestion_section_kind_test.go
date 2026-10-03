package references

import "testing"

func TestWave6SuggestionSectionDistinguishesFilesAndFolders(t *testing.T) {
	recent := map[string]struct{}{"docs": {}, "docs/readme.md": {}}
	open := map[string]struct{}{}
	context := map[string]struct{}{}

	if got := suggestionSectionFor("docs", true, recent, open, context); got != "Recent Folders" {
		t.Fatalf("expected recent folder section, got %q", got)
	}
	if got := suggestionSectionFor("docs/readme.md", false, recent, open, context); got != "Recent Files" {
		t.Fatalf("expected recent file section, got %q", got)
	}
}
