package commands

import (
	"context"
	"testing"
)

func TestAliasMarketplaceDispatchesPlugin(t *testing.T) {
	r := DefaultRegistry()
	_, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/marketplace")
	if err != nil {
		t.Fatalf("marketplace alias failed: %v", err)
	}
}
