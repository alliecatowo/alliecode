package types

import "testing"

func TestAgentRecoveryStateExtendedFields(t *testing.T) {
	r := AgentRecoveryState{
		Kind:        "prompt_too_long",
		Branch:      "prompt_too_long",
		FromPhase:   AgentTurnPhaseProviderRequest,
		ToPhase:     AgentTurnPhaseRetry,
		Attempt:     1,
		MaxAttempts: 2,
	}
	if r.Branch == "" || r.FromPhase == "" || r.ToPhase == "" {
		t.Fatalf("unexpected recovery state: %+v", r)
	}
}
