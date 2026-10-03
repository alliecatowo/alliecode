package commands

import (
	"context"
	"testing"
)

func TestAliasNewBreakCacheDispatch(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/break-cache status")
	if err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}
	if got := res.Message; len(got) < len("BREAK_CACHE_STATUS") || got[:len("BREAK_CACHE_STATUS")] != "BREAK_CACHE_STATUS" {
		t.Fatalf("unexpected output: %q", got)
	}
}
