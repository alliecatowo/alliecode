package commands

import (
	"context"
	"testing"
)

func TestAliasMarketplaceDispatchesPluginCommand(t *testing.T) {
	r := DefaultRegistry()
	_, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/marketplace list")
	if err != nil {
		t.Fatalf("marketplace alias failed: %v", err)
	}
}
