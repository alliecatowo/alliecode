package commands

import (
	"context"
	"testing"
)

func TestAliasWorkflowDispatchesWorkflows(t *testing.T) {
	r := DefaultRegistry()
	_, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/workflow status")
	if err != nil {
		t.Fatalf("workflow alias failed: %v", err)
	}
}
