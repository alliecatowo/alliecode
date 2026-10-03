package integration_test

import (
	"context"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/types"
)

func TestIntegrationStageJIntentFirstCommandMatrix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantKinds []types.RenderIntentKind
		forbid    []string
	}{
		{name: "status", input: "/status", wantKinds: []types.RenderIntentKind{types.RenderIntentSummaryCard}, forbid: []string{"STATUS_REPORT"}},
		{name: "provider list", input: "/provider list", wantKinds: []types.RenderIntentKind{types.RenderIntentTable}, forbid: []string{"PROVIDER_LIST"}},
		{name: "model list all", input: "/model list all", wantKinds: []types.RenderIntentKind{types.RenderIntentTable}, forbid: []string{"MODEL_LIST_ALL"}},
		{name: "permissions list", input: "/permissions list", wantKinds: []types.RenderIntentKind{types.RenderIntentChecklist}, forbid: []string{"PERMISSIONS_MODES"}},
		{name: "history status", input: "/history status", wantKinds: []types.RenderIntentKind{types.RenderIntentSummaryCard}, forbid: []string{"HISTORY_STATUS"}},
		{name: "session token", input: "/session token show", wantKinds: []types.RenderIntentKind{types.RenderIntentDetailRows}, forbid: []string{"SESSION_TOKEN"}},
		{name: "mcp status", input: "/mcp status github", wantKinds: []types.RenderIntentKind{types.RenderIntentOptionList}, forbid: []string{"MCP_STATUS"}},
		{name: "rewind status", input: "/rewind status", wantKinds: []types.RenderIntentKind{types.RenderIntentSummaryCard}, forbid: []string{"REWIND_STATUS"}},
		{name: "tag status", input: "/tag status", wantKinds: []types.RenderIntentKind{types.RenderIntentSummaryCard}, forbid: []string{"TAG_STATUS"}},
		{name: "remote env list", input: "/remote-env list", wantKinds: []types.RenderIntentKind{types.RenderIntentOptionList}, forbid: []string{"REMOTE_ENV_LIST"}},
		{name: "security review status", input: "/security-review status", wantKinds: []types.RenderIntentKind{types.RenderIntentSummaryCard}, forbid: []string{"SECURITY_REVIEW_STATUS"}},
		{name: "agents list", input: "/agents list", wantKinds: []types.RenderIntentKind{types.RenderIntentTable}, forbid: []string{"AGENTS_LIST"}},
		{name: "vim status", input: "/vim status", wantKinds: []types.RenderIntentKind{types.RenderIntentSummaryCard}, forbid: []string{"VIM_STATUS"}},
		{name: "voice status", input: "/voice status", wantKinds: []types.RenderIntentKind{types.RenderIntentSummaryCard}, forbid: []string{"VOICE_STATUS"}},
		{name: "buddy help", input: "/buddy help", wantKinds: []types.RenderIntentKind{types.RenderIntentOptionList}, forbid: []string{"BUDDY_HELP"}},
		{name: "statusline status", input: "/statusline status", wantKinds: []types.RenderIntentKind{types.RenderIntentSummaryCard}, forbid: []string{"STATUSLINE_STATUS"}},
		{name: "ide detect", input: "/ide detect", wantKinds: []types.RenderIntentKind{types.RenderIntentSummaryCard, types.RenderIntentOptionList}, forbid: []string{"IDE_DETECT"}},
		{name: "keybindings status", input: "/keybindings status", wantKinds: []types.RenderIntentKind{types.RenderIntentSummaryCard}, forbid: []string{"KEYBINDINGS_STATUS"}},
	}

	registry := commands.DefaultRegistry()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			state := scenarioCommandState("contract")
			res, err := registry.Dispatch(context.Background(), commands.Context{State: state}, tc.input)
			if err != nil {
				t.Fatalf("dispatch %q failed: %v", tc.input, err)
			}
			if len(res.RenderIntents) == 0 {
				t.Fatalf("expected intents for %q, got none", tc.input)
			}
			for _, intent := range res.RenderIntents {
				if intent.Kind == types.RenderIntentContract {
					t.Fatalf("expected intent-first payload for %q, got %#v", tc.input, res.RenderIntents)
				}
			}

			for _, wantKind := range tc.wantKinds {
				if !containsIntentKind(res.RenderIntents, wantKind) {
					t.Fatalf("expected kind %q for %q, got %#v", wantKind, tc.input, res.RenderIntents)
				}
			}

			rendered := renderIntentsFromResult(res)
			for _, token := range tc.forbid {
				if strings.Contains(rendered, token) {
					t.Fatalf("intent render unexpectedly leaked legacy token %q for %q:\n%s", token, tc.input, rendered)
				}
			}
		})
	}
}

func containsIntentKind(intents []types.RenderIntent, want types.RenderIntentKind) bool {
	for _, intent := range intents {
		if intent.Kind == want {
			return true
		}
	}
	return false
}

func renderIntentsFromResult(res commands.Result) string {
	parts := make([]string, 0, len(res.RenderIntents))
	for _, intent := range res.RenderIntents {
		lines := []string{intent.Title, intent.Summary, intent.Status}
		for _, f := range intent.Fields {
			lines = append(lines, f.Label, f.Value)
		}
		for _, g := range intent.Groups {
			lines = append(lines, g.Title, g.Summary)
			lines = append(lines, g.Items...)
		}
		for _, item := range intent.Items {
			lines = append(lines, item.Label, item.Detail)
		}
		lines = append(lines, intent.Columns...)
		for _, row := range intent.Rows {
			lines = append(lines, row.Cells...)
		}
		for _, row := range intent.DetailRows {
			lines = append(lines, row.Label, row.Value, row.Status, row.Detail)
		}
		for _, opt := range intent.Options {
			lines = append(lines, opt.Label, opt.Detail, opt.Status, opt.Hint)
		}
		for _, hint := range intent.Hints {
			lines = append(lines, hint.Label, hint.Command)
		}
		for _, action := range intent.Actions {
			lines = append(lines, action.Label, action.Command, action.Detail, action.Status)
		}
		for _, diag := range intent.Diagnostics {
			lines = append(lines, diag.Label, diag.Status, diag.Detail)
		}
		for _, c := range intent.Contract {
			lines = append(lines, c.Key, c.Value)
		}
		parts = append(parts, strings.Join(lines, "\n"))
	}
	return strings.Join(parts, "\n")
}
