package agent

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestQueryReplayStatsIncludeContinuityAndRewindState(t *testing.T) {
	a := New(Config{})
	a.emit(types.AgentEvent{Type: types.AgentEventReplayCheckpoint, Turn: 1, ReplayLabel: "turn_begin"})
	a.emit(types.AgentEvent{Type: types.AgentEventToolStart, Turn: 1, ToolUseID: "tool-1"})
	a.emit(types.AgentEvent{Type: types.AgentEventToolEnd, Turn: 1, ToolUseID: "tool-1"})

	view := a.QueryReplay(types.AgentReplayQuery{Continuity: types.AgentReplayContinuityStrict, RewindTo: 2})
	if view.Stats.Continuity.Mode != types.AgentReplayContinuityStrict {
		t.Fatalf("continuity mode = %q, want strict", view.Stats.Continuity.Mode)
	}
	if !view.Stats.Rewind.Applied || view.Stats.Rewind.TargetSeq != 2 {
		t.Fatalf("rewind state = %+v, want applied target=2", view.Stats.Rewind)
	}
	if len(view.Stats.Rewind.Checkpoints) == 0 {
		t.Fatalf("expected rewind checkpoints")
	}
}
