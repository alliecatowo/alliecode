package providers

import (
	"strings"

	"github.com/alliecatowo/alliecode/internal/types"
)

// TaskComplexity describes the expected reasoning effort for a task.
type TaskComplexity int

const (
	TaskComplexitySimple TaskComplexity = iota
	TaskComplexityNormal
	TaskComplexityComplex
)

// RoutingHints allow callers to express runtime constraints.
type RoutingHints struct {
	Provider                string
	Model                   string
	PreferFastModel         bool
	PreferLowCost           bool
	PreferredTier           ComplexityTier
	AllowedCostClass        CostClass
	AllowedCostClasses      []CostClass
	CostClassSelector       string
	ServiceTier             ServiceTier
	DisableProviderFallback bool
	FallbackPriorities      []string
	RequiresText            bool
	RequiresImage           bool
	RequiresAudio           bool
	RequiresToolUse         bool
	RequiresTools           bool
	RequiresThinking        bool
	RequiresVision          bool
	RequiresAttachments     bool
	MinimumContextWindow    int
	MinimumContextClass     ContextWindowClass
}

// RoutingRequest is evaluated by a RoutingPolicy.
type RoutingRequest struct {
	Complexity TaskComplexity
	Hints      RoutingHints
}

// RouteDecision is the selected provider/model pair.
type RouteDecision struct {
	Provider string
	Model    string
	Reason   string
}

// RoutingPolicy chooses provider/model for a task.
type RoutingPolicy interface {
	Select(route types.ModelRoute, req RoutingRequest) RouteDecision
}

// PolicyEngine selects provider/model using route config + capabilities.
type PolicyEngine struct {
	registry ModelMetadataRegistry
	matrix   CapabilityMatrix
}

// NewPolicyEngine creates a capability-aware routing policy engine.
func NewPolicyEngine(registry ModelMetadataRegistry, matrix CapabilityMatrix) *PolicyEngine {
	return &PolicyEngine{registry: registry, matrix: matrix}
}

// Select chooses provider/model by complexity and hints.
func (p *PolicyEngine) Select(route types.ModelRoute, req RoutingRequest) RouteDecision {
	provider := route.Provider
	if req.Hints.Provider != "" {
		provider = req.Hints.Provider
	}

	if req.Hints.Model != "" {
		modelProvider := provider
		if modelProvider == "" {
			if meta, ok := p.registry.FindByModel(req.Hints.Model); ok {
				modelProvider = meta.Provider
			}
		}
		if modelProvider != "" && p.supports(modelProvider, req.Hints.Model, req.Hints) {
			return RouteDecision{Provider: modelProvider, Model: req.Hints.Model, Reason: "explicit model hint"}
		}
	}

	primaryProvider := provider
	if primaryProvider == "" {
		if meta, ok := p.registry.FindByModel(route.Primary); ok {
			primaryProvider = meta.Provider
		}
	}

	if req.Hints.PreferFastModel || req.Complexity == TaskComplexitySimple {
		if route.Fast != "" && primaryProvider != "" && p.supports(primaryProvider, route.Fast, req.Hints) {
			return RouteDecision{Provider: primaryProvider, Model: route.Fast, Reason: "fast model for simple task"}
		}
	}

	if req.Complexity == TaskComplexityComplex && primaryProvider != "" {
		if candidate := p.firstTierCandidate(primaryProvider, ComplexityTierReasoning, req.Hints); candidate != "" {
			return RouteDecision{Provider: primaryProvider, Model: candidate, Reason: "reasoning model for complex task"}
		}
	}

	if route.Primary != "" && primaryProvider != "" && p.supports(primaryProvider, route.Primary, req.Hints) {
		return RouteDecision{Provider: primaryProvider, Model: route.Primary, Reason: "primary route model"}
	}

	if route.Fast != "" && primaryProvider != "" && p.supports(primaryProvider, route.Fast, req.Hints) {
		return RouteDecision{Provider: primaryProvider, Model: route.Fast, Reason: "fast fallback"}
	}

	fallbackOrder := resolveFallbackPriorities(req.Hints)
	for _, fallback := range fallbackOrder {
		switch fallback {
		case "provider":
			if primaryProvider != "" {
				if candidate := p.firstProviderCandidate(primaryProvider, req); candidate != "" {
					return RouteDecision{Provider: primaryProvider, Model: candidate, Reason: "provider capability fallback"}
				}
			}
		case "global":
			if req.Hints.DisableProviderFallback {
				continue
			}
			if fallbackProvider, fallbackModel := p.firstGlobalCandidate(req); fallbackProvider != "" && fallbackModel != "" {
				return RouteDecision{Provider: fallbackProvider, Model: fallbackModel, Reason: "global capability fallback"}
			}
		case "route":
			if route.Fast != "" && route.Provider != "" {
				return RouteDecision{Provider: route.Provider, Model: route.Fast, Reason: "route fallback"}
			}
		}
	}

	if provider == "" && route.Primary != "" {
		if meta, ok := p.registry.FindByModel(route.Primary); ok && p.supports(meta.Provider, route.Primary, req.Hints) {
			return RouteDecision{Provider: meta.Provider, Model: route.Primary, Reason: "provider inferred from model"}
		}
	}

	if route.Fast != "" && route.Provider != "" {
		return RouteDecision{Provider: route.Provider, Model: route.Fast, Reason: "route fallback"}
	}
	return RouteDecision{Provider: route.Provider, Model: route.Primary, Reason: "route default"}
}

func (p *PolicyEngine) supports(provider, model string, hints RoutingHints) bool {
	if hints.ServiceTier != "" {
		if tier, ok := p.registry.ServiceTier(provider, model); !ok || tier != hints.ServiceTier {
			return false
		}
	}
	if !p.costClassAllowed(provider, model, hints) {
		return false
	}
	if hints.MinimumContextClass != "" {
		if class, ok := p.registry.ContextWindowClass(provider, model); !ok || contextClassRank(class) < contextClassRank(hints.MinimumContextClass) {
			return false
		}
	}
	req := CapabilityRequirements{
		NeedsText:        hints.RequiresText,
		NeedsImage:       hints.RequiresImage,
		NeedsAudio:       hints.RequiresAudio,
		NeedsToolUse:     hints.RequiresToolUse,
		NeedsTools:       hints.RequiresTools,
		NeedsThinking:    hints.RequiresThinking,
		NeedsVision:      hints.RequiresVision,
		NeedsAttachments: hints.RequiresAttachments,
		MinContextWindow: hints.MinimumContextWindow,
	}
	return p.matrix.Supports(provider, model, req)
}

func (p *PolicyEngine) costClassAllowed(provider, model string, hints RoutingHints) bool {
	class, ok := p.registry.CostClass(provider, model)
	if !ok {
		return false
	}
	if hints.AllowedCostClass != "" && costClassRank(class) > costClassRank(hints.AllowedCostClass) {
		return false
	}
	if len(hints.AllowedCostClasses) > 0 {
		allowed := false
		for _, c := range hints.AllowedCostClasses {
			if class == c {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}
	selector := strings.ToLower(strings.TrimSpace(hints.CostClassSelector))
	switch selector {
	case "exact":
		if hints.AllowedCostClass == "" {
			return true
		}
		return class == hints.AllowedCostClass
	case "ceiling", "", "inclusive":
		return true
	default:
		return true
	}
}

func resolveFallbackPriorities(hints RoutingHints) []string {
	if len(hints.FallbackPriorities) == 0 {
		return []string{"provider", "global", "route"}
	}
	seen := make(map[string]struct{}, len(hints.FallbackPriorities))
	out := make([]string, 0, len(hints.FallbackPriorities))
	for _, v := range hints.FallbackPriorities {
		norm := strings.ToLower(strings.TrimSpace(v))
		switch norm {
		case "provider", "global", "route":
			if _, ok := seen[norm]; ok {
				continue
			}
			seen[norm] = struct{}{}
			out = append(out, norm)
		}
	}
	if len(out) == 0 {
		return []string{"provider", "global", "route"}
	}
	return out
}

func (p *PolicyEngine) firstGlobalCandidate(req RoutingRequest) (string, string) {
	for _, provider := range p.registry.Providers() {
		if candidate := p.firstProviderCandidate(provider, req); candidate != "" {
			return provider, candidate
		}
	}
	return "", ""
}

func (p *PolicyEngine) firstTierCandidate(provider string, tier ComplexityTier, hints RoutingHints) string {
	for _, m := range p.registry.ListByTier(provider, tier) {
		if p.supports(provider, m.Model, hints) {
			return m.Model
		}
	}
	return ""
}

func (p *PolicyEngine) firstProviderCandidate(provider string, req RoutingRequest) string {
	models := p.registry.ListByProvider(provider)
	bestModel := ""
	bestScore := -1
	for _, m := range models {
		if !p.supports(provider, m.Model, req.Hints) {
			continue
		}
		score := scoreCandidate(m, req)
		if score > bestScore {
			bestScore = score
			bestModel = m.Model
		}
	}
	return bestModel
}

func scoreCandidate(meta ModelMetadata, req RoutingRequest) int {
	score := 0
	if req.Hints.PreferredTier != "" && meta.Tier == req.Hints.PreferredTier {
		score += 40
	}
	if req.Complexity == TaskComplexityComplex && meta.Tier == ComplexityTierReasoning {
		score += 30
	}
	if req.Complexity == TaskComplexitySimple && meta.Tier == ComplexityTierFast {
		score += 30
	}
	if req.Hints.PreferLowCost {
		score += 20 - (costClassRank(meta.CostClass) * 5)
	}
	if req.Hints.MinimumContextWindow > 0 {
		score += min(meta.ContextWindow/10000, 20)
	}
	if req.Hints.MinimumContextClass != "" && contextClassRank(meta.ContextClass) >= contextClassRank(req.Hints.MinimumContextClass) {
		score += 10
	}
	if req.Hints.ServiceTier != "" && meta.ServiceTier == req.Hints.ServiceTier {
		score += 10
	}
	return score
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
