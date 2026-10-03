package commands

import (
	"fmt"
	"strings"
)

func providerList() []string {
	return []string{"anthropic", "gemini", "ollama", "openai"}
}

func providerListMessage() string {
	known := providerList()
	lines := []string{"PROVIDER_LIST", fmt.Sprintf("count=%d", len(known))}
	for i, name := range known {
		lines = append(lines, fmt.Sprintf("provider.%d=%s", i+1, name))
	}
	lines = append(lines, "next=use_/provider_set_<name>_to_switch")
	return strings.Join(lines, "\n")
}

// ProviderReadyFor reports whether a provider is currently ready.
func ProviderReadyFor(provider string, loggedIn bool) bool {
	return computeProviderReady(provider, loggedIn)
}

func providerRecoveryHint(provider string) string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return "/provider set ollama"
	}
	if provider == "ollama" {
		return "/provider status"
	}
	return "/login provider " + provider
}

func modelRecoveryHint(provider, model string, ready bool) string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	model = strings.TrimSpace(model)
	if provider == "" {
		return "/provider set ollama"
	}
	if model == "" {
		return "/model " + provider + "/<model>"
	}
	if !ready {
		return providerRecoveryHint(provider)
	}
	return "/status"
}

func unknownProviderGuidance(provider string) string {
	provider = strings.TrimSpace(provider)
	if provider == "" {
		return "unknown provider"
	}
	return fmt.Sprintf("unknown provider %q (run /provider list, then /provider set <name>)", provider)
}
