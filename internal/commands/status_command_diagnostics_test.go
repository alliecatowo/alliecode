package commands

import (
	"context"
	"strings"
	"testing"
)

func TestStatusDiagnosticsReportsCorrectiveLoops(t *testing.T) {
	cmd := NewStatusCommand()
	state := &RuntimeState{}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "status", Args: []string{"diagnostics"}})
	if err != nil {
		t.Fatalf("status diagnostics failed: %v", err)
	}
	if !strings.Contains(res.Message, "STATUS_DIAGNOSTICS") || !strings.Contains(res.Message, "loop_count=7") {
		t.Fatalf("unexpected status diagnostics: %q", res.Message)
	}
}
