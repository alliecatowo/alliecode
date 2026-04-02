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
