package commands

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/agent"
	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStatusCommandPrefersAgentRuntimeProviderModel(t *testing.T) {
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
	if err := ag.SetProviderModel("anthropic", "claude-opus-4-20250514"); err != nil {
		t.Fatalf("SetProviderModel failed: %v", err)
	}
	state := &RuntimeState{ProviderName: "openai", Model: "gpt-4o-mini", ModelRef: "openai/gpt-4o-mini", Agent: ag}

	res, err := NewStatusCommand().Execute(context.Background(), Context{State: state}, Invocation{Name: "status"})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !strings.Contains(res.Message, "provider=anthropic") {
		t.Fatalf("status should report runtime provider, got %q", res.Message)
	}
	if !strings.Contains(res.Message, "model=claude-opus-4-20250514") {
		t.Fatalf("status should report runtime model, got %q", res.Message)
	}
}

func TestStatuslineCommandReportsRuntimeProviderModelTuple(t *testing.T) {
	openaiProvider := &modelRuntimeProvider{name: "openai"}
	ag := agent.New(agent.Config{Provider: openaiProvider, ProviderName: "openai", Model: "gpt-4o-mini"})
	state := &RuntimeState{ProviderName: "openai", Model: "gpt-4o-mini", ModelRef: "openai/gpt-4o-mini", Agent: ag}

	if err := ag.SetProviderModel("openai", "gpt-4o"); err != nil {
		t.Fatalf("SetProviderModel failed: %v", err)
	}

	res, err := NewStatuslineCommand().Execute(context.Background(), Context{State: state}, Invocation{Name: "statusline", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("statusline status failed: %v", err)
	}
	for _, want := range []string{"provider=openai", "model=gpt-4o", "model_ref=openai/gpt-4o"} {
		if !strings.Contains(res.Message, want) {
			t.Fatalf("statusline output missing %q: %q", want, res.Message)
		}
	}
}
