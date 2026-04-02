package agent

import (
	"fmt"

	"github.com/alliecatowo/alliecode/internal/types"
)

func buildContextDependencyGraph(events []types.AgentEvent) types.AgentContextDependencyGraph {
	nodes := make([]types.AgentContextDependencyNode, 0, len(events))
	lastByTurn := map[int]string{}
	lastByTool := map[string]string{}

	for _, ev := range events {
		nodeID := fmt.Sprintf("seq:%d", ev.Sequence)
		node := types.AgentContextDependencyNode{
			ID:          nodeID,
			Kind:        types.AgentContextDependencyEvent,
			Sequence:    ev.Sequence,
			Turn:        ev.Turn,
			ReplayLabel: ev.ReplayLabel,
			ToolUseID:   ev.ToolUseID,
			EventType:   ev.Type,
		}

		if ev.Turn > 0 {
			if prev, ok := lastByTurn[ev.Turn]; ok {
				node.DependsOn = append(node.DependsOn, prev)
			}
			lastByTurn[ev.Turn] = nodeID
		}

		if ev.ToolUseID != "" {
			if prev, ok := lastByTool[ev.ToolUseID]; ok {
				node.DependsOn = append(node.DependsOn, prev)
			}
			lastByTool[ev.ToolUseID] = nodeID
			switch ev.Type {
			case types.AgentEventToolStart, types.AgentEventToolEnd, types.AgentEventToolRetry:
				node.Kind = types.AgentContextDependencyToolFlow
			default:
				node.Kind = types.AgentContextDependencyToolUse
			}
		} else if ev.Type == types.AgentEventTurnStart || ev.Type == types.AgentEventTurnEnd {
			node.Kind = types.AgentContextDependencyTurn
		}

		nodes = append(nodes, node)
	}

	return types.AgentContextDependencyGraph{Nodes: nodes}
}

func missingDependencyParents(events []types.AgentEvent) int {
	graph := buildContextDependencyGraph(events)
	ids := make(map[string]struct{}, len(graph.Nodes))
	for _, node := range graph.Nodes {
		ids[node.ID] = struct{}{}
	}
	missing := 0
	for _, node := range graph.Nodes {
		for _, dep := range node.DependsOn {
			if _, ok := ids[dep]; !ok {
				missing++
			}
		}
	}
	return missing
}
