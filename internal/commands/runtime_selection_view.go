package commands

import "strings"

type RuntimeSelectionView struct {
	ProviderName  string
	ModelName     string
	ModelRef      string
	LoggedIn      bool
	ProviderReady bool
}

func RuntimeSelectionTruth(state *RuntimeState) RuntimeSelectionView {
	if state == nil {
		return RuntimeSelectionView{}
	}

	providerName, modelName, modelRef := normalizeSelectionTuple(
		strings.ToLower(strings.TrimSpace(state.ProviderName)),
		strings.TrimSpace(state.Model),
		strings.TrimSpace(state.ModelRef),
	)

	providerName, modelName, modelRef = normalizeSelectionTuple(
		firstNonEmpty(strings.ToLower(strings.TrimSpace(state.Runtime.ProviderName)), providerName),
		firstNonEmpty(strings.TrimSpace(state.Runtime.Model), modelName),
		firstNonEmpty(strings.TrimSpace(state.Runtime.ModelRef), modelRef),
	)

	if state.Agent != nil {
		runtime := state.Agent.RuntimeSnapshot()
		providerName, modelName, modelRef = normalizeSelectionTuple(
			firstNonEmpty(strings.ToLower(strings.TrimSpace(runtime.ProviderName)), providerName),
			firstNonEmpty(strings.TrimSpace(runtime.Model), modelName),
			firstNonEmpty(strings.TrimSpace(runtime.ModelRef), modelRef),
		)
	}

	if providerName != "" && modelName != "" {
		if selection, err := ResolveProviderModelSelection(providerName, modelName); err == nil {
			providerName = selection.ProviderName
			modelName = selection.ModelName
			modelRef = selection.ModelRef
		}
	}
	if modelRef == "" && providerName != "" && modelName != "" {
		modelRef = providerName + "/" + modelName
	}

	loggedIn := state.LoggedIn
	providerReady := providerReadyForAuthProvider(providerName, loggedIn, state.AuthProvider)

	return RuntimeSelectionView{
		ProviderName:  providerName,
		ModelName:     modelName,
		ModelRef:      modelRef,
		LoggedIn:      loggedIn,
		ProviderReady: providerReady,
	}
}

func normalizeSelectionTuple(providerName, modelName, modelRef string) (string, string, string) {
	providerName = strings.ToLower(strings.TrimSpace(providerName))
	modelName = strings.TrimSpace(modelName)
	modelRef = strings.TrimSpace(modelRef)
	rawModelName := modelName

	if parsedProvider, parsedModel, ok := parseModelRef(modelRef); ok {
		providerName = parsedProvider
		modelName = parsedModel
	}
	if parsedProvider, parsedModel, ok := parseModelRef(rawModelName); ok {
		providerName = parsedProvider
		modelName = parsedModel
	}
	if providerName != "" && modelName != "" {
		modelRef = providerName + "/" + modelName
	}
	return providerName, modelName, modelRef
}

func parseModelRef(raw string) (string, string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", false
	}
	parts := strings.SplitN(raw, "/", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	providerName := strings.ToLower(strings.TrimSpace(parts[0]))
	modelName := strings.TrimSpace(parts[1])
	if providerName == "" || modelName == "" {
		return "", "", false
	}
	return providerName, modelName, true
}

func firstNonEmpty(value string, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func ProviderReadyForState(provider string, state *RuntimeState) bool {
	if state == nil {
		return computeProviderReady(provider, false)
	}
	return providerReadyForAuthProvider(provider, state.LoggedIn, state.AuthProvider)
}

func HydrateRuntimeSelection(state *RuntimeState) {
	if state == nil {
		return
	}
	selection := RuntimeSelectionTruth(state)
	state.ProviderName = selection.ProviderName
	state.Model = selection.ModelName
	state.ModelRef = selection.ModelRef
	state.ProviderReady = selection.ProviderReady
	state.Runtime.ProviderName = selection.ProviderName
	state.Runtime.Model = selection.ModelName
	state.Runtime.ModelRef = selection.ModelRef
	state.Runtime.ProviderReady = selection.ProviderReady
	state.Runtime.LoggedIn = selection.LoggedIn
}

func RuntimeSelectionNextAction(state *RuntimeState) string {
	selection := RuntimeSelectionTruth(state)
	providerName := strings.TrimSpace(selection.ProviderName)
	switch {
	case providerName == "":
		return "/provider set ollama"
	case !selection.ProviderReady:
		return "/login provider " + providerName
	default:
		return ""
	}
}

func providerReadyForAuthProvider(provider string, loggedIn bool, authProvider string) bool {
	provider = strings.ToLower(strings.TrimSpace(provider))
	authProvider = strings.ToLower(strings.TrimSpace(authProvider))
	if provider == "" {
		return false
	}
	if provider == "ollama" {
		return true
	}
	if !loggedIn {
		return false
	}
	if authProvider != "" && authProvider != provider {
		return false
	}
	return true
}
