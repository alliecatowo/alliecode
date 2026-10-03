package commands

import (
	"context"
	"testing"
)

func TestAliasNewDebugToolCallDispatch(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/debug-tool-call status")
	if err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}
	if got := res.Message; len(got) < len("DEBUG_TOOL_CALL_STATUS") || got[:len("DEBUG_TOOL_CALL_STATUS")] != "DEBUG_TOOL_CALL_STATUS" {
		t.Fatalf("unexpected output: %q", got)
	}
}
