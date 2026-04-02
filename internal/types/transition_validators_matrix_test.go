package types

import "testing"

func TestIsValidLifecycleTransitionMatrix(t *testing.T) {
	if !IsValidLifecycleTransition(AgentLifecycleProviderStream, AgentLifecycleToolExecution) {
		t.Fatalf("expected provider_stream -> tool_execution to be valid")
	}
	if IsValidLifecycleTransition(AgentLifecycleStopped, AgentLifecycleProviderRequest) {
		t.Fatalf("expected stopped -> provider_request to be invalid")
	}
}

func TestIsValidTurnPhaseTransitionMatrix(t *testing.T) {
	if !IsValidTurnPhaseTransition(AgentTurnPhaseRetry, AgentTurnPhaseProviderRequest) {
		t.Fatalf("expected retry -> provider_request to be valid")
	}
	if IsValidTurnPhaseTransition(AgentTurnPhaseTurnEnd, AgentTurnPhaseProviderRequest) {
		t.Fatalf("expected turn_end -> provider_request to be invalid")
	}
}
