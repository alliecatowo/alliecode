package types

import "testing"

func TestAgentTurnPhaseValuesAreNonEmpty(t *testing.T) {
	phases := []AgentTurnPhase{
		AgentTurnPhaseInit,
		AgentTurnPhaseProviderRequest,
		AgentTurnPhaseProviderStream,
		AgentTurnPhaseToolExecution,
		AgentTurnPhaseRetry,
		AgentTurnPhaseTurnEnd,
	}
	for _, phase := range phases {
		if phase == "" {
			t.Fatalf("found empty turn phase")
		}
	}
}
