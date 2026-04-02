package types

import "testing"

func TestValidateLifecycleTransition(t *testing.T) {
	valid := AgentTransition{From: AgentLifecycleIdle, To: AgentLifecycleTurnStarting, Reason: AgentTransitionInputAccepted}
	if err := ValidateLifecycleTransition(valid); err != nil {
		t.Fatalf("valid transition rejected: %v", err)
	}
	invalid := AgentTransition{From: AgentLifecycleIdle, To: AgentLifecycleToolExecution, Reason: AgentTransitionToolExecutionStart}
	if err := ValidateLifecycleTransition(invalid); err == nil {
		t.Fatalf("expected invalid transition to fail")
	}
}

func TestValidateTurnPhaseState(t *testing.T) {
	if err := ValidateTurnPhaseState(AgentTurnPhaseState{Phase: AgentTurnPhaseProviderRequest, PreviousPhase: AgentTurnPhaseCompactionCheck}); err != nil {
		t.Fatalf("expected valid phase transition: %v", err)
	}
	if err := ValidateTurnPhaseState(AgentTurnPhaseState{Phase: AgentTurnPhaseToolExecution, PreviousPhase: AgentTurnPhaseInit}); err == nil {
		t.Fatalf("expected invalid turn phase transition")
	}
}
