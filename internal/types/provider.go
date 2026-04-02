package types

import "context"

// Provider is the universal interface for LLM backends.
// Each provider (Anthropic, OpenAI, Ollama, etc.) implements this interface,
// translating AllieCode's universal message format to/from their native API.
type Provider interface {
	// Name returns the provider identifier (e.g., "anthropic", "openai", "ollama").
	Name() string

	// Chat sends a chat request and returns a stream of events.
	// The caller should consume the channel until it is closed.
	Chat(ctx context.Context, req ChatRequest) (<-chan StreamEvent, error)

	// ChatSync sends a chat request and waits for the complete response.
	ChatSync(ctx context.Context, req ChatRequest) (*ChatResponse, error)

	// ListModels returns available models from this provider.
	ListModels(ctx context.Context) ([]Model, error)

	// SupportsStreaming returns whether this provider supports streaming responses.
	SupportsStreaming() bool

	// SupportsTools returns whether this provider supports tool/function calling.
	SupportsTools() bool

	// SupportsThinking returns whether this provider supports thinking/reasoning blocks.
	SupportsThinking() bool
}

// ProviderMetadataSource is an optional interface for providers that can
// expose richer model metadata and capabilities.
type ProviderMetadataSource interface {
	ModelMetadata(ctx context.Context) ([]ModelMetadata, error)
}

// ProviderRetryPolicySource is an optional interface for providers that can
// expose retry policy behavior for diagnostics.
type ProviderRetryPolicySource interface {
	RetryPolicy() ProviderRetryPolicy
}

// ChatRequest is the universal request format sent to any provider.
type ChatRequest struct {
	Messages      []Message `json:"messages"`
	Tools         []ToolDef `json:"tools,omitempty"`
	System        string    `json:"system,omitempty"`
	MaxTokens     int       `json:"max_tokens,omitempty"`
	Temperature   *float64  `json:"temperature,omitempty"`
	TopP          *float64  `json:"top_p,omitempty"`
	Model         string    `json:"model,omitempty"`
	StopSequences []string  `json:"stop_sequences,omitempty"`

	ProviderHints ProviderHints `json:"provider_hints,omitempty"`
	Routing       RoutingPolicy `json:"routing,omitempty"`

	// Thinking/reasoning configuration
	EnableThinking bool `json:"enable_thinking,omitempty"`
	ThinkingBudget int  `json:"thinking_budget,omitempty"`
}

// ProviderHints carries provider-specific request hints in a provider-agnostic shape.
type ProviderHints struct {
	FastMode          bool   `json:"fast_mode,omitempty"`
	PreferLowLatency  bool   `json:"prefer_low_latency,omitempty"`
	PreferLowCost     bool   `json:"prefer_low_cost,omitempty"`
	DisableToolChoice bool   `json:"disable_tool_choice,omitempty"`
	CostClassCeiling  string `json:"cost_class_ceiling,omitempty"`
	CostClass         string `json:"cost_class,omitempty"`
	CostClassSelector string `json:"cost_class_selector,omitempty"`
	ContextClass      string `json:"context_class,omitempty"`
	ServiceTier       string `json:"service_tier,omitempty"`
	FallbackPriority  string `json:"fallback_priority,omitempty"`
}

// RoutingPolicy captures routing directives from loop/runtime.
type RoutingPolicy struct {
	Strategy             string   `json:"strategy,omitempty"`
	FallbackChain        []string `json:"fallback_chain,omitempty"`
	FallbackPriorities   []string `json:"fallback_priorities,omitempty"`
	MinimumContextWindow int      `json:"minimum_context_window,omitempty"`
	MinimumContextClass  string   `json:"minimum_context_class,omitempty"`
	RequireCapabilities  []string `json:"require_capabilities,omitempty"`
	CostClassCeiling     string   `json:"cost_class_ceiling,omitempty"`
	AllowedCostClasses   []string `json:"allowed_cost_classes,omitempty"`
	CostClassSelector    string   `json:"cost_class_selector,omitempty"`
	ServiceTier          string   `json:"service_tier,omitempty"`
}

// Model describes an available model from a provider.
type Model struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Provider       string  `json:"provider"`
	ContextWindow  int     `json:"context_window"`
	MaxOutput      int     `json:"max_output"`
	SupportsTools  bool    `json:"supports_tools"`
	SupportsVision bool    `json:"supports_vision"`
	CostPerMInput  float64 `json:"cost_per_m_input,omitempty"`
	CostPerMOutput float64 `json:"cost_per_m_output,omitempty"`
}

// ModelMetadata stores provider-agnostic model capability details.
type ModelMetadata struct {
	Provider            string  `json:"provider"`
	Model               string  `json:"model"`
	Tier                string  `json:"tier,omitempty"`
	ServiceTier         string  `json:"service_tier,omitempty"`
	CostClass           string  `json:"cost_class,omitempty"`
	ContextClass        string  `json:"context_class,omitempty"`
	ContextWindow       int     `json:"context_window,omitempty"`
	MaxOutput           int     `json:"max_output,omitempty"`
	SupportsText        bool    `json:"supports_text,omitempty"`
	SupportsImage       bool    `json:"supports_image,omitempty"`
	SupportsAudio       bool    `json:"supports_audio,omitempty"`
	SupportsToolUse     bool    `json:"supports_tool_use,omitempty"`
	SupportsTools       bool    `json:"supports_tools,omitempty"`
	SupportsThinking    bool    `json:"supports_thinking,omitempty"`
	SupportsVision      bool    `json:"supports_vision,omitempty"`
	SupportsAttachments bool    `json:"supports_attachments,omitempty"`
	CostPerMInput       float64 `json:"cost_per_m_input,omitempty"`
	CostPerMOutput      float64 `json:"cost_per_m_output,omitempty"`
}

// ProviderRetryPolicy exposes retry policy knobs used by a provider.
type ProviderRetryPolicy struct {
	Name          string `json:"name,omitempty"`
	MaxRetries    int    `json:"max_retries,omitempty"`
	BaseBackoffMS int    `json:"base_backoff_ms,omitempty"`
	MaxBackoffMS  int    `json:"max_backoff_ms,omitempty"`
}

// ModelRoute configures which model to use for different task types.
type ModelRoute struct {
	// Primary model for reasoning and complex tasks
	Primary string `yaml:"primary" json:"primary"`

	// Fast model for simple tool calls and quick responses
	Fast string `yaml:"fast,omitempty" json:"fast,omitempty"`

	// Provider to use (if not specified, auto-detect from model name)
	Provider string `yaml:"provider,omitempty" json:"provider,omitempty"`
}

// ProviderConfig holds configuration for a specific provider.
type ProviderConfig struct {
	APIKey  string `yaml:"api_key,omitempty" json:"api_key,omitempty"`
	BaseURL string `yaml:"base_url,omitempty" json:"base_url,omitempty"`
	OrgID   string `yaml:"org_id,omitempty" json:"org_id,omitempty"`

	// Provider-specific options
	Options map[string]string `yaml:"options,omitempty" json:"options,omitempty"`
}
