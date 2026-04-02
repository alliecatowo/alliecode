package commands

import (
	"context"
	"testing"
)

func TestAliasQuitDispatchesExitAdditional(t *testing.T) {
	r := DefaultRegistry()
	_, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/quit")
	if err != nil {
		t.Fatalf("quit alias failed: %v", err)
	}
}
