package tools

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestToolSearchToolKeywordAndSelect(t *testing.T) {
	DefaultRegistry()

	keywordIn, _ := json.Marshal(map[string]any{"query": "task", "max_results": 3})
	res, err := (&ToolSearchTool{}).Execute(context.Background(), keywordIn, types.ToolContext{})
	if err != nil || res.IsError {
		t.Fatalf("keyword search failed: err=%v content=%q", err, res.Content)
	}

	selectIn, _ := json.Marshal(map[string]any{"query": "select:task_get"})
	res, err = (&ToolSearchTool{}).Execute(context.Background(), selectIn, types.ToolContext{})
	if err != nil || res.IsError {
		t.Fatalf("select search failed: err=%v content=%q", err, res.Content)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(res.Content), &out); err != nil {
		t.Fatalf("invalid json output: %v", err)
	}
	matches, _ := out["matches"].([]any)
	if len(matches) != 1 {
		t.Fatalf("expected one select match, got %d", len(matches))
	}
}

func TestFindToolMatchesDeterministicRanking(t *testing.T) {
	defs := []types.ToolDef{
		{Name: "task_get", Description: "Get task status and details"},
		{Name: "task_output", Description: "Get task output and status"},
		{Name: "task_update", Description: "Update task status"},
		{Name: "memory", Description: "Store and retrieve memory"},
	}

	first := findToolMatches(defs, "task get", 3)
	second := findToolMatches(defs, "task get", 3)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("expected deterministic ranking, got different outputs")
	}
	if len(first) == 0 || first[0].Name != "task_get" {
		t.Fatalf("expected task_get ranked first, got %+v", first)
	}
}

func TestFindToolMatchesExactNameBoost(t *testing.T) {
	defs := []types.ToolDef{
		{Name: "task_output", Description: "Task output"},
		{Name: "task_get", Description: "Task retrieval"},
	}

	res := findToolMatches(defs, "task_output", 2)
	if len(res) == 0 {
		t.Fatalf("expected at least one result")
	}
	if res[0].Name != "task_output" {
		t.Fatalf("expected exact name to rank first, got %q", res[0].Name)
	}
}
