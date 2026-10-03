package commands

import (
	"context"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestSessionCommandTokenAndDisconnectEmitDetailRows(t *testing.T) {
	cmd := NewSessionCommand()
	state := &RuntimeState{SessionConnected: true, SessionConnectedAddr: "127.0.0.1:4317", SessionToken: "abcdef1234", SessionTokenSource: "manual", SessionTokenPrefix: "abcdef..."}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "session", Args: []string{"token", "show"}})
	if err != nil {
		t.Fatalf("session token show failed: %v", err)
	}
	if len(res.RenderIntents) == 0 || res.RenderIntents[0].Kind != types.RenderIntentDetailRows {
		t.Fatalf("expected detail rows intent, got %#v", res.RenderIntents)
	}
	res, err = cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "session", Args: []string{"disconnect"}})
	if err != nil {
		t.Fatalf("session disconnect failed: %v", err)
	}
	if len(res.RenderIntents) < 2 || res.RenderIntents[1].Kind != types.RenderIntentActionList {
		t.Fatalf("expected action list follow-up intent, got %#v", res.RenderIntents)
	}
}

func TestMCPCommandStatusAndReconnectEmitStructuredIntents(t *testing.T) {
	cmd := NewMCPCommand()
	state := &RuntimeState{MCPConnections: map[string]bool{"github": true}}
	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "mcp", Args: []string{"status", "github"}})
	if err != nil {
		t.Fatalf("mcp status failed: %v", err)
	}
	if len(statusRes.RenderIntents) == 0 || statusRes.RenderIntents[0].Kind != types.RenderIntentOptionList {
		t.Fatalf("expected option list status intent, got %#v", statusRes.RenderIntents)
	}
	reconnectRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "mcp", Args: []string{"reconnect", "github"}})
	if err != nil {
		t.Fatalf("mcp reconnect failed: %v", err)
	}
	if len(reconnectRes.RenderIntents) < 2 || reconnectRes.RenderIntents[1].Kind != types.RenderIntentActionList {
		t.Fatalf("expected structured reconnect follow-up, got %#v", reconnectRes.RenderIntents)
	}
}
