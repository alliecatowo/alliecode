package commands

import (
	"context"
	"strings"
	"testing"
)

func TestCommitSuggest(t *testing.T) {
	cmd := NewCommitCommand()
	state := &RuntimeState{DiffEntries: []DiffEntry{{Path: "a.go", Added: 2, Staged: true}}}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "commit", Args: []string{"suggest"}})
	if err != nil || !strings.Contains(res.Message, "COMMIT_SUGGEST") {
		t.Fatalf("commit suggest failed: %v %q", err, res.Message)
	}
}
