package commands

import (
	"context"
	"testing"
)

func TestAliasNewBackfillSessionsDispatch(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/backfill-sessions status")
	if err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}
	if got := res.Message; len(got) < len("BACKFILL_STATUS") || got[:len("BACKFILL_STATUS")] != "BACKFILL_STATUS" {
		t.Fatalf("unexpected output: %q", got)
	}
}
