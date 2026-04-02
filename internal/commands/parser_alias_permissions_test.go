package commands

import (
	"context"
	"testing"
)

func TestAliasPermissionDispatchesPermissions(t *testing.T) {
	r := DefaultRegistry()
	_, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/permission status")
	if err != nil {
		t.Fatalf("permission alias failed: %v", err)
	}
}
