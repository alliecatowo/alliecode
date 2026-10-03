package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/agent"
	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/permissions"
	"github.com/alliecatowo/alliecode/internal/types"
	tea "github.com/charmbracelet/bubbletea"
)

func TestSubmitStressMatrixProviderModelLoginAndSendPermutations(t *testing.T) {
	type scenario struct {
		name          string
		initial       commands.RuntimeState
		steps         []string
		wantContains  []string
		wantErrPrefix string
		wantStatusBar []string
		wantHint      []string
		wantCalls     map[string]int
	}

	cases := []scenario{
		{
			name: "startup_openai_missing_login_then_send",
			initial: commands.RuntimeState{
				ProviderName: "openai",
				Model:        "gpt-4o-mini",
				ModelRef:     "openai/gpt-4o-mini",
			},
			steps:         []string{"hello"},
			wantErrPrefix: "provider openai needs login; next: run /login provider openai and retry",
		},
		{
			name:         "login_openai_send",
			initial:      commands.RuntimeState{ProviderName: "openai", Model: "gpt-4o-mini", ModelRef: "openai/gpt-4o-mini"},
			steps:        []string{"/login provider openai", "hello"},
			wantContains: []string{"LOGIN_PROVIDER", "openai"},
		},
		{
			name: "provider_switch_anthropic_requires_relogin_before_send",
			initial: commands.RuntimeState{
				ProviderName:  "openai",
				Model:         "gpt-4o-mini",
				ModelRef:      "openai/gpt-4o-mini",
				LoggedIn:      true,
				AuthProvider:  "openai",
				ProviderReady: true,
			},
			steps:         []string{"/provider set anthropic", "hello"},
			wantErrPrefix: "provider anthropic needs login; next: run /login provider anthropic and retry",
		},
		{
			name: "provider_switch_then_login_then_send",
			initial: commands.RuntimeState{
				ProviderName:  "openai",
				Model:         "gpt-4o-mini",
				ModelRef:      "openai/gpt-4o-mini",
				LoggedIn:      true,
				AuthProvider:  "openai",
				ProviderReady: true,
			},
			steps:         []string{"/provider set anthropic", "/login provider anthropic", "hello"},
			wantContains:  []string{"LOGIN_PROVIDER", "anthropic"},
			wantStatusBar: []string{"anthropic/claude-sonnet-4-20250514"},
			wantHint:      []string{"provider anthropic"},
			wantCalls:     map[string]int{"openai": 0, "anthropic": 1, "ollama": 0},
		},
		{
			name: "model_switch_cross_provider_then_send_requires_login",
			initial: commands.RuntimeState{
				ProviderName:  "openai",
				Model:         "gpt-4o-mini",
				ModelRef:      "openai/gpt-4o-mini",
				LoggedIn:      true,
				AuthProvider:  "openai",
				ProviderReady: true,
			},
			steps:         []string{"/model anthropic/sonnet", "hello"},
			wantErrPrefix: "provider anthropic needs login; next: run /login provider anthropic and retry",
		},
		{
			name: "model_switch_same_provider_then_send",
			initial: commands.RuntimeState{
				ProviderName:  "openai",
				Model:         "gpt-4o-mini",
				ModelRef:      "openai/gpt-4o-mini",
				LoggedIn:      true,
				AuthProvider:  "openai",
				ProviderReady: true,
			},
			steps:        []string{"/model gpt-4o", "hello"},
			wantContains: []string{"Model set to gpt-4o"},
		},
		{
			name: "startup_anthropic_logged_in_send",
			initial: commands.RuntimeState{
				ProviderName:  "anthropic",
				Model:         "claude-opus-4-20250514",
				ModelRef:      "anthropic/claude-opus-4-20250514",
				LoggedIn:      true,
				AuthProvider:  "anthropic",
				ProviderReady: true,
			},
			steps:         []string{"hello"},
			wantContains:  []string{"hello"},
			wantStatusBar: []string{"anthropic/claude-opus-4-20250514"},
			wantCalls:     map[string]int{"openai": 0, "anthropic": 1, "ollama": 0},
		},
		{
			name: "startup_ollama_send_without_login",
			initial: commands.RuntimeState{
				ProviderName: "ollama",
				Model:        "llama3",
				ModelRef:     "ollama/llama3",
			},
			steps:         []string{"hello"},
			wantContains:  []string{"hello"},
			wantStatusBar: []string{"ollama/llama3"},
			wantHint:      []string{"provider ollama"},
			wantCalls:     map[string]int{"openai": 0, "anthropic": 0, "ollama": 1},
		},
		{
			name: "logout_then_send_requires_login",
			initial: commands.RuntimeState{
				ProviderName:  "openai",
				Model:         "gpt-4o-mini",
				ModelRef:      "openai/gpt-4o-mini",
				LoggedIn:      true,
				AuthProvider:  "openai",
				ProviderReady: true,
			},
			steps:         []string{"/logout", "hello"},
			wantErrPrefix: "provider openai needs login; next: run /login provider openai and retry",
		},
		{
			name: "logout_login_send_recovers",
			initial: commands.RuntimeState{
				ProviderName:  "openai",
				Model:         "gpt-4o-mini",
				ModelRef:      "openai/gpt-4o-mini",
				LoggedIn:      true,
				AuthProvider:  "openai",
				ProviderReady: true,
			},
			steps:        []string{"/logout", "/login provider openai", "hello"},
			wantContains: []string{"LOGOUT_RESULT", "LOGIN_PROVIDER"},
		},
		{
			name: "provider_set_ollama_after_logout_send",
			initial: commands.RuntimeState{
				ProviderName:  "openai",
				Model:         "gpt-4o-mini",
				ModelRef:      "openai/gpt-4o-mini",
				LoggedIn:      true,
				AuthProvider:  "openai",
				ProviderReady: true,
			},
			steps:         []string{"/logout", "/provider set ollama", "hello"},
			wantContains:  []string{"PROVIDER_SET", "provider=ollama"},
			wantStatusBar: []string{"ollama/llama3"},
			wantCalls:     map[string]int{"openai": 0, "anthropic": 0, "ollama": 1},
		},
		{
			name: "startup_openai_login_alias_send_uses_openai_chain",
			initial: commands.RuntimeState{
				ProviderName: "openai",
				Model:        "gpt-4o-mini",
				ModelRef:     "openai/gpt-4o-mini",
			},
			steps: []string{"/login provider openai", "/model gpt-4o", "hello"},
			wantContains: []string{
				"LOGIN_PROVIDER",
				"Model set to gpt-4o",
				"hello",
			},
			wantStatusBar: []string{"openai/gpt-4o"},
			wantHint:      []string{"provider openai"},
			wantCalls:     map[string]int{"openai": 1, "anthropic": 0, "ollama": 0},
		},
		{
			name: "startup_openai_cross_tuple_login_send_routes_anthropic_chain",
			initial: commands.RuntimeState{
				ProviderName: "openai",
				Model:        "gpt-4o-mini",
				ModelRef:     "openai/gpt-4o-mini",
			},
			steps: []string{"/model anthropic/sonnet", "/login provider anthropic", "hello"},
			wantContains: []string{
				"Model set to claude-sonnet-4-20250514",
				"LOGIN_PROVIDER",
				"anthropic",
				"hello",
			},
			wantStatusBar: []string{"anthropic/claude-sonnet-4-20250514"},
			wantHint:      []string{"provider anthropic"},
			wantCalls:     map[string]int{"openai": 0, "anthropic": 1, "ollama": 0},
		},
		{
			name: "startup_ollama_to_openai_send_requires_login_gate",
			initial: commands.RuntimeState{
				ProviderName: "ollama",
				Model:        "llama3",
				ModelRef:     "ollama/llama3",
			},
			steps:         []string{"/model openai/gpt-4o", "hello"},
			wantErrPrefix: "provider openai needs login; next: run /login provider openai and retry",
			wantStatusBar: []string{"openai/gpt-4o"},
			wantHint:      []string{"needs login via /login provider openai"},
			wantCalls:     map[string]int{"openai": 0, "anthropic": 0, "ollama": 0},
		},
		{
			name: "startup_openai_login_then_switch_ollama_send_routes_ollama",
			initial: commands.RuntimeState{
				ProviderName: "openai",
				Model:        "gpt-4o-mini",
				ModelRef:     "openai/gpt-4o-mini",
			},
			steps: []string{"/login provider openai", "/model ollama/llama3", "hello"},
			wantContains: []string{
				"LOGIN_PROVIDER",
				"Model set to llama3",
				"hello",
			},
			wantStatusBar: []string{"ollama/llama3"},
			wantHint:      []string{"provider ollama"},
			wantCalls:     map[string]int{"openai": 0, "anthropic": 0, "ollama": 1},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			state := tc.initial
			state.PermissionMode = permissions.ModeDefault

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
			providerName := strings.TrimSpace(state.ProviderName)
			modelName := strings.TrimSpace(state.Model)
			if providerName == "" {
				providerName = "openai"
			}
			if modelName == "" {
				modelName = "gpt-4o-mini"
			}
			if err := ag.SetProviderModel(providerName, modelName); err != nil {
				t.Fatalf("SetProviderModel(%s,%s) failed: %v", providerName, modelName, err)
			}
			state.Agent = ag
			app := New(Config{Agent: ag, InitialState: state})
			updated, _ := app.Update(tea.WindowSizeMsg{Width: 160, Height: 28})
			app = updated.(*App)

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

			if tc.wantErrPrefix != "" {
				if len(app.timeline) == 0 {
					t.Fatalf("expected timeline error row")
				}
				last := app.timeline[len(app.timeline)-1]
				if last.kind != timelineError {
					t.Fatalf("expected last timeline row to be error, got kind=%s text=%q", last.kind, last.text)
				}
				if !strings.HasPrefix(last.text, tc.wantErrPrefix) {
					t.Fatalf("error = %q, want prefix %q", last.text, tc.wantErrPrefix)
				}
				if app.stateValue() != stateIdle {
					t.Fatalf("expected idle state after blocked submit, got %s", app.stateLabel())
				}
				if !app.input.focused {
					t.Fatalf("expected input focused after blocked submit")
				}
			}

			for _, want := range tc.wantContains {
				joined := ""
				for _, row := range app.timeline {
					joined += row.text + "\n"
				}
				if !strings.Contains(joined, want) {
					t.Fatalf("timeline missing %q\n%s", want, joined)
				}
			}

			statusBar := stripANSIForTest(app.renderStatusBar())
			for _, want := range tc.wantStatusBar {
				if !strings.Contains(statusBar, want) {
					t.Fatalf("status bar missing %q: %q", want, statusBar)
				}
			}
			hints := stripANSIForTest(app.renderStatusHints())
			for _, want := range tc.wantHint {
				if !strings.Contains(hints, want) {
					t.Fatalf("status hints missing %q: %q", want, hints)
				}
			}
			for providerName, wantCalls := range tc.wantCalls {
				var gotCalls int
				switch providerName {
				case "openai":
					gotCalls = openaiProvider.chatCalls
				case "anthropic":
					gotCalls = anthropicProvider.chatCalls
				case "ollama":
					gotCalls = ollamaProvider.chatCalls
				default:
					t.Fatalf("unknown provider in wantCalls: %q", providerName)
				}
				if gotCalls != wantCalls {
					t.Fatalf("provider %s chat calls = %d, want %d", providerName, gotCalls, wantCalls)
				}
			}
		})
	}
}
