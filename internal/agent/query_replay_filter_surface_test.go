package agent

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestQueryReplayFiltersByClassToolIDAndTurnIndex(t *testing.T) {
	a := New(Config{})
	a.emit(types.AgentEvent{Type: types.AgentEventTurnStart, Turn: 1})
	a.emit(types.AgentEvent{Type: types.AgentEventToolStart, Turn: 1, ToolUseID: "tool-1"})
	a.emit(types.AgentEvent{Type: types.AgentEventToolEnd, Turn: 1, ToolUseID: "tool-1"})
	a.emit(types.AgentEvent{Type: types.AgentEventToolStart, Turn: 2, ToolUseID: "tool-2"})
	a.emit(types.AgentEvent{Type: types.AgentEventPermissionResult, Turn: 2, ToolUseID: "tool-2"})

	view := a.QueryReplay(types.AgentReplayQuery{
		Classes:   []types.AgentEventClass{types.AgentEventClassTool},
		ToolID:    "tool-2",
		TurnIndex: 2,
	})

	if len(view.Events) != 1 {
		t.Fatalf("events = %d, want 1", len(view.Events))
	}
	if view.Events[0].Type != types.AgentEventToolStart || view.Events[0].ToolUseID != "tool-2" {
		t.Fatalf("unexpected event: %+v", view.Events[0])
	}
	if view.Stats.ToolsSeen != 1 {
		t.Fatalf("tools_seen = %d, want 1", view.Stats.ToolsSeen)
	}
}

func TestQueryReplayClassCounts(t *testing.T) {
	a := New(Config{})
	a.emit(types.AgentEvent{Type: types.AgentEventTurnStart, Turn: 1})
	a.emit(types.AgentEvent{Type: types.AgentEventTurnPhase, Turn: 1})
	a.emit(types.AgentEvent{Type: types.AgentEventToolStart, Turn: 1, ToolUseID: "x"})
	a.emit(types.AgentEvent{Type: types.AgentEventStop, Turn: 1})

	view := a.QueryReplay(types.AgentReplayQuery{})
	if len(view.Stats.ClassCounts) == 0 {
		t.Fatalf("expected class counts")
	}
}
