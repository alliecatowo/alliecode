package commands

import (
	"context"
	"strings"
	"testing"
)

func TestMCPReconnectSubcommand(t *testing.T) {
	cmd := NewMCPCommand()
	state := &RuntimeState{MCPConnections: map[string]bool{"alpha": false}}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "mcp", Args: []string{"reconnect", "alpha"}})
	if err != nil {
		t.Fatalf("mcp reconnect failed: %v", err)
	}
	if !strings.Contains(res.Message, "MCP_RECONNECT") || !state.MCPConnections["alpha"] {
		t.Fatalf("unexpected reconnect output: %q state=%v", res.Message, state.MCPConnections)
	}
}
