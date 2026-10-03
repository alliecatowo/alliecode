package e2e_test

import (
	"context"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
)

func TestE2EProviderModelSubmitStressFlow(t *testing.T) {
	state := &commands.RuntimeState{ProviderName: "openai", Model: "gpt-4o-mini", ModelRef: "openai/gpt-4o-mini"}
	r := commands.DefaultRegistry()
	ctx := commands.Context{State: state}

	steps := []string{
		"/login provider openai",
		"/model gpt-4o",
		"/model anthropic/sonnet",
		"/provider set anthropic",
		"/login provider anthropic",
		"/model opus",
		"/status",
		"/statusline status",
	}

	var statusMsg, statuslineMsg string
	for _, step := range steps {
		res, err := r.Dispatch(context.Background(), ctx, step)
		if err != nil {
			t.Fatalf("dispatch %q failed: %v", step, err)
		}
		if step == "/status" {
			statusMsg = res.Message
		}
		if step == "/statusline status" {
			statuslineMsg = res.Message
		}
	}

	for _, want := range []string{"provider=anthropic", "model=claude-opus-4-20250514", "logged_in=true", "provider_ready=true"} {
		if !strings.Contains(statusMsg, want) {
			t.Fatalf("status missing %q: %q", want, statusMsg)
		}
		if !strings.Contains(statuslineMsg, want) {
			t.Fatalf("statusline missing %q: %q", want, statuslineMsg)
		}
	}
}

func TestE2EProviderStatusIncludesAuthAndModelRecoveryHints(t *testing.T) {
	state := &commands.RuntimeState{ProviderName: "anthropic", Model: "claude-sonnet-4-20250514", ModelRef: "anthropic/claude-sonnet-4-20250514", LoggedIn: false}
	r := commands.DefaultRegistry()
	ctx := commands.Context{State: state}

	res, err := r.Dispatch(context.Background(), ctx, "/provider status")
	if err != nil {
		t.Fatalf("dispatch provider status failed: %v", err)
	}
	if !strings.Contains(res.Message, "quick_fix_auth=") || !strings.Contains(res.Message, "quick_fix_model=") {
		t.Fatalf("provider status missing recovery hints: %q", res.Message)
	}
}
