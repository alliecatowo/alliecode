package commands

import (
	"context"
	"testing"
)

func TestAliasStateDispatchesStatus(t *testing.T) {
	r := DefaultRegistry()
	_, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/state")
	if err != nil {
		t.Fatalf("status alias failed: %v", err)
	}
}
