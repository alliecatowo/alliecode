package agent

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestQueryReplayFiltersAndStats(t *testing.T) {
	a := New(Config{})
	a.emit(types.AgentEvent{Type: types.AgentEventTurnStart, Turn: 1})
	a.emit(types.AgentEvent{Type: types.AgentEventToolStart, Turn: 1})
	a.emit(types.AgentEvent{Type: types.AgentEventToolEnd, Turn: 1})
	a.emit(types.AgentEvent{Type: types.AgentEventStop, Turn: 1})

	view := a.QueryReplay(types.AgentReplayQuery{
		Types: []types.AgentEventType{types.AgentEventToolStart, types.AgentEventToolEnd},
		Turn:  1,
		Limit: 10,
	})

	if len(view.Events) != 2 {
		t.Fatalf("events = %d, want 2", len(view.Events))
	}
	if view.Stats.Matched != 2 {
		t.Fatalf("matched = %d, want 2", view.Stats.Matched)
	}
}

func TestStatusViewIncludesRuntimeAndReplay(t *testing.T) {
	a := New(Config{})
	a.emit(types.AgentEvent{Type: types.AgentEventTurnStart, Turn: 1})
	status := a.StatusView()
	if status.Replay.Size == 0 {
		t.Fatalf("expected replay cursor size > 0")
	}
}
