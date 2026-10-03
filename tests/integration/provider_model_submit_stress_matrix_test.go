package integration_test

import (
	"context"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/permissions"
)

func TestIntegrationProviderModelSubmitStressMatrix(t *testing.T) {
	type matrixCase struct {
		name           string
		initial        commands.RuntimeState
		steps          []string
		wantStatus     []string
		wantStatusline []string
	}

	cases := []matrixCase{
		{
			name:           "openai_to_anthropic_with_reauth",
			initial:        commands.RuntimeState{ProviderName: "openai", Model: "gpt-4o-mini", ModelRef: "openai/gpt-4o-mini", PermissionMode: permissions.ModeDefault},
			steps:          []string{"/login provider openai", "/model anthropic/sonnet", "/login provider anthropic", "/model opus"},
			wantStatus:     []string{"provider=anthropic", "model=claude-opus-4-20250514", "provider_ready=true", "logged_in=true"},
			wantStatusline: []string{"provider=anthropic", "model=claude-opus-4-20250514", "model_ref=anthropic/claude-opus-4-20250514", "provider_ready=true", "logged_in=true"},
		},
		{
			name:           "logout_keeps_selection_but_unready",
			initial:        commands.RuntimeState{ProviderName: "anthropic", Model: "claude-sonnet-4-20250514", ModelRef: "anthropic/claude-sonnet-4-20250514", PermissionMode: permissions.ModeDefault},
			steps:          []string{"/login provider anthropic", "/model haiku", "/logout"},
			wantStatus:     []string{"provider=anthropic", "model=claude-haiku-3-5-20241022", "provider_ready=false", "logged_in=false"},
			wantStatusline: []string{"provider=anthropic", "model=claude-haiku-3-5-20241022", "model_ref=anthropic/claude-haiku-3-5-20241022", "provider_ready=false", "logged_in=false"},
		},
		{
			name:           "ollama_switch_is_ready_without_login",
			initial:        commands.RuntimeState{ProviderName: "openai", Model: "gpt-4o-mini", ModelRef: "openai/gpt-4o-mini", PermissionMode: permissions.ModeDefault},
			steps:          []string{"/provider set ollama", "/model llama3"},
			wantStatus:     []string{"provider=ollama", "model=llama3", "provider_ready=true"},
			wantStatusline: []string{"provider=ollama", "model=llama3", "model_ref=ollama/llama3", "provider_ready=true"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			state := tc.initial
			r := commands.DefaultRegistry()
			ctx := commands.Context{State: &state}
			for _, step := range tc.steps {
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

			for _, want := range tc.wantStatus {
				if !strings.Contains(statusRes.Message, want) {
					t.Fatalf("status missing %q: %q", want, statusRes.Message)
				}
			}
			for _, want := range tc.wantStatusline {
				if !strings.Contains(statuslineRes.Message, want) {
					t.Fatalf("statusline missing %q: %q", want, statuslineRes.Message)
				}
			}
		})
	}
}

func TestIntegrationProviderModelDoctorAndStatusIncludeRecoveryHints(t *testing.T) {
	state := commands.RuntimeState{ProviderName: "openai", Model: "gpt-4o-mini", ModelRef: "openai/gpt-4o-mini", PermissionMode: permissions.ModeDefault}
	r := commands.DefaultRegistry()
	ctx := commands.Context{State: &state}

	if _, err := r.Dispatch(context.Background(), ctx, "/provider set anthropic"); err != nil {
		t.Fatalf("provider set failed: %v", err)
	}

	providerStatus, err := r.Dispatch(context.Background(), ctx, "/provider status")
	if err != nil {
		t.Fatalf("provider status failed: %v", err)
	}
	if !strings.Contains(providerStatus.Message, "quick_fix_auth=") || !strings.Contains(providerStatus.Message, "quick_fix_model=") {
		t.Fatalf("provider status missing recovery hints: %q", providerStatus.Message)
	}

	providerDoctor, err := r.Dispatch(context.Background(), ctx, "/provider doctor")
	if err != nil {
		t.Fatalf("provider doctor failed: %v", err)
	}
	if !strings.Contains(providerDoctor.Message, "quick_fix=") {
		t.Fatalf("provider doctor missing quick fix hint: %q", providerDoctor.Message)
	}

	modelDoctor, err := r.Dispatch(context.Background(), ctx, "/model doctor")
	if err != nil {
		t.Fatalf("model doctor failed: %v", err)
	}
	if !strings.Contains(modelDoctor.Message, "quick_fix=") {
		t.Fatalf("model doctor missing quick fix hint: %q", modelDoctor.Message)
	}
}
