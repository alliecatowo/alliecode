package types

import "testing"

func TestAgentTurnPhaseStateFields(t *testing.T) {
	state := AgentTurnPhaseState{
		Turn:           1,
		TurnIndex:      1,
		Phase:          AgentTurnPhaseProviderRequest,
		PreviousPhase:  AgentTurnPhaseCompactionCheck,
		Transition:     "provider_request",
		RecoveryBranch: "provider_error",
	}
	if state.PreviousPhase == "" || state.Transition == "" {
		t.Fatalf("unexpected phase state: %+v", state)
	}
}
