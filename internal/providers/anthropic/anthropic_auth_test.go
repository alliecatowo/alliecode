package anthropic

import "testing"

func TestResolveAuthTokenPrecedence(t *testing.T) {
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "env-auth")
	t.Setenv("ANTHROPIC_ACCESS_TOKEN", "env-access")

	got := resolveAuthToken("cfg-auth", "cfg-access")
	if got != "env-auth" {
		t.Fatalf("resolveAuthToken() = %q, want env auth token", got)
	}
}

func TestResolveAuthTokenFallsBackToAccessToken(t *testing.T) {
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "")
	t.Setenv("ANTHROPIC_ACCESS_TOKEN", "")

	got := resolveAuthToken("", "cfg-access")
	if got != "cfg-access" {
		t.Fatalf("resolveAuthToken() = %q, want cfg access token", got)
	}
}

func TestAuthHeaderUsesBearerTokenWhenPresent(t *testing.T) {
	p := &Provider{apiKey: "api-key", authToken: "oauth-token"}

	header, value := p.authHeader()
	if header != "Authorization" {
		t.Fatalf("header = %q, want Authorization", header)
	}
	if value != "Bearer oauth-token" {
		t.Fatalf("value = %q, want Bearer oauth-token", value)
	}
}
