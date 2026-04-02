package providers

import "sort"

var defaultRegistry = NewDefaultModelMetadataRegistry()

// LookupModelMetadata returns metadata for a provider/model pair.
func LookupModelMetadata(provider, model string) (ModelMetadata, bool) {
	return defaultRegistry.Get(provider, model)
}

// LookupModelContextWindow returns context window for a provider/model pair.
func LookupModelContextWindow(provider, model string) (int, bool) {
	return defaultRegistry.ContextWindow(provider, model)
}

// LookupModelPricing returns input/output pricing for a provider/model pair.
func LookupModelPricing(provider, model string) (float64, float64, bool) {
	return defaultRegistry.Pricing(provider, model)
}

// LookupModelMaxOutput returns max output tokens for provider/model.
func LookupModelMaxOutput(provider, model string) (int, bool) {
	return defaultRegistry.MaxOutput(provider, model)
}

// LookupModelCostClass returns cost class for provider/model.
func LookupModelCostClass(provider, model string) (CostClass, bool) {
	return defaultRegistry.CostClass(provider, model)
}

// LookupModelContextClass returns context window class for provider/model.
func LookupModelContextClass(provider, model string) (ContextWindowClass, bool) {
	return defaultRegistry.ContextWindowClass(provider, model)
}

// LookupModelServiceTier returns service tier for provider/model.
func LookupModelServiceTier(provider, model string) (ServiceTier, bool) {
	return defaultRegistry.ServiceTier(provider, model)
}

// LookupModelCapabilitySummary returns a stable capability token list.
func LookupModelCapabilitySummary(provider, model string) (string, bool) {
	meta, ok := defaultRegistry.Get(provider, model)
	if !ok {
		return "", false
	}
	return meta.CapabilitySummary(), true
}

// ListModelsByMinimumContextWindow returns models with at least minimum tokens.
func ListModelsByMinimumContextWindow(provider string, minimum int) []ModelMetadata {
	return defaultRegistry.ListByMinimumContextWindow(provider, minimum)
}

// ListModelsByCostClass returns models up to a cost class ceiling.
func ListModelsByCostClass(provider string, max CostClass) []ModelMetadata {
	return defaultRegistry.ListByCostClass(provider, max)
}

// ListModelsByServiceTier returns provider models at a service tier.
func ListModelsByServiceTier(provider string, tier ServiceTier) []ModelMetadata {
	return defaultRegistry.ListByServiceTier(provider, tier)
}

// ListModelsByCostClasses returns models for any of the provided classes.
func ListModelsByCostClasses(provider string, classes ...CostClass) []ModelMetadata {
	return defaultRegistry.ListByCostClasses(provider, classes...)
}

// ListModelsByMinimumContextClass returns models at or above a context class.
func ListModelsByMinimumContextClass(provider string, class ContextWindowClass) []ModelMetadata {
	return defaultRegistry.ListByMinimumContextClass(provider, class)
}

// ListModelsByProvider returns known models for a provider.
func ListModelsByProvider(provider string) []ModelMetadata {
	return defaultRegistry.ListByProvider(provider)
}

// SupportedProviderNames returns stable provider identifiers.
func SupportedProviderNames() []string {
	names := defaultRegistry.Providers()
	sort.Strings(names)
	return names
}
