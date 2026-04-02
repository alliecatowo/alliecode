package agent

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestQueryReplayFilterByStopReason(t *testing.T) {
	a := New(Config{})
	a.emit(types.AgentEvent{Type: types.AgentEventTurnStart, Turn: 1})
	a.emit(types.AgentEvent{Type: types.AgentEventStop, Turn: 1, StopReason: types.AgentStopBudgetUSD})
	a.emit(types.AgentEvent{Type: types.AgentEventStop, Turn: 2, StopReason: types.AgentStopProviderError})

	view := a.QueryReplay(types.AgentReplayQuery{StopReasons: []types.AgentStopReason{types.AgentStopBudgetUSD}})
	if len(view.Events) != 1 || view.Events[0].StopReason != types.AgentStopBudgetUSD {
		t.Fatalf("unexpected stop reason filter result: %+v", view.Events)
	}
}
