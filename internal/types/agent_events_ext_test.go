package types

import "testing"

func TestAgentEventExtendedViews(t *testing.T) {
	ev := AgentEvent{
		TurnPlan:   AgentTurnPlan{Turn: 2, BudgetCheck: true},
		RetryPlan:  AgentRetryPlan{Kind: "provider_error", Attempt: 1, Max: 1},
		Recovery:   AgentRecoveryState{Kind: "prompt_too_long", Recovered: true},
		Compaction: AgentCompactionView{Forced: true, Applied: true},
	}
	if ev.TurnPlan.Turn != 2 || !ev.Compaction.Applied {
		t.Fatalf("unexpected extended event payload: %+v", ev)
	}
}
