package commands

import (
	"context"
	"testing"
)

func TestAliasNewPerfIssueDispatch(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/perf-issue status")
	if err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}
	if got := res.Message; len(got) < len("PERF_ISSUE_STATUS") || got[:len("PERF_ISSUE_STATUS")] != "PERF_ISSUE_STATUS" {
		t.Fatalf("unexpected output: %q", got)
	}
}
