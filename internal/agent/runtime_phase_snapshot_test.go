package agent

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestEmitTurnPhaseUpdatesRuntimeSnapshot(t *testing.T) {
	a := New(Config{})
	a.emitTurnPhase(3, types.AgentTurnPhaseRetry, types.AgentTurnPhaseProviderStream, "retry_provider_error", false, "provider_error")

	rt := a.RuntimeSnapshot()
	if rt.TurnIndex != 3 {
		t.Fatalf("turn_index = %d, want 3", rt.TurnIndex)
	}
	if rt.Phase != types.AgentTurnPhaseRetry {
		t.Fatalf("phase = %q, want retry", rt.Phase)
	}
}

func TestEmitStopEmitsTurnEndPhase(t *testing.T) {
	a := New(Config{})
	a.emitStop(types.AgentStopEndTurn, types.StopEndTurn, "", 1)

	view := a.QueryReplay(types.AgentReplayQuery{Classes: []types.AgentEventClass{types.AgentEventClassLifecycle}})
	found := false
	for _, ev := range view.Events {
		if ev.Type == types.AgentEventTurnPhase && ev.TurnPhase.Phase == types.AgentTurnPhaseTurnEnd {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected turn_end phase event in replay")
	}
}
