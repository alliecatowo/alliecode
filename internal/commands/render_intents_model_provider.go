package commands

import (
	"fmt"
	"strings"

	"github.com/alliecatowo/alliecode/internal/providers"
	"github.com/alliecatowo/alliecode/internal/types"
)

func modelDoctorIntents(providerName, modelName string, ready bool, quickFix string) []types.RenderIntent {
	provider := strings.TrimSpace(providerName)
	if provider == "" {
		provider = "-"
	}
	model := strings.TrimSpace(modelName)
	if model == "" {
		model = "-"
	}
	return []types.RenderIntent{
		diagnosticsIntent("Model doctor", boolState(ready, "ready", "needs attention"), "Validate the active model/provider selection.", []types.RenderDiagnostic{
			diagnostic("Provider", boolState(strings.TrimSpace(providerName) != "", "ok", "missing"), provider),
			diagnostic("Model", boolState(strings.TrimSpace(modelName) != "", "ok", "missing"), model),
			diagnostic("Runtime readiness", boolState(ready, "ok", "warn"), boolState(ready, "provider and model are ready", "selection is incomplete or provider is not ready")),
		}, []types.RenderActionHint{hint("Quick fix", quickFix), hint("Login provider", "/login provider <name>")}),
	}
}

func modelListProviderIntents(providerName string, ready bool) []types.RenderIntent {
	models := providers.ListModelsByProvider(providerName)
	rows := make([]types.RenderTableRow, 0, len(models))
	for _, model := range models {
		rows = append(rows, tableRow(providerName, model.Model, fmt.Sprintf("%d", model.ContextWindow), boolState(ready, "ready", "not ready")))
	}
	return []types.RenderIntent{
		tableIntent("Models", fmt.Sprintf("Provider %s", providerName), []string{"Provider", "Model", "Context", "State"}, rows...),
		actionHintsIntent("Actions", hint("Switch model", "/model "+providerName+"/<model>")),
	}
}

func modelListProvidersIntents(state *RuntimeState) []types.RenderIntent {
	providersList := providerList()
	rows := make([]types.RenderTableRow, 0, len(providersList))
	for _, providerName := range providersList {
		models := providers.ListModelsByProvider(providerName)
		rows = append(rows, tableRow(providerName, fmt.Sprintf("%d", len(models)), boolState(ProviderReadyForState(providerName, state), "ready", "not ready")))
	}
	return []types.RenderIntent{
		tableIntent("Model providers", "Browse providers before choosing a model.", []string{"Provider", "Models", "State"}, rows...),
		actionHintsIntent("Actions", hint("Select a model", "/model <provider>/<model>")),
	}
}

func modelListAllIntents() []types.RenderIntent {
	providersList := providerList()
	rows := []types.RenderTableRow{}
	for _, providerName := range providersList {
		for _, model := range providers.ListModelsByProvider(providerName) {
			rows = append(rows, tableRow(providerName, model.Model, fmt.Sprintf("%d", model.ContextWindow)))
		}
	}
	return []types.RenderIntent{tableIntent("All models", "Cross-provider model inventory.", []string{"Provider", "Model", "Context"}, rows...)}
}

func modelStatusIntents(providerName, modelName, capabilities, quickFix string, ready bool) []types.RenderIntent {
	return []types.RenderIntent{
		summaryCardIntent("Model status", "Current runtime model selection.",
			field("Provider", defaultDash(providerName)),
			field("Model", defaultDash(modelName)),
			field("Capabilities", defaultDash(capabilities)),
			field("Provider ready", boolState(ready, "yes", "no")),
		),
		actionHintsIntent("Actions", hint("Quick fix", quickFix), hint("Provider status", "/provider status"), hint("Browse models", "/model list")),
	}
}

func modelRepairIntents(providerName, modelName string, ready bool) []types.RenderIntent {
	return []types.RenderIntent{
		summaryCardIntent("Model set to "+modelName, "Updated the active model selection.", field("Provider", providerName), field("Model", modelName), field("Provider ready", boolState(ready, "yes", "no"))),
		actionHintsIntent("Actions", hint("Verify runtime", "/status")),
	}
}

func providerStatusIntents(providerName, modelHint string, ready bool) []types.RenderIntent {
	return []types.RenderIntent{
		summaryCardIntent("Provider status", "Current provider availability for this session.", field("Provider", defaultDash(providerName)), field("Ready", boolState(ready, "yes", "no")), field("Model hint", defaultDash(modelHint))),
		actionHintsIntent("Actions", hint("Set provider", "/provider set <name>"), hint("Login provider", "/login provider <name>"), hint("Set model", "/model "+modelHint)),
	}
}

func providerDoctorIntents(providerName, modelName string, ready, modelValid bool, quickFix string) []types.RenderIntent {
	return []types.RenderIntent{
		diagnosticsIntent("Provider doctor", boolState(ready, "ready", "needs attention"), "Validate provider auth and model compatibility.", []types.RenderDiagnostic{
			diagnostic("Provider", boolState(strings.TrimSpace(providerName) != "" && providerName != "-", "ok", "missing"), defaultDash(providerName)),
			diagnostic("Ready", boolState(ready, "ok", "warn"), boolState(ready, "provider is ready", "provider needs setup or auth")),
			diagnostic("Model", boolState(modelValid, "ok", "warn"), defaultDash(modelName)),
		}, []types.RenderActionHint{hint("Quick fix", quickFix), hint("Login provider", "/login provider <name>")}),
	}
}

func providerModelsIntents(providerName string) []types.RenderIntent {
	models := providers.ListModelsByProvider(providerName)
	rows := make([]types.RenderTableRow, 0, len(models))
	for _, model := range models {
		rows = append(rows, tableRow(model.Model))
	}
	return []types.RenderIntent{tableIntent("Provider models", fmt.Sprintf("Provider %s", providerName), []string{"Model"}, rows...)}
}

func providerSetIntents(providerName string, ready bool) []types.RenderIntent {
	return []types.RenderIntent{
		summaryCardIntent("Provider updated", "Changed the active provider for this session.", field("Provider", providerName), field("Ready", boolState(ready, "yes", "no"))),
		actionHintsIntent("Actions", hint("Pick a model", "/model "+providerName+"/<model>"), hint("Provider status", "/provider status")),
	}
}

func correctiveLoopIntents(title, summary string, loops []correctiveLoop) []types.RenderIntent {
	items := make([]types.RenderChecklistItem, 0, len(loops))
	for _, loop := range loops {
		items = append(items, checklistItem(loop.Area, false, fmt.Sprintf("state=%s next=%s action=%s", loop.State, loop.Next, loop.Action)))
	}
	return []types.RenderIntent{checklistIntent(title, summary, items...)}
}

func boolState(ok bool, yes, no string) string {
	if ok {
		return yes
	}
	return no
}

func defaultDash(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "-"
	}
	return v
}
