package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTeamStatusIncludesContractWave6(t *testing.T) {
	resetOrchestrationStateForTests()
	_, _ = (&TeamCreateTool{}).Execute(context.Background(), []byte(`{"team_name":"wave6-status"}`), types.ToolContext{})
	res, err := (&TeamStatusTool{}).Execute(context.Background(), []byte(`{"team_name":"wave6-status"}`), types.ToolContext{})
	if err != nil || res.IsError {
		t.Fatalf("team_status failed: err=%v content=%q", err, res.Content)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(res.Content), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["contract"] == nil {
		t.Fatalf("expected contract metadata")
	}
}
