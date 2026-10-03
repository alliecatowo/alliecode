package commands

import (
	"context"
	"testing"
)

func TestAliasNewUltraplanDispatch(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/ultraplan status")
	if err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}
	if got := res.Message; len(got) < len("ULTRAPLAN_STATUS") || got[:len("ULTRAPLAN_STATUS")] != "ULTRAPLAN_STATUS" {
		t.Fatalf("unexpected output: %q", got)
	}
}
