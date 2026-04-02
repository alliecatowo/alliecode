package commands

import (
	"context"
	"strings"
	"testing"
)

func TestMCPToggleEnableDisable(t *testing.T) {
	cmd := NewMCPCommand()
	state := &RuntimeState{MCPConnections: map[string]bool{"alpha": false, "beta": true}}
	enableRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "mcp", Args: []string{"enable", "alpha"}})
	if err != nil {
		t.Fatalf("mcp enable failed: %v", err)
	}
	if !strings.Contains(enableRes.Message, "MCP_TOGGLE") || !state.MCPConnections["alpha"] {
		t.Fatalf("unexpected enable output: %q state=%v", enableRes.Message, state.MCPConnections)
	}
	disableRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "mcp", Args: []string{"disable", "beta"}})
	if err != nil {
		t.Fatalf("mcp disable failed: %v", err)
	}
	if !strings.Contains(disableRes.Message, "action=disable") || state.MCPConnections["beta"] {
		t.Fatalf("unexpected disable output: %q state=%v", disableRes.Message, state.MCPConnections)
	}
}
