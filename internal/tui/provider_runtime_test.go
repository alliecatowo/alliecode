package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/agent"
	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/types"
)

type tuiRuntimeProvider struct {
	name      string
	chatCalls int
	requests  []types.ChatRequest
}

func (p *tuiRuntimeProvider) Name() string { return p.name }

func (p *tuiRuntimeProvider) Chat(_ context.Context, req types.ChatRequest) (<-chan types.StreamEvent, error) {
	p.chatCalls++
	p.requests = append(p.requests, req)
	ch := make(chan types.StreamEvent, 2)
	ch <- types.StreamEvent{Type: types.StreamContentDelta, Delta: p.name}
	ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopEndTurn}
	close(ch)
	return ch, nil
}

func (p *tuiRuntimeProvider) ChatSync(context.Context, types.ChatRequest) (*types.ChatResponse, error) {
	return nil, errors.New("not implemented")
}

func (p *tuiRuntimeProvider) ListModels(context.Context) ([]types.Model, error) { return nil, nil }
func (p *tuiRuntimeProvider) SupportsStreaming() bool                           { return true }
func (p *tuiRuntimeProvider) SupportsTools() bool                               { return false }
func (p *tuiRuntimeProvider) SupportsThinking() bool                            { return false }

func TestTUIUsesSelectedProviderForPromptSubmission(t *testing.T) {
	openaiProvider := &tuiRuntimeProvider{name: "openai"}
	anthropicProvider := &tuiRuntimeProvider{name: "anthropic"}
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
	app := New(Config{
		Agent: ag,
		InitialState: commands.RuntimeState{
			ProviderName: "openai",
			Model:        "gpt-4o-mini",
			ModelRef:     "openai/gpt-4o-mini",
		},
	})
	app.width = 180

	updated, _ := app.Update(submitMsg{text: "/model anthropic/claude-opus-4-20250514"})
	app = updated.(*App)
	updated, _ = app.Update(submitMsg{text: "/login provider anthropic"})
	app = updated.(*App)
	updated, cmd := app.Update(submitMsg{text: "hello"})
	app = updated.(*App)
	app.width = 120
	if cmd == nil {
		t.Fatalf("expected prompt submission command")
	}
	if msg := cmd(); msg == nil {
		t.Fatalf("expected agent completion message")
	}
	if anthropicProvider.chatCalls != 1 {
		t.Fatalf("anthropic provider calls = %d, want 1", anthropicProvider.chatCalls)
	}
	if openaiProvider.chatCalls != 0 {
		t.Fatalf("openai provider calls = %d, want 0", openaiProvider.chatCalls)
	}
	if got := anthropicProvider.requests[0].Model; got != "claude-opus-4-20250514" {
		t.Fatalf("anthropic request model = %q, want claude-opus-4-20250514", got)
	}
	if !strings.Contains(stripANSIForTest(app.renderStatusBar()), "anthropic/claude-opus-4-20250514") {
		t.Fatalf("expected status bar provider source-of-truth, got %q", app.renderStatusBar())
	}
}

func TestTUIHydratesStartupProviderModelState(t *testing.T) {
	app := New(Config{InitialState: commands.RuntimeState{
		ProviderName:  "anthropic",
		Model:         "claude-opus-4-20250514",
		ModelRef:      "anthropic/claude-opus-4-20250514",
		LoggedIn:      true,
		AuthProvider:  "anthropic",
		AuthAccount:   "dev@acme",
		ProviderReady: true,
		SessionID:     "startup-session",
	}})
	app.width = 180

	if got := app.statusProviderLabel(); got != "anthropic" {
		t.Fatalf("statusProviderLabel() = %q, want anthropic", got)
	}
	if got := app.statusModelLabel(); got != "claude-opus-4-20250514" {
		t.Fatalf("statusModelLabel() = %q, want claude-opus-4-20250514", got)
	}
	bar := stripANSIForTest(app.renderStatusBar())
	if !strings.Contains(bar, "anthropic/claude-opus-4-20250514") {
		t.Fatalf("expected hydrated provider in status bar, got %q", app.renderStatusBar())
	}
	if !strings.Contains(stripANSIForTest(app.renderStatusHints()), "provider anthropic") {
		t.Fatalf("expected hydrated provider in context hint, got %q", app.renderStatusHints())
	}
}

func TestTUIStatusHintUsesActualProviderReadiness(t *testing.T) {
	app := New(Config{InitialState: commands.RuntimeState{
		ProviderName: "anthropic",
		Model:        "claude-opus-4-20250514",
		ModelRef:     "anthropic/claude-opus-4-20250514",
		LoggedIn:     true,
		AuthProvider: "openai",
	}})
	app.width = 120
	app.syncInputMode()

	if !strings.Contains(app.renderStatusHints(), "needs login via /login provider anthropic") {
		t.Fatalf("expected readiness guidance in status hint, got %q", app.renderStatusHints())
	}
}

func TestTUIStatusLabelsFollowAgentRuntimeSnapshotOnDrift(t *testing.T) {
	openaiProvider := &tuiRuntimeProvider{name: "openai"}
	anthropicProvider := &tuiRuntimeProvider{name: "anthropic"}
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

	app := New(Config{InitialState: commands.RuntimeState{ProviderName: "openai", Model: "gpt-4o-mini", ModelRef: "openai/gpt-4o-mini", Agent: ag}})
	app.width = 140

	if got := app.statusProviderLabel(); got != "openai" {
		t.Fatalf("statusProviderLabel() = %q, want openai", got)
	}
	if got := app.statusModelLabel(); got != "gpt-4o-mini" {
		t.Fatalf("statusModelLabel() = %q, want gpt-4o-mini", got)
	}
	if !strings.Contains(stripANSIForTest(app.renderStatusBar()), "openai/gpt-4o-mini") {
		t.Fatalf("expected status bar to reflect startup runtime tuple, got %q", app.renderStatusBar())
	}
}
