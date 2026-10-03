package commands

import (
	"context"
	"strings"
	"testing"
)

func TestProviderModelRecoveryGuidancePitfalls(t *testing.T) {
	r := DefaultRegistry()
	ctx := Context{State: &RuntimeState{ProviderName: "openai"}}

	_, err := r.Dispatch(context.Background(), ctx, "/model not-a-model")
	if err == nil {
		t.Fatalf("expected unknown model error")
	}
	if !strings.Contains(err.Error(), "/model list openai") || !strings.Contains(err.Error(), "/model openai/<model>") {
		t.Fatalf("missing model recovery guidance: %v", err)
	}

	_, err = r.Dispatch(context.Background(), ctx, "/provider set wrong-provider")
	if err == nil {
		t.Fatalf("expected unknown provider error")
	}
	if !strings.Contains(err.Error(), "/provider list") || !strings.Contains(err.Error(), "/provider set <name>") {
		t.Fatalf("missing provider recovery guidance: %v", err)
	}

	if _, err = r.Dispatch(context.Background(), ctx, "/provider set anthropic"); err != nil {
		t.Fatalf("provider set anthropic failed: %v", err)
	}
	status, err := r.Dispatch(context.Background(), ctx, "/provider status")
	if err != nil {
		t.Fatalf("provider status failed: %v", err)
	}
	if !strings.Contains(status.Message, "quick_fix_auth=/login provider anthropic") {
		t.Fatalf("missing auth recovery hint for unauthenticated provider: %q", status.Message)
	}
}
