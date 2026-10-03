package config

import "testing"

func TestProviderCredentialPresenceMatrix(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.Providers["openai"] = &ProviderSettings{APIKey: "config-key"}

	t.Setenv("OPENAI_API_KEY", "env-key")
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "anthropic-auth")

	openai := cfg.ProviderCredentialPresence("openai")
	if !openai.FromConfig || !openai.FromEnv || !openai.HasCredentials {
		t.Fatalf("unexpected openai presence: %+v", openai)
	}

	anthropic := cfg.ProviderCredentialPresence("anthropic")
	if anthropic.FromConfig || !anthropic.FromEnv || !anthropic.HasCredentials {
		t.Fatalf("unexpected anthropic presence: %+v", anthropic)
	}

	gemini := cfg.ProviderCredentialPresence("gemini")
	if gemini.HasCredentials {
		t.Fatalf("expected no gemini credentials: %+v", gemini)
	}
}
