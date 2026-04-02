package common

import "testing"

func TestResolveAuthValuePrecedence(t *testing.T) {
	t.Setenv("TEST_PROVIDER_KEY", "env-key")

	got := ResolveAuthValue("TEST_PROVIDER_KEY", "cfg-key", "fallback")
	if got != "env-key" {
		t.Fatalf("expected env value, got %q", got)
	}
}

func TestResolveAuthValueConfigFallback(t *testing.T) {
	t.Setenv("TEST_PROVIDER_KEY", "")

	got := ResolveAuthValue("TEST_PROVIDER_KEY", "cfg-key", "fallback")
	if got != "cfg-key" {
		t.Fatalf("expected config value, got %q", got)
	}

	got = ResolveAuthValue("TEST_PROVIDER_KEY", "", "fallback")
	if got != "fallback" {
		t.Fatalf("expected fallback value, got %q", got)
	}
}
