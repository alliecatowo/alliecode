package commands

import (
	"context"
	"testing"
)

func TestAliasSettingsDispatchesConfig(t *testing.T) {
	r := DefaultRegistry()
	_, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/settings status")
	if err != nil {
		t.Fatalf("settings alias failed: %v", err)
	}
}
