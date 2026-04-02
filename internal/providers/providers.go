// Package providers implements LLM provider management and multi-model routing.
package providers

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/alliecatowo/alliecode/internal/config"
	"github.com/alliecatowo/alliecode/internal/providers/anthropic"
	"github.com/alliecatowo/alliecode/internal/providers/gemini"
	"github.com/alliecatowo/alliecode/internal/providers/ollama"
	"github.com/alliecatowo/alliecode/internal/providers/openai"
	"github.com/alliecatowo/alliecode/internal/providers/openaicompat"
	"github.com/alliecatowo/alliecode/internal/types"
)

// ProviderValidationStatus reports provider credential/model readiness.
type ProviderValidationStatus struct {
	Provider       string
	Ready          bool
	ModelAvailable bool
	ModelCount     int
	Models         []string
	Reason         string
}

// New creates a provider instance by name.
func New(name string, cfg *config.ProviderSettings) (types.Provider, error) {
	if cfg == nil {
		cfg = &config.ProviderSettings{}
	}

	switch name {
	case "anthropic":
		return anthropic.New(cfg.APIKey, cfg.AuthToken, cfg.AccessToken, cfg.BaseURL, cfg.Options)
	case "gemini":
		return gemini.New(cfg.APIKey, cfg.BaseURL, cfg.Options)
	case "openai":
		return openai.New(cfg.APIKey, cfg.BaseURL, cfg.OrgID, cfg.Options)
	case "ollama":
		baseURL := cfg.BaseURL
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		return ollama.New(baseURL, cfg.Options)
	case "openai-compat", "openaicompat":
		if cfg.BaseURL == "" {
			return nil, fmt.Errorf("openai-compat provider requires base_url")
		}
		return openaicompat.New(cfg.APIKey, cfg.BaseURL, cfg.Options)
	default:
		return nil, fmt.Errorf("unknown provider: %q (available: anthropic, gemini, openai, ollama, openai-compat)", name)
	}
}

// ValidateProviderModelReadiness checks provider setup and model availability.
func ValidateProviderModelReadiness(ctx context.Context, providerName string, cfg *config.ProviderSettings, targetModel string) ProviderValidationStatus {
	providerName = strings.ToLower(strings.TrimSpace(providerName))
	modelName := strings.TrimSpace(targetModel)
	status := ProviderValidationStatus{Provider: providerName}
	if providerName == "" {
		status.Reason = "provider is empty"
		return status
	}
	if _, ok := LookupModelMetadata(providerName, modelName); modelName != "" && !ok {
		status.Reason = fmt.Sprintf("model %q is unknown for provider %q", modelName, providerName)
		return status
	}
	prov, err := New(providerName, cfg)
	if err != nil {
		status.Reason = err.Error()
		return status
	}
	models, err := prov.ListModels(ctx)
	if err != nil {
		status.Reason = err.Error()
		return status
	}
	ids := make([]string, 0, len(models))
	seen := make(map[string]struct{}, len(models))
	for _, m := range models {
		id := strings.TrimSpace(m.ID)
		if id == "" {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	status.Models = ids
	status.ModelCount = len(ids)
	status.Ready = true
	if modelName == "" {
		status.ModelAvailable = len(ids) > 0
	} else {
		for _, id := range ids {
			if strings.EqualFold(id, modelName) {
				status.ModelAvailable = true
				break
			}
		}
	}
	if !status.ModelAvailable {
		if modelName == "" {
			status.Reason = fmt.Sprintf("provider %q returned no models", providerName)
		} else {
			status.Reason = fmt.Sprintf("model %q not found in provider %q list", modelName, providerName)
		}
	}
	return status
}

// Router implements multi-model routing — cheap model for simple tasks,
// expensive model for reasoning.
type Router struct {
	Primary types.Provider
	Fast    types.Provider
	Route   types.ModelRoute

	providers map[string]types.Provider
	policy    RoutingPolicy
}

// NewRouter creates a multi-model router.
func NewRouter(route types.ModelRoute, providerSettings map[string]*config.ProviderSettings) (*Router, error) {
	if providerSettings == nil {
		providerSettings = make(map[string]*config.ProviderSettings)
	}

	registry := NewDefaultModelMetadataRegistry()
	matrix := NewRegistryCapabilityMatrix(registry)
	policy := NewPolicyEngine(registry, matrix)

	decision := policy.Select(route, RoutingRequest{Complexity: TaskComplexityNormal})
	if decision.Provider == "" {
		return nil, fmt.Errorf("router: unable to resolve provider for route")
	}

	primary, err := New(decision.Provider, providerSettings[decision.Provider])
	if err != nil {
		return nil, fmt.Errorf("primary provider: %w", err)
	}

	router := &Router{
		Primary: primary,
		Route:   route,
		providers: map[string]types.Provider{
			decision.Provider: primary,
		},
		policy: policy,
	}

	if route.Fast != "" && decision.Provider != "" {
		fast, err := New(decision.Provider, providerSettings[decision.Provider])
		if err != nil {
			return nil, fmt.Errorf("fast provider: %w", err)
		}
		router.Fast = fast
		router.providers[decision.Provider] = fast
	}

	return router, nil
}

// ForTask returns the appropriate provider and model for a given task complexity.
func (r *Router) ForTask(isSimple bool) (types.Provider, string) {
	complexity := TaskComplexityNormal
	if isSimple {
		complexity = TaskComplexitySimple
	}

	decision := r.policy.Select(r.Route, RoutingRequest{Complexity: complexity})
	if isSimple && r.Fast != nil && r.Route.Fast != "" && decision.Model == r.Route.Fast {
		return r.Fast, r.Route.Fast
	}
	if decision.Model == "" {
		return r.Primary, r.Route.Primary
	}
	provider := r.Primary
	if p, ok := r.providers[decision.Provider]; ok {
		provider = p
	}
	return provider, decision.Model
}
