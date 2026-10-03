package commands

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/agent"
	"github.com/alliecatowo/alliecode/internal/types"
)

type modelRuntimeProvider struct {
	name      string
	chatCalls int
}

func (p *modelRuntimeProvider) Name() string { return p.name }

func (p *modelRuntimeProvider) Chat(_ context.Context, _ types.ChatRequest) (<-chan types.StreamEvent, error) {
	p.chatCalls++
	ch := make(chan types.StreamEvent, 2)
	ch <- types.StreamEvent{Type: types.StreamContentDelta, Delta: "ok"}
	ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopEndTurn}
	close(ch)
	return ch, nil
}

func (p *modelRuntimeProvider) ChatSync(context.Context, types.ChatRequest) (*types.ChatResponse, error) {
	return nil, errors.New("not implemented")
}

func (p *modelRuntimeProvider) ListModels(context.Context) ([]types.Model, error) { return nil, nil }
func (p *modelRuntimeProvider) SupportsStreaming() bool                           { return true }
func (p *modelRuntimeProvider) SupportsTools() bool                               { return false }
func (p *modelRuntimeProvider) SupportsThinking() bool                            { return false }

func TestModelCommandStoresCanonicalProviderModelRef(t *testing.T) {
	cmd := NewModelCommand()
	state := &RuntimeState{ProviderName: "anthropic"}

	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "model", Args: []string{"anthropic/CLAUDE-OPUS-4-20250514"}})
	if err != nil {
		t.Fatalf("set failed: %v", err)
	}
	if !res.Handled {
		t.Fatalf("expected handled result")
	}
	if got := state.ProviderName; got != "anthropic" {
		t.Fatalf("provider = %q, want anthropic", got)
	}
	if got := state.Model; got != "claude-opus-4-20250514" {
		t.Fatalf("model = %q, want canonical claude-opus-4-20250514", got)
	}
	if got := state.ModelRef; got != "anthropic/claude-opus-4-20250514" {
		t.Fatalf("model ref = %q, want anthropic/claude-opus-4-20250514", got)
	}
}

func TestModelCommandSwitchesAgentRuntimeProviderForNextTurn(t *testing.T) {
	openaiProvider := &modelRuntimeProvider{name: "openai"}
	anthropicProvider := &modelRuntimeProvider{name: "anthropic"}
	ag := agent.New(agent.Config{
		Provider:     openaiProvider,
		ProviderName: "openai",
		Model:        "gpt-4o-mini",
		MaxTurns:     1,
		ResolveProvider: func(name string) (types.Provider, error) {
			switch strings.ToLower(strings.TrimSpace(name)) {
			case "openai":
				return openaiProvider, nil
			case "anthropic":
				return anthropicProvider, nil
			default:
				return nil, errors.New("unknown provider")
			}
		},
	})

	state := &RuntimeState{ProviderName: "openai", Model: "gpt-4o-mini", Agent: ag}
	cmd := NewModelCommand()

	if _, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "model", Args: []string{"anthropic/claude-opus-4-20250514"}}); err != nil {
		t.Fatalf("model set failed: %v", err)
	}

	if err := ag.Run(context.Background(), "hello"); err != nil {
		t.Fatalf("agent run failed: %v", err)
	}
	if anthropicProvider.chatCalls != 1 {
		t.Fatalf("anthropic provider calls = %d, want 1", anthropicProvider.chatCalls)
	}
	if openaiProvider.chatCalls != 0 {
		t.Fatalf("openai provider calls = %d, want 0", openaiProvider.chatCalls)
	}
}

func TestModelCommandRollbackKeepsStateWhenProviderSwitchFails(t *testing.T) {
	openaiProvider := &modelRuntimeProvider{name: "openai"}
	ag := agent.New(agent.Config{
		Provider:     openaiProvider,
		ProviderName: "openai",
		Model:        "gpt-4o-mini",
		ResolveProvider: func(name string) (types.Provider, error) {
			return nil, errors.New("provider unavailable")
		},
	})
	state := &RuntimeState{
		ProviderName:  "openai",
		Model:         "gpt-4o-mini",
		ModelRef:      "openai/gpt-4o-mini",
		LoggedIn:      true,
		AuthProvider:  "openai",
		AuthAccount:   "dev@acme",
		ProviderReady: true,
		Agent:         ag,
	}
	cmd := NewModelCommand()

	_, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "model", Args: []string{"anthropic/claude-opus-4-20250514"}})
	if err == nil {
		t.Fatalf("expected provider switch error")
	}
	if state.ProviderName != "openai" || state.Model != "gpt-4o-mini" || state.ModelRef != "openai/gpt-4o-mini" {
		t.Fatalf("selection changed after rollback: %+v", state)
	}
	if !state.LoggedIn || state.AuthProvider != "openai" || state.AuthAccount != "dev@acme" || !state.ProviderReady {
		t.Fatalf("auth state changed after rollback: %+v", state)
	}
}

func TestLoginProviderUsesTransactionalProviderModelApplyPath(t *testing.T) {
	openaiProvider := &modelRuntimeProvider{name: "openai"}
	anthropicProvider := &modelRuntimeProvider{name: "anthropic"}
	ag := agent.New(agent.Config{
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
				return nil, errors.New("unknown provider")
			}
		},
	})
	state := &RuntimeState{ProviderName: "openai", Model: "gpt-4o-mini", ModelRef: "openai/gpt-4o-mini", Agent: ag}
	cmd := NewLoginCommand()

	if _, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "login", Args: []string{"provider", "anthropic"}}); err != nil {
		t.Fatalf("login provider failed: %v", err)
	}
	if got := state.ProviderName; got != "anthropic" {
		t.Fatalf("provider = %q, want anthropic", got)
	}
	if got := state.ModelRef; !strings.HasPrefix(got, "anthropic/") {
		t.Fatalf("model ref = %q, want anthropic/*", got)
	}
	if !state.LoggedIn || state.AuthProvider != "anthropic" || !state.ProviderReady {
		t.Fatalf("expected authenticated anthropic runtime, got %+v", state)
	}

	if err := ag.Run(context.Background(), "hello"); err != nil {
		t.Fatalf("agent run failed: %v", err)
	}
	if anthropicProvider.chatCalls != 1 {
		t.Fatalf("anthropic provider calls = %d, want 1", anthropicProvider.chatCalls)
	}
	if openaiProvider.chatCalls != 0 {
		t.Fatalf("openai provider calls = %d, want 0", openaiProvider.chatCalls)
	}
}

func TestModelCommandAliasSelectionKeepsRuntimeAndAuthAligned(t *testing.T) {
	state := &RuntimeState{
		ProviderName: "anthropic",
		Model:        "claude-sonnet-4-20250514",
		ModelRef:     "anthropic/claude-sonnet-4-20250514",
		LoggedIn:     true,
		AuthProvider: "anthropic",
	}
	cmd := NewModelCommand()

	if _, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "model", Args: []string{"opus"}}); err != nil {
		t.Fatalf("model alias set failed: %v", err)
	}
	if got := state.ProviderName; got != "anthropic" {
		t.Fatalf("provider = %q, want anthropic", got)
	}
	if got := state.Model; got != "claude-opus-4-20250514" {
		t.Fatalf("model = %q, want canonical opus", got)
	}
	if got := state.ModelRef; got != "anthropic/claude-opus-4-20250514" {
		t.Fatalf("model ref = %q, want anthropic/claude-opus-4-20250514", got)
	}
	if !state.LoggedIn || state.AuthProvider != "anthropic" || !state.ProviderReady {
		t.Fatalf("expected auth/readiness preserved after alias selection, got %+v", state)
	}
}
