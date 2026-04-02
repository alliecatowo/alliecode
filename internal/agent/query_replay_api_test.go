package agent

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestReplayEventsByType(t *testing.T) {
	a := New(Config{})
	a.emit(types.AgentEvent{Type: types.AgentEventTurnStart, Turn: 1})
	a.emit(types.AgentEvent{Type: types.AgentEventToolStart, Turn: 1})
	a.emit(types.AgentEvent{Type: types.AgentEventToolEnd, Turn: 1})

	got := a.ReplayEventsByType(types.AgentEventToolStart, 10)
	if len(got) != 1 || got[0].Type != types.AgentEventToolStart {
		t.Fatalf("unexpected replay by type: %+v", got)
	}
}

func TestReplayEventsForTurn(t *testing.T) {
	a := New(Config{})
	a.emit(types.AgentEvent{Type: types.AgentEventTurnStart, Turn: 1})
	a.emit(types.AgentEvent{Type: types.AgentEventTurnStart, Turn: 2})

	got := a.ReplayEventsForTurn(2, 10)
	if len(got) != 1 || got[0].Turn != 2 {
		t.Fatalf("unexpected replay for turn: %+v", got)
	}
}
