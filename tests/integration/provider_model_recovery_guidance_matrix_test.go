package integration_test

import (
	"context"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/permissions"
)

func TestIntegrationProviderModelRecoveryGuidanceMatrix(t *testing.T) {
	type tc struct {
		name string
		run  func(t *testing.T, r *commands.Registry, ctx commands.Context)
	}

	cases := []tc{
		{
			name: "bad_model_id_suggests_list_and_set",
			run: func(t *testing.T, r *commands.Registry, ctx commands.Context) {
				_, err := r.Dispatch(context.Background(), ctx, "/model not-real")
				if err == nil {
					t.Fatalf("expected unknown model error")
				}
				if !strings.Contains(err.Error(), "/model list openai") || !strings.Contains(err.Error(), "/model openai/<model>") {
					t.Fatalf("missing model guidance: %v", err)
				}
			},
		},
		{
			name: "wrong_provider_suggests_list_and_set",
			run: func(t *testing.T, r *commands.Registry, ctx commands.Context) {
				_, err := r.Dispatch(context.Background(), ctx, "/provider set not-a-provider")
				if err == nil {
					t.Fatalf("expected unknown provider error")
				}
				if !strings.Contains(err.Error(), "/provider list") || !strings.Contains(err.Error(), "/provider set <name>") {
					t.Fatalf("missing provider guidance: %v", err)
				}
			},
		},
		{
			name: "missing_auth_surfaces_login_recovery",
			run: func(t *testing.T, r *commands.Registry, ctx commands.Context) {
				if _, err := r.Dispatch(context.Background(), ctx, "/provider set anthropic"); err != nil {
					t.Fatalf("provider set failed: %v", err)
				}
				res, err := r.Dispatch(context.Background(), ctx, "/provider status")
				if err != nil {
					t.Fatalf("provider status failed: %v", err)
				}
				if !strings.Contains(res.Message, "quick_fix_auth=/login provider anthropic") {
					t.Fatalf("missing auth quick fix: %q", res.Message)
				}
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			r := commands.DefaultRegistry()
			ctx := commands.Context{State: &commands.RuntimeState{ProviderName: "openai", Model: "gpt-4o-mini", ModelRef: "openai/gpt-4o-mini", PermissionMode: permissions.ModeDefault}}
			c.run(t, r, ctx)
		})
	}
}
