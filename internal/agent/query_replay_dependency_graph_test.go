package agent

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestBuildContextDependencyGraphTracksToolFlow(t *testing.T) {
	events := []types.AgentEvent{
		{Sequence: 1, Type: types.AgentEventTurnStart, Turn: 1},
		{Sequence: 2, Type: types.AgentEventToolStart, Turn: 1, ToolUseID: "tool-1"},
		{Sequence: 3, Type: types.AgentEventToolEnd, Turn: 1, ToolUseID: "tool-1"},
	}
	g := buildContextDependencyGraph(events)
	if len(g.Nodes) != 3 {
		t.Fatalf("nodes = %d, want 3", len(g.Nodes))
	}
	if g.Nodes[2].Kind != types.AgentContextDependencyToolFlow {
		t.Fatalf("kind = %q, want %q", g.Nodes[2].Kind, types.AgentContextDependencyToolFlow)
	}
	if len(g.Nodes[2].DependsOn) == 0 {
		t.Fatalf("expected tool dependency edge")
	}
}
