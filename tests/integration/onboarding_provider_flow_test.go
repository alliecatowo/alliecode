package integration_test

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/config"
	"github.com/alliecatowo/alliecode/tests/integration/fixtures"
)

func TestOnboardingProviderChoiceOutcomesWithCLIInvocationFixture(t *testing.T) {
	cases := []struct {
		name            string
		provider        string
		modelSetArg     string
		wantModel       string
		wantProviderSet bool
	}{
		{
			name:            "openai choice uses explicit provider model",
			provider:        "openai",
			modelSetArg:     "openai/gpt-4o-mini",
			wantModel:       "gpt-4o-mini",
			wantProviderSet: true,
		},
		{
			name:            "anthropic choice uses explicit provider model",
			provider:        "anthropic",
			modelSetArg:     "anthropic/claude-sonnet-4-20250514",
			wantModel:       "claude-sonnet-4-20250514",
			wantProviderSet: true,
		},
		{
			name:            "ollama choice allows providerless model selection",
			provider:        "ollama",
			modelSetArg:     "llama3",
			wantModel:       "llama3",
			wantProviderSet: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.NewDefaultConfig()
			cfg.HasCompletedOnboarding = true
			cfg.OnboardingProviderChoice = tc.provider

			state := &commands.RuntimeState{}
			if tc.wantProviderSet {
				state.ProviderName = cfg.OnboardingProviderChoice
			}

			fx := fixtures.NewCLIInvocationFixture(state)
			res, err := fx.Invoke("/model " + tc.modelSetArg)
			if err != nil {
				t.Fatalf("Invoke(/model %s) error = %v", tc.modelSetArg, err)
			}
			if !res.Handled {
				t.Fatalf("/model %s should be handled", tc.modelSetArg)
			}
			if got := state.Model; got != tc.wantModel {
				t.Fatalf("state.Model = %q, want %q", got, tc.wantModel)
			}
		})
	}
}

func TestProviderSelectionFallbackDefaultsToOllamaWhenUnset(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.LoadLayered(project)
	if err != nil {
		t.Fatalf("LoadLayered() error = %v", err)
	}

	if got := cfg.DefaultProvider; got != "ollama" {
		t.Fatalf("DefaultProvider = %q, want %q", got, "ollama")
	}
}

func TestProviderSelectionMissingCredsBehaviorForRemoteProvider(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.DefaultProvider = "openai"

	if cfg.HasDefaultProviderCredentials() {
		t.Fatalf("expected openai default provider to report missing credentials")
	}

	cfg.Providers["openai"] = &config.ProviderSettings{APIKey: "sk-test"}
	if !cfg.HasDefaultProviderCredentials() {
		t.Fatalf("expected openai default provider credentials to be detected")
	}
}
