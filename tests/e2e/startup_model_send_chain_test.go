package e2e_test

import (
	"context"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
)

func TestE2EStartupModelSendChainOpenAICrossProviderRelogin(t *testing.T) {
	state := &commands.RuntimeState{ProviderName: "openai", Model: "gpt-4o-mini", ModelRef: "openai/gpt-4o-mini"}
	r := commands.DefaultRegistry()
	ctx := commands.Context{State: state}

	for _, step := range []string{"/login provider openai", "/model anthropic/sonnet", "/login provider anthropic", "/statusline status"} {
		if _, err := r.Dispatch(context.Background(), ctx, step); err != nil {
			t.Fatalf("dispatch %q failed: %v", step, err)
		}
	}

	if got := state.ProviderName; got != "anthropic" {
		t.Fatalf("ProviderName = %q, want anthropic", got)
	}
	if got := state.ModelRef; got != "anthropic/claude-sonnet-4-20250514" {
		t.Fatalf("ModelRef = %q, want anthropic/claude-sonnet-4-20250514", got)
	}
	if !state.LoggedIn || !state.ProviderReady || state.AuthProvider != "anthropic" {
		t.Fatalf("expected relogin to restore ready auth state: %+v", state)
	}
}

func TestE2EStartupModelSendChainOllamaToOpenAIGateThenRecover(t *testing.T) {
	state := &commands.RuntimeState{ProviderName: "ollama", Model: "llama3", ModelRef: "ollama/llama3"}
	r := commands.DefaultRegistry()
	ctx := commands.Context{State: state}

	if _, err := r.Dispatch(context.Background(), ctx, "/model openai/gpt-4o"); err != nil {
		t.Fatalf("dispatch /model openai/gpt-4o failed: %v", err)
	}
	if state.ProviderReady || state.LoggedIn {
		t.Fatalf("expected openai tuple switch from ollama to require login: %+v", state)
	}

	res, err := r.Dispatch(context.Background(), ctx, "/statusline status")
	if err != nil {
		t.Fatalf("dispatch /statusline status failed: %v", err)
	}
	for _, want := range []string{"provider=openai", "model=gpt-4o", "provider_ready=false", "logged_in=false"} {
		if !strings.Contains(res.Message, want) {
			t.Fatalf("statusline missing %q: %q", want, res.Message)
		}
	}

	if _, err := r.Dispatch(context.Background(), ctx, "/login provider openai"); err != nil {
		t.Fatalf("dispatch /login provider openai failed: %v", err)
	}
	if !state.ProviderReady || !state.LoggedIn || state.AuthProvider != "openai" {
		t.Fatalf("expected openai login recovery after tuple switch: %+v", state)
	}
}
