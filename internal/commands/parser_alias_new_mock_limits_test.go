package commands

import (
	"context"
	"testing"
)

func TestAliasNewMockLimitsDispatch(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/mock-limits status")
	if err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}
	if got := res.Message; len(got) < len("MOCK_LIMITS_STATUS") || got[:len("MOCK_LIMITS_STATUS")] != "MOCK_LIMITS_STATUS" {
		t.Fatalf("unexpected output: %q", got)
	}
}
