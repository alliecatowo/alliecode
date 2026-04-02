package agent

import (
	"context"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTurnStartIncludesPhaseMetadata(t *testing.T) {
	provider := &scriptedProvider{scripts: []func(chan types.StreamEvent){func(ch chan types.StreamEvent) {
		ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopEndTurn}
	}}}
	a := New(Config{Provider: provider, MaxTurns: 1})
	var plan types.AgentTurnPlan
	a.SetEventCallback(func(ev types.AgentEvent) {
		if ev.Type == types.AgentEventTurnStart {
			plan = ev.TurnPlan
		}
	})
	if err := a.Run(context.Background(), "phase"); err != nil {
		t.Fatalf("Run() err = %v", err)
	}
	if plan.Phase != types.AgentTurnPhaseInit {
		t.Fatalf("expected init phase, got %q", plan.Phase)
	}
}
