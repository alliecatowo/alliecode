package commands

import (
	"context"
	"strings"
	"testing"
)

func TestCommitCheck(t *testing.T) {
	cmd := NewCommitCommand()
	state := &RuntimeState{DiffEntries: []DiffEntry{{Path: "a.go", Added: 2, Staged: true}}}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "commit", Args: []string{"check"}})
	if err != nil || !strings.Contains(res.Message, "COMMIT_CHECK") {
		t.Fatalf("commit check failed: %v %q", err, res.Message)
	}
}
