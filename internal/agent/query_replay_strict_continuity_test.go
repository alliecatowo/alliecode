package agent

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestQueryReplayStrictContinuityBackfillsFilteredGaps(t *testing.T) {
	a := New(Config{})
	a.emit(types.AgentEvent{Type: types.AgentEventTurnStart, Turn: 1})
	a.emit(types.AgentEvent{Type: types.AgentEventToolStart, Turn: 1, ToolUseID: "t-1"})
	a.emit(types.AgentEvent{Type: types.AgentEventToolEnd, Turn: 1, ToolUseID: "t-1"})
	a.emit(types.AgentEvent{Type: types.AgentEventStop, Turn: 1})

	view := a.QueryReplay(types.AgentReplayQuery{
		Types:      []types.AgentEventType{types.AgentEventTurnStart, types.AgentEventStop},
		Continuity: types.AgentReplayContinuityStrict,
	})
	if len(view.Events) != 4 {
		t.Fatalf("events = %d, want 4", len(view.Events))
	}
	if !view.Stats.Continuity.Strict {
		t.Fatalf("expected strict continuity stats")
	}
}
