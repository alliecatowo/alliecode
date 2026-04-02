package commands

import (
	"context"
	"strings"
	"testing"
)

func TestMCPDoctorWithoutManager(t *testing.T) {
	cmd := NewMCPCommand()
	state := &RuntimeState{MCPConnections: map[string]bool{"local": true}}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "mcp", Args: []string{"doctor"}})
	if err != nil {
		t.Fatalf("mcp doctor failed: %v", err)
	}
	if !strings.Contains(res.Message, "MCP_DOCTOR") {
		t.Fatalf("unexpected doctor message: %q", res.Message)
	}
}
