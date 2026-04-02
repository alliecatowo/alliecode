package types

import "testing"

func TestAgentContextDependencyGraphStructs(t *testing.T) {
	g := AgentContextDependencyGraph{Nodes: []AgentContextDependencyNode{{ID: "seq:1", Kind: AgentContextDependencyEvent}}}
	if len(g.Nodes) != 1 || g.Nodes[0].Kind != AgentContextDependencyEvent {
		t.Fatalf("unexpected graph shape: %+v", g)
	}
}
