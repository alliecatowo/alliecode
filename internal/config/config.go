// Package config manages AllieCode's configuration loading and persistence.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	defaultProvider               = "openai"
	defaultLayeredProvider        = "ollama"
	defaultModel                  = "gpt-4o-mini"
	defaultMaxRetries             = 2
	defaultMaxTokens              = 4096
	defaultRequestTimeoutSeconds  = 60
	maxCompanionNameLength        = 80
	maxCompanionPersonalityLength = 500
	remoteModeDisabled            = "disabled"
	remoteModeLocal               = "local"
	remoteModeSelfHosted          = "self_hosted"
	remoteModeP2P                 = "p2p"
	remoteTokenSourceNone         = "none"
	remoteTokenSourceEnv          = "env"
	remoteTokenSourceConfig       = "config"
	remoteTokenSourceKeychain     = "keychain"
)

// Config holds AllieCode's runtime configuration.
type Config struct {
	DefaultProvider          string                       `yaml:"default_provider"`
	DefaultModel             string                       `yaml:"default_model"`
	HasCompletedOnboarding   bool                         `yaml:"has_completed_onboarding,omitempty"`
	OnboardingProviderChoice string                       `yaml:"onboarding_provider_choice,omitempty"`
	Providers                map[string]*ProviderSettings `yaml:"providers,omitempty"`
	ProviderAllow            []string                     `yaml:"provider_allow,omitempty"`
	ModelAllow               []string                     `yaml:"model_allow,omitempty"`
	ToolPermissions          map[string]bool              `yaml:"tool_permissions,omitempty"`
	Limits                   NumericLimits                `yaml:"limits,omitempty"`
	Remote                   RemoteSettings               `yaml:"remote,omitempty"`
	Companion                CompanionStoredState         `yaml:"companion,omitempty"`
	CompanionMuted           *bool                        `yaml:"companion_muted,omitempty"`
	TerminalProfile          string                       `yaml:"terminal_profile,omitempty"`
	IDEEditor                string                       `yaml:"ide_editor,omitempty"`
	MigrationVersion         int                          `yaml:"migration_version,omitempty"`
	Runtime                  RuntimeSettings              `yaml:"runtime,omitempty"`
	Startup                  StartupSettings              `yaml:"startup,omitempty"`

	path string // file path for saving

	defaultProviderSet          bool
	defaultModelSet             bool
	hasCompletedOnboardingSet   bool
	onboardingProviderChoiceSet bool
	providerAllowSet            bool
	modelAllowSet               bool
	toolPermissionsSet          bool
	limitsMaxRetriesSet         bool
	limitsMaxTokensSet          bool
	limitsTimeoutSecondsSet     bool
	remoteModeSet               bool
	remoteListenAddrSet         bool
	remoteConnectAddrSet        bool
	remoteTokenSourceSet        bool
	remoteEnabledSet            bool
	companionNameSet            bool
	companionPersonalitySet     bool
	companionHatchedAtSet       bool
	companionMutedSet           bool
	terminalProfileSet          bool
	ideEditorSet                bool
	migrationVersionSet         bool
	runtimeSurfaceSet           bool
	runtimeCommandSurfaceSet    bool
	startupHydrationModeSet     bool
	startupHydrationStrictSet   bool
}

func (c *Config) UnmarshalYAML(value *yaml.Node) error {
	type rawConfig Config
	var decoded rawConfig
	if err := value.Decode(&decoded); err != nil {
		return err
	}
	*c = Config(decoded)

	keys := topLevelKeys(value)
	if _, ok := keys["default_provider"]; ok {
		c.defaultProviderSet = true
	}
	if _, ok := keys["default_model"]; ok {
		c.defaultModelSet = true
	}
	if _, ok := keys["has_completed_onboarding"]; ok {
		c.hasCompletedOnboardingSet = true
	}
	if _, ok := keys["onboarding_provider_choice"]; ok {
		c.onboardingProviderChoiceSet = true
	}
	if _, ok := keys["provider_allow"]; ok {
		c.providerAllowSet = true
	}
	if _, ok := keys["model_allow"]; ok {
		c.modelAllowSet = true
	}
	if _, ok := keys["tool_permissions"]; ok {
		c.toolPermissionsSet = true
	}
	if _, ok := keys["companion_muted"]; ok {
		c.companionMutedSet = true
	}
	if _, ok := keys["terminal_profile"]; ok {
		c.terminalProfileSet = true
	}
	if _, ok := keys["ide_editor"]; ok {
		c.ideEditorSet = true
	}
	if _, ok := keys["migration_version"]; ok {
		c.migrationVersionSet = true
	}

	if limitsNode, ok := keys["limits"]; ok {
		for k := range topLevelKeys(limitsNode) {
			switch k {
			case "max_retries":
				c.limitsMaxRetriesSet = true
			case "max_tokens":
				c.limitsMaxTokensSet = true
			case "request_timeout_seconds":
				c.limitsTimeoutSecondsSet = true
			}
		}
	}

	if remoteNode, ok := keys["remote"]; ok {
		for k := range topLevelKeys(remoteNode) {
			switch k {
			case "mode":
				c.remoteModeSet = true
			case "listen_addr":
				c.remoteListenAddrSet = true
			case "connect_addr":
				c.remoteConnectAddrSet = true
			case "token_source":
				c.remoteTokenSourceSet = true
			case "enabled":
				c.remoteEnabledSet = true
			}
		}
	}

	if companionNode, ok := keys["companion"]; ok {
		for k := range topLevelKeys(companionNode) {
			switch k {
			case "name":
				c.companionNameSet = true
			case "personality":
				c.companionPersonalitySet = true
			case "hatched_at":
				c.companionHatchedAtSet = true
			}
		}
	}

	if runtimeNode, ok := keys["runtime"]; ok {
		for k := range topLevelKeys(runtimeNode) {
			switch k {
			case "surface":
				c.runtimeSurfaceSet = true
			case "command_surface":
				c.runtimeCommandSurfaceSet = true
			}
		}
	}

	if startupNode, ok := keys["startup"]; ok {
		for k := range topLevelKeys(startupNode) {
			switch k {
			case "hydration_mode":
				c.startupHydrationModeSet = true
			case "strict_hydration":
				c.startupHydrationStrictSet = true
			}
		}
	}

	return nil
}

// CompanionStoredState stores persisted companion profile attributes.
type CompanionStoredState struct {
	Name        string `yaml:"name,omitempty"`
	Personality string `yaml:"personality,omitempty"`
	HatchedAt   int64  `yaml:"hatched_at,omitempty"`
}

// NumericLimits holds numeric guardrails for runtime behavior.
type NumericLimits struct {
	MaxRetries            int `yaml:"max_retries,omitempty"`
	MaxTokens             int `yaml:"max_tokens,omitempty"`
	RequestTimeoutSeconds int `yaml:"request_timeout_seconds,omitempty"`
}

type RuntimeSettings struct {
	Surface        string `yaml:"surface,omitempty"`
	CommandSurface string `yaml:"command_surface,omitempty"`
}

type StartupSettings struct {
	HydrationMode   string `yaml:"hydration_mode,omitempty"`
	StrictHydration bool   `yaml:"strict_hydration,omitempty"`
}

// RemoteSettings holds remote session transport configuration.
type RemoteSettings struct {
	Mode        string `yaml:"mode,omitempty"`
	ListenAddr  string `yaml:"listen_addr,omitempty"`
	ConnectAddr string `yaml:"connect_addr,omitempty"`
	TokenSource string `yaml:"token_source,omitempty"`
	Enabled     bool   `yaml:"enabled,omitempty"`
}

// ProviderSettings holds provider-specific configuration.
type ProviderSettings struct {
	APIKey      string            `yaml:"api_key,omitempty"`
	AuthToken   string            `yaml:"auth_token,omitempty"`
	AccessToken string            `yaml:"access_token,omitempty"`
	BaseURL     string            `yaml:"base_url,omitempty"`
	OrgID       string            `yaml:"org_id,omitempty"`
	Options     map[string]string `yaml:"options,omitempty"`
}

type PersistedAccountProviderReadiness struct {
	Provider           string `json:"provider"`
	Account            string `json:"account,omitempty"`
	AccountPersisted   bool   `json:"account_persisted"`
	ProviderConfigured bool   `json:"provider_configured"`
	ProviderReady      bool   `json:"provider_ready"`
}

type LayerDiagnostics struct {
	EffectivePath     string
	GlobalPath        string
	ProjectPath       string
	LayeringMode      string
	DefaultProvider   string
	DefaultModel      string
	ProviderFromEnv   bool
	ModelFromEnv      bool
	ValidationPassed  bool
	ValidationError   string
	RemoteMode        string
	HydrationMode     string
	StrictHydration   bool
	MigrationVersion  int
	ProviderAllowSize int
	ModelAllowSize    int
}

type MigrationPatch struct {
	Description string
	Apply       func(*Config) bool
}

// ProviderCredentialPresence summarizes where provider credentials were found.
type ProviderCredentialPresence struct {
	Provider       string
	FromConfig     bool
	FromEnv        bool
	HasCredentials bool
	EnvVars        []string
}

// Load reads configuration from the given path, or the default location if empty.
func Load(path string) (*Config, error) {
	if path == "" {
		resolved, err := globalConfigPath()
		if err != nil {
			return NewDefaultConfig(), nil
		}
		path = resolved
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := NewDefaultConfig()
			cfg.path = path
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	cfg = *Merge(NewDefaultConfig(), &cfg)
	cfg.path = path
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// NewDefaultConfig returns default config values.
func NewDefaultConfig() *Config {
	return &Config{
		DefaultProvider: defaultProvider,
		DefaultModel:    defaultModel,
		Providers:       make(map[string]*ProviderSettings),
		ToolPermissions: make(map[string]bool),
		Limits: NumericLimits{
			MaxRetries:            defaultMaxRetries,
			MaxTokens:             defaultMaxTokens,
			RequestTimeoutSeconds: defaultRequestTimeoutSeconds,
		},
	}
}

// ConventionalPaths returns config lookup paths in precedence order.
func ConventionalPaths(projectDir string) (string, string, error) {
	globalPath, err := globalConfigPath()
	if err != nil {
		return "", "", err
	}
	projectPath := filepath.Join(projectDir, ".alliecode", "config.yaml")
	return globalPath, projectPath, nil
}

// LoadLayered reads defaults + global + project + env overrides.
func LoadLayered(projectDir string) (*Config, error) {
	globalPath, projectPath, err := ConventionalPaths(projectDir)
	if err != nil {
		return nil, err
	}

	globalCfg, err := loadLayer(globalPath)
	if err != nil {
		return nil, fmt.Errorf("loading global config: %w", err)
	}
	projectCfg, err := loadLayer(projectPath)
	if err != nil {
		return nil, fmt.Errorf("loading project config: %w", err)
	}
	envCfg, err := envOverrides()
	if err != nil {
		return nil, err
	}

	cfg := Merge(NewDefaultConfig(), globalCfg, projectCfg, envCfg)
	if !hasExplicitDefaultProvider(globalCfg, projectCfg, envCfg) {
		cfg.DefaultProvider = defaultLayeredProvider
	}
	cfg.path = projectPath
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Merge combines config layers from lowest to highest precedence.
func Merge(layers ...*Config) *Config {
	merged := &Config{
		Providers:       make(map[string]*ProviderSettings),
		ToolPermissions: make(map[string]bool),
	}

	for _, layer := range layers {
		if layer == nil {
			continue
		}
		if layer.defaultProviderSet || layer.DefaultProvider != "" {
			merged.DefaultProvider = layer.DefaultProvider
		}
		if layer.defaultModelSet || layer.DefaultModel != "" {
			merged.DefaultModel = layer.DefaultModel
		}
		if layer.hasCompletedOnboardingSet || layer.HasCompletedOnboarding {
			merged.HasCompletedOnboarding = layer.HasCompletedOnboarding
		}
		if layer.onboardingProviderChoiceSet || layer.OnboardingProviderChoice != "" {
			merged.OnboardingProviderChoice = layer.OnboardingProviderChoice
		}
		if layer.providerAllowSet || len(layer.ProviderAllow) > 0 {
			merged.ProviderAllow = append([]string(nil), layer.ProviderAllow...)
		}
		if layer.modelAllowSet || len(layer.ModelAllow) > 0 {
			merged.ModelAllow = append([]string(nil), layer.ModelAllow...)
		}
		if layer.toolPermissionsSet {
			merged.ToolPermissions = make(map[string]bool)
		}
		for name, provider := range layer.Providers {
			if provider == nil {
				continue
			}
			merged.Providers[name] = cloneProviderSettings(provider)
		}
		for tool, allowed := range layer.ToolPermissions {
			merged.ToolPermissions[tool] = allowed
		}

		if layer.limitsMaxRetriesSet || layer.Limits.MaxRetries != 0 {
			merged.Limits.MaxRetries = layer.Limits.MaxRetries
		}
		if layer.limitsMaxTokensSet || layer.Limits.MaxTokens != 0 {
			merged.Limits.MaxTokens = layer.Limits.MaxTokens
		}
		if layer.limitsTimeoutSecondsSet || layer.Limits.RequestTimeoutSeconds != 0 {
			merged.Limits.RequestTimeoutSeconds = layer.Limits.RequestTimeoutSeconds
		}

		if layer.remoteModeSet || layer.Remote.Mode != "" {
			merged.Remote.Mode = layer.Remote.Mode
		}
		if layer.remoteListenAddrSet || layer.Remote.ListenAddr != "" {
			merged.Remote.ListenAddr = layer.Remote.ListenAddr
		}
		if layer.remoteConnectAddrSet || layer.Remote.ConnectAddr != "" {
			merged.Remote.ConnectAddr = layer.Remote.ConnectAddr
		}
		if layer.remoteTokenSourceSet || layer.Remote.TokenSource != "" {
			merged.Remote.TokenSource = layer.Remote.TokenSource
		}
		if layer.remoteEnabledSet || layer.Remote.Enabled {
			merged.Remote.Enabled = layer.Remote.Enabled
		}

		if layer.companionNameSet || layer.Companion.Name != "" {
			merged.Companion.Name = layer.Companion.Name
		}
		if layer.companionPersonalitySet || layer.Companion.Personality != "" {
			merged.Companion.Personality = layer.Companion.Personality
		}
		if layer.companionHatchedAtSet || layer.Companion.HatchedAt != 0 {
			merged.Companion.HatchedAt = layer.Companion.HatchedAt
		}
		if layer.companionMutedSet || layer.CompanionMuted != nil {
			merged.CompanionMuted = cloneBoolPointer(layer.CompanionMuted)
		}
		if layer.terminalProfileSet || layer.TerminalProfile != "" {
			merged.TerminalProfile = layer.TerminalProfile
		}
		if layer.ideEditorSet || layer.IDEEditor != "" {
			merged.IDEEditor = layer.IDEEditor
		}
		if layer.migrationVersionSet || layer.MigrationVersion != 0 {
			merged.MigrationVersion = layer.MigrationVersion
		}
		if layer.runtimeSurfaceSet || layer.Runtime.Surface != "" {
			merged.Runtime.Surface = layer.Runtime.Surface
		}
		if layer.runtimeCommandSurfaceSet || layer.Runtime.CommandSurface != "" {
			merged.Runtime.CommandSurface = layer.Runtime.CommandSurface
		}
		if layer.startupHydrationModeSet || layer.Startup.HydrationMode != "" {
			merged.Startup.HydrationMode = layer.Startup.HydrationMode
		}
		if layer.startupHydrationStrictSet || layer.Startup.StrictHydration {
			merged.Startup.StrictHydration = layer.Startup.StrictHydration
		}
	}

	if merged.Providers == nil {
		merged.Providers = make(map[string]*ProviderSettings)
	}
	if merged.ToolPermissions == nil {
		merged.ToolPermissions = make(map[string]bool)
	}

	return merged
}

// Validate ensures configuration is coherent and safe.
func (c *Config) Validate() error {
	var errs []error

	if strings.TrimSpace(c.DefaultProvider) == "" {
		errs = append(errs, errors.New("default_provider must not be empty"))
	}
	if strings.TrimSpace(c.DefaultModel) == "" {
		errs = append(errs, errors.New("default_model must not be empty"))
	}
	if c.OnboardingProviderChoice != "" {
		if strings.TrimSpace(c.OnboardingProviderChoice) == "" {
			errs = append(errs, errors.New("onboarding_provider_choice must not be blank"))
		}
	}

	for name := range c.Providers {
		if strings.TrimSpace(name) == "" {
			errs = append(errs, errors.New("provider name must not be empty"))
		}
	}

	if len(c.ProviderAllow) > 0 {
		if !contains(c.ProviderAllow, c.DefaultProvider) {
			errs = append(errs, fmt.Errorf("default_provider %q is not in provider_allow", c.DefaultProvider))
		}
	}
	if len(c.ModelAllow) > 0 {
		if !contains(c.ModelAllow, c.DefaultModel) {
			errs = append(errs, fmt.Errorf("default_model %q is not in model_allow", c.DefaultModel))
		}
	}

	for tool := range c.ToolPermissions {
		if strings.TrimSpace(tool) == "" {
			errs = append(errs, errors.New("tool permission name must not be empty"))
		}
	}

	if c.Limits.MaxRetries < 0 || c.Limits.MaxRetries > 20 {
		errs = append(errs, fmt.Errorf("limits.max_retries must be between 0 and 20, got %d", c.Limits.MaxRetries))
	}
	if c.Limits.MaxTokens <= 0 || c.Limits.MaxTokens > 1_000_000 {
		errs = append(errs, fmt.Errorf("limits.max_tokens must be between 1 and 1000000, got %d", c.Limits.MaxTokens))
	}
	if c.Limits.RequestTimeoutSeconds <= 0 || c.Limits.RequestTimeoutSeconds > 3600 {
		errs = append(errs, fmt.Errorf("limits.request_timeout_seconds must be between 1 and 3600, got %d", c.Limits.RequestTimeoutSeconds))
	}

	remoteMode := strings.TrimSpace(c.Remote.Mode)
	remoteListenAddr := strings.TrimSpace(c.Remote.ListenAddr)
	remoteConnectAddr := strings.TrimSpace(c.Remote.ConnectAddr)
	remoteTokenSource := strings.TrimSpace(c.Remote.TokenSource)
	hasRemoteConfig := c.remoteModeSet || c.remoteListenAddrSet || c.remoteConnectAddrSet || c.remoteTokenSourceSet || c.remoteEnabledSet || remoteMode != "" || remoteListenAddr != "" || remoteConnectAddr != "" || remoteTokenSource != "" || c.Remote.Enabled
	if hasRemoteConfig {
		if remoteMode == "" {
			errs = append(errs, errors.New("remote.mode must be set when remote config is present"))
		} else if !isSupportedRemoteMode(remoteMode) {
			errs = append(errs, fmt.Errorf("remote.mode must be one of %q, %q, %q, %q, got %q", remoteModeDisabled, remoteModeLocal, remoteModeSelfHosted, remoteModeP2P, remoteMode))
		}

		if remoteListenAddr == "" && c.remoteListenAddrSet {
			errs = append(errs, errors.New("remote.listen_addr must not be blank"))
		}
		if remoteConnectAddr == "" && c.remoteConnectAddrSet {
			errs = append(errs, errors.New("remote.connect_addr must not be blank"))
		}

		if remoteTokenSource != "" && !isSupportedRemoteTokenSource(remoteTokenSource) {
			errs = append(errs, fmt.Errorf("remote.token_source must be one of %q, %q, %q, %q, got %q", remoteTokenSourceNone, remoteTokenSourceEnv, remoteTokenSourceConfig, remoteTokenSourceKeychain, remoteTokenSource))
		}

		switch remoteMode {
		case remoteModeDisabled:
			if c.Remote.Enabled {
				errs = append(errs, errors.New("remote.enabled must be false when remote.mode is disabled"))
			}
			if remoteListenAddr != "" || remoteConnectAddr != "" || remoteTokenSource != "" {
				errs = append(errs, errors.New("remote.listen_addr, remote.connect_addr, and remote.token_source must be empty when remote.mode is disabled"))
			}
		case remoteModeLocal:
			if remoteConnectAddr != "" {
				errs = append(errs, errors.New("remote.connect_addr must be empty when remote.mode is local"))
			}
		case remoteModeSelfHosted, remoteModeP2P:
			if remoteListenAddr == "" && remoteConnectAddr == "" {
				errs = append(errs, fmt.Errorf("remote.listen_addr or remote.connect_addr is required when remote.mode is %s", remoteMode))
			}
		}
	}

	if c.Companion.HatchedAt < 0 {
		errs = append(errs, fmt.Errorf("companion.hatched_at must be non-negative, got %d", c.Companion.HatchedAt))
	}
	if c.Companion.Name != "" {
		trimmed := strings.TrimSpace(c.Companion.Name)
		if trimmed == "" {
			errs = append(errs, errors.New("companion.name must not be blank"))
		} else if len([]rune(trimmed)) > maxCompanionNameLength {
			errs = append(errs, fmt.Errorf("companion.name must be at most %d characters", maxCompanionNameLength))
		}
	}
	if c.Companion.Personality != "" {
		trimmed := strings.TrimSpace(c.Companion.Personality)
		if trimmed == "" {
			errs = append(errs, errors.New("companion.personality must not be blank"))
		} else if len([]rune(trimmed)) > maxCompanionPersonalityLength {
			errs = append(errs, fmt.Errorf("companion.personality must be at most %d characters", maxCompanionPersonalityLength))
		}
	}
	if c.TerminalProfile != "" && strings.TrimSpace(c.TerminalProfile) == "" {
		errs = append(errs, errors.New("terminal_profile must not be blank"))
	}
	if c.IDEEditor != "" && strings.TrimSpace(c.IDEEditor) == "" {
		errs = append(errs, errors.New("ide_editor must not be blank"))
	}
	if c.MigrationVersion < 0 {
		errs = append(errs, fmt.Errorf("migration_version must be non-negative, got %d", c.MigrationVersion))
	}

	runtimeSurface := strings.TrimSpace(c.Runtime.Surface)
	commandSurface := strings.TrimSpace(c.Runtime.CommandSurface)
	if c.runtimeSurfaceSet || c.runtimeCommandSurfaceSet || runtimeSurface != "" || commandSurface != "" {
		if c.runtimeSurfaceSet && runtimeSurface == "" {
			errs = append(errs, errors.New("runtime.surface must not be blank"))
		}
		if c.runtimeCommandSurfaceSet && commandSurface == "" {
			errs = append(errs, errors.New("runtime.command_surface must not be blank"))
		}
	}

	hydrationMode := strings.TrimSpace(c.Startup.HydrationMode)
	if c.startupHydrationModeSet || hydrationMode != "" || c.startupHydrationStrictSet {
		if hydrationMode != "" {
			switch hydrationMode {
			case "legacy", "compat", "strict":
			default:
				errs = append(errs, fmt.Errorf("startup.hydration_mode must be one of %q, %q, %q, got %q", "legacy", "compat", "strict", hydrationMode))
			}
		} else if c.startupHydrationModeSet {
			errs = append(errs, errors.New("startup.hydration_mode must not be blank"))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// Get retrieves a config value by dot-separated key.
func (c *Config) Get(key string) (string, bool) {
	switch key {
	case "provider":
		return c.DefaultProvider, c.DefaultProvider != ""
	case "model":
		return c.DefaultModel, c.DefaultModel != ""
	default:
		return "", false
	}
}

// Set updates a config value by dot-separated key.
func (c *Config) Set(key, value string) {
	switch key {
	case "provider":
		c.DefaultProvider = value
	case "model":
		c.DefaultModel = value
	}
}

// Save writes the configuration to disk.
func (c *Config) Save() error {
	if c.path == "" {
		return fmt.Errorf("no config path set")
	}
	if err := c.Validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(c.path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0o644)
}

// GetProviderSettings returns settings for the named provider, or nil.
func (c *Config) GetProviderSettings(name string) *ProviderSettings {
	if ps, ok := c.Providers[name]; ok {
		return ps
	}
	return &ProviderSettings{}
}

// HasProviderCredentials reports whether a provider has usable credentials.
func (c *Config) HasProviderCredentials(name string) bool {
	return c.ProviderCredentialPresence(name).HasCredentials
}

// HasDefaultProviderCredentials reports whether the default provider is configured.
func (c *Config) HasDefaultProviderCredentials() bool {
	provider := strings.TrimSpace(c.DefaultProvider)
	if provider == "" {
		return false
	}
	return c.HasProviderCredentials(provider)
}

// ProviderCredentialPresence returns provider credential presence across config + env.
func (c *Config) ProviderCredentialPresence(name string) ProviderCredentialPresence {
	providerName := strings.ToLower(strings.TrimSpace(name))
	presence := ProviderCredentialPresence{
		Provider: providerName,
		EnvVars:  providerCredentialEnvVars(providerName),
	}
	provider := c.GetProviderSettings(providerName)
	if provider != nil {
		presence.FromConfig = strings.TrimSpace(provider.APIKey) != "" ||
			strings.TrimSpace(provider.AuthToken) != "" ||
			strings.TrimSpace(provider.AccessToken) != ""
	}
	for _, envVar := range presence.EnvVars {
		if v, ok := os.LookupEnv(envVar); ok && strings.TrimSpace(v) != "" {
			presence.FromEnv = true
			break
		}
	}
	presence.HasCredentials = presence.FromConfig || presence.FromEnv
	return presence
}

func providerCredentialEnvVars(provider string) []string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "anthropic":
		return []string{"ANTHROPIC_AUTH_TOKEN", "ANTHROPIC_ACCESS_TOKEN", "ANTHROPIC_API_KEY"}
	case "openai":
		return []string{"OPENAI_API_KEY"}
	case "gemini":
		return []string{"GEMINI_API_KEY"}
	case "ollama":
		return nil
	default:
		return nil
	}
}

func (c *Config) PersistedAccountProviderReadiness(provider, account string) PersistedAccountProviderReadiness {
	resolvedProvider := strings.TrimSpace(provider)
	if resolvedProvider == "" {
		resolvedProvider = strings.TrimSpace(c.DefaultProvider)
	}
	resolvedAccount := strings.TrimSpace(account)
	providerConfigured := false
	if resolvedProvider != "" {
		providerConfigured = c.HasProviderCredentials(resolvedProvider)
	}
	accountPersisted := resolvedAccount != ""
	return PersistedAccountProviderReadiness{
		Provider:           resolvedProvider,
		Account:            resolvedAccount,
		AccountPersisted:   accountPersisted,
		ProviderConfigured: providerConfigured,
		ProviderReady:      accountPersisted && providerConfigured,
	}
}

func (c *Config) DefaultPersistedAccountProviderReadiness(account string) PersistedAccountProviderReadiness {
	return c.PersistedAccountProviderReadiness(c.DefaultProvider, account)
}

func globalConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".config", "alliecode", "config.yaml"), nil
}

func loadLayer(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return &cfg, nil
}

func envOverrides() (*Config, error) {
	cfg := &Config{}

	if provider, ok := os.LookupEnv("ALLIECODE_DEFAULT_PROVIDER"); ok {
		cfg.DefaultProvider = strings.TrimSpace(provider)
		cfg.defaultProviderSet = true
	}
	if model, ok := os.LookupEnv("ALLIECODE_DEFAULT_MODEL"); ok {
		cfg.DefaultModel = strings.TrimSpace(model)
		cfg.defaultModelSet = true
	}

	if raw, ok := os.LookupEnv("ALLIECODE_PROVIDER_ALLOW"); ok {
		cfg.ProviderAllow = splitCSV(raw)
		cfg.providerAllowSet = true
	}
	if raw, ok := os.LookupEnv("ALLIECODE_MODEL_ALLOW"); ok {
		cfg.ModelAllow = splitCSV(raw)
		cfg.modelAllowSet = true
	}

	if raw, ok := os.LookupEnv("ALLIECODE_TOOL_PERMISSIONS"); ok {
		cfg.toolPermissionsSet = true
		if strings.TrimSpace(raw) != "" {
			permissions, err := parseToolPermissions(raw)
			if err != nil {
				return nil, err
			}
			cfg.ToolPermissions = permissions
		}
	}

	if raw, ok := os.LookupEnv("ALLIECODE_MAX_RETRIES"); ok {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return nil, errors.New("invalid ALLIECODE_MAX_RETRIES: value is empty")
		}
		value, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid ALLIECODE_MAX_RETRIES: %w", err)
		}
		cfg.Limits.MaxRetries = value
		cfg.limitsMaxRetriesSet = true
	}
	if raw, ok := os.LookupEnv("ALLIECODE_MAX_TOKENS"); ok {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return nil, errors.New("invalid ALLIECODE_MAX_TOKENS: value is empty")
		}
		value, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid ALLIECODE_MAX_TOKENS: %w", err)
		}
		cfg.Limits.MaxTokens = value
		cfg.limitsMaxTokensSet = true
	}
	if raw, ok := os.LookupEnv("ALLIECODE_REQUEST_TIMEOUT_SECONDS"); ok {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return nil, errors.New("invalid ALLIECODE_REQUEST_TIMEOUT_SECONDS: value is empty")
		}
		value, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid ALLIECODE_REQUEST_TIMEOUT_SECONDS: %w", err)
		}
		cfg.Limits.RequestTimeoutSeconds = value
		cfg.limitsTimeoutSecondsSet = true
	}

	if raw, ok := os.LookupEnv("ALLIECODE_REMOTE_MODE"); ok {
		cfg.Remote.Mode = strings.TrimSpace(raw)
		cfg.remoteModeSet = true
	}
	if raw, ok := os.LookupEnv("ALLIECODE_REMOTE_LISTEN_ADDR"); ok {
		cfg.Remote.ListenAddr = strings.TrimSpace(raw)
		cfg.remoteListenAddrSet = true
	}
	if raw, ok := os.LookupEnv("ALLIECODE_REMOTE_CONNECT_ADDR"); ok {
		cfg.Remote.ConnectAddr = strings.TrimSpace(raw)
		cfg.remoteConnectAddrSet = true
	}
	if raw, ok := os.LookupEnv("ALLIECODE_REMOTE_TOKEN_SOURCE"); ok {
		cfg.Remote.TokenSource = strings.TrimSpace(raw)
		cfg.remoteTokenSourceSet = true
	}
	if raw, ok := os.LookupEnv("ALLIECODE_REMOTE_ENABLED"); ok {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return nil, errors.New("invalid ALLIECODE_REMOTE_ENABLED: value is empty")
		}
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid ALLIECODE_REMOTE_ENABLED: %w", err)
		}
		cfg.Remote.Enabled = value
		cfg.remoteEnabledSet = true
	}

	if raw, ok := os.LookupEnv("ALLIECODE_COMPANION_NAME"); ok {
		cfg.Companion.Name = strings.TrimSpace(raw)
		cfg.companionNameSet = true
	}
	if raw, ok := os.LookupEnv("ALLIECODE_COMPANION_PERSONALITY"); ok {
		cfg.Companion.Personality = strings.TrimSpace(raw)
		cfg.companionPersonalitySet = true
	}
	if raw, ok := os.LookupEnv("ALLIECODE_COMPANION_HATCHED_AT"); ok {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return nil, errors.New("invalid ALLIECODE_COMPANION_HATCHED_AT: value is empty")
		}
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid ALLIECODE_COMPANION_HATCHED_AT: %w", err)
		}
		cfg.Companion.HatchedAt = value
		cfg.companionHatchedAtSet = true
	}
	if raw, ok := os.LookupEnv("ALLIECODE_COMPANION_MUTED"); ok {
		cfg.companionMutedSet = true
		raw = strings.TrimSpace(raw)
		if raw != "" {
			value, err := strconv.ParseBool(raw)
			if err != nil {
				return nil, fmt.Errorf("invalid ALLIECODE_COMPANION_MUTED: %w", err)
			}
			cfg.CompanionMuted = &value
		}
	}
	if raw, ok := os.LookupEnv("ALLIECODE_TERMINAL_PROFILE"); ok {
		cfg.TerminalProfile = strings.TrimSpace(raw)
		cfg.terminalProfileSet = true
	}
	if raw, ok := os.LookupEnv("ALLIECODE_IDE_EDITOR"); ok {
		cfg.IDEEditor = strings.TrimSpace(raw)
		cfg.ideEditorSet = true
	}
	if raw, ok := os.LookupEnv("ALLIECODE_MIGRATION_VERSION"); ok {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return nil, errors.New("invalid ALLIECODE_MIGRATION_VERSION: value is empty")
		}
		value, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid ALLIECODE_MIGRATION_VERSION: %w", err)
		}
		cfg.MigrationVersion = value
		cfg.migrationVersionSet = true
	}
	if raw, ok := os.LookupEnv("ALLIECODE_RUNTIME_SURFACE"); ok {
		cfg.Runtime.Surface = strings.TrimSpace(raw)
		cfg.runtimeSurfaceSet = true
	}
	if raw, ok := os.LookupEnv("ALLIECODE_RUNTIME_COMMAND_SURFACE"); ok {
		cfg.Runtime.CommandSurface = strings.TrimSpace(raw)
		cfg.runtimeCommandSurfaceSet = true
	}
	if raw, ok := os.LookupEnv("ALLIECODE_STARTUP_HYDRATION_MODE"); ok {
		cfg.Startup.HydrationMode = strings.TrimSpace(raw)
		cfg.startupHydrationModeSet = true
	}
	if raw, ok := os.LookupEnv("ALLIECODE_STARTUP_STRICT_HYDRATION"); ok {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return nil, errors.New("invalid ALLIECODE_STARTUP_STRICT_HYDRATION: value is empty")
		}
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid ALLIECODE_STARTUP_STRICT_HYDRATION: %w", err)
		}
		cfg.Startup.StrictHydration = value
		cfg.startupHydrationStrictSet = true
	}

	return cfg, nil
}

func parseToolPermissions(raw string) (map[string]bool, error) {
	permissions := make(map[string]bool)
	entries := splitCSV(raw)
	for _, entry := range entries {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid tool permission %q; expected name=true|false", entry)
		}
		name := strings.TrimSpace(parts[0])
		if name == "" {
			return nil, fmt.Errorf("invalid tool permission %q; tool name is empty", entry)
		}
		allowed, err := strconv.ParseBool(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid tool permission %q: %w", entry, err)
		}
		permissions[name] = allowed
	}
	return permissions, nil
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func cloneProviderSettings(in *ProviderSettings) *ProviderSettings {
	clone := &ProviderSettings{
		APIKey:      in.APIKey,
		AuthToken:   in.AuthToken,
		AccessToken: in.AccessToken,
		BaseURL:     in.BaseURL,
		OrgID:       in.OrgID,
	}
	if len(in.Options) > 0 {
		clone.Options = make(map[string]string, len(in.Options))
		for k, v := range in.Options {
			clone.Options[k] = v
		}
	}
	return clone
}

func (c *Config) LayerDiagnostics(projectDir string) LayerDiagnostics {
	globalPath, projectPath, _ := ConventionalPaths(projectDir)
	d := LayerDiagnostics{
		EffectivePath:     c.path,
		GlobalPath:        globalPath,
		ProjectPath:       projectPath,
		DefaultProvider:   strings.TrimSpace(c.DefaultProvider),
		DefaultModel:      strings.TrimSpace(c.DefaultModel),
		RemoteMode:        strings.TrimSpace(c.Remote.Mode),
		HydrationMode:     strings.TrimSpace(c.Startup.HydrationMode),
		StrictHydration:   c.Startup.StrictHydration,
		MigrationVersion:  c.MigrationVersion,
		ProviderAllowSize: len(c.ProviderAllow),
		ModelAllowSize:    len(c.ModelAllow),
	}
	if d.HydrationMode == "" {
		d.HydrationMode = "compat"
	}
	if strings.TrimSpace(c.path) == strings.TrimSpace(projectPath) {
		d.LayeringMode = "layered_project"
	} else {
		d.LayeringMode = "single_file"
	}
	if _, ok := os.LookupEnv("ALLIECODE_DEFAULT_PROVIDER"); ok {
		d.ProviderFromEnv = true
	}
	if _, ok := os.LookupEnv("ALLIECODE_DEFAULT_MODEL"); ok {
		d.ModelFromEnv = true
	}
	if err := c.Validate(); err != nil {
		d.ValidationPassed = false
		d.ValidationError = err.Error()
	} else {
		d.ValidationPassed = true
	}
	return d
}

func ApplyMigrationPatches(cfg *Config, patches []MigrationPatch) []string {
	if cfg == nil || len(patches) == 0 {
		return nil
	}
	applied := make([]string, 0, len(patches))
	for _, patch := range patches {
		if patch.Apply == nil {
			continue
		}
		if patch.Apply(cfg) {
			desc := strings.TrimSpace(patch.Description)
			if desc != "" {
				applied = append(applied, desc)
			}
		}
	}
	return applied
}

func DefaultMigrationPatches() []MigrationPatch {
	return []MigrationPatch{
		{
			Description: "set startup hydration mode default",
			Apply: func(cfg *Config) bool {
				if strings.TrimSpace(cfg.Startup.HydrationMode) != "" {
					return false
				}
				cfg.Startup.HydrationMode = "compat"
				cfg.startupHydrationModeSet = true
				return true
			},
		},
		{
			Description: "set runtime surface default",
			Apply: func(cfg *Config) bool {
				if strings.TrimSpace(cfg.Runtime.Surface) != "" {
					return false
				}
				cfg.Runtime.Surface = "cli"
				cfg.runtimeSurfaceSet = true
				return true
			},
		},
		{
			Description: "ensure migration version floor",
			Apply: func(cfg *Config) bool {
				if cfg.MigrationVersion >= 1 {
					return false
				}
				cfg.MigrationVersion = 1
				cfg.migrationVersionSet = true
				return true
			},
		},
	}
}

func hasExplicitDefaultProvider(layers ...*Config) bool {
	for _, layer := range layers {
		if layer == nil {
			continue
		}
		if layer.defaultProviderSet {
			return true
		}
		if strings.TrimSpace(layer.DefaultProvider) != "" {
			return true
		}
	}
	return false
}

func cloneBoolPointer(in *bool) *bool {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func isSupportedRemoteMode(mode string) bool {
	switch mode {
	case remoteModeDisabled, remoteModeLocal, remoteModeSelfHosted, remoteModeP2P:
		return true
	default:
		return false
	}
}

func isSupportedRemoteTokenSource(source string) bool {
	switch source {
	case remoteTokenSourceNone, remoteTokenSourceEnv, remoteTokenSourceConfig, remoteTokenSourceKeychain:
		return true
	default:
		return false
	}
}

func topLevelKeys(node *yaml.Node) map[string]*yaml.Node {
	keys := make(map[string]*yaml.Node)
	if node == nil {
		return keys
	}
	target := node
	if target.Kind == yaml.DocumentNode && len(target.Content) > 0 {
		target = target.Content[0]
	}
	if target.Kind != yaml.MappingNode {
		return keys
	}
	for i := 0; i+1 < len(target.Content); i += 2 {
		keyNode := target.Content[i]
		valueNode := target.Content[i+1]
		keys[strings.TrimSpace(keyNode.Value)] = valueNode
	}
	return keys
}
