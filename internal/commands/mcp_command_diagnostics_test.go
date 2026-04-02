package commands

import (
	"context"
	"strings"
	"testing"
)

func TestMCPDiagnosticsAndRepair(t *testing.T) {
	cmd := NewMCPCommand()
	state := &RuntimeState{MCPConnections: map[string]bool{"alpha": false}}
	diagRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "mcp", Args: []string{"diagnostics"}})
	if err != nil {
		t.Fatalf("mcp diagnostics failed: %v", err)
	}
	if !strings.Contains(diagRes.Message, "MCP_DIAGNOSTICS") {
		t.Fatalf("unexpected diagnostics message: %q", diagRes.Message)
	}
	repairRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "mcp", Args: []string{"repair", "reconnect"}})
	if err != nil {
		t.Fatalf("mcp repair failed: %v", err)
	}
	if !strings.Contains(repairRes.Message, "MCP_REPAIR") || !state.MCPConnections["alpha"] {
		t.Fatalf("expected reconnect repair to connect alpha: %q state=%v", repairRes.Message, state.MCPConnections)
	}
}
