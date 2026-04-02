package types

import "testing"

func TestAgentStatusViewFields(t *testing.T) {
	v := AgentStatusView{
		Lifecycle: AgentLifecycleIdle,
		Runtime:   AgentRuntimeSnapshot{Turns: 1},
		Replay:    AgentReplayCursor{Start: 1, End: 2, Size: 2},
	}
	if v.Runtime.Turns != 1 {
		t.Fatalf("turns = %d, want 1", v.Runtime.Turns)
	}
}
