package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/agent"
	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/permissions"
	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStartupModelSendVisibleStateAndChainCoherence(t *testing.T) {
	type chainCase struct {
		name            string
		initial         commands.RuntimeState
		steps           []string
		wantStatusBar   []string
		wantHints       []string
		wantTimeline    []string
		wantErrorPrefix string
		wantCalls       map[string]int
	}

	cases := []chainCase{
		{
			name:    "openai_startup_login_model_send",
			initial: commands.RuntimeState{ProviderName: "openai", Model: "gpt-4o-mini", ModelRef: "openai/gpt-4o-mini", PermissionMode: permissions.ModeDefault},
			steps:   []string{"/login provider openai", "/model gpt-4o", "ship it"},
			wantStatusBar: []string{
				"openai/gpt-4o",
			},
			wantHints:    []string{"provider openai"},
			wantTimeline: []string{"LOGIN_PROVIDER", "Model set to gpt-4o", "ship it"},
			wantCalls:    map[string]int{"openai": 1, "anthropic": 0, "ollama": 0},
		},
		{
			name:    "cross_provider_model_requires_login_before_send",
			initial: commands.RuntimeState{ProviderName: "openai", Model: "gpt-4o-mini", ModelRef: "openai/gpt-4o-mini", LoggedIn: true, AuthProvider: "openai", ProviderReady: true, PermissionMode: permissions.ModeDefault},
			steps:   []string{"/model anthropic/sonnet", "ship it"},
			wantStatusBar: []string{
				"anthropic/claude-sonnet-4-20250514",
			},
			wantHints:       []string{"needs login via /login provider anthropic"},
			wantTimeline:    []string{"Model set to claude-sonnet-4-20250514"},
			wantErrorPrefix: "provider anthropic needs login; next: run /login provider anthropic and retry",
			wantCalls:       map[string]int{"openai": 0, "anthropic": 0, "ollama": 0},
		},
		{
			name:    "ollama_startup_send_without_login",
			initial: commands.RuntimeState{ProviderName: "ollama", Model: "llama3", ModelRef: "ollama/llama3", PermissionMode: permissions.ModeDefault},
			steps:   []string{"ship it"},
			wantStatusBar: []string{
				"ollama/llama3",
			},
			wantHints:    []string{"provider ollama"},
			wantTimeline: []string{"ship it"},
			wantCalls:    map[string]int{"openai": 0, "anthropic": 0, "ollama": 1},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			openaiProvider := &tuiRuntimeProvider{name: "openai"}
			anthropicProvider := &tuiRuntimeProvider{name: "anthropic"}
			ollamaProvider := &tuiRuntimeProvider{name: "ollama"}
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
					case "ollama":
						return ollamaProvider, nil
					default:
						return nil, errors.New("unknown provider")
					}
				},
			})

			state := tc.initial
			if state.ProviderName == "" {
				state.ProviderName = "openai"
			}
			if state.Model == "" {
				state.Model = "gpt-4o-mini"
			}
			if err := ag.SetProviderModel(state.ProviderName, state.Model); err != nil {
				t.Fatalf("SetProviderModel(%s,%s) failed: %v", state.ProviderName, state.Model, err)
			}
			state.Agent = ag

			app := New(Config{Agent: ag, InitialState: state})
			app.width = 180
			app.height = 30

			for _, step := range tc.steps {
				updated, cmd := app.Update(submitMsg{text: step})
				app = updated.(*App)
				if cmd != nil {
					event := cmd()
					if event != nil {
						updated, _ = app.Update(event)
						app = updated.(*App)
					}
				}
			}

			joined := ""
			for _, row := range app.timeline {
				joined += row.text + "\n"
			}
			for _, want := range tc.wantTimeline {
				if !strings.Contains(joined, want) {
					t.Fatalf("timeline missing %q\n%s", want, joined)
				}
			}
			if tc.wantErrorPrefix != "" {
				if len(app.timeline) == 0 {
					t.Fatalf("expected error timeline row")
				}
				last := app.timeline[len(app.timeline)-1]
				if last.kind != timelineError {
					t.Fatalf("expected error row, got kind=%s", last.kind)
				}
				if !strings.HasPrefix(last.text, tc.wantErrorPrefix) {
					t.Fatalf("error = %q, want prefix %q", last.text, tc.wantErrorPrefix)
				}
			}

			statusBar := stripANSIForTest(app.renderStatusBar())
			for _, want := range tc.wantStatusBar {
				if !strings.Contains(statusBar, want) {
					t.Fatalf("status bar missing %q: %q", want, statusBar)
				}
			}
			hints := stripANSIForTest(app.renderStatusHints())
			for _, want := range tc.wantHints {
				if !strings.Contains(hints, want) {
					t.Fatalf("status hints missing %q: %q", want, hints)
				}
			}

			for providerName, wantCalls := range tc.wantCalls {
				var got int
				switch providerName {
				case "openai":
					got = openaiProvider.chatCalls
				case "anthropic":
					got = anthropicProvider.chatCalls
				case "ollama":
					got = ollamaProvider.chatCalls
				}
				if got != wantCalls {
					t.Fatalf("provider %s chat calls = %d, want %d", providerName, got, wantCalls)
				}
			}
		})
	}
}
