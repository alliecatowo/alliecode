package agent

import (
	"context"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestLoopEmitsValidatedTransitionAndTurnPhase(t *testing.T) {
	provider := &scriptedProvider{scripts: []func(chan types.StreamEvent){func(ch chan types.StreamEvent) {
		ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopEndTurn}
	}}}
	a := New(Config{Provider: provider, MaxTurns: 1})

	var transitionEvent *types.AgentEvent
	var phaseEvent *types.AgentEvent
	a.SetEventCallback(func(ev types.AgentEvent) {
		if ev.Type == types.AgentEventStateTransition && transitionEvent == nil {
			cp := ev
			transitionEvent = &cp
		}
		if ev.Type == types.AgentEventTurnPhase && phaseEvent == nil {
			cp := ev
			phaseEvent = &cp
		}
	})

	if err := a.Run(context.Background(), "hello"); err != nil {
		t.Fatalf("Run() err = %v", err)
	}
	if transitionEvent == nil || phaseEvent == nil {
		t.Fatalf("expected transition and turn_phase events")
	}
	if transitionEvent.Details != "" {
		t.Fatalf("expected valid transition without validator details, got %q", transitionEvent.Details)
	}
	if phaseEvent.Details != "" {
		t.Fatalf("expected valid turn phase without validator details, got %q", phaseEvent.Details)
	}
}
