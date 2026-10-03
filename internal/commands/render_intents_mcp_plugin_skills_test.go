package commands

import (
	"context"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestMCPDoctorEmitsDiagnosticsIntent(t *testing.T) {
	state := &RuntimeState{MCPConnections: map[string]bool{"local": true}}
	res, err := NewMCPCommand().Execute(context.Background(), Context{State: state}, Invocation{Name: "mcp", Args: []string{"doctor"}})
	if err != nil {
		t.Fatalf("mcp doctor failed: %v", err)
	}
	if len(res.RenderIntents) == 0 || res.RenderIntents[0].Kind != types.RenderIntentDiagnostics {
		t.Fatalf("expected diagnostics intent, got %#v", res.RenderIntents)
	}
}

func TestSkillsListEmitsTableIntent(t *testing.T) {
	state := &RuntimeState{Skills: []string{"openclaw-status"}, SkillsSources: map[string]string{"openclaw-status": "file"}, SkillsOrigins: map[string]string{"openclaw-status": "/tmp/openclaw-status.md"}, SkillsEnabled: map[string]bool{"openclaw-status": true}}
	res, err := NewSkillsCommand().Execute(context.Background(), Context{State: state}, Invocation{Name: "skills", Args: []string{"list"}})
	if err != nil {
		t.Fatalf("skills list failed: %v", err)
	}
	foundTable := false
	for _, intent := range res.RenderIntents {
		if intent.Kind == types.RenderIntentTable {
			foundTable = true
			break
		}
	}
	if !foundTable {
		t.Fatalf("expected table intent, got %#v", res.RenderIntents)
	}
}
