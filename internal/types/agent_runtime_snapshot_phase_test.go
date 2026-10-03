package types

import "testing"

func TestAgentRuntimeSnapshotPhaseFields(t *testing.T) {
	r := AgentRuntimeSnapshot{ProviderName: "anthropic", Model: "claude-opus-4-20250514", ModelRef: "anthropic/claude-opus-4-20250514", Turns: 4, TurnIndex: 2, Phase: AgentTurnPhaseToolExecution}
	if r.TurnIndex != 2 || r.Phase != AgentTurnPhaseToolExecution || r.ProviderName != "anthropic" || r.ModelRef != "anthropic/claude-opus-4-20250514" {
		t.Fatalf("unexpected runtime snapshot: %+v", r)
	}
}
