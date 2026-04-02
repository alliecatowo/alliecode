package types

import "testing"

func TestAgentReplayQueryRewindFields(t *testing.T) {
	q := AgentReplayQuery{RewindTo: 42, StopReasons: []AgentStopReason{AgentStopCanceled}}
	if q.RewindTo != 42 || len(q.StopReasons) != 1 {
		t.Fatalf("unexpected rewind query fields: %+v", q)
	}
	v := AgentReplayView{Stats: AgentReplayStats{Rewind: AgentReplayRewindState{Applied: true, TargetSeq: 42}}}
	if !v.Stats.Rewind.Applied || v.Stats.Rewind.TargetSeq != 42 {
		t.Fatalf("unexpected rewind view fields: %+v", v.Stats.Rewind)
	}
}
