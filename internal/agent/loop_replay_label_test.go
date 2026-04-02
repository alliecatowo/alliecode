package agent

import (
	"context"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestLoopEmitsNamedReplayCheckpoints(t *testing.T) {
	provider := &scriptedProvider{scripts: []func(chan types.StreamEvent){func(ch chan types.StreamEvent) {
		ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopEndTurn}
	}}}
	a := New(Config{Provider: provider, MaxTurns: 1})
	if err := a.Run(context.Background(), "hello"); err != nil {
		t.Fatalf("Run() err = %v", err)
	}
	events := a.ReplayEvents(0, 0)
	foundStart := false
	foundEnd := false
	for _, ev := range events {
		if ev.Type != types.AgentEventReplayCheckpoint {
			continue
		}
		if ev.ReplayLabel == "turn_begin" || ev.ReplayLabel == "turn_start" {
			foundStart = true
		}
		if ev.ReplayLabel == "turn_end" {
			foundEnd = true
		}
	}
	if !foundStart || !foundEnd {
		t.Fatalf("expected start and end replay labels")
	}
}
