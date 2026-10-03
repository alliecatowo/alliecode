package commands

import (
	"context"
	"testing"
)

func TestAliasNewBridgeDispatch(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/bridge status")
	if err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}
	if got := res.Message; len(got) < len("BRIDGE_STATUS") || got[:len("BRIDGE_STATUS")] != "BRIDGE_STATUS" {
		t.Fatalf("unexpected output: %q", got)
	}
}
