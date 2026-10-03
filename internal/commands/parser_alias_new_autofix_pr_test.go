package commands

import (
	"context"
	"testing"
)

func TestAliasNewAutofixPrDispatch(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/autofix-pr status")
	if err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}
	if got := res.Message; len(got) < len("AUTOFIX_PR_STATUS") || got[:len("AUTOFIX_PR_STATUS")] != "AUTOFIX_PR_STATUS" {
		t.Fatalf("unexpected output: %q", got)
	}
}
