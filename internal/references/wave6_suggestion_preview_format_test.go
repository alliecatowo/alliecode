package references

import "testing"

func TestWave6SuggestionPreviewIncludesKindAndDepth(t *testing.T) {
	filePreview := buildSuggestionPreview("internal/tui/app.go", false)
	if filePreview == "" {
		t.Fatalf("expected file preview text")
	}
	if !containsAllParts(filePreview, []string{".go", "depth:"}) {
		t.Fatalf("expected file preview to include extension and depth, got %q", filePreview)
	}

	dirPreview := buildSuggestionPreview("internal/tui", true)
	if !containsAllParts(dirPreview, []string{"directory", "depth:"}) {
		t.Fatalf("expected folder preview to include folder marker and depth, got %q", dirPreview)
	}
}

func containsAllParts(s string, parts []string) bool {
	for _, part := range parts {
		if part == "" {
			continue
		}
		if !containsPart(s, part) {
			return false
		}
	}
	return true
}

func containsPart(s, part string) bool {
	return len(part) == 0 || (len(s) >= len(part) && (indexPart(s, part) >= 0))
}

func indexPart(s, part string) int {
	for i := 0; i+len(part) <= len(s); i++ {
		if s[i:i+len(part)] == part {
			return i
		}
	}
	return -1
}
