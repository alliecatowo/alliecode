package commands

import (
	"context"
	"testing"
)

func TestAliasAssistDispatchesAssistant(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/assist status")
	if err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}
	if res.Message == "" {
		t.Fatalf("expected output")
	}
}
