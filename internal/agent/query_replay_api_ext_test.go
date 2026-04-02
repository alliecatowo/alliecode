package agent

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestReplayEventsByClass(t *testing.T) {
	a := New(Config{})
	a.emit(types.AgentEvent{Type: types.AgentEventPermissionAsk, Turn: 1})
	a.emit(types.AgentEvent{Type: types.AgentEventToolStart, Turn: 1})

	got := a.ReplayEventsByClass(types.AgentEventClassPermission, 10)
	if len(got) != 1 || got[0].Type != types.AgentEventPermissionAsk {
		t.Fatalf("unexpected replay by class: %+v", got)
	}
}

func TestReplayEventsByToolIDAndTurnIndex(t *testing.T) {
	a := New(Config{})
	a.emit(types.AgentEvent{Type: types.AgentEventToolStart, Turn: 1, ToolUseID: "t-1"})
	a.emit(types.AgentEvent{Type: types.AgentEventToolEnd, Turn: 2, ToolUseID: "t-2"})

	toolEvents := a.ReplayEventsByToolID("t-2", 10)
	if len(toolEvents) != 1 || toolEvents[0].ToolUseID != "t-2" {
		t.Fatalf("unexpected replay by tool id: %+v", toolEvents)
	}

	turnEvents := a.ReplayEventsForTurnIndex(2, 10)
	if len(turnEvents) != 1 || turnEvents[0].Turn != 2 {
		t.Fatalf("unexpected replay by turn index: %+v", turnEvents)
	}
}
