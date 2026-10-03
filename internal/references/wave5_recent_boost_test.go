package references

import (
	"path/filepath"
	"testing"
)

func TestWave5RecentSuggestionBoostsRanking(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "src"))
	mustWriteFile(t, filepath.Join(root, "src", "alpha.go"), "package src\n")
	mustWriteFile(t, filepath.Join(root, "src", "beta.go"), "package src\n")

	r := NewResolver(root)
	suggestions := r.Suggest("src", []string{"src/beta.go"}, nil, nil, 8)
	if len(suggestions) == 0 {
		t.Fatalf("expected ranked suggestions")
	}
	betaPos := -1
	alphaPos := -1
	for i, item := range suggestions {
		if item.Path == "src/beta.go" {
			betaPos = i
		}
		if item.Path == "src/alpha.go" {
			alphaPos = i
		}
	}
	if betaPos == -1 || alphaPos == -1 {
		t.Fatalf("expected both beta and alpha suggestions, got %#v", suggestions)
	}
	if betaPos > alphaPos {
		t.Fatalf("expected recent beta.go to rank before alpha.go, got %#v", suggestions)
	}
}
