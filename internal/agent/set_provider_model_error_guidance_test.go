package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

type providerOnlyStub struct{ name string }

func (p providerOnlyStub) Name() string { return p.name }
func (p providerOnlyStub) Chat(_ context.Context, _ types.ChatRequest) (<-chan types.StreamEvent, error) {
	return nil, nil
}
func (p providerOnlyStub) ChatSync(_ context.Context, _ types.ChatRequest) (*types.ChatResponse, error) {
	return nil, nil
}
func (p providerOnlyStub) ListModels(_ context.Context) ([]types.Model, error) { return nil, nil }
func (p providerOnlyStub) SupportsStreaming() bool                             { return true }
func (p providerOnlyStub) SupportsTools() bool                                 { return true }
func (p providerOnlyStub) SupportsThinking() bool                              { return false }

func TestSetProviderModelErrorsExposeActionableGuidance(t *testing.T) {
	a := New(Config{})

	if err := a.SetProviderModel("", "gpt-4o"); err == nil || !strings.Contains(err.Error(), "next: run /provider set <name>") {
		t.Fatalf("expected provider guidance, got %v", err)
	}
	if err := a.SetProviderModel("openai", ""); err == nil || !strings.Contains(err.Error(), "next: run /model openai/<model>") {
		t.Fatalf("expected model guidance, got %v", err)
	}
}

func TestSetProviderModelReportsResolverGuidanceWhenSwitchNotConfigured(t *testing.T) {
	a := New(Config{Provider: providerOnlyStub{name: "openai"}, ProviderName: "openai", Model: "gpt-4o-mini"})
	err := a.SetProviderModel("anthropic", "claude-opus-4-20250514")
	if err == nil {
		t.Fatalf("expected provider switch error")
	}
	if !strings.Contains(err.Error(), "provider switch to \"anthropic\" is not configured") {
		t.Fatalf("unexpected error text: %v", err)
	}
	if !strings.Contains(err.Error(), "next: configure provider resolver or keep current provider") {
		t.Fatalf("expected actionable next step, got: %v", err)
	}
}
