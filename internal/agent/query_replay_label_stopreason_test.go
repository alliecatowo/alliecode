package agent

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestReplayEventsForLabel(t *testing.T) {
	a := New(Config{})
	a.emit(types.AgentEvent{Type: types.AgentEventReplayCheckpoint, ReplayLabel: "turn_begin"})
	a.emit(types.AgentEvent{Type: types.AgentEventReplayCheckpoint, ReplayLabel: "turn_end"})

	got := a.ReplayEventsForLabel("turn_end", 10)
	if len(got) != 1 || got[0].ReplayLabel != "turn_end" {
		t.Fatalf("unexpected events for label: %+v", got)
	}
}

func TestReplayEventsForStopReason(t *testing.T) {
	a := New(Config{})
	a.emit(types.AgentEvent{Type: types.AgentEventStop, StopReason: types.AgentStopProviderError})
	a.emit(types.AgentEvent{Type: types.AgentEventStop, StopReason: types.AgentStopBudgetToken})

	got := a.ReplayEventsForStopReason(types.AgentStopBudgetToken, 10)
	if len(got) != 1 || got[0].StopReason != types.AgentStopBudgetToken {
		t.Fatalf("unexpected events for stop reason: %+v", got)
	}
}
