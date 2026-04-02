package snapshots_test

import (
	"regexp"
	"testing"

	"github.com/alliecatowo/alliecode/internal/tui"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiPattern.ReplaceAllString(s, "")
}

func TestSnapshot_RenderDiffSingleLineEdit(t *testing.T) {
	got := tui.RenderDiff("alpha\nbeta\n", "alpha\nBETA\n", "demo.txt")
	plain := stripANSI(got)

	const want = "--- a/demo.txt\n" +
		"+++ b/demo.txt\n" +
		"@@ -1,3 +1,3 @@\n" +
		" alpha\n" +
		"-beta\n" +
		"+BETA\n" +
		" \n"

	if plain != want {
		t.Fatalf("snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", plain, want)
	}
}

func TestSnapshot_RenderDiffInsertion(t *testing.T) {
	got := tui.RenderDiff("one\nthree\n", "one\ntwo\nthree\n", "insert.txt")
	plain := stripANSI(got)

	const want = "--- a/insert.txt\n" +
		"+++ b/insert.txt\n" +
		"@@ -1,3 +1,4 @@\n" +
		" one\n" +
		"+two\n" +
		" three\n" +
		" \n"

	if plain != want {
		t.Fatalf("snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", plain, want)
	}
}
