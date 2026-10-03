package commands

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/permissions"
	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStageJInteractivePanelFamiliesExposeIntentBackedRows(t *testing.T) {
	t.Parallel()

	state := &RuntimeState{
		ProviderName:       "openai",
		Model:              "gpt-4o",
		ModelRef:           "openai/gpt-4o",
		LoggedIn:           false,
		ProviderReady:      false,
		PermissionMode:     permissions.ModeAuto,
		SessionToken:       "abcdef123456",
		SessionTokenSource: "manual",
		SessionTokenPrefix: "abcdef...",
		MCPConnections:     map[string]bool{"github": true},
		Tasks:              []string{"stabilize intent rows"},
		Issues:             []IssueRecord{{ID: "ISSUE-1", Title: "Panel drift", Status: "open", Provider: "github"}},
	}

	for _, name := range []string{"status", "doctor", "model", "permissions", "session", "mcp", "history", "tasks", "issue"} {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			panel := requirePanel(t, name, state)
			assertNonContractIntents(t, name+" header", panel.HeaderIntents)
			for _, item := range panel.Items {
				if item.ApplyInput == "" {
					t.Fatalf("expected apply input for %q item %+v", name, item)
				}
				assertNonContractIntents(t, name+" item "+item.Key, item.PreviewIntents)
			}
		})
	}
}

func assertNonContractIntents(t *testing.T, label string, intents []types.RenderIntent) {
	t.Helper()
	if len(intents) == 0 {
		t.Fatalf("expected intents for %s", label)
	}
	for _, intent := range intents {
		if !intent.HasContent() {
			t.Fatalf("expected contentful intent for %s, got %#v", label, intents)
		}
		if intent.Kind == types.RenderIntentContract {
			t.Fatalf("expected intent-first panel payload for %s, got contract %#v", label, intents)
		}
	}
}
