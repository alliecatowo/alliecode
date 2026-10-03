package commands

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/agent"
	"github.com/alliecatowo/alliecode/internal/types"
)

func TestProviderSetReconcilesModelAndUpdatesAgentRuntime(t *testing.T) {
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
	cmd := NewProviderCommand()

	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "provider", Args: []string{"set", "anthropic"}})
	if err != nil {
		t.Fatalf("provider set failed: %v", err)
	}
	if !res.Handled {
		t.Fatalf("expected handled result")
	}
	if state.ProviderName != "anthropic" {
		t.Fatalf("ProviderName = %q, want anthropic", state.ProviderName)
	}
	if state.Model == "gpt-4o-mini" || state.ModelRef == "openai/gpt-4o-mini" {
		t.Fatalf("expected provider set to reconcile model, got model=%q ref=%q", state.Model, state.ModelRef)
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

func TestProviderSetRollbackKeepsStateWhenProviderSwitchFails(t *testing.T) {
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
	cmd := NewProviderCommand()

	_, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "provider", Args: []string{"set", "anthropic"}})
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

func TestProviderSetToOllamaKeepsCanonicalRefAndReadyState(t *testing.T) {
	state := &RuntimeState{
		ProviderName: "openai",
		Model:        "gpt-4o-mini",
		ModelRef:     "openai/gpt-4o-mini",
		LoggedIn:     true,
		AuthProvider: "openai",
	}
	cmd := NewProviderCommand()

	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "provider", Args: []string{"set", "ollama"}})
	if err != nil {
		t.Fatalf("provider set failed: %v", err)
	}
	if !res.Handled {
		t.Fatalf("expected handled result")
	}
	if state.ProviderName != "ollama" {
		t.Fatalf("ProviderName = %q, want ollama", state.ProviderName)
	}
	if state.ModelRef == "" || !strings.HasPrefix(state.ModelRef, "ollama/") {
		t.Fatalf("ModelRef = %q, want ollama/*", state.ModelRef)
	}
	if !state.LoggedIn {
		t.Fatalf("expected login flag to remain set for existing session")
	}
	if state.AuthProvider != "openai" {
		t.Fatalf("expected existing auth provider to remain unchanged, got %q", state.AuthProvider)
	}
	if !state.ProviderReady {
		t.Fatalf("expected ollama provider ready without login")
	}
}
