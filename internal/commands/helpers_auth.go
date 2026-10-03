package commands

import "strings"

func defaultLoginProvider(current string) string {
	provider := strings.TrimSpace(current)
	if provider == "" {
		return "anthropic"
	}
	return provider
}

func normalizeAccountName(raw string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(raw), " "))
}

func computeProviderReady(provider string, loggedIn bool) bool {
	return providerReadyForAuthProvider(provider, loggedIn, provider)
}

func syncProviderReadiness(state *RuntimeState) {
	if state == nil {
		return
	}
	HydrateRuntimeSelection(state)
}

func resetAuthIdentityForProviderSwitch(state *RuntimeState, nextProvider string) {
	if state == nil {
		return
	}
	current := strings.ToLower(strings.TrimSpace(state.ProviderName))
	next := strings.ToLower(strings.TrimSpace(nextProvider))
	if current == "" || next == "" || current == next {
		return
	}
	if next == "ollama" {
		return
	}
	state.LoggedIn = false
	state.AuthProvider = ""
	state.AuthAccount = ""
}
