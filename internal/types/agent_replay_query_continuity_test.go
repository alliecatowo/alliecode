package types

import "testing"

func TestAgentReplayQueryContinuityFields(t *testing.T) {
	q := AgentReplayQuery{Continuity: AgentReplayContinuityStrict, ReplayLabel: "turn_end", RequireTools: true}
	if q.Continuity != AgentReplayContinuityStrict || q.ReplayLabel != "turn_end" || !q.RequireTools {
		t.Fatalf("unexpected continuity query fields: %+v", q)
	}
}
