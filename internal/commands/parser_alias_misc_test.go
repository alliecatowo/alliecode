package commands

import (
	"context"
	"testing"
)

func TestAliasQuitDispatchesExit(t *testing.T) {
	r := DefaultRegistry()
	_, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/quit")
	if err != nil {
		t.Fatalf("quit alias failed: %v", err)
	}
}
