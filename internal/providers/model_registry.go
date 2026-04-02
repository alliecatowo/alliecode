package providers

import (
	"sort"
	"strings"
)

// CostClass describes relative model cost for routing.
type CostClass string

const (
	CostClassFree   CostClass = "free"
	CostClassLow    CostClass = "low"
	CostClassMedium CostClass = "medium"
	CostClassHigh   CostClass = "high"
)

// ServiceTier describes operator-facing service tier labels.
type ServiceTier string

const (
	ServiceTierDefault  ServiceTier = "default"
	ServiceTierStandard ServiceTier = "standard"
	ServiceTierPremium  ServiceTier = "premium"
)

// ComplexityTier describes where a model is strongest.
type ComplexityTier string

const (
	ComplexityTierFast      ComplexityTier = "fast"
	ComplexityTierBalanced  ComplexityTier = "balanced"
	ComplexityTierReasoning ComplexityTier = "reasoning"
)

// ContextWindowClass groups models by context window capacity.
type ContextWindowClass string

const (
	ContextWindowClassTiny   ContextWindowClass = "tiny"
	ContextWindowClassSmall  ContextWindowClass = "small"
	ContextWindowClassMedium ContextWindowClass = "medium"
	ContextWindowClassLarge  ContextWindowClass = "large"
	ContextWindowClassXLarge ContextWindowClass = "xlarge"
)

// ModelMetadata stores routing-relevant model capabilities.
type ModelMetadata struct {
	Provider            string
	Model               string
	Tier                ComplexityTier
	ServiceTier         ServiceTier
	CostClass           CostClass
	ContextClass        ContextWindowClass
	ContextWindow       int
	MaxOutput           int
	SupportsText        bool
	SupportsImage       bool
	SupportsAudio       bool
	SupportsToolUse     bool
	SupportsTools       bool
	SupportsThinking    bool
	SupportsVision      bool
	SupportsAttachments bool
	CostPerMInput       float64
	CostPerMOutput      float64
}

// CapabilitySummary returns a stable comma-delimited capability list.
func (m ModelMetadata) CapabilitySummary() string {
	capabilities := make([]string, 0, 6)
	if m.SupportsText {
		capabilities = append(capabilities, "text")
	}
	if m.SupportsImage {
		capabilities = append(capabilities, "image")
	}
	if m.SupportsAudio {
		capabilities = append(capabilities, "audio")
	}
	if m.SupportsToolUse || m.SupportsTools {
		capabilities = append(capabilities, "tool_use")
	}
	if m.SupportsVision {
		capabilities = append(capabilities, "vision")
	}
	if m.SupportsAttachments {
		capabilities = append(capabilities, "attachments")
	}
	if len(capabilities) == 0 {
		return "-"
	}
	return strings.Join(capabilities, ",")
}

// ModelMetadataRegistry stores and resolves model metadata.
type ModelMetadataRegistry interface {
	Get(provider, model string) (ModelMetadata, bool)
	FindByModel(model string) (ModelMetadata, bool)
	Providers() []string
	ListByProvider(provider string) []ModelMetadata
	ListByTier(provider string, tier ComplexityTier) []ModelMetadata
	ListByCostClass(provider string, max CostClass) []ModelMetadata
	ListByCostClasses(provider string, classes ...CostClass) []ModelMetadata
	ListByServiceTier(provider string, tier ServiceTier) []ModelMetadata
	ListByMinimumContextWindow(provider string, minimum int) []ModelMetadata
	ListByMinimumContextClass(provider string, class ContextWindowClass) []ModelMetadata
	ContextWindow(provider, model string) (int, bool)
	ContextWindowClass(provider, model string) (ContextWindowClass, bool)
	MaxOutput(provider, model string) (int, bool)
	Pricing(provider, model string) (float64, float64, bool)
	CostClass(provider, model string) (CostClass, bool)
	ServiceTier(provider, model string) (ServiceTier, bool)
}

// StaticModelMetadataRegistry is an in-memory metadata registry.
type StaticModelMetadataRegistry struct {
	byProviderModel map[string]ModelMetadata
	byModel         map[string][]ModelMetadata
}

// NewStaticModelMetadataRegistry creates an in-memory model metadata registry.
func NewStaticModelMetadataRegistry(metadata ...ModelMetadata) *StaticModelMetadataRegistry {
	r := &StaticModelMetadataRegistry{
		byProviderModel: make(map[string]ModelMetadata),
		byModel:         make(map[string][]ModelMetadata),
	}
	r.Register(metadata...)
	return r
}

// Register inserts metadata entries.
func (r *StaticModelMetadataRegistry) Register(metadata ...ModelMetadata) {
	for _, m := range metadata {
		if m.Provider == "" || m.Model == "" {
			continue
		}
		key := providerModelKey(m.Provider, m.Model)
		if m.ContextClass == "" {
			m.ContextClass = classifyContextWindow(m.ContextWindow)
		}
		r.byProviderModel[key] = m
		r.byModel[m.Model] = append(r.byModel[m.Model], m)
	}
}

// Get returns metadata for provider/model.
func (r *StaticModelMetadataRegistry) Get(provider, model string) (ModelMetadata, bool) {
	m, ok := r.byProviderModel[providerModelKey(provider, model)]
	return m, ok
}

// FindByModel resolves metadata by model name.
func (r *StaticModelMetadataRegistry) FindByModel(model string) (ModelMetadata, bool) {
	entries, ok := r.byModel[model]
	if !ok || len(entries) == 0 {
		return ModelMetadata{}, false
	}
	// Keep selection deterministic across providers.
	sorted := make([]ModelMetadata, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Provider < sorted[j].Provider
	})
	return sorted[0], true
}

// Providers returns all providers present in metadata.
func (r *StaticModelMetadataRegistry) Providers() []string {
	providers := make([]string, 0)
	seen := make(map[string]struct{})
	for _, m := range r.byProviderModel {
		if _, ok := seen[m.Provider]; ok {
			continue
		}
		seen[m.Provider] = struct{}{}
		providers = append(providers, m.Provider)
	}
	sort.Strings(providers)
	return providers
}

// ListByProvider returns all metadata for a provider.
func (r *StaticModelMetadataRegistry) ListByProvider(provider string) []ModelMetadata {
	models := make([]ModelMetadata, 0)
	for _, m := range r.byProviderModel {
		if m.Provider == provider {
			models = append(models, m)
		}
	}
	sort.Slice(models, func(i, j int) bool {
		return models[i].Model < models[j].Model
	})
	return models
}

// ListByTier returns models for a provider at a given tier.
func (r *StaticModelMetadataRegistry) ListByTier(provider string, tier ComplexityTier) []ModelMetadata {
	models := make([]ModelMetadata, 0)
	for _, m := range r.byProviderModel {
		if m.Provider == provider && m.Tier == tier {
			models = append(models, m)
		}
	}
	sort.Slice(models, func(i, j int) bool {
		return models[i].Model < models[j].Model
	})
	return models
}

// ListByCostClass returns models up to the given cost class.
func (r *StaticModelMetadataRegistry) ListByCostClass(provider string, max CostClass) []ModelMetadata {
	models := make([]ModelMetadata, 0)
	maxRank := costClassRank(max)
	for _, m := range r.byProviderModel {
		if m.Provider != provider {
			continue
		}
		if max != "" && costClassRank(m.CostClass) > maxRank {
			continue
		}
		models = append(models, m)
	}
	sort.Slice(models, func(i, j int) bool {
		if costClassRank(models[i].CostClass) == costClassRank(models[j].CostClass) {
			return models[i].Model < models[j].Model
		}
		return costClassRank(models[i].CostClass) < costClassRank(models[j].CostClass)
	})
	return models
}

// ListByCostClasses returns models in any provided cost class.
func (r *StaticModelMetadataRegistry) ListByCostClasses(provider string, classes ...CostClass) []ModelMetadata {
	if len(classes) == 0 {
		return r.ListByProvider(provider)
	}
	set := make(map[CostClass]struct{}, len(classes))
	for _, class := range classes {
		set[class] = struct{}{}
	}
	models := make([]ModelMetadata, 0)
	for _, m := range r.byProviderModel {
		if m.Provider != provider {
			continue
		}
		if _, ok := set[m.CostClass]; !ok {
			continue
		}
		models = append(models, m)
	}
	sort.Slice(models, func(i, j int) bool {
		if costClassRank(models[i].CostClass) == costClassRank(models[j].CostClass) {
			return models[i].Model < models[j].Model
		}
		return costClassRank(models[i].CostClass) < costClassRank(models[j].CostClass)
	})
	return models
}

// ListByServiceTier returns models for a given service tier.
func (r *StaticModelMetadataRegistry) ListByServiceTier(provider string, tier ServiceTier) []ModelMetadata {
	models := make([]ModelMetadata, 0)
	for _, m := range r.byProviderModel {
		if m.Provider != provider {
			continue
		}
		if tier != "" && m.ServiceTier != tier {
			continue
		}
		models = append(models, m)
	}
	sort.Slice(models, func(i, j int) bool {
		return models[i].Model < models[j].Model
	})
	return models
}

// ListByMinimumContextWindow returns provider models with required context length.
func (r *StaticModelMetadataRegistry) ListByMinimumContextWindow(provider string, minimum int) []ModelMetadata {
	models := make([]ModelMetadata, 0)
	for _, m := range r.byProviderModel {
		if m.Provider != provider {
			continue
		}
		if minimum > 0 && m.ContextWindow > 0 && m.ContextWindow < minimum {
			continue
		}
		models = append(models, m)
	}
	sort.Slice(models, func(i, j int) bool {
		if models[i].ContextWindow == models[j].ContextWindow {
			return models[i].Model < models[j].Model
		}
		return models[i].ContextWindow > models[j].ContextWindow
	})
	return models
}

// ListByMinimumContextClass returns provider models at or above the class.
func (r *StaticModelMetadataRegistry) ListByMinimumContextClass(provider string, class ContextWindowClass) []ModelMetadata {
	models := make([]ModelMetadata, 0)
	minimumRank := contextClassRank(class)
	for _, m := range r.byProviderModel {
		if m.Provider != provider {
			continue
		}
		if class != "" && contextClassRank(m.ContextClass) < minimumRank {
			continue
		}
		models = append(models, m)
	}
	sort.Slice(models, func(i, j int) bool {
		if contextClassRank(models[i].ContextClass) == contextClassRank(models[j].ContextClass) {
			return models[i].Model < models[j].Model
		}
		return contextClassRank(models[i].ContextClass) > contextClassRank(models[j].ContextClass)
	})
	return models
}

// ContextWindow returns the context window for provider/model.
func (r *StaticModelMetadataRegistry) ContextWindow(provider, model string) (int, bool) {
	m, ok := r.Get(provider, model)
	if !ok {
		return 0, false
	}
	return m.ContextWindow, true
}

// ContextWindowClass returns context window class for provider/model.
func (r *StaticModelMetadataRegistry) ContextWindowClass(provider, model string) (ContextWindowClass, bool) {
	m, ok := r.Get(provider, model)
	if !ok {
		return "", false
	}
	if m.ContextClass != "" {
		return m.ContextClass, true
	}
	return classifyContextWindow(m.ContextWindow), true
}

// MaxOutput returns the max output tokens for provider/model.
func (r *StaticModelMetadataRegistry) MaxOutput(provider, model string) (int, bool) {
	m, ok := r.Get(provider, model)
	if !ok {
		return 0, false
	}
	return m.MaxOutput, true
}

// Pricing returns input/output pricing per million tokens.
func (r *StaticModelMetadataRegistry) Pricing(provider, model string) (float64, float64, bool) {
	m, ok := r.Get(provider, model)
	if !ok {
		return 0, 0, false
	}
	return m.CostPerMInput, m.CostPerMOutput, true
}

// CostClass returns model cost class for provider/model.
func (r *StaticModelMetadataRegistry) CostClass(provider, model string) (CostClass, bool) {
	m, ok := r.Get(provider, model)
	if !ok {
		return "", false
	}
	return m.CostClass, true
}

// ServiceTier returns model service tier for provider/model.
func (r *StaticModelMetadataRegistry) ServiceTier(provider, model string) (ServiceTier, bool) {
	m, ok := r.Get(provider, model)
	if !ok {
		return "", false
	}
	return m.ServiceTier, true
}

func providerModelKey(provider, model string) string {
	return provider + "::" + model
}

func costClassRank(class CostClass) int {
	switch class {
	case CostClassFree:
		return 0
	case CostClassLow:
		return 1
	case CostClassMedium:
		return 2
	case CostClassHigh:
		return 3
	default:
		return 2
	}
}

func contextClassRank(class ContextWindowClass) int {
	switch class {
	case ContextWindowClassTiny:
		return 0
	case ContextWindowClassSmall:
		return 1
	case ContextWindowClassMedium:
		return 2
	case ContextWindowClassLarge:
		return 3
	case ContextWindowClassXLarge:
		return 4
	default:
		return 0
	}
}

func classifyContextWindow(window int) ContextWindowClass {
	switch {
	case window <= 0:
		return ContextWindowClassSmall
	case window <= 8192:
		return ContextWindowClassTiny
	case window <= 65536:
		return ContextWindowClassSmall
	case window <= 200000:
		return ContextWindowClassMedium
	case window <= 512000:
		return ContextWindowClassLarge
	default:
		return ContextWindowClassXLarge
	}
}

func defaultModelMetadata() []ModelMetadata {
	return []ModelMetadata{
		{Provider: "anthropic", Model: "claude-opus-4-20250514", Tier: ComplexityTierReasoning, ServiceTier: ServiceTierPremium, CostClass: CostClassHigh, ContextWindow: 200000, MaxOutput: 32000, SupportsText: true, SupportsImage: true, SupportsAudio: false, SupportsToolUse: true, SupportsTools: true, SupportsThinking: true, SupportsVision: true, SupportsAttachments: true, CostPerMInput: 15.00, CostPerMOutput: 75.00},
		{Provider: "anthropic", Model: "claude-sonnet-4-20250514", Tier: ComplexityTierBalanced, ServiceTier: ServiceTierPremium, CostClass: CostClassMedium, ContextWindow: 200000, MaxOutput: 64000, SupportsText: true, SupportsImage: true, SupportsAudio: false, SupportsToolUse: true, SupportsTools: true, SupportsThinking: true, SupportsVision: true, SupportsAttachments: true, CostPerMInput: 3.00, CostPerMOutput: 15.00},
		{Provider: "anthropic", Model: "claude-haiku-3-5-20241022", Tier: ComplexityTierFast, ServiceTier: ServiceTierStandard, CostClass: CostClassLow, ContextWindow: 200000, MaxOutput: 8192, SupportsText: true, SupportsImage: true, SupportsAudio: false, SupportsToolUse: true, SupportsTools: true, SupportsThinking: true, SupportsVision: true, SupportsAttachments: true, CostPerMInput: 0.80, CostPerMOutput: 4.00},
		{Provider: "gemini", Model: "gemini-2.5-pro", Tier: ComplexityTierReasoning, ServiceTier: ServiceTierPremium, CostClass: CostClassMedium, ContextWindow: 1048576, MaxOutput: 65536, SupportsText: true, SupportsImage: true, SupportsAudio: true, SupportsToolUse: true, SupportsTools: true, SupportsThinking: false, SupportsVision: true, SupportsAttachments: true, CostPerMInput: 1.25, CostPerMOutput: 10.00},
		{Provider: "gemini", Model: "gemini-2.5-flash", Tier: ComplexityTierFast, ServiceTier: ServiceTierStandard, CostClass: CostClassLow, ContextWindow: 1048576, MaxOutput: 65536, SupportsText: true, SupportsImage: true, SupportsAudio: true, SupportsToolUse: true, SupportsTools: true, SupportsThinking: false, SupportsVision: true, SupportsAttachments: true, CostPerMInput: 0.30, CostPerMOutput: 2.50},
		{Provider: "openai", Model: "gpt-4o", Tier: ComplexityTierBalanced, ServiceTier: ServiceTierPremium, CostClass: CostClassMedium, ContextWindow: 128000, MaxOutput: 16384, SupportsText: true, SupportsImage: true, SupportsAudio: true, SupportsToolUse: true, SupportsTools: true, SupportsThinking: false, SupportsVision: true, SupportsAttachments: true, CostPerMInput: 5.00, CostPerMOutput: 15.00},
		{Provider: "openai", Model: "gpt-4o-mini", Tier: ComplexityTierFast, ServiceTier: ServiceTierStandard, CostClass: CostClassLow, ContextWindow: 128000, MaxOutput: 16384, SupportsText: true, SupportsImage: true, SupportsAudio: true, SupportsToolUse: true, SupportsTools: true, SupportsThinking: false, SupportsVision: true, SupportsAttachments: true, CostPerMInput: 0.15, CostPerMOutput: 0.60},
		{Provider: "openai", Model: "o3", Tier: ComplexityTierReasoning, ServiceTier: ServiceTierPremium, CostClass: CostClassHigh, ContextWindow: 200000, MaxOutput: 100000, SupportsText: true, SupportsImage: true, SupportsAudio: false, SupportsToolUse: true, SupportsTools: true, SupportsThinking: false, SupportsVision: true, SupportsAttachments: true, CostPerMInput: 10.00, CostPerMOutput: 40.00},
		{Provider: "openai", Model: "o4-mini", Tier: ComplexityTierFast, ServiceTier: ServiceTierStandard, CostClass: CostClassMedium, ContextWindow: 200000, MaxOutput: 100000, SupportsText: true, SupportsImage: true, SupportsAudio: false, SupportsToolUse: true, SupportsTools: true, SupportsThinking: false, SupportsVision: true, SupportsAttachments: true, CostPerMInput: 1.10, CostPerMOutput: 4.40},
		{Provider: "ollama", Model: "llama3", Tier: ComplexityTierFast, ServiceTier: ServiceTierDefault, CostClass: CostClassFree, ContextWindow: 8192, MaxOutput: 4096, SupportsText: true, SupportsImage: false, SupportsAudio: false, SupportsToolUse: true, SupportsTools: true, SupportsThinking: false, SupportsVision: false, SupportsAttachments: false, CostPerMInput: 0.00, CostPerMOutput: 0.00},
	}
}

// NewDefaultModelMetadataRegistry returns an in-memory registry seeded with defaults.
func NewDefaultModelMetadataRegistry() *StaticModelMetadataRegistry {
	return NewStaticModelMetadataRegistry(defaultModelMetadata()...)
}
