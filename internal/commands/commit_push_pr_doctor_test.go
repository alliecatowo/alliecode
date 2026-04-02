package commands

import (
	"context"
	"strings"
	"testing"
)

func TestCommitPushPRDoctor(t *testing.T) {
	cmd := NewCommitPushPRCommand()
	state := &RuntimeState{DiffEntries: []DiffEntry{{Path: "a.go", Added: 1, Staged: true}}}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "commit-push-pr", Args: []string{"doctor"}})
	if err != nil || !strings.Contains(res.Message, "COMMIT_PUSH_PR_DOCTOR") {
		t.Fatalf("doctor failed: %v %q", err, res.Message)
	}
}
