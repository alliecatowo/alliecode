package integration_test

import (
	"context"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/permissions"
)

func TestIntegration_StatusAndStatuslineStayAlignedAfterProviderModelPermutations(t *testing.T) {
	r := commands.DefaultRegistry()
	state := &commands.RuntimeState{ProviderName: "openai", Model: "gpt-4o-mini", ModelRef: "openai/gpt-4o-mini", PermissionMode: permissions.ModeDefault}
	ctx := commands.Context{State: state}

	steps := []string{
		"/login provider openai",
		"/model gpt-4o",
		"/model anthropic/sonnet",
		"/provider set anthropic",
		"/model opus",
		"/logout",
	}
	for _, step := range steps {
		if _, err := r.Dispatch(context.Background(), ctx, step); err != nil {
			t.Fatalf("dispatch %q failed: %v", step, err)
		}
	}

	statusRes, err := r.Dispatch(context.Background(), ctx, "/status")
	if err != nil {
		t.Fatalf("/status failed: %v", err)
	}
	statuslineRes, err := r.Dispatch(context.Background(), ctx, "/statusline status")
	if err != nil {
		t.Fatalf("/statusline status failed: %v", err)
	}
	for _, want := range []string{"provider=anthropic", "model=claude-opus-4-20250514", "provider_ready=false", "logged_in=false"} {
		if !strings.Contains(statusRes.Message, want) {
			t.Fatalf("status output missing %q: %q", want, statusRes.Message)
		}
	}
	for _, want := range []string{"provider=anthropic", "model=claude-opus-4-20250514", "model_ref=anthropic/claude-opus-4-20250514", "provider_ready=false", "logged_in=false"} {
		if !strings.Contains(statuslineRes.Message, want) {
			t.Fatalf("statusline output missing %q: %q", want, statuslineRes.Message)
		}
	}
}

func TestIntegration_ProviderAndModelErrorsOfferRecoveryGuidance(t *testing.T) {
	r := commands.DefaultRegistry()
	ctx := commands.Context{State: &commands.RuntimeState{ProviderName: "openai", PermissionMode: permissions.ModeDefault}}

	_, err := r.Dispatch(context.Background(), ctx, "/model not-a-model")
	if err == nil {
		t.Fatalf("expected unknown model error")
	}
	if !strings.Contains(err.Error(), "run /model list openai") || !strings.Contains(err.Error(), "/model openai/<model>") {
		t.Fatalf("expected actionable model recovery guidance, got %v", err)
	}

	_, err = r.Dispatch(context.Background(), ctx, "/provider set nope")
	if err == nil {
		t.Fatalf("expected unknown provider error")
	}
	if !strings.Contains(err.Error(), "run /provider list") || !strings.Contains(err.Error(), "/provider set <name>") {
		t.Fatalf("expected actionable provider recovery guidance, got %v", err)
	}
}
