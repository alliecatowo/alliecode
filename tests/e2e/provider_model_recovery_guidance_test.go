package e2e_test

import (
	"context"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
)

func TestE2EProviderModelRecoveryGuidancePitfalls(t *testing.T) {
	state := &commands.RuntimeState{ProviderName: "openai", Model: "gpt-4o-mini", ModelRef: "openai/gpt-4o-mini"}
	r := commands.DefaultRegistry()
	ctx := commands.Context{State: state}

	_, err := r.Dispatch(context.Background(), ctx, "/model bogus-model")
	if err == nil {
		t.Fatalf("expected unknown model error")
	}
	if !strings.Contains(err.Error(), "/model list openai") {
		t.Fatalf("expected model recovery hint, got %v", err)
	}

	if _, err := r.Dispatch(context.Background(), ctx, "/provider set anthropic"); err != nil {
		t.Fatalf("provider switch failed: %v", err)
	}
	status, err := r.Dispatch(context.Background(), ctx, "/provider status")
	if err != nil {
		t.Fatalf("provider status failed: %v", err)
	}
	if !strings.Contains(status.Message, "quick_fix_auth=/login provider anthropic") {
		t.Fatalf("expected login recovery hint after provider switch, got %q", status.Message)
	}
}
