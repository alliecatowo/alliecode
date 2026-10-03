package commands

import (
	"context"
	"testing"
)

func TestAliasNewInstallDispatch(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/install status")
	if err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}
	if got := res.Message; len(got) < len("INSTALL_STATUS") || got[:len("INSTALL_STATUS")] != "INSTALL_STATUS" {
		t.Fatalf("unexpected output: %q", got)
	}
}
