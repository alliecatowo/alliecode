package integration_test

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMain pins the ambient environment the contract goldens depend on so the
// suite gives the same answer on a laptop, in a worktree and on a bare CI
// runner:
//   - the terminal profile goldens expect Ghostty;
//   - /plugin status expects an (empty) plugin root to exist in the working
//     directory rather than reporting a missing-root diagnostic.
func TestMain(m *testing.M) {
	_ = os.Setenv("TERM_PROGRAM", "ghostty")
	if wd, err := os.Getwd(); err == nil {
		_ = os.MkdirAll(filepath.Join(wd, ".alliecode", "plugins"), 0o755)
	}
	os.Exit(m.Run())
}
