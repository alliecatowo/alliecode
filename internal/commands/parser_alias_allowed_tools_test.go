package commands

import (
	"context"
	"testing"
)

func TestAliasAllowedToolsDispatchesPermissionsStatus(t *testing.T) {
	r := DefaultRegistry()
	_, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/allowed-tools status")
	if err != nil {
		t.Fatalf("allowed-tools alias failed: %v", err)
	}
}
