package commands

import (
	"context"
	"strings"
	"testing"
)

func TestWorkflowsCancelAndRerun(t *testing.T) {
	cmd := NewWorkflowsCommand()
	state := &RuntimeState{}
	_, _ = cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "workflows", Args: []string{"run", "ci"}})
	cancelRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "workflows", Args: []string{"cancel", "ci"}})
	if err != nil || !strings.Contains(cancelRes.Message, "status=cancelled") {
		t.Fatalf("cancel failed: %v %q", err, cancelRes.Message)
	}
	rerunRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "workflows", Args: []string{"rerun", "ci"}})
	if err != nil || !strings.Contains(rerunRes.Message, "status=running") {
		t.Fatalf("rerun failed: %v %q", err, rerunRes.Message)
	}
}
