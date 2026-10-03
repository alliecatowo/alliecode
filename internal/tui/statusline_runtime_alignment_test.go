package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/agent"
	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTUIRuntimePaneFollowsRuntimeSubmitTuple(t *testing.T) {
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

	app := New(Config{Agent: ag, InitialState: commands.RuntimeState{ProviderName: "openai", Model: "gpt-4o-mini", ModelRef: "openai/gpt-4o-mini"}})
	app.width = 160
	app.setSearchModeValue(searchModeTimeline)
	app.setState(stateSearch)

	if got := stripANSIForTest(app.renderStatusBar()); !strings.Contains(got, "anthropic/claude-opus-4-20250514") {
		t.Fatalf("status bar should reflect runtime tuple: %q", got)
	}
	if got := stripANSIForTest(app.renderStatusRuntimePanes()); !strings.Contains(got, "runtime: state=search") {
		t.Fatalf("expected runtime pane for search state: %q", got)
	}
	app.exitSearch()
	updated, _ := app.Update(submitMsg{text: "/login provider anthropic"})
	app = updated.(*App)

	updated, cmd := app.Update(submitMsg{text: "hello"})
	app = updated.(*App)
	if cmd == nil {
		t.Fatalf("expected submit command")
	}
	if msg := cmd(); msg == nil {
		t.Fatalf("expected agent completion message")
	}
	if anthropicProvider.chatCalls != 1 {
		t.Fatalf("anthropic provider calls = %d, want 1", anthropicProvider.chatCalls)
	}
	if got := anthropicProvider.requests[0].Model; got != "claude-opus-4-20250514" {
		t.Fatalf("submit model = %q, want claude-opus-4-20250514", got)
	}
}
