package commands

import (
	"context"
	"strings"
	"testing"
)

func TestIssueLabelAndUnlabel(t *testing.T) {
	cmd := NewIssueCommand()
	state := &RuntimeState{Issues: []IssueRecord{{ID: "ISSUE-1", Title: "x", Status: "open"}}}
	labelRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "issue", Args: []string{"label", "ISSUE-1", "bug"}})
	if err != nil || !strings.Contains(labelRes.Message, "ISSUE_LABEL") {
		t.Fatalf("label failed: %v %q", err, labelRes.Message)
	}
	unlabelRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "issue", Args: []string{"unlabel", "ISSUE-1", "bug"}})
	if err != nil || !strings.Contains(unlabelRes.Message, "ISSUE_UNLABEL") {
		t.Fatalf("unlabel failed: %v %q", err, unlabelRes.Message)
	}
}
