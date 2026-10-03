package commands

import (
	"context"
	"testing"
)

func TestAliasNewBughunterDispatch(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/bughunter status")
	if err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}
	if got := res.Message; len(got) < len("BUGHUNTER_STATUS") || got[:len("BUGHUNTER_STATUS")] != "BUGHUNTER_STATUS" {
		t.Fatalf("unexpected output: %q", got)
	}
}
