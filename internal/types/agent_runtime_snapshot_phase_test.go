package types

import "testing"

func TestAgentRuntimeSnapshotPhaseFields(t *testing.T) {
	r := AgentRuntimeSnapshot{Turns: 4, TurnIndex: 2, Phase: AgentTurnPhaseToolExecution}
	if r.TurnIndex != 2 || r.Phase != AgentTurnPhaseToolExecution {
		t.Fatalf("unexpected runtime snapshot: %+v", r)
	}
}
