package commands

import (
	"context"
	"testing"
)

func TestAliasRemoteDispatchesSession(t *testing.T) {
	r := DefaultRegistry()
	_, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/remote status")
	if err != nil {
		t.Fatalf("remote alias failed: %v", err)
	}
}
