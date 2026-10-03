package commands

import (
	"context"
	"testing"
)

func TestAliasNewHeapdumpDispatch(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/heapdump status")
	if err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}
	if got := res.Message; len(got) < len("HEAPDUMP_STATUS") || got[:len("HEAPDUMP_STATUS")] != "HEAPDUMP_STATUS" {
		t.Fatalf("unexpected output: %q", got)
	}
}
