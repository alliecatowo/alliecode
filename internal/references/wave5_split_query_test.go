package references

import "testing"

func TestWave5SplitSuggestionQueryParsesLineSuffix(t *testing.T) {
	path, suffix := splitSuggestionQuery("internal/tui/app.go:42")
	if path != "internal/tui/app.go" || suffix != ":42" {
		t.Fatalf("unexpected split result path=%q suffix=%q", path, suffix)
	}
}

func TestWave5SplitSuggestionQueryParsesHashLineSuffix(t *testing.T) {
	path, suffix := splitSuggestionQuery("internal/tui/app.go#L9")
	if path != "internal/tui/app.go" || suffix != "#L9" {
		t.Fatalf("unexpected split result path=%q suffix=%q", path, suffix)
	}
}
