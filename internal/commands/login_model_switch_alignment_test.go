package commands

import (
	"context"
	"strings"
	"testing"
)

func TestModelSwitchAcrossProvidersClearsIncompatibleLoginState(t *testing.T) {
	state := &RuntimeState{
		ProviderName:  "openai",
		Model:         "gpt-4o-mini",
		ModelRef:      "openai/gpt-4o-mini",
		LoggedIn:      true,
		AuthProvider:  "openai",
		AuthAccount:   "dev@acme",
		ProviderReady: true,
	}

	res, err := NewModelCommand().Execute(context.Background(), Context{State: state}, Invocation{Name: "model", Args: []string{"anthropic/claude-sonnet-4-20250514"}})
	if err != nil {
		t.Fatalf("model set failed: %v", err)
	}
	if !res.Handled {
		t.Fatalf("expected handled model switch")
	}
	if state.ProviderName != "anthropic" || state.ModelRef != "anthropic/claude-sonnet-4-20250514" {
		t.Fatalf("unexpected selection after switch: %+v", state)
	}
	if state.LoggedIn {
		t.Fatalf("expected login cleared after provider switch")
	}
	if state.AuthProvider != "" || state.AuthAccount != "" {
		t.Fatalf("expected auth identity cleared after provider switch: %+v", state)
	}
	if state.ProviderReady {
		t.Fatalf("expected provider readiness false until relogin")
	}
}

func TestUnknownModelErrorOffersActionableRecoveryGuidance(t *testing.T) {
	state := &RuntimeState{ProviderName: "openai"}
	_, err := NewModelCommand().Execute(context.Background(), Context{State: state}, Invocation{Name: "model", Args: []string{"not-a-model"}})
	if err == nil {
		t.Fatalf("expected unknown model error")
	}
	if !strings.Contains(err.Error(), "run /model list openai") || !strings.Contains(err.Error(), "/model openai/<model>") {
		t.Fatalf("expected actionable model recovery guidance, got %v", err)
	}
}

func TestUnknownProviderErrorOffersActionableRecoveryGuidance(t *testing.T) {
	state := &RuntimeState{}
	_, err := NewProviderCommand().Execute(context.Background(), Context{State: state}, Invocation{Name: "provider", Args: []string{"set", "bogus"}})
	if err == nil {
		t.Fatalf("expected unknown provider error")
	}
	if !strings.Contains(err.Error(), "run /provider list") || !strings.Contains(err.Error(), "/provider set <name>") {
		t.Fatalf("expected actionable provider recovery guidance, got %v", err)
	}
}
