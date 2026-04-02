package commands

import (
	"context"
	"testing"
)

func TestAliasOAuthDispatchesOAuthRefresh(t *testing.T) {
	r := DefaultRegistry()
	_, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/oauth status")
	if err != nil {
		t.Fatalf("oauth alias failed: %v", err)
	}
}
