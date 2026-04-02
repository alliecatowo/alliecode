package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTeamStatusMessageAndHistoryLimit(t *testing.T) {
	resetOrchestrationStateForTests()
	ctx := types.ToolContext{}
	_, _ = (&TeamCreateTool{}).Execute(context.Background(), []byte(`{"team_name":"delta","members":["dev"]}`), ctx)
	for i := 0; i < 3; i++ {
		_, _ = (&SendMessageTool{}).Execute(context.Background(), []byte(`{"to":"dev","from":"team_lead","message":"hello"}`), ctx)
	}
	res, err := (&TeamStatusTool{}).Execute(context.Background(), []byte(`{"team_name":"delta","message_limit":1,"history_limit":2}`), ctx)
	if err != nil || res.IsError {
		t.Fatalf("team status failed: err=%v content=%q", err, res.Content)
	}
	var out struct {
		Team struct {
			Messages []any `json:"messages"`
			History  []any `json:"history"`
		} `json:"team"`
	}
	if err := json.Unmarshal([]byte(res.Content), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Team.Messages) != 1 || len(out.Team.History) != 2 {
		t.Fatalf("unexpected trimmed team status: %+v", out.Team)
	}
}
