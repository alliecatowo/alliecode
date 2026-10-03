package references

import (
	"path/filepath"
	"testing"
)

func TestWave5SuggestPreservesLineSuffixForRecentEntries(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "internal", "tui"))
	mustWriteFile(t, filepath.Join(root, "internal", "tui", "app.go"), "package tui\n")

	r := NewResolver(root)
	suggestions := r.Suggest("internal/tui/app.go:9", []string{"internal/tui/app.go"}, nil, nil, 4)
	if len(suggestions) == 0 {
		t.Fatalf("expected suggestions")
	}
	if suggestions[0].Path != "internal/tui/app.go:9" {
		t.Fatalf("expected line suffix to remain on ranked suggestion, got %q", suggestions[0].Path)
	}
}
