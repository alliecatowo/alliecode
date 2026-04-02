package commands

import (
	"context"
	"strings"
	"testing"
)

func TestMCPToggleAllNoServers(t *testing.T) {
	cmd := NewMCPCommand()
	state := &RuntimeState{MCPConnections: map[string]bool{}}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "mcp", Args: []string{"enable"}})
	if err != nil {
		t.Fatalf("mcp enable all failed: %v", err)
	}
	if !strings.Contains(res.Message, "target=all") || !strings.Contains(res.Message, "changed=0") {
		t.Fatalf("unexpected toggle-all output: %q", res.Message)
	}
}
