package providers

import (
	"context"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

// GetProviderRetryPolicy returns provider retry policy details when available.
func GetProviderRetryPolicy(provider types.Provider) types.ProviderRetryPolicy {
	if src, ok := provider.(types.ProviderRetryPolicySource); ok {
		return src.RetryPolicy()
	}
	return types.ProviderRetryPolicy{}
}

// GetProviderModelMetadata returns provider model metadata when available.
func GetProviderModelMetadata(ctx context.Context, provider types.Provider) ([]types.ModelMetadata, error) {
	if src, ok := provider.(types.ProviderMetadataSource); ok {
		return src.ModelMetadata(ctx)
	}
	models, err := provider.ListModels(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]types.ModelMetadata, 0, len(models))
	for _, m := range models {
		meta, _ := LookupModelMetadata(m.Provider, m.ID)
		out = append(out, types.ModelMetadata{
			Provider:        m.Provider,
			Model:           m.ID,
			Tier:            string(meta.Tier),
			ServiceTier:     string(meta.ServiceTier),
			CostClass:       string(meta.CostClass),
			ContextClass:    string(meta.ContextClass),
			ContextWindow:   m.ContextWindow,
			MaxOutput:       m.MaxOutput,
			SupportsText:    true,
			SupportsTools:   m.SupportsTools,
			SupportsToolUse: m.SupportsTools,
			SupportsVision:  m.SupportsVision,
			CostPerMInput:   m.CostPerMInput,
			CostPerMOutput:  m.CostPerMOutput,
		})
	}
	return out, nil
}

// RetryPolicyFromDefault converts common defaults to provider-facing policy.
func RetryPolicyFromDefault(maxRetries int, baseBackoff, maxBackoff time.Duration) types.ProviderRetryPolicy {
	return types.ProviderRetryPolicy{
		Name:          "default",
		MaxRetries:    maxRetries,
		BaseBackoffMS: int(baseBackoff.Milliseconds()),
		MaxBackoffMS:  int(maxBackoff.Milliseconds()),
	}
}
