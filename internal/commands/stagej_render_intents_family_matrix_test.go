package commands

import (
	"context"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/permissions"
	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStageJRenderIntentFamilyMatrixForMigratedFlows(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantKinds []types.RenderIntentKind
	}{
		{name: "status", input: "/status", wantKinds: []types.RenderIntentKind{types.RenderIntentSummaryCard}},
		{name: "doctor fix", input: "/doctor fix", wantKinds: []types.RenderIntentKind{types.RenderIntentChecklist}},
		{name: "model", input: "/model", wantKinds: []types.RenderIntentKind{types.RenderIntentSummaryCard}},
		{name: "provider list", input: "/provider list", wantKinds: []types.RenderIntentKind{types.RenderIntentTable}},
		{name: "permissions list", input: "/permissions list", wantKinds: []types.RenderIntentKind{types.RenderIntentChecklist}},
		{name: "session token show", input: "/session token show", wantKinds: []types.RenderIntentKind{types.RenderIntentDetailRows}},
		{name: "mcp status", input: "/mcp status github", wantKinds: []types.RenderIntentKind{types.RenderIntentOptionList}},
		{name: "history status", input: "/history status", wantKinds: []types.RenderIntentKind{types.RenderIntentSummaryCard}},
		{name: "help model", input: "/help model", wantKinds: []types.RenderIntentKind{types.RenderIntentOptionList}},
		{name: "config show", input: "/config show", wantKinds: []types.RenderIntentKind{types.RenderIntentOptionList}},
		{name: "branch list", input: "/branch list", wantKinds: []types.RenderIntentKind{types.RenderIntentTable}},
		{name: "files list", input: "/files list", wantKinds: []types.RenderIntentKind{types.RenderIntentTable}},
		{name: "output style list", input: "/output-style list", wantKinds: []types.RenderIntentKind{types.RenderIntentOptionList}},
		{name: "theme list", input: "/theme list", wantKinds: []types.RenderIntentKind{types.RenderIntentOptionList}},
		{name: "privacy settings list", input: "/privacy-settings list", wantKinds: []types.RenderIntentKind{types.RenderIntentOptionList}},
		{name: "upgrade status", input: "/upgrade status", wantKinds: []types.RenderIntentKind{types.RenderIntentSummaryCard}},
		{name: "plan status", input: "/plan status", wantKinds: []types.RenderIntentKind{types.RenderIntentSummaryCard}},
		{name: "context show", input: "/context show", wantKinds: []types.RenderIntentKind{types.RenderIntentDetailRows}},
		{name: "memory list", input: "/memory list", wantKinds: []types.RenderIntentKind{types.RenderIntentChecklist}},
		{name: "resume status", input: "/resume status", wantKinds: []types.RenderIntentKind{types.RenderIntentSummaryCard}},
		{name: "diff status", input: "/diff status", wantKinds: []types.RenderIntentKind{types.RenderIntentSummaryCard}},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			registry := DefaultRegistry()
			state := stageJRuntimeState()
			res, err := registry.Dispatch(context.Background(), Context{State: state}, tc.input)
			if err != nil {
				t.Fatalf("dispatch %q failed: %v", tc.input, err)
			}
			if !res.Handled {
				t.Fatalf("command %q was not handled", tc.input)
			}
			if len(res.RenderIntents) == 0 {
				t.Fatalf("expected render intents for %q, got none (message=%q)", tc.input, res.Message)
			}

			foundKinds := map[types.RenderIntentKind]bool{}
			for _, intent := range res.RenderIntents {
				if !intent.HasContent() {
					t.Fatalf("expected contentful render intents for %q, got %#v", tc.input, res.RenderIntents)
				}
				if intent.Kind == types.RenderIntentContract {
					t.Fatalf("expected intent-first payload (not contract dump) for %q, got %#v", tc.input, res.RenderIntents)
				}
				foundKinds[intent.Kind] = true
			}

			for _, want := range tc.wantKinds {
				if !foundKinds[want] {
					t.Fatalf("expected intent kind %q for %q, got %#v", want, tc.input, res.RenderIntents)
				}
			}

			if strings.TrimSpace(res.Message) == "" {
				t.Fatalf("expected compatibility message payload for %q", tc.input)
			}
		})
	}
}

func stageJRuntimeState() *RuntimeState {
	return &RuntimeState{
		ProviderName:       "openai",
		Model:              "gpt-4o",
		ModelRef:           "openai/gpt-4o",
		PermissionMode:     permissions.ModeAuto,
		LoggedIn:           false,
		ProviderReady:      false,
		Branches:           []string{"main", "feature/intent-panels"},
		ActiveBranch:       "main",
		DiffMode:           "working",
		DiffEntries:        []DiffEntry{{Path: "internal/tui/app.go", Added: 12, Removed: 3}},
		ConfigValues:       map[string]string{"settings.output-style": "default", "settings.transport": "local"},
		ContextFiles:       []string{"internal/tui/app.go"},
		MemoryEntries:      []string{"prefer intent-backed rendering"},
		MCPConnections:     map[string]bool{"github": true},
		SessionToken:       "abcdef123456",
		SessionTokenSource: "manual",
		SessionTokenPrefix: "abcdef...",
		OutputStyle:        "default",
		OutputFormat:       "default",
		PrivacyTelemetry:   true,
	}
}
