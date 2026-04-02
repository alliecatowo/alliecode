package providers

import (
	"context"
	"testing"

	"github.com/alliecatowo/alliecode/internal/config"
)

func TestNewSupportsGeminiProvider(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "env-gem-key")

	p, err := New("gemini", &config.ProviderSettings{})
	if err != nil {
		t.Fatalf("expected gemini provider, got error: %v", err)
	}
	if p.Name() != "gemini" {
		t.Fatalf("provider name = %q, want gemini", p.Name())
	}
}

func TestNewAnthropicAuthTokenTakesPrecedenceOverAPIKey(t *testing.T) {
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "env-auth-token")
	t.Setenv("ANTHROPIC_API_KEY", "env-api-key")

	p, err := New("anthropic", &config.ProviderSettings{})
	if err != nil {
		t.Fatalf("expected anthropic provider, got error: %v", err)
	}
	if p.Name() != "anthropic" {
		t.Fatalf("provider name = %q, want anthropic", p.Name())
	}
}

func TestValidateProviderModelReadinessFailsWithoutCredentials(t *testing.T) {
	status := ValidateProviderModelReadiness(context.Background(), "openai", &config.ProviderSettings{}, "gpt-4o-mini")
	if status.Ready {
		t.Fatalf("expected not ready without credentials")
	}
	if status.ModelAvailable {
		t.Fatalf("expected model unavailable when provider is not ready")
	}
	if status.Reason == "" {
		t.Fatalf("expected readiness failure reason")
	}
}

func TestValidateProviderModelReadinessReportsUnknownModel(t *testing.T) {
	status := ValidateProviderModelReadiness(context.Background(), "openai", &config.ProviderSettings{APIKey: "sk-openai"}, "not-a-model")
	if status.Ready {
		t.Fatalf("expected not ready for unknown model")
	}
	if status.Reason == "" || status.Reason == "openai: API key required (set OPENAI_API_KEY)" {
		t.Fatalf("expected unknown-model reason, got %q", status.Reason)
	}
}
