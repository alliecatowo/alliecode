package commands

import (
	"context"
	"testing"

	"github.com/alliecatowo/alliecode/internal/permissions"
	"github.com/alliecatowo/alliecode/internal/types"
)

func TestPermissionsSummaryEmitsSummaryIntent(t *testing.T) {
	state := &RuntimeState{PermissionMode: permissions.ModeAuto, PermissionRules: []string{"session:allow bash"}}
	res, err := NewPermissionsCommand().Execute(context.Background(), Context{State: state}, Invocation{Name: "permissions"})
	if err != nil {
		t.Fatalf("permissions summary failed: %v", err)
	}
	if len(res.RenderIntents) == 0 || res.RenderIntents[0].Kind != types.RenderIntentSummaryCard {
		t.Fatalf("expected summary card intent, got %#v", res.RenderIntents)
	}
}

func TestPermissionsRulesEmitsChecklistIntent(t *testing.T) {
	state := &RuntimeState{PermissionRules: []string{"session:allow bash"}}
	res, err := NewPermissionsCommand().Execute(context.Background(), Context{State: state}, Invocation{Name: "permissions", Args: []string{"rules"}})
	if err != nil {
		t.Fatalf("permissions rules failed: %v", err)
	}
	if len(res.RenderIntents) == 0 || res.RenderIntents[0].Kind != types.RenderIntentChecklist {
		t.Fatalf("expected checklist intent, got %#v", res.RenderIntents)
	}
}
