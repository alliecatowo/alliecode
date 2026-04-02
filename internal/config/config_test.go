package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMergePrecedence(t *testing.T) {
	defaults := NewDefaultConfig()
	defaults.DefaultProvider = "default-provider"
	defaults.DefaultModel = "default-model"
	defaults.ProviderAllow = []string{"default-provider"}
	defaults.ModelAllow = []string{"default-model"}
	defaults.ToolPermissions["read"] = true
	defaults.Limits.MaxRetries = 1
	defaults.Remote = RemoteSettings{Mode: remoteModeDisabled}
	defaults.Companion = CompanionStoredState{Name: "default-name", Personality: "default-personality", HatchedAt: 100}
	defaults.CompanionMuted = boolPtr(false)
	defaults.HasCompletedOnboarding = true
	defaults.OnboardingProviderChoice = "openai"

	global := &Config{
		DefaultProvider:          "global-provider",
		Companion:                CompanionStoredState{Name: "global-name", HatchedAt: 200},
		HasCompletedOnboarding:   true,
		OnboardingProviderChoice: "anthropic",
		Providers: map[string]*ProviderSettings{
			"shared": {
				APIKey:    "global-key",
				AuthToken: "global-auth-token",
			},
		},
		ToolPermissions: map[string]bool{
			"run": false,
		},
		Limits: NumericLimits{MaxRetries: 3},
		Remote: RemoteSettings{
			Mode:       remoteModeLocal,
			ListenAddr: "127.0.0.1:7777",
			Enabled:    true,
		},
	}

	project := &Config{
		DefaultModel: "project-model",
		Companion: CompanionStoredState{
			Personality: "project-personality",
		},
		HasCompletedOnboarding:   false,
		OnboardingProviderChoice: "ollama",
		CompanionMuted:           boolPtr(true),
		Providers: map[string]*ProviderSettings{
			"shared": {
				APIKey:      "project-key",
				AccessToken: "project-access-token",
			},
		},
		ToolPermissions: map[string]bool{
			"run": true,
		},
		ProviderAllow:             []string{"project-provider"},
		ModelAllow:                []string{"project-model"},
		Limits:                    NumericLimits{MaxRetries: 5},
		Remote:                    RemoteSettings{Mode: remoteModeSelfHosted, ConnectAddr: "example.internal:8443", TokenSource: remoteTokenSourceEnv, Enabled: true},
		remoteModeSet:             true,
		remoteConnectAddrSet:      true,
		remoteTokenSourceSet:      true,
		remoteEnabledSet:          true,
		hasCompletedOnboardingSet: true,
	}

	env := &Config{
		DefaultProvider: "env-provider",
		Companion:       CompanionStoredState{Name: "env-name", HatchedAt: 300},
		CompanionMuted:  boolPtr(false),
		ToolPermissions: map[string]bool{
			"shell": true,
		},
		Limits:               NumericLimits{RequestTimeoutSeconds: 10},
		Remote:               RemoteSettings{Mode: remoteModeP2P, ListenAddr: "0.0.0.0:9000", TokenSource: remoteTokenSourceKeychain, Enabled: false},
		remoteModeSet:        true,
		remoteListenAddrSet:  true,
		remoteTokenSourceSet: true,
		remoteEnabledSet:     true,
	}

	merged := Merge(defaults, global, project, env)

	if got, want := merged.DefaultProvider, "env-provider"; got != want {
		t.Fatalf("default provider mismatch: got %q want %q", got, want)
	}
	if got, want := merged.DefaultModel, "project-model"; got != want {
		t.Fatalf("default model mismatch: got %q want %q", got, want)
	}
	if got := merged.Providers["shared"].APIKey; got != "project-key" {
		t.Fatalf("provider overlay mismatch: got %q", got)
	}
	if got := merged.Providers["shared"].AuthToken; got != "" {
		t.Fatalf("provider auth token should be cleared by project override, got %q", got)
	}
	if got := merged.Providers["shared"].AccessToken; got != "project-access-token" {
		t.Fatalf("provider access token mismatch: got %q", got)
	}
	if !merged.ToolPermissions["read"] {
		t.Fatalf("expected inherited read permission to be true")
	}
	if !merged.ToolPermissions["run"] {
		t.Fatalf("expected project layer to override run permission")
	}
	if !merged.ToolPermissions["shell"] {
		t.Fatalf("expected env layer permission to be applied")
	}
	if got := merged.Limits.MaxRetries; got != 5 {
		t.Fatalf("max retries mismatch: got %d", got)
	}
	if got := merged.Limits.RequestTimeoutSeconds; got != 10 {
		t.Fatalf("timeout mismatch: got %d", got)
	}
	if got, want := merged.Remote.Mode, remoteModeP2P; got != want {
		t.Fatalf("remote mode mismatch: got %q want %q", got, want)
	}
	if got, want := merged.Remote.ListenAddr, "0.0.0.0:9000"; got != want {
		t.Fatalf("remote listen addr mismatch: got %q want %q", got, want)
	}
	if got, want := merged.Remote.ConnectAddr, "example.internal:8443"; got != want {
		t.Fatalf("remote connect addr mismatch: got %q want %q", got, want)
	}
	if got, want := merged.Remote.TokenSource, remoteTokenSourceKeychain; got != want {
		t.Fatalf("remote token source mismatch: got %q want %q", got, want)
	}
	if merged.Remote.Enabled {
		t.Fatalf("expected env remote enabled override to be false")
	}
	if got, want := merged.ProviderAllow[0], "project-provider"; got != want {
		t.Fatalf("provider allow mismatch: got %q want %q", got, want)
	}
	if got, want := merged.Companion.Name, "env-name"; got != want {
		t.Fatalf("companion name mismatch: got %q want %q", got, want)
	}
	if got, want := merged.Companion.Personality, "project-personality"; got != want {
		t.Fatalf("companion personality mismatch: got %q want %q", got, want)
	}
	if got, want := merged.Companion.HatchedAt, int64(300); got != want {
		t.Fatalf("companion hatched_at mismatch: got %d want %d", got, want)
	}
	if merged.CompanionMuted == nil || *merged.CompanionMuted {
		t.Fatalf("expected env companion mute setting to be false")
	}
	if merged.HasCompletedOnboarding {
		t.Fatalf("expected project onboarding completion to override true to false")
	}
	if got, want := merged.OnboardingProviderChoice, "ollama"; got != want {
		t.Fatalf("onboarding provider choice mismatch: got %q want %q", got, want)
	}
}

func TestValidateFailures(t *testing.T) {
	cfg := &Config{
		DefaultProvider: "",
		DefaultModel:    "model-a",
		ProviderAllow:   []string{"provider-a"},
		ModelAllow:      []string{"model-b"},
		Companion: CompanionStoredState{
			Name:        "   ",
			Personality: strings.Repeat("x", maxCompanionPersonalityLength+1),
			HatchedAt:   -1,
		},
		ToolPermissions: map[string]bool{"": true},
		Limits: NumericLimits{
			MaxRetries:            99,
			MaxTokens:             0,
			RequestTimeoutSeconds: -1,
		},
		Remote: RemoteSettings{
			Mode:        "unsupported",
			TokenSource: "mystery",
			Enabled:     true,
		},
		remoteModeSet:        true,
		remoteTokenSourceSet: true,
		remoteEnabledSet:     true,
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatalf("expected validation error")
	}

	message := err.Error()
	wants := []string{
		"default_provider must not be empty",
		"default_model \"model-a\" is not in model_allow",
		"tool permission name must not be empty",
		"limits.max_retries must be between 0 and 20",
		"limits.max_tokens must be between 1 and 1000000",
		"limits.request_timeout_seconds must be between 1 and 3600",
		"remote.mode must be one of",
		"remote.token_source must be one of",
		"companion.hatched_at must be non-negative",
		"companion.name must not be blank",
		"companion.personality must be at most",
	}
	for _, want := range wants {
		if !strings.Contains(message, want) {
			t.Fatalf("validation error missing %q in %q", want, message)
		}
	}
}

func TestValidateRemoteConfigRules(t *testing.T) {
	t.Run("disabled mode rejects extra fields", func(t *testing.T) {
		cfg := NewDefaultConfig()
		cfg.Remote = RemoteSettings{
			Mode:       remoteModeDisabled,
			ListenAddr: "127.0.0.1:7777",
			Enabled:    true,
		}
		cfg.remoteModeSet = true
		cfg.remoteListenAddrSet = true
		cfg.remoteEnabledSet = true

		err := cfg.Validate()
		if err == nil {
			t.Fatalf("expected validation error")
		}
		message := err.Error()
		if !strings.Contains(message, "remote.enabled must be false when remote.mode is disabled") {
			t.Fatalf("expected disabled-mode enabled flag error, got %q", message)
		}
		if !strings.Contains(message, "remote.listen_addr, remote.connect_addr, and remote.token_source must be empty when remote.mode is disabled") {
			t.Fatalf("expected disabled-mode field error, got %q", message)
		}
	})

	t.Run("self-hosted mode requires endpoint", func(t *testing.T) {
		cfg := NewDefaultConfig()
		cfg.Remote = RemoteSettings{Mode: remoteModeSelfHosted, Enabled: true}
		cfg.remoteModeSet = true
		cfg.remoteEnabledSet = true

		err := cfg.Validate()
		if err == nil {
			t.Fatalf("expected validation error")
		}
		if !strings.Contains(err.Error(), "remote.listen_addr or remote.connect_addr is required when remote.mode is self_hosted") {
			t.Fatalf("expected endpoint requirement error, got %q", err.Error())
		}
	})

	t.Run("local mode rejects connect_addr", func(t *testing.T) {
		cfg := NewDefaultConfig()
		cfg.Remote = RemoteSettings{Mode: remoteModeLocal, ConnectAddr: "example.internal:8080", Enabled: true}
		cfg.remoteModeSet = true
		cfg.remoteConnectAddrSet = true
		cfg.remoteEnabledSet = true

		err := cfg.Validate()
		if err == nil {
			t.Fatalf("expected validation error")
		}
		if !strings.Contains(err.Error(), "remote.connect_addr must be empty when remote.mode is local") {
			t.Fatalf("expected local connect_addr error, got %q", err.Error())
		}
	})
}

func TestLoadLayered_RemoteEnvOverrides(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)

	projectPath := filepath.Join(project, ".alliecode", "config.yaml")
	writeFile(t, projectPath, `
default_provider: openai
default_model: gpt-4o-mini
remote:
  mode: local
  enabled: true
  listen_addr: 127.0.0.1:7777
`)

	t.Setenv("ALLIECODE_REMOTE_MODE", "self_hosted")
	t.Setenv("ALLIECODE_REMOTE_CONNECT_ADDR", "gateway.internal:9443")
	t.Setenv("ALLIECODE_REMOTE_LISTEN_ADDR", "")
	t.Setenv("ALLIECODE_REMOTE_TOKEN_SOURCE", "env")
	t.Setenv("ALLIECODE_REMOTE_ENABLED", "true")

	cfg, err := LoadLayered(project)
	if err != nil {
		t.Fatalf("LoadLayered failed: %v", err)
	}

	if got, want := cfg.Remote.Mode, remoteModeSelfHosted; got != want {
		t.Fatalf("remote mode mismatch: got %q want %q", got, want)
	}
	if got := cfg.Remote.ListenAddr; got != "" {
		t.Fatalf("expected listen_addr to be cleared by env, got %q", got)
	}
	if got, want := cfg.Remote.ConnectAddr, "gateway.internal:9443"; got != want {
		t.Fatalf("remote connect addr mismatch: got %q want %q", got, want)
	}
	if got, want := cfg.Remote.TokenSource, remoteTokenSourceEnv; got != want {
		t.Fatalf("remote token source mismatch: got %q want %q", got, want)
	}
	if !cfg.Remote.Enabled {
		t.Fatalf("expected remote enabled true")
	}
}

func TestLoadLayered_RemoteEmptyEnabledEnvFails(t *testing.T) {
	t.Setenv("ALLIECODE_REMOTE_ENABLED", "")

	_, err := envOverrides()
	if err == nil {
		t.Fatalf("expected parse error for empty remote enabled env")
	}
	if !strings.Contains(err.Error(), "ALLIECODE_REMOTE_ENABLED") {
		t.Fatalf("expected remote enabled env parse error, got %v", err)
	}
}

func TestLoadLayeredPrecedence(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)

	globalPath := filepath.Join(home, ".config", "alliecode", "config.yaml")
	projectPath := filepath.Join(project, ".alliecode", "config.yaml")

	writeFile(t, globalPath, `
default_provider: global
default_model: global-model
has_completed_onboarding: true
onboarding_provider_choice: openai
companion:
  name: global-companion
  personality: global-personality
  hatched_at: 111
companion_muted: true
limits:
  max_retries: 4
tool_permissions:
  run: false
`)

	writeFile(t, projectPath, `
default_model: project-model
has_completed_onboarding: false
onboarding_provider_choice: ollama
companion:
  personality: project-personality
companion_muted: false
limits:
  max_retries: 7
tool_permissions:
  run: true
`)

	t.Setenv("ALLIECODE_DEFAULT_PROVIDER", "env-provider")
	t.Setenv("ALLIECODE_MAX_RETRIES", "9")
	t.Setenv("ALLIECODE_COMPANION_NAME", "env-companion")
	t.Setenv("ALLIECODE_COMPANION_HATCHED_AT", "222")
	t.Setenv("ALLIECODE_COMPANION_MUTED", "true")

	cfg, err := LoadLayered(project)
	if err != nil {
		t.Fatalf("LoadLayered failed: %v", err)
	}

	if got, want := cfg.DefaultProvider, "env-provider"; got != want {
		t.Fatalf("provider mismatch: got %q want %q", got, want)
	}
	if got, want := cfg.DefaultModel, "project-model"; got != want {
		t.Fatalf("model mismatch: got %q want %q", got, want)
	}
	if got, want := cfg.Limits.MaxRetries, 9; got != want {
		t.Fatalf("max retries mismatch: got %d want %d", got, want)
	}
	if got, want := cfg.Companion.Name, "env-companion"; got != want {
		t.Fatalf("companion name mismatch: got %q want %q", got, want)
	}
	if got, want := cfg.Companion.Personality, "project-personality"; got != want {
		t.Fatalf("companion personality mismatch: got %q want %q", got, want)
	}
	if got, want := cfg.Companion.HatchedAt, int64(222); got != want {
		t.Fatalf("companion hatched_at mismatch: got %d want %d", got, want)
	}
	if cfg.CompanionMuted == nil || !*cfg.CompanionMuted {
		t.Fatalf("expected companion muted env override to be true")
	}
	if cfg.HasCompletedOnboarding {
		t.Fatalf("expected project to override onboarding completion to false")
	}
	if got, want := cfg.OnboardingProviderChoice, "ollama"; got != want {
		t.Fatalf("onboarding provider choice mismatch: got %q want %q", got, want)
	}
	if got := cfg.ToolPermissions["run"]; !got {
		t.Fatalf("project tool permissions should override global")
	}
	if got, want := cfg.path, projectPath; got != want {
		t.Fatalf("path mismatch: got %q want %q", got, want)
	}
}

func TestSaveLoadPreservesCompanionState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	cfg := NewDefaultConfig()
	cfg.path = path
	cfg.Companion = CompanionStoredState{
		Name:        "Mochi",
		Personality: "Cheerful and focused",
		HatchedAt:   12345,
	}
	cfg.CompanionMuted = boolPtr(true)
	cfg.HasCompletedOnboarding = true
	cfg.OnboardingProviderChoice = "anthropic"
	cfg.Providers["anthropic"] = &ProviderSettings{AuthToken: "token-123"}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got, want := loaded.Companion.Name, cfg.Companion.Name; got != want {
		t.Fatalf("companion name mismatch: got %q want %q", got, want)
	}
	if got, want := loaded.Companion.Personality, cfg.Companion.Personality; got != want {
		t.Fatalf("companion personality mismatch: got %q want %q", got, want)
	}
	if got, want := loaded.Companion.HatchedAt, cfg.Companion.HatchedAt; got != want {
		t.Fatalf("companion hatched_at mismatch: got %d want %d", got, want)
	}
	if loaded.CompanionMuted == nil || !*loaded.CompanionMuted {
		t.Fatalf("expected companion muted to persist as true")
	}
	if !loaded.HasCompletedOnboarding {
		t.Fatalf("expected has_completed_onboarding to persist as true")
	}
	if got, want := loaded.OnboardingProviderChoice, "anthropic"; got != want {
		t.Fatalf("onboarding provider choice mismatch: got %q want %q", got, want)
	}
	if got := loaded.Providers["anthropic"].AuthToken; got != "token-123" {
		t.Fatalf("provider auth token mismatch: got %q", got)
	}
}

func TestLoadLayered_LocalFirstProviderWhenUnset(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)

	projectPath := filepath.Join(project, ".alliecode", "config.yaml")
	writeFile(t, projectPath, `
default_model: project-model
`)

	cfg, err := LoadLayered(project)
	if err != nil {
		t.Fatalf("LoadLayered failed: %v", err)
	}

	if got, want := cfg.DefaultProvider, "ollama"; got != want {
		t.Fatalf("default provider mismatch: got %q want %q", got, want)
	}
}

func TestLoadLayered_ExplicitProviderPreventsLocalFirstOverride(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)

	projectPath := filepath.Join(project, ".alliecode", "config.yaml")
	writeFile(t, projectPath, `
default_provider: anthropic
default_model: project-model
`)

	cfg, err := LoadLayered(project)
	if err != nil {
		t.Fatalf("LoadLayered failed: %v", err)
	}

	if got, want := cfg.DefaultProvider, "anthropic"; got != want {
		t.Fatalf("default provider mismatch: got %q want %q", got, want)
	}
}

func TestProviderCredentialHelpers(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.DefaultProvider = "anthropic"
	cfg.Providers["anthropic"] = &ProviderSettings{AuthToken: "auth-token"}
	cfg.Providers["openai"] = &ProviderSettings{APIKey: "api-key"}
	cfg.Providers["azure"] = &ProviderSettings{AccessToken: "access-token"}

	if !cfg.HasProviderCredentials("anthropic") {
		t.Fatalf("expected auth_token to count as credentials")
	}
	if !cfg.HasProviderCredentials("openai") {
		t.Fatalf("expected api_key to count as credentials")
	}
	if !cfg.HasProviderCredentials("azure") {
		t.Fatalf("expected access_token to count as credentials")
	}
	if cfg.HasProviderCredentials("missing") {
		t.Fatalf("expected missing provider to have no credentials")
	}
	if !cfg.HasDefaultProviderCredentials() {
		t.Fatalf("expected default provider credentials to be detected")
	}

	cfg.DefaultProvider = "missing"
	if cfg.HasDefaultProviderCredentials() {
		t.Fatalf("expected missing default provider credentials to be false")
	}
}

func TestPersistedAccountProviderReadinessHelpers(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.DefaultProvider = "anthropic"
	cfg.Providers["anthropic"] = &ProviderSettings{AuthToken: "token"}

	summary := cfg.DefaultPersistedAccountProviderReadiness("  dev@acme  ")
	if summary.Provider != "anthropic" {
		t.Fatalf("Provider = %q, want anthropic", summary.Provider)
	}
	if summary.Account != "dev@acme" {
		t.Fatalf("Account = %q, want dev@acme", summary.Account)
	}
	if !summary.AccountPersisted {
		t.Fatalf("expected account to be persisted")
	}
	if !summary.ProviderConfigured {
		t.Fatalf("expected provider to be configured")
	}
	if !summary.ProviderReady {
		t.Fatalf("expected provider readiness to be true")
	}

	notReady := cfg.PersistedAccountProviderReadiness("openai", "")
	if notReady.ProviderReady {
		t.Fatalf("expected readiness false without account/provider credentials")
	}
	if notReady.AccountPersisted {
		t.Fatalf("expected account persisted to be false for empty account")
	}
	if notReady.ProviderConfigured {
		t.Fatalf("expected provider configured false for missing credentials")
	}
}

func TestProviderCredentialPresencePrefersEnvAndConfig(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.Providers["openai"] = &ProviderSettings{}

	presence := cfg.ProviderCredentialPresence("openai")
	if presence.HasCredentials {
		t.Fatalf("expected no credentials without config/env")
	}
	if presence.FromConfig || presence.FromEnv {
		t.Fatalf("expected no config/env credential sources, got %+v", presence)
	}

	t.Setenv("OPENAI_API_KEY", "env-openai-key")
	presence = cfg.ProviderCredentialPresence("openai")
	if !presence.HasCredentials || !presence.FromEnv {
		t.Fatalf("expected env credentials to be detected, got %+v", presence)
	}
	if presence.FromConfig {
		t.Fatalf("expected config source false for env-only credentials")
	}

	cfg.Providers["openai"] = &ProviderSettings{APIKey: "config-openai-key"}
	presence = cfg.ProviderCredentialPresence("openai")
	if !presence.HasCredentials || !presence.FromEnv || !presence.FromConfig {
		t.Fatalf("expected both config+env credentials to be detected, got %+v", presence)
	}

	if !cfg.HasProviderCredentials("openai") {
		t.Fatalf("expected HasProviderCredentials to honor env/config precedence")
	}
}

func TestProviderCredentialPresenceAnthropicTokenVars(t *testing.T) {
	cfg := NewDefaultConfig()
	presence := cfg.ProviderCredentialPresence("anthropic")
	if len(presence.EnvVars) != 3 {
		t.Fatalf("expected anthropic env vars, got %+v", presence.EnvVars)
	}
	t.Setenv("ANTHROPIC_ACCESS_TOKEN", "env-access")
	presence = cfg.ProviderCredentialPresence("anthropic")
	if !presence.FromEnv || !presence.HasCredentials {
		t.Fatalf("expected anthropic env token to satisfy credentials, got %+v", presence)
	}
}

func TestLoadLayeredValidationFailure(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("ALLIECODE_DEFAULT_PROVIDER", "env-provider")
	t.Setenv("ALLIECODE_MODEL_ALLOW", "other-model")

	_, err := LoadLayered(project)
	if err == nil {
		t.Fatalf("expected validation failure")
	}
	if !strings.Contains(err.Error(), "default_model") {
		t.Fatalf("expected default_model validation error, got %v", err)
	}
}

func TestLoadLayered_EnvEmptyStringClearsValues(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)

	globalPath := filepath.Join(home, ".config", "alliecode", "config.yaml")
	projectPath := filepath.Join(project, ".alliecode", "config.yaml")

	writeFile(t, globalPath, `
default_provider: global
default_model: global-model
provider_allow:
  - global
model_allow:
  - global-model
tool_permissions:
  read: true
companion:
  name: Global Buddy
  personality: Always helpful
`)

	writeFile(t, projectPath, `
default_provider: project
default_model: project-model
provider_allow:
  - project
model_allow:
  - project-model
tool_permissions:
  run: true
companion:
  name: Project Buddy
  personality: Project personality
`)

	t.Setenv("ALLIECODE_DEFAULT_PROVIDER", "")
	t.Setenv("ALLIECODE_DEFAULT_MODEL", "")
	t.Setenv("ALLIECODE_PROVIDER_ALLOW", "")
	t.Setenv("ALLIECODE_MODEL_ALLOW", "")
	t.Setenv("ALLIECODE_TOOL_PERMISSIONS", "")
	t.Setenv("ALLIECODE_COMPANION_NAME", "")
	t.Setenv("ALLIECODE_COMPANION_PERSONALITY", "")

	_, err := LoadLayered(project)
	if err == nil {
		t.Fatalf("expected validation failure from cleared defaults")
	}

	message := err.Error()
	if !strings.Contains(message, "default_provider must not be empty") {
		t.Fatalf("expected cleared provider error, got %q", message)
	}
	if !strings.Contains(message, "default_model must not be empty") {
		t.Fatalf("expected cleared model error, got %q", message)
	}
}

func TestLoadLayered_EnvCompanionHatchedAtZeroOverrides(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)

	globalPath := filepath.Join(home, ".config", "alliecode", "config.yaml")
	projectPath := filepath.Join(project, ".alliecode", "config.yaml")

	writeFile(t, globalPath, `
companion:
  hatched_at: 42
`)

	writeFile(t, projectPath, `
companion:
  hatched_at: 7
`)

	t.Setenv("ALLIECODE_COMPANION_HATCHED_AT", "0")

	cfg, err := LoadLayered(project)
	if err != nil {
		t.Fatalf("LoadLayered failed: %v", err)
	}

	if got := cfg.Companion.HatchedAt; got != 0 {
		t.Fatalf("expected env override to set hatched_at to 0, got %d", got)
	}
}

func TestLoadLayered_EmptyNumericEnvFails(t *testing.T) {
	t.Setenv("ALLIECODE_MAX_RETRIES", "")

	_, err := envOverrides()
	if err == nil {
		t.Fatalf("expected parse error for empty numeric env")
	}
	if !strings.Contains(err.Error(), "ALLIECODE_MAX_RETRIES") {
		t.Fatalf("expected max retries env parse error, got %v", err)
	}
}

func TestMergePrecedence_ExplicitEmptyCompanionMutedClearsValue(t *testing.T) {
	base := &Config{CompanionMuted: boolPtr(true)}
	override := &Config{companionMutedSet: true}

	merged := Merge(base, override)
	if merged.CompanionMuted != nil {
		t.Fatalf("expected companion muted to be cleared, got %v", *merged.CompanionMuted)
	}
}

func TestLoadLayered_EnvEmptyCompanionMutedClearsValue(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)

	projectPath := filepath.Join(project, ".alliecode", "config.yaml")
	writeFile(t, projectPath, `
default_provider: openai
default_model: gpt-4o-mini
companion_muted: true
`)

	t.Setenv("ALLIECODE_COMPANION_MUTED", "")

	cfg, err := LoadLayered(project)
	if err != nil {
		t.Fatalf("LoadLayered failed: %v", err)
	}
	if cfg.CompanionMuted != nil {
		t.Fatalf("expected empty env var to clear companion_muted")
	}
}

func TestMergeIncludesTerminalAndIDEFields(t *testing.T) {
	base := &Config{TerminalProfile: "vscode", IDEEditor: "vscode"}
	override := &Config{
		TerminalProfile:    "ghostty",
		IDEEditor:          "cursor",
		terminalProfileSet: true,
		ideEditorSet:       true,
	}

	merged := Merge(base, override)
	if got, want := merged.TerminalProfile, "ghostty"; got != want {
		t.Fatalf("terminal profile mismatch: got %q want %q", got, want)
	}
	if got, want := merged.IDEEditor, "cursor"; got != want {
		t.Fatalf("ide editor mismatch: got %q want %q", got, want)
	}
}

func TestEnvOverridesTerminalAndIDE(t *testing.T) {
	t.Setenv("ALLIECODE_TERMINAL_PROFILE", "iterm")
	t.Setenv("ALLIECODE_IDE_EDITOR", "zed")

	cfg, err := envOverrides()
	if err != nil {
		t.Fatalf("envOverrides failed: %v", err)
	}
	if got, want := cfg.TerminalProfile, "iterm"; got != want {
		t.Fatalf("terminal profile env mismatch: got %q want %q", got, want)
	}
	if got, want := cfg.IDEEditor, "zed"; got != want {
		t.Fatalf("ide editor env mismatch: got %q want %q", got, want)
	}
}

func TestEnvOverridesRuntimeStartupAndMigration(t *testing.T) {
	t.Setenv("ALLIECODE_MIGRATION_VERSION", "7")
	t.Setenv("ALLIECODE_RUNTIME_SURFACE", "cli")
	t.Setenv("ALLIECODE_RUNTIME_COMMAND_SURFACE", "repl")
	t.Setenv("ALLIECODE_STARTUP_HYDRATION_MODE", "strict")
	t.Setenv("ALLIECODE_STARTUP_STRICT_HYDRATION", "true")

	cfg, err := envOverrides()
	if err != nil {
		t.Fatalf("envOverrides failed: %v", err)
	}
	if cfg.MigrationVersion != 7 {
		t.Fatalf("MigrationVersion = %d, want 7", cfg.MigrationVersion)
	}
	if cfg.Runtime.Surface != "cli" || cfg.Runtime.CommandSurface != "repl" {
		t.Fatalf("unexpected runtime settings: %+v", cfg.Runtime)
	}
	if cfg.Startup.HydrationMode != "strict" || !cfg.Startup.StrictHydration {
		t.Fatalf("unexpected startup settings: %+v", cfg.Startup)
	}
}

func TestValidateRuntimeStartupAndMigrationFields(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.Runtime.Surface = " "
	cfg.runtimeSurfaceSet = true
	cfg.Startup.HydrationMode = "invalid"
	cfg.startupHydrationModeSet = true
	cfg.MigrationVersion = -1

	err := cfg.Validate()
	if err == nil {
		t.Fatalf("expected validation error")
	}
	msg := err.Error()
	for _, want := range []string{"runtime.surface must not be blank", "startup.hydration_mode must be one of", "migration_version must be non-negative"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("missing %q in %q", want, msg)
		}
	}
}

func TestApplyMigrationPatchesAndLayerDiagnostics(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.path = "/tmp/project/.alliecode/config.yaml"
	applied := ApplyMigrationPatches(cfg, DefaultMigrationPatches())
	if len(applied) == 0 {
		t.Fatalf("expected migration patches to apply")
	}
	if cfg.MigrationVersion < 1 || cfg.Startup.HydrationMode == "" || cfg.Runtime.Surface == "" {
		t.Fatalf("expected migration patches to set defaults, got %+v", cfg)
	}

	diag := cfg.LayerDiagnostics("/tmp/project")
	if diag.LayeringMode == "" || diag.HydrationMode == "" {
		t.Fatalf("expected diagnostics to include layering/hydration fields, got %+v", diag)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	trimmed := strings.TrimSpace(content) + "\n"
	if err := os.WriteFile(path, []byte(trimmed), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
}

func boolPtr(v bool) *bool {
	return &v
}
