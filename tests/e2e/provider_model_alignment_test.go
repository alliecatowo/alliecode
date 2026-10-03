package e2e_test

import (
	"context"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
)

func TestE2EProviderModelAlignmentAcrossLoginSwitchLogout(t *testing.T) {
	state := &commands.RuntimeState{ProviderName: "openai", Model: "gpt-4o-mini", ModelRef: "openai/gpt-4o-mini"}
	r := commands.DefaultRegistry()
	ctx := commands.Context{State: state}

	for _, cmd := range []string{"/login provider openai", "/model anthropic/sonnet", "/logout", "/status"} {
		if _, err := r.Dispatch(context.Background(), ctx, cmd); err != nil {
			t.Fatalf("dispatch %q failed: %v", cmd, err)
		}
	}

	if got := state.ProviderName; got != "anthropic" {
		t.Fatalf("ProviderName = %q, want anthropic", got)
	}
	if got := state.ModelRef; got != "anthropic/claude-sonnet-4-20250514" {
		t.Fatalf("ModelRef = %q, want anthropic/claude-sonnet-4-20250514", got)
	}
	if state.LoggedIn || state.ProviderReady {
		t.Fatalf("expected logged out/unready state after provider switch + logout: %+v", state)
	}

	res, err := r.Dispatch(context.Background(), ctx, "/statusline status")
	if err != nil {
		t.Fatalf("/statusline status failed: %v", err)
	}
	if !strings.Contains(res.Message, "provider=anthropic") || !strings.Contains(res.Message, "model=claude-sonnet-4-20250514") {
		t.Fatalf("statusline drifted from runtime state: %q", res.Message)
	}
}

func TestE2EProviderDoctorAndModelDoctorExposeRecoveryHints(t *testing.T) {
	state := &commands.RuntimeState{ProviderName: "anthropic", Model: "claude-sonnet-4-20250514", ModelRef: "anthropic/claude-sonnet-4-20250514", LoggedIn: false}
	r := commands.DefaultRegistry()
	ctx := commands.Context{State: state}

	providerDoctor, err := r.Dispatch(context.Background(), ctx, "/provider doctor")
	if err != nil {
		t.Fatalf("dispatch provider doctor failed: %v", err)
	}
	if !strings.Contains(providerDoctor.Message, "quick_fix=") {
		t.Fatalf("provider doctor missing quick fix: %q", providerDoctor.Message)
	}

	modelDoctor, err := r.Dispatch(context.Background(), ctx, "/model doctor")
	if err != nil {
		t.Fatalf("dispatch model doctor failed: %v", err)
	}
	if !strings.Contains(modelDoctor.Message, "quick_fix=") {
		t.Fatalf("model doctor missing quick fix: %q", modelDoctor.Message)
	}
}
