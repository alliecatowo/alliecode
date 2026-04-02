package types

import "testing"

func TestAgentReplayQueryExtendedFilters(t *testing.T) {
	q := AgentReplayQuery{
		Classes:   []AgentEventClass{AgentEventClassTool},
		ToolID:    "tool-9",
		TurnIndex: 2,
		Limit:     25,
	}
	if q.ToolID != "tool-9" || q.TurnIndex != 2 || len(q.Classes) != 1 {
		t.Fatalf("unexpected replay query: %+v", q)
	}
}
