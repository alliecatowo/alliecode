package commands

import (
	"context"
	"testing"
)

func TestAliasNewGoodClaudeDispatch(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/good-claude status")
	if err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}
	if got := res.Message; len(got) < len("GOOD_CLAUDE_STATUS") || got[:len("GOOD_CLAUDE_STATUS")] != "GOOD_CLAUDE_STATUS" {
		t.Fatalf("unexpected output: %q", got)
	}
}
