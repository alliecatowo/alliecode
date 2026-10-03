package commands

import (
	"context"
	"testing"
)

func TestAliasNewCtxVizDispatch(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/ctx-viz status")
	if err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}
	if got := res.Message; len(got) < len("CTX_VIZ_STATUS") || got[:len("CTX_VIZ_STATUS")] != "CTX_VIZ_STATUS" {
		t.Fatalf("unexpected output: %q", got)
	}
}
