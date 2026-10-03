package integration_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/agent"
	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/tools"
	"github.com/alliecatowo/alliecode/internal/types"
)

type integrationToolLoopProvider struct {
	chatCalls int
}

func (p *integrationToolLoopProvider) Name() string { return "integration-loop" }

func (p *integrationToolLoopProvider) Chat(_ context.Context, _ types.ChatRequest) (<-chan types.StreamEvent, error) {
	p.chatCalls++
	ch := make(chan types.StreamEvent, 8)
	go func(call int) {
		defer close(ch)
		if call == 1 {
			ch <- types.StreamEvent{Type: types.StreamToolUseStart, ToolUseID: "tool-1", ToolName: "fake"}
			ch <- types.StreamEvent{Type: types.StreamToolUseDelta, ToolUseID: "tool-1", ToolName: "fake", Delta: `{"value":1}`}
			ch <- types.StreamEvent{Type: types.StreamToolUseDone, ToolUseID: "tool-1", ToolName: "fake"}
			ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopToolUse}
			return
		}
		msg := types.NewTextMessage(types.RoleAssistant, "complete")
		ch <- types.StreamEvent{Type: types.StreamContentDelta, Delta: "complete"}
		ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopEndTurn, Message: &msg}
	}(p.chatCalls)
	return ch, nil
}

func (p *integrationToolLoopProvider) ChatSync(context.Context, types.ChatRequest) (*types.ChatResponse, error) {
	return nil, nil
}

func (p *integrationToolLoopProvider) ListModels(context.Context) ([]types.Model, error) {
	return nil, nil
}
func (p *integrationToolLoopProvider) SupportsStreaming() bool { return true }
func (p *integrationToolLoopProvider) SupportsTools() bool     { return true }
func (p *integrationToolLoopProvider) SupportsThinking() bool  { return false }

type integrationFakeTool struct {
	executeCount int
}

func (t *integrationFakeTool) Name() string        { return "fake" }
func (t *integrationFakeTool) Description() string { return "integration fake tool" }
func (t *integrationFakeTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{Type: "object"}
}
func (t *integrationFakeTool) Execute(context.Context, types.ToolInput, types.ToolContext) (types.ToolResult, error) {
	t.executeCount++
	return types.ToolResult{Content: "ok"}, nil
}
func (t *integrationFakeTool) IsReadOnly(types.ToolInput) bool        { return false }
func (t *integrationFakeTool) IsDestructive(types.ToolInput) bool     { return false }
func (t *integrationFakeTool) IsConcurrencySafe(types.ToolInput) bool { return false }
func (t *integrationFakeTool) CheckPermissions(types.ToolInput, types.ToolContext) types.ToolPermission {
	return types.PermissionAllowed
}

func TestIntegrationAgentToolLoopContinuesPastConfiguredSingleTurn(t *testing.T) {
	provider := &integrationToolLoopProvider{}
	tool := &integrationFakeTool{}
	ag := agent.New(agent.Config{Provider: provider, Tools: []types.Tool{tool}, MaxTurns: 1, WorkingDir: t.TempDir()})

	if err := ag.Run(context.Background(), "run tool then continue"); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if provider.chatCalls != 2 {
		t.Fatalf("chat calls = %d, want 2", provider.chatCalls)
	}
	if tool.executeCount != 1 {
		t.Fatalf("tool execute count = %d, want 1", tool.executeCount)
	}
}

func TestIntegrationStatusAndSandboxReportWorkspacePathScope(t *testing.T) {
	workingDir := t.TempDir()
	ag := agent.New(agent.Config{WorkingDir: workingDir})
	state := &commands.RuntimeState{Agent: ag, ProviderName: "openai", Model: "gpt-4o-mini"}
	ctx := commands.Context{State: state}

	statusRes, err := commands.NewStatusCommand().Execute(context.Background(), ctx, commands.Invocation{Name: "status"})
	if err != nil {
		t.Fatalf("status execute failed: %v", err)
	}
	for _, want := range []string{"working_dir=" + workingDir, "workspace_root=" + workingDir, "path_scope=workspace"} {
		if !strings.Contains(statusRes.Message, want) {
			t.Fatalf("status output missing %q: %q", want, statusRes.Message)
		}
	}

	sandboxRes, err := commands.NewSandboxCommand().Execute(context.Background(), ctx, commands.Invocation{Name: "sandbox", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("sandbox status execute failed: %v", err)
	}
	for _, want := range []string{"scope.working_dir=" + workingDir, "scope.workspace_root=" + workingDir, "scope.path_scope=workspace"} {
		if !strings.Contains(sandboxRes.Message, want) {
			t.Fatalf("sandbox output missing %q: %q", want, sandboxRes.Message)
		}
	}
}

func TestIntegrationShellExecutionMetadataIncludesWorkspaceScope(t *testing.T) {
	workingDir := t.TempDir()
	in, err := json.Marshal(map[string]any{"command": "printf 'ok'"})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	res, err := (&tools.BashTool{}).Execute(context.Background(), in, types.ToolContext{WorkingDir: workingDir})
	if err != nil {
		t.Fatalf("Bash.Execute() error = %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got error: %q", res.Content)
	}
	if !strings.Contains(res.Content, "workspace_root: "+workingDir) || !strings.Contains(res.Content, "path_scope: workspace") {
		t.Fatalf("shell execution metadata missing workspace scope: %q", res.Content)
	}
}
