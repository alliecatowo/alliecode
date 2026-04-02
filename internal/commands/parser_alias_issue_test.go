package commands

import (
	"context"
	"testing"
)

func TestAliasIssuesDispatchesIssue(t *testing.T) {
	r := DefaultRegistry()
	_, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/issues status")
	if err != nil {
		t.Fatalf("issues alias failed: %v", err)
	}
}
