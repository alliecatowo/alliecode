package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestPowerShellToolExecutesLocalCommand(t *testing.T) {
	tool := &PowerShellTool{}
	res, err := tool.Execute(context.Background(), []byte(`{"command":"printf 'ok'"}`), types.ToolContext{WorkingDir: t.TempDir()})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected successful result, got error: %s", res.Content)
	}
	if !strings.Contains(res.Content, "ok") {
		t.Fatalf("unexpected output: %q", res.Content)
	}
}

func TestPowerShellToolDenyRulesMatchBashSafety(t *testing.T) {
	tool := &PowerShellTool{}
	res, err := tool.Execute(context.Background(), []byte(`{"command":"git status; git diff"}`), types.ToolContext{WorkingDir: t.TempDir()})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected deny result")
	}
	if !strings.Contains(res.Content, "Command denied") {
		t.Fatalf("expected deny message, got: %q", res.Content)
	}
}

func TestPlanModeEnterExitLifecycle(t *testing.T) {
	resetPlanModeStateForTests()

	enter := &EnterPlanModeTool{}
	exit := &ExitPlanModeTool{}

	enterRes, err := enter.Execute(context.Background(), []byte(`{}`), types.ToolContext{})
	if err != nil || enterRes.IsError {
		t.Fatalf("enter_plan_mode failed: err=%v content=%q", err, enterRes.Content)
	}

	exitRes, err := exit.Execute(context.Background(), []byte(`{"allowedPrompts":[{"tool":"Bash","prompt":"run tests"}]}`), types.ToolContext{})
	if err != nil || exitRes.IsError {
		t.Fatalf("exit_plan_mode failed: err=%v content=%q", err, exitRes.Content)
	}
	if !strings.Contains(exitRes.Content, `"allowed_prompts_count":1`) {
		t.Fatalf("expected prompt count in response: %s", exitRes.Content)
	}
}

func TestExitPlanModeRequiresActivePlanMode(t *testing.T) {
	resetPlanModeStateForTests()
	res, err := (&ExitPlanModeTool{}).Execute(context.Background(), []byte(`{}`), types.ToolContext{})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected error when exiting without active plan mode")
	}
}

func TestREPLToolRelaysToPrimitiveTool(t *testing.T) {
	tool := &REPLTool{}
	input := []byte(`{"tool_name":"Bash","input":{"command":"printf relay"}}`)
	res, err := tool.Execute(context.Background(), input, types.ToolContext{WorkingDir: t.TempDir(), IsNonInteractive: true})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %s", res.Content)
	}
	if !strings.Contains(res.Content, `"result":"relay`) {
		t.Fatalf("expected relayed command output, got: %s", res.Content)
	}
}

func TestREPLToolRejectsDisallowedTarget(t *testing.T) {
	tool := &REPLTool{}
	res, err := tool.Execute(context.Background(), []byte(`{"tool_name":"task_list","input":{}}`), types.ToolContext{})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected disallowed target to error")
	}
}

func TestSyntheticOutputToolPassesThroughStructuredPayload(t *testing.T) {
	tool := &SyntheticOutputTool{}
	res, err := tool.Execute(context.Background(), []byte(`{"kind":"final","count":2}`), types.ToolContext{})
	if err != nil || res.IsError {
		t.Fatalf("synthetic_output failed: err=%v content=%q", err, res.Content)
	}

	var out map[string]any
	if err := json.Unmarshal([]byte(res.Content), &out); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if out["data"] != "Structured output provided successfully" {
		t.Fatalf("unexpected data field: %v", out["data"])
	}
	structured, ok := out["structured_output"].(map[string]any)
	if !ok {
		t.Fatalf("expected structured_output object, got %T", out["structured_output"])
	}
	if structured["kind"] != "final" {
		t.Fatalf("expected passthrough payload, got %v", structured)
	}
}

func TestSyntheticOutputToolRejectsNonObjectInput(t *testing.T) {
	tool := &SyntheticOutputTool{}
	res, err := tool.Execute(context.Background(), []byte(`[]`), types.ToolContext{})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected error for non-object input")
	}
}
