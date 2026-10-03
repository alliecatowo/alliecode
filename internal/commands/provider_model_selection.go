package commands

import (
	"fmt"
	"strings"

	"github.com/alliecatowo/alliecode/internal/providers"
)

type ProviderModelSelection struct {
	ProviderName string
	ModelName    string
	ModelRef     string
}

type runtimeSelectionSnapshot struct {
	providerName  string
	modelName     string
	modelRef      string
	loggedIn      bool
	authProvider  string
	authAccount   string
	providerReady bool
}

func PreferredProvider(raw string) string {
	providerName := strings.ToLower(strings.TrimSpace(raw))
	if providerName == "" {
		return "ollama"
	}
	if !isSupportedProviderName(providerName) {
		return "ollama"
	}
	if len(providers.ListModelsByProvider(providerName)) == 0 {
		return "ollama"
	}
	return providerName
}

func PreferredModelForProvider(providerName, raw string) string {
	providerName = PreferredProvider(providerName)
	modelName := strings.TrimSpace(raw)
	if modelName != "" {
		if selection, err := ResolveProviderModelSelection(providerName, modelName); err == nil {
			return selection.ModelName
		}
	}
	defaultByProvider := map[string]string{
		"ollama":    "llama3",
		"openai":    "gpt-4o-mini",
		"anthropic": "claude-sonnet-4-20250514",
		"gemini":    "gemini-2.5-flash",
	}
	if fallback, ok := defaultByProvider[providerName]; ok {
		return fallback
	}
	models := providers.ListModelsByProvider(providerName)
	if len(models) > 0 {
		return strings.TrimSpace(models[0].Model)
	}
	return "llama3"
}

func ResolveProviderModelSelection(providerName, modelName string) (ProviderModelSelection, error) {
	providerName = strings.ToLower(strings.TrimSpace(providerName))
	modelName = strings.TrimSpace(modelName)
	if providerName == "" {
		return ProviderModelSelection{}, fmt.Errorf("provider cannot be empty (next: run /provider set <name>)")
	}
	if !isSupportedProviderName(providerName) {
		return ProviderModelSelection{}, fmt.Errorf("%s", unknownProviderGuidance(providerName))
	}
	if modelName == "" {
		return ProviderModelSelection{}, fmt.Errorf("model cannot be empty (next: run /model %s/<model>)", providerName)
	}
	if mapped, ok := friendlyModelAlias(providerName, modelName); ok {
		modelName = mapped
	}
	if canonical, ok := providers.CanonicalModelName(providerName, modelName); ok {
		modelName = canonical
	}
	if _, ok := providers.LookupModelMetadata(providerName, modelName); !ok {
		return ProviderModelSelection{}, unknownModelForProviderError(providerName, modelName)
	}
	return ProviderModelSelection{
		ProviderName: providerName,
		ModelName:    modelName,
		ModelRef:     providerName + "/" + modelName,
	}, nil
}

func ReconcileProviderSelection(providerName, currentModel string) (ProviderModelSelection, error) {
	providerName = PreferredProvider(providerName)
	if providerName == "" {
		return ProviderModelSelection{}, fmt.Errorf("provider cannot be empty (next: run /provider set <name>)")
	}
	if _, modelName, ok := resolveModelForStatus(currentModel, providerName); ok {
		if selection, err := ResolveProviderModelSelection(providerName, modelName); err == nil {
			return selection, nil
		}
	}
	return ResolveProviderModelSelection(providerName, PreferredModelForProvider(providerName, currentModel))
}

func ApplyProviderModelSelection(state *RuntimeState, selection ProviderModelSelection) error {
	if state == nil {
		return nil
	}
	selection, err := ResolveProviderModelSelection(selection.ProviderName, selection.ModelName)
	if err != nil {
		return err
	}
	prev := snapshotRuntimeSelection(state)
	next := prev
	next.providerName = selection.ProviderName
	next.modelName = selection.ModelName
	next.modelRef = selection.ModelRef
	next.providerReady = computeProviderReady(selection.ProviderName, next.loggedIn)
	if shouldClearAuthForSelection(prev, selection.ProviderName) {
		next.loggedIn = false
		next.authAccount = ""
		next.providerReady = computeProviderReady(selection.ProviderName, false)
	}
	if state.Agent != nil {
		if err := state.Agent.SetProviderModel(selection.ProviderName, selection.ModelName); err != nil {
			return err
		}
	}
	commitRuntimeSelection(state, next)
	return nil
}

func snapshotRuntimeSelection(state *RuntimeState) runtimeSelectionSnapshot {
	return runtimeSelectionSnapshot{
		providerName:  strings.ToLower(strings.TrimSpace(state.ProviderName)),
		modelName:     strings.TrimSpace(state.Model),
		modelRef:      strings.TrimSpace(state.ModelRef),
		loggedIn:      state.LoggedIn,
		authProvider:  strings.ToLower(strings.TrimSpace(state.AuthProvider)),
		authAccount:   strings.TrimSpace(state.AuthAccount),
		providerReady: state.ProviderReady,
	}
}

func shouldClearAuthForSelection(prev runtimeSelectionSnapshot, nextProvider string) bool {
	nextProvider = strings.ToLower(strings.TrimSpace(nextProvider))
	if nextProvider == "" || nextProvider == "ollama" {
		return false
	}
	if prev.authProvider != "" {
		return prev.authProvider != nextProvider
	}
	if prev.providerName == "" {
		return false
	}
	return prev.providerName != nextProvider
}

func commitRuntimeSelection(state *RuntimeState, next runtimeSelectionSnapshot) {
	state.ProviderName = next.providerName
	state.Model = next.modelName
	state.ModelRef = next.modelRef
	state.LoggedIn = next.loggedIn
	state.AuthAccount = next.authAccount
	state.ProviderReady = next.providerReady
	if next.loggedIn && next.providerName != "" && next.providerName != "ollama" {
		state.AuthProvider = next.providerName
	} else if !next.loggedIn {
		state.AuthProvider = ""
	}
	state.Runtime.ProviderName = next.providerName
	state.Runtime.Model = next.modelName
	state.Runtime.ModelRef = next.modelRef
	state.Runtime.LoggedIn = next.loggedIn
	state.Runtime.ProviderReady = next.providerReady
}
