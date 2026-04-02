package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTeamListQuerySummary(t *testing.T) {
	resetOrchestrationStateForTests()
	ctx := types.ToolContext{}
	_, _ = (&TeamCreateTool{}).Execute(context.Background(), []byte(`{"team_name":"alpha","members":["dev1"]}`), ctx)
	_, _ = (&TeamCreateTool{}).Execute(context.Background(), []byte(`{"team_name":"beta","members":["dev2"]}`), ctx)

	res, err := (&TeamListTool{}).Execute(context.Background(), []byte(`{"member":"dev1","limit":1}`), ctx)
	if err != nil || res.IsError {
		t.Fatalf("team list failed: err=%v content=%q", err, res.Content)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(res.Content), &out); err != nil {
		t.Fatal(err)
	}
	if out["query_summary"] == nil {
		t.Fatalf("expected query_summary in output: %s", res.Content)
	}
}
