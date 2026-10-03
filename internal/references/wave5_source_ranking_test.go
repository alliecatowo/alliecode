package references

import (
	"path/filepath"
	"testing"
)

func TestWave5SourceRankingRecentBeforeOpenBeforeContextBeforeWorkspace(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "src"))
	mustWriteFile(t, filepath.Join(root, "src", "alpha.go"), "package src\n")
	mustWriteFile(t, filepath.Join(root, "src", "beta.go"), "package src\n")
	mustWriteFile(t, filepath.Join(root, "src", "gamma.go"), "package src\n")
	mustWriteFile(t, filepath.Join(root, "src", "delta.go"), "package src\n")

	r := NewResolver(root)
	suggestions := r.Suggest("src", []string{"src/alpha.go"}, []string{"src/beta.go"}, []string{"src/gamma.go"}, 8)
	if len(suggestions) < 4 {
		t.Fatalf("expected at least four suggestions, got %#v", suggestions)
	}
	find := func(path string) int {
		for i, s := range suggestions {
			if s.Path == path {
				return i
			}
		}
		return -1
	}
	alpha := find("src/alpha.go")
	beta := find("src/beta.go")
	gamma := find("src/gamma.go")
	delta := find("src/delta.go")
	if alpha < 0 || beta < 0 || gamma < 0 || delta < 0 {
		t.Fatalf("expected all source tiers in results, got %#v", suggestions)
	}
	if !(alpha < beta && beta < gamma && gamma < delta) {
		t.Fatalf("expected recent<open<context<workspace ordering, got %#v", suggestions)
	}
}
