package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTeamListAndStatusIncludeSummary(t *testing.T) {
	resetOrchestrationStateForTests()
	ctx := types.ToolContext{}

	_, _ = (&TeamCreateTool{}).Execute(context.Background(), []byte(`{"team_name":"omega","members":["dev1"]}`), ctx)

	listRes, err := (&TeamListTool{}).Execute(context.Background(), []byte(`{}`), ctx)
	if err != nil || listRes.IsError {
		t.Fatalf("team_list failed: err=%v content=%q", err, listRes.Content)
	}
	var listed map[string]any
	if err := json.Unmarshal([]byte(listRes.Content), &listed); err != nil {
		t.Fatalf("invalid list json: %v", err)
	}
	if listed["summary"] == nil {
		t.Fatalf("expected summary in team_list output")
	}

	statusRes, err := (&TeamStatusTool{}).Execute(context.Background(), []byte(`{"team_name":"omega"}`), ctx)
	if err != nil || statusRes.IsError {
		t.Fatalf("team_status failed: err=%v content=%q", err, statusRes.Content)
	}
	var status map[string]any
	if err := json.Unmarshal([]byte(statusRes.Content), &status); err != nil {
		t.Fatalf("invalid status json: %v", err)
	}
	if status["summary"] == nil {
		t.Fatalf("expected summary in team_status output")
	}
}

func TestTeamToolsIncludeContractMetadata(t *testing.T) {
	resetOrchestrationStateForTests()
	ctx := types.ToolContext{}

	createRes, err := (&TeamCreateTool{}).Execute(context.Background(), []byte(`{"team_name":"contract-team"}`), ctx)
	if err != nil || createRes.IsError {
		t.Fatalf("team_create failed: err=%v content=%q", err, createRes.Content)
	}

	for _, tc := range []struct {
		name  string
		input string
		tool  interface {
			Execute(context.Context, types.ToolInput, types.ToolContext) (types.ToolResult, error)
		}
	}{
		{name: "team_list", tool: &TeamListTool{}, input: `{}`},
		{name: "team_status", tool: &TeamStatusTool{}, input: `{"team_name":"contract-team"}`},
		{name: "team_update", tool: &TeamUpdateTool{}, input: `{"team_name":"contract-team","description":"updated"}`},
		{name: "team_delete", tool: &TeamDeleteTool{}, input: `{"team_name":"contract-team"}`},
	} {
		res, err := tc.tool.Execute(context.Background(), []byte(tc.input), ctx)
		if err != nil || res.IsError {
			t.Fatalf("%s failed: err=%v content=%q", tc.name, err, res.Content)
		}
		var out map[string]any
		if err := json.Unmarshal([]byte(res.Content), &out); err != nil {
			t.Fatalf("%s invalid JSON: %v", tc.name, err)
		}
		if out["contract"] == nil {
			t.Fatalf("%s missing contract metadata", tc.name)
		}
	}
}
