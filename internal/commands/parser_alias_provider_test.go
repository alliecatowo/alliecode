package commands

import (
	"context"
	"testing"
)

func TestAliasAllowedToolsDispatchesPermissions(t *testing.T) {
	r := DefaultRegistry()
	state := &RuntimeState{}
	res, err := r.Dispatch(context.Background(), Context{State: state}, "/allowed-tools list")
	if err != nil {
		t.Fatalf("alias dispatch failed: %v", err)
	}
	if res.Message == "" {
		t.Fatalf("expected output")
	}
}
