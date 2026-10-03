package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alliecatowo/alliecode/internal/hooks"
	"github.com/alliecatowo/alliecode/internal/permissions"
	"github.com/alliecatowo/alliecode/internal/types"
)

type scriptedProvider struct {
	mu       sync.Mutex
	name     string
	scripts  []func(chan types.StreamEvent)
	errs     []error
	requests []types.ChatRequest
	syncResp *types.ChatResponse
	syncErr  error
	chatCall int
}

func (p *scriptedProvider) Name() string {
	if strings.TrimSpace(p.name) == "" {
		return "scripted"
	}
	return p.name
}

func (p *scriptedProvider) Chat(_ context.Context, req types.ChatRequest) (<-chan types.StreamEvent, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.chatCall++
	p.requests = append(p.requests, req)
	if len(p.errs) > 0 {
		err := p.errs[0]
		p.errs = p.errs[1:]
		if err != nil {
			return nil, err
		}
	}
	if len(p.scripts) == 0 {
		return nil, errors.New("no scripted stream available")
	}
	script := p.scripts[0]
	p.scripts = p.scripts[1:]
	ch := make(chan types.StreamEvent)
	go func() {
		defer close(ch)
		script(ch)
	}()
	return ch, nil
}

func (p *scriptedProvider) ChatSync(context.Context, types.ChatRequest) (*types.ChatResponse, error) {
	if p.syncResp != nil || p.syncErr != nil {
		return p.syncResp, p.syncErr
	}
	return nil, errors.New("not implemented")
}

func (p *scriptedProvider) ListModels(context.Context) ([]types.Model, error) { return nil, nil }
func (p *scriptedProvider) SupportsStreaming() bool                           { return true }
func (p *scriptedProvider) SupportsTools() bool                               { return true }
func (p *scriptedProvider) SupportsThinking() bool                            { return false }

type fakeTool struct {
	name         string
	executeCount int
	result       string
}

func (t *fakeTool) Name() string        { return t.name }
func (t *fakeTool) Description() string { return "fake tool" }
func (t *fakeTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{Type: "object"}
}
func (t *fakeTool) Execute(context.Context, types.ToolInput, types.ToolContext) (types.ToolResult, error) {
	t.executeCount++
	return types.ToolResult{Content: t.result}, nil
}
func (t *fakeTool) IsReadOnly(types.ToolInput) bool        { return false }
func (t *fakeTool) IsDestructive(types.ToolInput) bool     { return false }
func (t *fakeTool) IsConcurrencySafe(types.ToolInput) bool { return false }
func (t *fakeTool) CheckPermissions(types.ToolInput, types.ToolContext) types.ToolPermission {
	return types.PermissionAllowed
}

type delayTool struct {
	name            string
	delay           time.Duration
	concurrencySafe bool
	result          string
}

type flakyTool struct {
	name       string
	failCount  int
	called     int
	result     string
	failErrMsg string
}

func (t *flakyTool) Name() string                  { return t.name }
func (t *flakyTool) Description() string           { return "flaky tool" }
func (t *flakyTool) InputSchema() types.ToolSchema { return types.ToolSchema{Type: "object"} }
func (t *flakyTool) Execute(_ context.Context, _ types.ToolInput, _ types.ToolContext) (types.ToolResult, error) {
	t.called++
	if t.called <= t.failCount {
		return types.ToolResult{}, errors.New(t.failErrMsg)
	}
	return types.ToolResult{Content: t.result}, nil
}
func (t *flakyTool) IsReadOnly(types.ToolInput) bool        { return false }
func (t *flakyTool) IsDestructive(types.ToolInput) bool     { return false }
func (t *flakyTool) IsConcurrencySafe(types.ToolInput) bool { return false }
func (t *flakyTool) CheckPermissions(types.ToolInput, types.ToolContext) types.ToolPermission {
	return types.PermissionAllowed
}

func (t *delayTool) Name() string                  { return t.name }
func (t *delayTool) Description() string           { return "delay tool" }
func (t *delayTool) InputSchema() types.ToolSchema { return types.ToolSchema{Type: "object"} }
func (t *delayTool) Execute(_ context.Context, _ types.ToolInput, _ types.ToolContext) (types.ToolResult, error) {
	time.Sleep(t.delay)
	return types.ToolResult{Content: t.result}, nil
}
func (t *delayTool) IsReadOnly(types.ToolInput) bool        { return true }
func (t *delayTool) IsDestructive(types.ToolInput) bool     { return false }
func (t *delayTool) IsConcurrencySafe(types.ToolInput) bool { return t.concurrencySafe }
func (t *delayTool) CheckPermissions(types.ToolInput, types.ToolContext) types.ToolPermission {
	return types.PermissionAllowed
}

type fixedPermissionChecker struct {
	decision permissions.Decision
}

func (c fixedPermissionChecker) Check(string, json.RawMessage) permissions.Decision {
	return c.decision
}

type hookRecorder struct {
	mu     sync.Mutex
	events []hooks.Event
}

func (r *hookRecorder) Fire(_ context.Context, event hooks.Event, _ map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, event)
	return nil
}

func (r *hookRecorder) HasHooks(event hooks.Event) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, ev := range r.events {
		if ev == event {
			return true
		}
	}
	return false
}

func (r *hookRecorder) contains(event hooks.Event) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, ev := range r.events {
		if ev == event {
			return true
		}
	}
	return false
}

func toolUseScript(toolName, toolUseID string, input json.RawMessage) func(chan types.StreamEvent) {
	return func(ch chan types.StreamEvent) {
		ch <- types.StreamEvent{Type: types.StreamToolUseStart, ToolUseID: toolUseID, ToolName: toolName}
		ch <- types.StreamEvent{Type: types.StreamToolUseDelta, Delta: string(input)}
		ch <- types.StreamEvent{Type: types.StreamToolUseDone, ToolUseID: toolUseID, ToolName: toolName}
	}
}

func textScript(text string) func(chan types.StreamEvent) {
	return func(ch chan types.StreamEvent) {
		ch <- types.StreamEvent{Type: types.StreamContentDelta, Delta: text}
		ch <- types.StreamEvent{Type: types.StreamContentDone}
	}
}

func twoToolUseScript(firstName, firstID string, firstInput json.RawMessage, secondName, secondID string, secondInput json.RawMessage) func(chan types.StreamEvent) {
	return func(ch chan types.StreamEvent) {
		ch <- types.StreamEvent{Type: types.StreamToolUseStart, ToolUseID: firstID, ToolName: firstName}
		ch <- types.StreamEvent{Type: types.StreamToolUseDelta, Delta: string(firstInput)}
		ch <- types.StreamEvent{Type: types.StreamToolUseDone, ToolUseID: firstID, ToolName: firstName}
		ch <- types.StreamEvent{Type: types.StreamToolUseStart, ToolUseID: secondID, ToolName: secondName}
		ch <- types.StreamEvent{Type: types.StreamToolUseDelta, Delta: string(secondInput)}
		ch <- types.StreamEvent{Type: types.StreamToolUseDone, ToolUseID: secondID, ToolName: secondName}
	}
}

func anthropicStyleTwoToolScript(firstName, firstID string, firstInput json.RawMessage, secondName, secondID string, secondInput json.RawMessage) func(chan types.StreamEvent) {
	return func(ch chan types.StreamEvent) {
		ch <- types.StreamEvent{Type: types.StreamToolUseStart, ToolUseID: firstID, ToolName: firstName}
		ch <- types.StreamEvent{Type: types.StreamToolUseDelta, ToolUseID: firstID, ToolName: firstName, Delta: string(firstInput)}
		ch <- types.StreamEvent{Type: types.StreamContentDone}
		ch <- types.StreamEvent{Type: types.StreamToolUseStart, ToolUseID: secondID, ToolName: secondName}
		ch <- types.StreamEvent{Type: types.StreamToolUseDelta, ToolUseID: secondID, ToolName: secondName, Delta: string(secondInput)}
		ch <- types.StreamEvent{Type: types.StreamContentDone}
		ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopToolUse}
	}
}

func TestAgentPermissionDenyPath(t *testing.T) {
	tool := &fakeTool{name: "fake", result: "ok"}
	provider := &scriptedProvider{scripts: []func(chan types.StreamEvent){
		toolUseScript("fake", "tool-1", json.RawMessage(`{"x":1}`)),
		textScript("done"),
	}}

	a := New(Config{
		Provider:          provider,
		Tools:             []types.Tool{tool},
		MaxTurns:          4,
		PermissionChecker: fixedPermissionChecker{decision: permissions.DecisionDeny},
	})

	var events []types.AgentEvent
	a.SetEventCallback(func(ev types.AgentEvent) {
		events = append(events, ev)
	})

	if err := a.Run(context.Background(), "hello"); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if tool.executeCount != 0 {
		t.Fatalf("tool executed %d times, want 0", tool.executeCount)
	}

	hasDeniedResult := false
	for _, msg := range a.Messages() {
		for _, block := range msg.Content {
			if block.Type == types.ContentToolResult && block.IsError && strings.Contains(block.Content, "permission denied") {
				hasDeniedResult = true
			}
		}
	}
	if !hasDeniedResult {
		t.Fatalf("expected permission denied tool result in transcript")
	}

	hasPermissionDenyEvent := false
	for _, ev := range events {
		if ev.Type == types.AgentEventPermissionResult && ev.PermissionDecision == types.AgentPermissionDeny {
			hasPermissionDenyEvent = true
			break
		}
	}
	if !hasPermissionDenyEvent {
		t.Fatalf("expected permission deny event")
	}
}

func TestAgentToolAndStopHooksAreInvoked(t *testing.T) {
	tool := &fakeTool{name: "fake", result: "ok"}
	provider := &scriptedProvider{scripts: []func(chan types.StreamEvent){
		toolUseScript("fake", "tool-1", json.RawMessage(`{"x":1}`)),
		textScript("done"),
	}}
	hookRec := &hookRecorder{}

	a := New(Config{
		Provider: provider,
		Tools:    []types.Tool{tool},
		MaxTurns: 4,
		Hooks:    hookRec,
	})

	if err := a.Run(context.Background(), "hello"); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if !hookRec.contains(hooks.EventPreTool) {
		t.Fatalf("expected pre_tool hook invocation")
	}
	if !hookRec.contains(hooks.EventPostTool) {
		t.Fatalf("expected post_tool hook invocation")
	}
	if !hookRec.contains(hooks.EventStop) {
		t.Fatalf("expected stop hook invocation")
	}
}

func TestAgentStopsOnTokenBudget(t *testing.T) {
	provider := &scriptedProvider{}
	hookRec := &hookRecorder{}

	a := New(Config{
		Provider:       provider,
		Tools:          nil,
		MaxTurns:       2,
		MaxTokenBudget: 10,
		Hooks:          hookRec,
	})

	a.mu.Lock()
	a.usage = types.Usage{InputTokens: 10}
	a.mu.Unlock()

	var stopEvent *types.AgentEvent
	a.SetEventCallback(func(ev types.AgentEvent) {
		if ev.Type == types.AgentEventStop {
			cp := ev
			stopEvent = &cp
		}
	})

	err := a.Run(context.Background(), "hello")
	if err == nil || !strings.Contains(err.Error(), "token budget reached") {
		t.Fatalf("expected token budget error, got %v", err)
	}

	if stopEvent == nil {
		t.Fatalf("expected stop event")
	}
	if stopEvent.StopReason != types.AgentStopBudgetToken {
		t.Fatalf("stop reason = %q, want %q", stopEvent.StopReason, types.AgentStopBudgetToken)
	}

	if provider.chatCall != 0 {
		t.Fatalf("provider called %d times, want 0", provider.chatCall)
	}

	if !hookRec.contains(hooks.EventStop) {
		t.Fatalf("expected stop hook invocation")
	}
}

func TestAgentToolTelemetryIsEmitted(t *testing.T) {
	tool := &fakeTool{name: "fake", result: "ok"}
	provider := &scriptedProvider{scripts: []func(chan types.StreamEvent){
		toolUseScript("fake", "tool-1", json.RawMessage(`{"x":1}`)),
		textScript("done"),
	}}

	a := New(Config{Provider: provider, Tools: []types.Tool{tool}, MaxTurns: 4})
	var start, end *types.AgentEvent
	a.SetEventCallback(func(ev types.AgentEvent) {
		if ev.Type == types.AgentEventToolStart {
			cp := ev
			start = &cp
		}
		if ev.Type == types.AgentEventToolEnd {
			cp := ev
			end = &cp
		}
	})

	if err := a.Run(context.Background(), "hello"); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if start == nil || end == nil {
		t.Fatalf("expected tool start/end events")
	}
	if start.ToolTelemetryID == "" {
		t.Fatalf("start telemetry id should be populated")
	}
	if start.ToolTelemetryID != end.ToolTelemetryID {
		t.Fatalf("telemetry IDs differ: %q != %q", start.ToolTelemetryID, end.ToolTelemetryID)
	}
	if start.ToolStartedAtMS == 0 {
		t.Fatalf("start timestamp should be populated")
	}
	if end.ToolEndedAtMS == 0 {
		t.Fatalf("end timestamp should be populated")
	}
	if end.ToolDurationMS < 0 {
		t.Fatalf("duration = %d, want >= 0", end.ToolDurationMS)
	}
}

func TestAgentRunLoopEmitsInputErrorStopReason(t *testing.T) {
	a := New(Config{Provider: &scriptedProvider{}, MaxTurns: 2})
	var stop *types.AgentEvent
	a.SetEventCallback(func(ev types.AgentEvent) {
		if ev.Type == types.AgentEventStop {
			cp := ev
			stop = &cp
		}
	})

	err := a.RunLoop(context.Background(), func() (string, error) {
		return "", fmt.Errorf("input failed")
	})
	if err == nil || !strings.Contains(err.Error(), "input failed") {
		t.Fatalf("expected input error, got %v", err)
	}
	if stop == nil {
		t.Fatalf("expected stop event")
	}
	if stop.StopReason != types.AgentStopInputError {
		t.Fatalf("stop reason = %q, want %q", stop.StopReason, types.AgentStopInputError)
	}
}

func TestAgentExtendsTurnBudgetForToolFollowUpWhenConfiguredMaxTurnsIsOne(t *testing.T) {
	tool := &fakeTool{name: "fake", result: "ok"}
	provider := &scriptedProvider{scripts: []func(chan types.StreamEvent){
		toolUseScript("fake", "tool-1", json.RawMessage(`{"x":1}`)),
		textScript("done"),
	}}

	a := New(Config{Provider: provider, Tools: []types.Tool{tool}, MaxTurns: 1})
	if err := a.Run(context.Background(), "hello"); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if provider.chatCall != 2 {
		t.Fatalf("provider chat calls = %d, want 2", provider.chatCall)
	}
	if tool.executeCount != 1 {
		t.Fatalf("tool execute count = %d, want 1", tool.executeCount)
	}
}

func TestAgentWorkingDirDefaultsToAbsoluteProcessDir(t *testing.T) {
	a := New(Config{})
	workingDir := a.WorkingDir()
	if strings.TrimSpace(workingDir) == "" {
		t.Fatalf("working dir should not be empty")
	}
	if !filepath.IsAbs(workingDir) {
		t.Fatalf("working dir = %q, want absolute path", workingDir)
	}
	runtime := a.RuntimeSnapshot()
	if runtime.Turns != 0 {
		t.Fatalf("expected zero turns on fresh runtime snapshot, got %+v", runtime)
	}
}

func TestAgentConcurrentToolResultsRemainOrdered(t *testing.T) {
	fast := &delayTool{name: "fast", delay: 5 * time.Millisecond, concurrencySafe: true, result: "fast-result"}
	slow := &delayTool{name: "slow", delay: 25 * time.Millisecond, concurrencySafe: true, result: "slow-result"}
	provider := &scriptedProvider{scripts: []func(chan types.StreamEvent){
		twoToolUseScript("slow", "tool-slow", json.RawMessage(`{"x":1}`), "fast", "tool-fast", json.RawMessage(`{"y":2}`)),
		textScript("done"),
	}}

	a := New(Config{Provider: provider, Tools: []types.Tool{slow, fast}, MaxTurns: 4})
	if err := a.Run(context.Background(), "hello"); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	var toolResults []types.ContentBlock
	for _, msg := range a.Messages() {
		for _, block := range msg.Content {
			if block.Type == types.ContentToolResult {
				toolResults = append(toolResults, block)
			}
		}
	}
	if len(toolResults) < 2 {
		t.Fatalf("expected at least 2 tool results, got %d", len(toolResults))
	}
	if got := toolResults[0].ForToolUseID; got != "tool-slow" {
		t.Fatalf("first tool result id = %q, want %q", got, "tool-slow")
	}
	if got := toolResults[1].ForToolUseID; got != "tool-fast" {
		t.Fatalf("second tool result id = %q, want %q", got, "tool-fast")
	}
}

func TestDoChatFlushesToolBoundariesAcrossEventStyles(t *testing.T) {
	provider := &scriptedProvider{scripts: []func(chan types.StreamEvent){
		anthropicStyleTwoToolScript("slow", "tool-slow", json.RawMessage(`{"x":1}`), "fast", "tool-fast", json.RawMessage(`{"y":2}`)),
	}}

	a := New(Config{Provider: provider, MaxTurns: 1})
	msg, stopReason, err := a.doChat(context.Background(), provider, types.ChatRequest{}, 1)
	if err != nil {
		t.Fatalf("doChat returned error: %v", err)
	}
	if stopReason != types.StopToolUse {
		t.Fatalf("stop reason = %q, want %q", stopReason, types.StopToolUse)
	}
	toolUses := msg.GetToolUses()
	if len(toolUses) != 2 {
		t.Fatalf("tool uses = %d, want 2", len(toolUses))
	}
	if toolUses[0].ToolUseID != "tool-slow" || toolUses[1].ToolUseID != "tool-fast" {
		t.Fatalf("tool use ids = [%s %s], want [tool-slow tool-fast]", toolUses[0].ToolUseID, toolUses[1].ToolUseID)
	}
}

func TestAgentStopEventIncludesProviderStopReason(t *testing.T) {
	provider := &scriptedProvider{scripts: []func(chan types.StreamEvent){
		func(ch chan types.StreamEvent) {
			ch <- types.StreamEvent{Type: types.StreamContentDelta, Delta: "done"}
			ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopMaxTokens}
		},
	}}

	a := New(Config{Provider: provider, MaxTurns: 1})
	var stop *types.AgentEvent
	a.SetEventCallback(func(ev types.AgentEvent) {
		if ev.Type == types.AgentEventStop {
			cp := ev
			stop = &cp
		}
	})

	if err := a.Run(context.Background(), "hello"); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if stop == nil {
		t.Fatalf("expected stop event")
	}
	if stop.StopReason != types.AgentStopMaxTokens {
		t.Fatalf("agent stop reason = %q, want %q", stop.StopReason, types.AgentStopMaxTokens)
	}
	if stop.ProviderStopReason != types.StopMaxTokens {
		t.Fatalf("provider stop reason = %q, want %q", stop.ProviderStopReason, types.StopMaxTokens)
	}
}

func TestDoChatAccumulatesCumulativeUsageWithoutDoubleCounting(t *testing.T) {
	provider := &scriptedProvider{scripts: []func(chan types.StreamEvent){
		func(ch chan types.StreamEvent) {
			ch <- types.StreamEvent{Usage: &types.Usage{InputTokens: 10, OutputTokens: 2}, UsageCumulative: true}
			ch <- types.StreamEvent{Usage: &types.Usage{InputTokens: 10, OutputTokens: 5}, UsageCumulative: true}
			ch <- types.StreamEvent{Usage: &types.Usage{InputTokens: 12, OutputTokens: 5}, UsageCumulative: true}
			ch <- types.StreamEvent{Type: types.StreamContentDelta, Delta: "ok"}
			ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopEndTurn}
		},
	}}

	a := New(Config{Provider: provider, MaxTurns: 1})
	if _, _, err := a.doChat(context.Background(), provider, types.ChatRequest{}, 1); err != nil {
		t.Fatalf("doChat returned error: %v", err)
	}
	usage := a.Usage()
	if usage.InputTokens != 12 {
		t.Fatalf("input tokens = %d, want 12", usage.InputTokens)
	}
	if usage.OutputTokens != 5 {
		t.Fatalf("output tokens = %d, want 5", usage.OutputTokens)
	}
}

func TestDoChatAdjustsFirstCumulativeUsageAgainstPriorIncrementalEvents(t *testing.T) {
	provider := &scriptedProvider{scripts: []func(chan types.StreamEvent){
		func(ch chan types.StreamEvent) {
			ch <- types.StreamEvent{Usage: &types.Usage{InputTokens: 2, OutputTokens: 1}}
			ch <- types.StreamEvent{Usage: &types.Usage{InputTokens: 5, OutputTokens: 2}, UsageCumulative: true}
			ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopEndTurn}
		},
	}}

	a := New(Config{Provider: provider, MaxTurns: 1})
	if _, _, err := a.doChat(context.Background(), provider, types.ChatRequest{}, 1); err != nil {
		t.Fatalf("doChat returned error: %v", err)
	}
	usage := a.Usage()
	if usage.InputTokens != 5 {
		t.Fatalf("input tokens = %d, want 5", usage.InputTokens)
	}
	if usage.OutputTokens != 2 {
		t.Fatalf("output tokens = %d, want 2", usage.OutputTokens)
	}
}

func TestAgentRecoversFromProviderTooLongByForcedCompaction(t *testing.T) {
	messages := []types.Message{
		types.NewTextMessage(types.RoleUser, strings.Repeat("u1 ", 9000)),
		types.NewTextMessage(types.RoleAssistant, strings.Repeat("a1 ", 9000)),
		types.NewTextMessage(types.RoleUser, strings.Repeat("u2 ", 9000)),
		types.NewTextMessage(types.RoleAssistant, strings.Repeat("a2 ", 9000)),
		types.NewTextMessage(types.RoleUser, strings.Repeat("u3 ", 9000)),
		types.NewTextMessage(types.RoleAssistant, strings.Repeat("a3 ", 9000)),
	}

	provider := &scriptedProvider{
		errs:     []error{errors.New("openai: context length exceeded")},
		syncResp: &types.ChatResponse{Message: types.NewTextMessage(types.RoleAssistant, "compact summary")},
		scripts: []func(chan types.StreamEvent){
			func(ch chan types.StreamEvent) {
				ch <- types.StreamEvent{Type: types.StreamContentDelta, Delta: "recovered"}
				ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopEndTurn}
			},
		},
	}

	a := New(Config{Provider: provider, MaxTurns: 3})
	a.messages = append(a.messages, messages...)

	if err := a.agentLoop(context.Background()); err != nil {
		t.Fatalf("agentLoop returned error: %v", err)
	}
	if provider.chatCall != 2 {
		t.Fatalf("provider chat calls = %d, want 2", provider.chatCall)
	}
	if len(provider.requests) < 2 {
		t.Fatalf("expected second request after recovery")
	}
	if got := provider.requests[1].Messages[1].GetText(); !strings.Contains(got, "Conversation compacted") {
		t.Fatalf("expected recovery request to include compaction summary marker, got %q", got)
	}
}

func TestAgentContinuesAfterMaxTokensWhenLikelyTruncated(t *testing.T) {
	provider := &scriptedProvider{scripts: []func(chan types.StreamEvent){
		func(ch chan types.StreamEvent) {
			ch <- types.StreamEvent{Type: types.StreamContentDelta, Delta: strings.Repeat("partial ", 12)}
			ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopMaxTokens}
		},
		func(ch chan types.StreamEvent) {
			ch <- types.StreamEvent{Type: types.StreamContentDelta, Delta: "continued"}
			ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopEndTurn}
		},
	}}

	a := New(Config{Provider: provider, MaxTurns: 3})
	var finalStop *types.AgentEvent
	a.SetEventCallback(func(ev types.AgentEvent) {
		if ev.Type == types.AgentEventStop {
			cp := ev
			finalStop = &cp
		}
	})

	if err := a.Run(context.Background(), "hello"); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if provider.chatCall != 2 {
		t.Fatalf("provider chat calls = %d, want 2", provider.chatCall)
	}
	if finalStop == nil || finalStop.ProviderStopReason != types.StopEndTurn {
		t.Fatalf("expected final provider stop reason end_turn, got %+v", finalStop)
	}
	if len(provider.requests) < 2 || len(provider.requests[1].Messages) == 0 {
		t.Fatalf("expected second provider request")
	}
	last := provider.requests[1].Messages[len(provider.requests[1].Messages)-1]
	if last.Role != types.RoleUser || !strings.Contains(strings.ToLower(last.GetText()), "resume") {
		t.Fatalf("expected continuation prompt, got role=%q text=%q", last.Role, last.GetText())
	}
}

func TestAgentRecoversFromPromptTooLongStopReasonByCompactionAndRecoveryMessage(t *testing.T) {
	messages := []types.Message{
		types.NewTextMessage(types.RoleUser, strings.Repeat("u1 ", 9000)),
		types.NewTextMessage(types.RoleAssistant, strings.Repeat("a1 ", 9000)),
		types.NewTextMessage(types.RoleUser, strings.Repeat("u2 ", 9000)),
		types.NewTextMessage(types.RoleAssistant, strings.Repeat("a2 ", 9000)),
		types.NewTextMessage(types.RoleUser, strings.Repeat("u3 ", 9000)),
	}

	provider := &scriptedProvider{
		syncResp: &types.ChatResponse{Message: types.NewTextMessage(types.RoleAssistant, "compact summary")},
		scripts: []func(chan types.StreamEvent){
			func(ch chan types.StreamEvent) {
				ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopContextLimit}
			},
			func(ch chan types.StreamEvent) {
				ch <- types.StreamEvent{Type: types.StreamContentDelta, Delta: "recovered"}
				ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopEndTurn}
			},
		},
	}

	a := New(Config{Provider: provider, MaxTurns: 3})
	a.messages = append(a.messages, messages...)

	if err := a.agentLoop(context.Background()); err != nil {
		t.Fatalf("agentLoop returned error: %v", err)
	}
	if provider.chatCall != 2 {
		t.Fatalf("provider chat calls = %d, want 2", provider.chatCall)
	}
	if len(provider.requests) < 2 || len(provider.requests[1].Messages) == 0 {
		t.Fatalf("expected retry request after stop_reason recovery")
	}
	last := provider.requests[1].Messages[len(provider.requests[1].Messages)-1]
	if last.Role != types.RoleUser || !strings.Contains(strings.ToLower(last.GetText()), "compacted") {
		t.Fatalf("expected prompt-too-long recovery message, got role=%q text=%q", last.Role, last.GetText())
	}
}

func TestDoChatMixedIncrementalAndCumulativeUsageDoesNotDoubleCount(t *testing.T) {
	provider := &scriptedProvider{scripts: []func(chan types.StreamEvent){
		func(ch chan types.StreamEvent) {
			ch <- types.StreamEvent{Usage: &types.Usage{InputTokens: 2, OutputTokens: 1}}
			ch <- types.StreamEvent{Usage: &types.Usage{InputTokens: 5, OutputTokens: 3}, UsageCumulative: true}
			ch <- types.StreamEvent{Usage: &types.Usage{InputTokens: 1, OutputTokens: 1}}
			ch <- types.StreamEvent{Usage: &types.Usage{InputTokens: 8, OutputTokens: 6}, UsageCumulative: true}
			ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopEndTurn}
		},
	}}

	a := New(Config{Provider: provider, MaxTurns: 1})
	if _, _, err := a.doChat(context.Background(), provider, types.ChatRequest{}, 1); err != nil {
		t.Fatalf("doChat returned error: %v", err)
	}
	usage := a.Usage()
	if usage.InputTokens != 8 {
		t.Fatalf("input tokens = %d, want 8", usage.InputTokens)
	}
	if usage.OutputTokens != 6 {
		t.Fatalf("output tokens = %d, want 6", usage.OutputTokens)
	}
}

func TestMapProviderStopReasonExpandedTaxonomy(t *testing.T) {
	tests := []struct {
		name string
		in   types.StopReason
		want types.AgentStopReason
	}{
		{name: "context_limit", in: types.StopContextLimit, want: types.AgentStopContextLimit},
		{name: "content_filter", in: types.StopContentFilter, want: types.AgentStopContentFilter},
		{name: "refusal", in: types.StopRefusal, want: types.AgentStopRefusal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mapProviderStopReason(tt.in); got != tt.want {
				t.Fatalf("mapProviderStopReason(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestAgentSetProviderModelSwitchesNextTurnProvider(t *testing.T) {
	openaiProvider := &scriptedProvider{name: "openai", scripts: []func(chan types.StreamEvent){
		func(ch chan types.StreamEvent) {
			ch <- types.StreamEvent{Type: types.StreamContentDelta, Delta: "openai"}
			ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopEndTurn}
		},
	}}
	anthropicProvider := &scriptedProvider{name: "anthropic", scripts: []func(chan types.StreamEvent){
		func(ch chan types.StreamEvent) {
			ch <- types.StreamEvent{Type: types.StreamContentDelta, Delta: "anthropic"}
			ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopEndTurn}
		},
	}}

	a := New(Config{
		Provider:     openaiProvider,
		ProviderName: "openai",
		Model:        "gpt-4o-mini",
		MaxTurns:     2,
		ResolveProvider: func(name string) (types.Provider, error) {
			switch strings.ToLower(strings.TrimSpace(name)) {
			case "openai":
				return openaiProvider, nil
			case "anthropic":
				return anthropicProvider, nil
			default:
				return nil, fmt.Errorf("unknown provider %q", name)
			}
		},
	})

	if err := a.Run(context.Background(), "first"); err != nil {
		t.Fatalf("first run returned error: %v", err)
	}
	if openaiProvider.chatCall != 1 {
		t.Fatalf("openai chat calls = %d, want 1", openaiProvider.chatCall)
	}

	if err := a.SetProviderModel("anthropic", "claude-opus-4-20250514"); err != nil {
		t.Fatalf("SetProviderModel returned error: %v", err)
	}

	if err := a.Run(context.Background(), "second"); err != nil {
		t.Fatalf("second run returned error: %v", err)
	}
	if anthropicProvider.chatCall != 1 {
		t.Fatalf("anthropic chat calls = %d, want 1", anthropicProvider.chatCall)
	}
	if got := anthropicProvider.requests[0].Model; got != "claude-opus-4-20250514" {
		t.Fatalf("anthropic request model = %q, want %q", got, "claude-opus-4-20250514")
	}
}

func TestAgentRuntimeSnapshotTracksProviderModelRefAcrossSwitches(t *testing.T) {
	openaiProvider := &scriptedProvider{name: "openai", scripts: []func(chan types.StreamEvent){
		func(ch chan types.StreamEvent) {
			ch <- types.StreamEvent{Type: types.StreamContentDelta, Delta: "openai"}
			ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopEndTurn}
		},
	}}
	anthropicProvider := &scriptedProvider{name: "anthropic", scripts: []func(chan types.StreamEvent){
		func(ch chan types.StreamEvent) {
			ch <- types.StreamEvent{Type: types.StreamContentDelta, Delta: "anthropic"}
			ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopEndTurn}
		},
	}}
	a := New(Config{
		Provider:     openaiProvider,
		ProviderName: "openai",
		Model:        "gpt-4o-mini",
		ResolveProvider: func(name string) (types.Provider, error) {
			switch strings.ToLower(strings.TrimSpace(name)) {
			case "openai":
				return openaiProvider, nil
			case "anthropic":
				return anthropicProvider, nil
			default:
				return nil, fmt.Errorf("unknown provider %q", name)
			}
		},
	})

	before := a.RuntimeSnapshot()
	if before.ProviderName != "openai" || before.Model != "gpt-4o-mini" || before.ModelRef != "openai/gpt-4o-mini" {
		t.Fatalf("unexpected initial runtime snapshot: %+v", before)
	}
	if err := a.SetProviderModel("anthropic", "claude-opus-4-20250514"); err != nil {
		t.Fatalf("SetProviderModel failed: %v", err)
	}
	after := a.RuntimeSnapshot()
	if after.ProviderName != "anthropic" || after.Model != "claude-opus-4-20250514" || after.ModelRef != "anthropic/claude-opus-4-20250514" {
		t.Fatalf("unexpected switched runtime snapshot: %+v", after)
	}
	if err := a.Run(context.Background(), "check"); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	post := a.RuntimeSnapshot()
	if post.ProviderName != "anthropic" || post.ModelRef != "anthropic/claude-opus-4-20250514" {
		t.Fatalf("runtime snapshot drifted after submit: %+v", post)
	}
}

func TestAgentStopReasonMaxTokensRecoveryExhausted(t *testing.T) {
	provider := &scriptedProvider{scripts: []func(chan types.StreamEvent){
		func(ch chan types.StreamEvent) {
			ch <- types.StreamEvent{Type: types.StreamContentDelta, Delta: strings.Repeat("partial ", 12)}
			ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopMaxTokens}
		},
		func(ch chan types.StreamEvent) {
			ch <- types.StreamEvent{Type: types.StreamContentDelta, Delta: strings.Repeat("partial ", 12)}
			ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopMaxTokens}
		},
		func(ch chan types.StreamEvent) {
			ch <- types.StreamEvent{Type: types.StreamContentDelta, Delta: strings.Repeat("partial ", 12)}
			ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopMaxTokens}
		},
	}}

	a := New(Config{Provider: provider, MaxTurns: 5})
	var stop *types.AgentEvent
	a.SetEventCallback(func(ev types.AgentEvent) {
		if ev.Type == types.AgentEventStop {
			cp := ev
			stop = &cp
		}
	})

	if err := a.Run(context.Background(), "hello"); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if provider.chatCall != 3 {
		t.Fatalf("provider chat calls = %d, want 3", provider.chatCall)
	}
	if stop == nil {
		t.Fatalf("expected stop event")
	}
	if stop.StopReason != types.AgentStopMaxTokensExhausted {
		t.Fatalf("stop reason = %q, want %q", stop.StopReason, types.AgentStopMaxTokensExhausted)
	}
	if stop.ProviderStopReason != types.StopMaxTokens {
		t.Fatalf("provider stop reason = %q, want %q", stop.ProviderStopReason, types.StopMaxTokens)
	}
	if !strings.Contains(stop.Details, "continuation_recoveries=2") {
		t.Fatalf("details = %q, want continuation recoveries", stop.Details)
	}
}

func TestAgentToolQueueLifecycleRuntimeCounters(t *testing.T) {
	tool := &fakeTool{name: "fake", result: "ok"}
	provider := &scriptedProvider{scripts: []func(chan types.StreamEvent){
		anthropicStyleTwoToolScript("fake", "tool-1", json.RawMessage(`{"x":1}`), "fake", "tool-2", json.RawMessage(`{"y":2}`)),
		textScript("done"),
	}}
	a := New(Config{Provider: provider, Tools: []types.Tool{tool}, MaxTurns: 4})

	var queued []types.AgentEvent
	var starts []types.AgentEvent
	var ends []types.AgentEvent
	a.SetEventCallback(func(ev types.AgentEvent) {
		switch ev.Type {
		case types.AgentEventToolQueued:
			queued = append(queued, ev)
		case types.AgentEventToolStart:
			starts = append(starts, ev)
		case types.AgentEventToolEnd:
			ends = append(ends, ev)
		}
	})

	if err := a.Run(context.Background(), "hello"); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(queued) != 2 || len(starts) != 2 || len(ends) != 2 {
		t.Fatalf("expected 2 queued/start/end events, got queued=%d start=%d end=%d", len(queued), len(starts), len(ends))
	}
	if queued[0].ToolQueueIndex != 1 || queued[0].ToolQueueTotal != 2 || queued[1].ToolQueueIndex != 2 || queued[1].ToolQueueTotal != 2 {
		t.Fatalf("unexpected queue indexes: %+v %+v", queued[0], queued[1])
	}
	if queued[0].Runtime.ToolInflight != 0 || queued[1].Runtime.ToolInflight != 0 {
		t.Fatalf("queued runtime inflight should be 0, got %d and %d", queued[0].Runtime.ToolInflight, queued[1].Runtime.ToolInflight)
	}
	if starts[0].Runtime.ToolInflight != 1 || starts[1].Runtime.ToolInflight != 1 {
		t.Fatalf("tool start runtime inflight should be 1, got %d and %d", starts[0].Runtime.ToolInflight, starts[1].Runtime.ToolInflight)
	}
	if ends[0].Runtime.ToolInflight != 0 || ends[1].Runtime.ToolInflight != 0 {
		t.Fatalf("tool end runtime inflight should be 0, got %d and %d", ends[0].Runtime.ToolInflight, ends[1].Runtime.ToolInflight)
	}
}

func TestAgentEmitsLifecycleChunksForTeamAndMessageTools(t *testing.T) {
	provider := &scriptedProvider{scripts: []func(chan types.StreamEvent){
		toolUseScript("team_create", "tool-1", json.RawMessage(`{"team_name":"alpha"}`)),
		toolUseScript("send_message", "tool-2", json.RawMessage(`{"to":"*","message":"hi"}`)),
		textScript("done"),
	}}

	a := New(Config{Provider: provider, Tools: []types.Tool{&TeamCreateStubTool{}, &SendMessageStubTool{}}, MaxTurns: 4})

	var chunks []string
	a.SetEventCallback(func(ev types.AgentEvent) {
		if ev.Type == types.AgentEventAssistantChunk && strings.HasPrefix(ev.AssistantChunk, "<lifecycle ") {
			chunks = append(chunks, ev.AssistantChunk)
		}
	})

	if err := a.Run(context.Background(), "hello"); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if len(chunks) < 4 {
		t.Fatalf("expected lifecycle start/end chunks, got %v", chunks)
	}
	joined := strings.Join(chunks, "\n")
	if !strings.Contains(joined, `kind="team" action="start" tool="team_create"`) {
		t.Fatalf("missing team lifecycle start chunk: %s", joined)
	}
	if !strings.Contains(joined, `kind="message" action="end" tool="send_message"`) {
		t.Fatalf("missing message lifecycle end chunk: %s", joined)
	}
}

func TestAgentEmitsTurnAndCompactionLifecycleEvents(t *testing.T) {
	provider := &scriptedProvider{scripts: []func(chan types.StreamEvent){
		func(ch chan types.StreamEvent) {
			ch <- types.StreamEvent{Type: types.StreamContentDelta, Delta: "done"}
			ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopEndTurn}
		},
	}}

	a := New(Config{Provider: provider, MaxTurns: 2, ContextWindow: 64})
	a.messages = append(a.messages,
		types.NewTextMessage(types.RoleUser, strings.Repeat("hello ", 80)),
		types.NewTextMessage(types.RoleAssistant, strings.Repeat("world ", 80)),
		types.NewTextMessage(types.RoleUser, strings.Repeat("again ", 80)),
		types.NewTextMessage(types.RoleAssistant, strings.Repeat("done ", 80)),
		types.NewTextMessage(types.RoleUser, strings.Repeat("tail ", 80)),
		types.NewTextMessage(types.RoleAssistant, strings.Repeat("tail2 ", 80)),
	)

	var sawTurnStart bool
	var sawTurnEnd bool
	var sawCompactionStart bool
	var sawCompactionEnd bool
	a.SetEventCallback(func(ev types.AgentEvent) {
		switch ev.Type {
		case types.AgentEventTurnStart:
			sawTurnStart = true
		case types.AgentEventTurnEnd:
			sawTurnEnd = true
		case types.AgentEventCompactionStart:
			sawCompactionStart = true
		case types.AgentEventCompactionEnd:
			sawCompactionEnd = true
		}
	})

	if err := a.Run(context.Background(), "hello"); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !sawTurnStart || !sawTurnEnd {
		t.Fatalf("expected turn lifecycle events, got start=%v end=%v", sawTurnStart, sawTurnEnd)
	}
	if !sawCompactionStart || !sawCompactionEnd {
		t.Fatalf("expected compaction lifecycle events, got start=%v end=%v", sawCompactionStart, sawCompactionEnd)
	}
}

func TestAgentRetriesRetryableToolError(t *testing.T) {
	tool := &flakyTool{name: "flaky", failCount: 1, result: "ok", failErrMsg: "temporary timeout"}
	provider := &scriptedProvider{scripts: []func(chan types.StreamEvent){
		toolUseScript("flaky", "tool-1", json.RawMessage(`{"x":1}`)),
		textScript("done"),
	}}
	a := New(Config{Provider: provider, Tools: []types.Tool{tool}, MaxTurns: 4})

	var retryEvent *types.AgentEvent
	a.SetEventCallback(func(ev types.AgentEvent) {
		if ev.Type == types.AgentEventToolRetry {
			cp := ev
			retryEvent = &cp
		}
	})

	if err := a.Run(context.Background(), "hello"); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if tool.called != 2 {
		t.Fatalf("tool called %d times, want 2", tool.called)
	}
	if retryEvent == nil {
		t.Fatalf("expected tool retry event")
	}
	if retryEvent.ToolAttempt != 2 || retryEvent.ToolMaxAttempts != toolMaxAttempts {
		t.Fatalf("unexpected retry metadata: %+v", *retryEvent)
	}
}

type TeamCreateStubTool struct{}

func (t *TeamCreateStubTool) Name() string                  { return "team_create" }
func (t *TeamCreateStubTool) Description() string           { return "stub" }
func (t *TeamCreateStubTool) InputSchema() types.ToolSchema { return types.ToolSchema{Type: "object"} }
func (t *TeamCreateStubTool) Execute(context.Context, types.ToolInput, types.ToolContext) (types.ToolResult, error) {
	return types.ToolResult{Content: `{"ok":true}`}, nil
}
func (t *TeamCreateStubTool) IsReadOnly(types.ToolInput) bool        { return false }
func (t *TeamCreateStubTool) IsDestructive(types.ToolInput) bool     { return false }
func (t *TeamCreateStubTool) IsConcurrencySafe(types.ToolInput) bool { return true }
func (t *TeamCreateStubTool) CheckPermissions(types.ToolInput, types.ToolContext) types.ToolPermission {
	return types.PermissionAllowed
}

type SendMessageStubTool struct{}

func (t *SendMessageStubTool) Name() string                  { return "send_message" }
func (t *SendMessageStubTool) Description() string           { return "stub" }
func (t *SendMessageStubTool) InputSchema() types.ToolSchema { return types.ToolSchema{Type: "object"} }
func (t *SendMessageStubTool) Execute(context.Context, types.ToolInput, types.ToolContext) (types.ToolResult, error) {
	return types.ToolResult{Content: `{"ok":true}`}, nil
}
func (t *SendMessageStubTool) IsReadOnly(types.ToolInput) bool        { return false }
func (t *SendMessageStubTool) IsDestructive(types.ToolInput) bool     { return false }
func (t *SendMessageStubTool) IsConcurrencySafe(types.ToolInput) bool { return true }
func (t *SendMessageStubTool) CheckPermissions(types.ToolInput, types.ToolContext) types.ToolPermission {
	return types.PermissionAllowed
}
