package types

import "testing"

func TestAgentReplayQueryStructsCompile(t *testing.T) {
	q := AgentReplayQuery{FromSequence: 1, ToSequence: 2, Turn: 1, Limit: 5}
	v := AgentReplayView{Query: q}
	if v.Query.Limit != 5 {
		t.Fatalf("limit = %d, want 5", v.Query.Limit)
	}
}
