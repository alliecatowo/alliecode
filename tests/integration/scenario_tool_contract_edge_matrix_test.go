package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/tools"
	"github.com/alliecatowo/alliecode/internal/types"
)

type toolContractCase struct {
	Name         string   `json:"name"`
	Kind         string   `json:"kind"`
	Input        string   `json:"input"`
	WantIsError  bool     `json:"want_is_error"`
	WantExact    string   `json:"want_exact"`
	WantContains []string `json:"want_contains"`
	NetworkOnly  bool     `json:"network_only"`
}

func TestScenarioMatrix_ToolContractEdges(t *testing.T) {
	t.Parallel()

	paths := loadScenarioMatrixPaths(t, "tool_contract_edge", "*.json")
	for _, path := range paths {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()

			var tc toolContractCase
			decodeScenarioCase(t, path, &tc)
			if tc.NetworkOnly {
				t.Skip("network-only scenario")
			}

			res, err := executeToolContractCase(t, tc)
			if err != nil {
				t.Fatalf("execute scenario %s error = %v", tc.Name, err)
			}
			if res.IsError != tc.WantIsError {
				t.Fatalf("IsError = %t, want %t (content=%q)", res.IsError, tc.WantIsError, res.Content)
			}
			if tc.WantExact != "" {
				assertToolContractExact(t, tc, res.Content)
			}
			for _, needle := range tc.WantContains {
				if !strings.Contains(res.Content, needle) {
					t.Fatalf("content missing %q in %q", needle, res.Content)
				}
			}
		})
	}
}

func assertToolContractExact(t *testing.T, tc toolContractCase, got string) {
	t.Helper()

	if tc.Name != "task_missing_repeat" {
		if got != tc.WantExact {
			t.Fatalf("content mismatch\n got: %q\nwant: %q", got, tc.WantExact)
		}
		return
	}

	var gotPayload map[string]any
	if err := json.Unmarshal([]byte(got), &gotPayload); err != nil {
		t.Fatalf("unmarshal got content: %v", err)
	}
	var wantPayload map[string]any
	if err := json.Unmarshal([]byte(tc.WantExact), &wantPayload); err != nil {
		t.Fatalf("unmarshal want_exact content: %v", err)
	}

	gotRuntime, gotOK := gotPayload["runtime"].(map[string]any)
	wantRuntime, wantOK := wantPayload["runtime"].(map[string]any)
	if !gotOK || !wantOK {
		t.Fatalf("runtime payload missing\n got: %q\nwant: %q", got, tc.WantExact)
	}

	duration, ok := gotRuntime["duration_ms"].(float64)
	if !ok || duration < 0 {
		t.Fatalf("runtime.duration_ms must be >= 0, got %v", gotRuntime["duration_ms"])
	}

	delete(gotRuntime, "duration_ms")
	delete(wantRuntime, "duration_ms")

	gotNorm, err := json.Marshal(gotPayload)
	if err != nil {
		t.Fatalf("marshal normalized got content: %v", err)
	}
	wantNorm, err := json.Marshal(wantPayload)
	if err != nil {
		t.Fatalf("marshal normalized want content: %v", err)
	}
	if string(gotNorm) != string(wantNorm) {
		t.Fatalf("content mismatch after normalizing runtime.duration_ms\n got: %q\nwant: %q", string(gotNorm), string(wantNorm))
	}
}

func executeToolContractCase(t *testing.T, tc toolContractCase) (types.ToolResult, error) {
	t.Helper()

	switch tc.Kind {
	case "ask_user_question":
		return (&tools.AskUserQuestionTool{}).Execute(context.Background(), []byte(tc.Input), types.ToolContext{})
	case "task_output":
		return (&tools.TaskOutputTool{}).Execute(context.Background(), []byte(tc.Input), types.ToolContext{})
	case "sleep":
		return (&tools.SleepTool{}).Execute(context.Background(), []byte(tc.Input), types.ToolContext{})
	case "sleep_cancelled":
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		return (&tools.SleepTool{}).Execute(ctx, []byte(tc.Input), types.ToolContext{})
	case "bash":
		in, err := json.Marshal(map[string]any{"command": tc.Input})
		if err != nil {
			return types.ToolResult{}, err
		}
		return (&tools.BashTool{}).Execute(context.Background(), in, types.ToolContext{})
	case "glob_truncation":
		dir := t.TempDir()
		for i := 0; i < 520; i++ {
			name := filepath.Join(dir, "f"+strconv.Itoa(i)+".txt")
			if err := os.WriteFile(name, []byte("x"), 0o644); err != nil {
				return types.ToolResult{}, err
			}
		}
		in, err := json.Marshal(map[string]any{"pattern": "*.txt", "path": dir})
		if err != nil {
			return types.ToolResult{}, err
		}
		return (&tools.GlobTool{}).Execute(context.Background(), in, types.ToolContext{WorkingDir: dir})
	case "grep_truncation":
		dir := t.TempDir()
		path := filepath.Join(dir, "big.txt")
		var b strings.Builder
		for i := 0; i < 260; i++ {
			b.WriteString("hit line\n")
		}
		if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
			return types.ToolResult{}, err
		}
		in, err := json.Marshal(map[string]any{"pattern": "hit", "path": dir, "output_mode": "content"})
		if err != nil {
			return types.ToolResult{}, err
		}
		return (&tools.GrepTool{}).Execute(context.Background(), in, types.ToolContext{WorkingDir: dir})
	default:
		return types.ToolResult{}, fmt.Errorf("unknown scenario kind %q", tc.Kind)
	}
}
