package agent

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestQueryReplayRewindClipsAfterTargetSequence(t *testing.T) {
	a := New(Config{})
	a.emit(types.AgentEvent{Type: types.AgentEventTurnStart, Turn: 1})
	a.emit(types.AgentEvent{Type: types.AgentEventReplayCheckpoint, Turn: 1, ReplayLabel: "turn_begin"})
	a.emit(types.AgentEvent{Type: types.AgentEventStop, Turn: 1, StopReason: types.AgentStopEndTurn})

	view := a.ReplayRewoundTo(2, types.AgentReplayQuery{})
	if len(view.Events) != 2 {
		t.Fatalf("events = %d, want 2", len(view.Events))
	}
	if !view.Stats.Rewind.Applied || view.Stats.Rewind.TargetSeq != 2 {
		t.Fatalf("unexpected rewind state: %+v", view.Stats.Rewind)
	}
}
