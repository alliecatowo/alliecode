package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/alliecatowo/alliecode/internal/tools"
	"github.com/alliecatowo/alliecode/internal/types"
)

func TestToolMetadataContractGoldenOutputs(t *testing.T) {
	t.Parallel()

	r := tools.DefaultRegistry()
	defs := r.ToolDefs()
	byName := make(map[string]types.ToolDef, len(defs))
	for _, def := range defs {
		byName[def.Name] = def
	}

	tests := []struct {
		name     string
		toolName string
		golden   string
	}{
		{name: "ask_user_question", toolName: "AskUserQuestion", golden: "ask_user_question.golden.json"},
		{name: "sleep", toolName: "Sleep", golden: "sleep.golden.json"},
		{name: "task_output", toolName: "task_output", golden: "task_output.golden.json"},
		{name: "tool_search", toolName: "tool_search", golden: "tool_search.golden.json"},
		{name: "bash", toolName: "Bash", golden: "bash.golden.json"},
		{name: "grep", toolName: "Grep", golden: "grep.golden.json"},
		{name: "glob", toolName: "Glob", golden: "glob.golden.json"},
		{name: "skill", toolName: "skill", golden: "skill.golden.json"},
		{name: "config", toolName: "config", golden: "config.golden.json"},
		{name: "lsp", toolName: "lsp", golden: "lsp.golden.json"},
		{name: "worktree_enter", toolName: "worktree_enter", golden: "worktree_enter.golden.json"},
		{name: "worktree_exit", toolName: "worktree_exit", golden: "worktree_exit.golden.json"},
		{name: "send_message", toolName: "send_message", golden: "send_message.golden.json"},
		{name: "team_create", toolName: "team_create", golden: "team_create.golden.json"},
		{name: "team_delete", toolName: "team_delete", golden: "team_delete.golden.json"},
		{name: "cron_create", toolName: "cron_create", golden: "cron_create.golden.json"},
		{name: "cron_delete", toolName: "cron_delete", golden: "cron_delete.golden.json"},
		{name: "cron_list", toolName: "cron_list", golden: "cron_list.golden.json"},
		{name: "remote_trigger", toolName: "remote_trigger", golden: "remote_trigger.golden.json"},
		{name: "brief", toolName: "brief", golden: "brief.golden.json"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, ok := byName[tc.toolName]
			if !ok {
				t.Fatalf("tool %q not found in default registry", tc.toolName)
			}

			want := readToolGolden(t, tc.golden)
			if !toolDefsEqual(got, want) {
				gotJSON, _ := json.MarshalIndent(got, "", "  ")
				wantJSON, _ := json.MarshalIndent(want, "", "  ")
				t.Fatalf("tool metadata mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", tc.toolName, string(gotJSON), string(wantJSON))
			}
		})
	}
}

func TestToolMetadataContractGoldenInventory(t *testing.T) {
	t.Parallel()

	r := tools.DefaultRegistry()
	defs := r.ToolDefs()
	sort.Slice(defs, func(i, j int) bool {
		return defs[i].Name < defs[j].Name
	})

	want := readToolInventoryGolden(t, "all_tools.golden.json")
	if !toolDefListsEqual(defs, want) {
		gotJSON, _ := json.MarshalIndent(defs, "", "  ")
		wantJSON, _ := json.MarshalIndent(want, "", "  ")
		t.Fatalf("tool metadata inventory mismatch\n--- got ---\n%s\n--- want ---\n%s", string(gotJSON), string(wantJSON))
	}
}

func readToolGolden(t *testing.T, name string) types.ToolDef {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime.Caller(0) failed")
	}
	path := filepath.Join(filepath.Dir(filename), "testdata", "tool_metadata_contract", name)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v", path, err)
	}

	var def types.ToolDef
	if err := json.Unmarshal(b, &def); err != nil {
		t.Fatalf("unmarshal golden %s: %v", path, err)
	}
	return def
}

func toolDefsEqual(a, b types.ToolDef) bool {
	ab, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bb, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(ab) == string(bb)
}

func readToolInventoryGolden(t *testing.T, name string) []types.ToolDef {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime.Caller(0) failed")
	}
	path := filepath.Join(filepath.Dir(filename), "testdata", "tool_metadata_contract", name)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v", path, err)
	}

	var defs []types.ToolDef
	if err := json.Unmarshal(b, &defs); err != nil {
		t.Fatalf("unmarshal golden %s: %v", path, err)
	}
	return defs
}

func toolDefListsEqual(a, b []types.ToolDef) bool {
	ab, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bb, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(ab) == string(bb)
}
