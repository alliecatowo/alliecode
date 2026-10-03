package commands

import (
	"fmt"
	"strings"

	"github.com/alliecatowo/alliecode/internal/providers"
	"github.com/alliecatowo/alliecode/internal/types"
)

func buildProviderInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	syncProviderReadiness(state)
	selection := RuntimeSelectionTruth(state)
	providerName := strings.TrimSpace(selection.ProviderName)
	if providerName == "" {
		providerName = "-"
	}
	modelHint := strings.TrimSpace(selection.ModelName)
	if modelHint == "" {
		modelHint = "<provider/model>"
	}

	panel := InteractivePanel{
		Command:       "provider",
		Title:         "provider panel: /provider",
		Subtitle:      fmt.Sprintf("provider=%s ready=%t", normalizeToken(providerName), selection.ProviderReady),
		HeaderIntents: providerStatusIntents(providerName, modelHint, selection.ProviderReady),
		Items: []InteractivePanelItem{
			{Key: "status", Section: "Overview", Label: "Provider status", Detail: "Inspect selected provider readiness", Status: statusWord(selection.ProviderReady, "ready", "needs-login"), ApplyInput: "/provider status", ApplyMode: PanelApplySubmit, PreviewIntents: providerStatusIntents(providerName, modelHint, selection.ProviderReady), Preview: previewLines(fmt.Sprintf("Provider: %s", providerName), fmt.Sprintf("Ready: %t", selection.ProviderReady))},
			{Key: "list", Section: "Overview", Label: "List providers", Detail: "Show supported providers and readiness", Status: "list", ApplyInput: "/provider list", ApplyMode: PanelApplySubmit, PreviewIntents: modelListProvidersIntents(state), Preview: previewLines("/provider list shows all supported providers and readiness.")},
			{Key: "doctor", Section: "Diagnostics", Label: "Provider doctor", Detail: "Validate provider auth and model compatibility", Status: statusWord(selection.ProviderReady, "healthy", "repair"), ApplyInput: "/provider doctor", ApplyMode: PanelApplySubmit, PreviewIntents: providerDoctorIntents(providerName, strings.TrimSpace(selection.ModelName), selection.ProviderReady, strings.TrimSpace(selection.ModelName) != "", "/provider set ollama"), Preview: previewLines("/provider doctor returns provider/model compatibility and quick fix guidance.")},
			{Key: "repair", Section: "Diagnostics", Label: "Repair provider", Detail: "Re-apply provider and refresh runtime readiness", Status: "repair", ApplyInput: "/provider repair", ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{actionHintsIntent("Provider repair", hint("Repair", "/provider repair"), hint("Verify runtime", "/status"))}, Preview: previewLines("/provider repair re-applies provider defaults and refreshes readiness.")},
		},
	}

	for _, name := range providers.SupportedProviderNames() {
		ready := ProviderReadyForState(name, state)
		setStatus := statusWord(strings.EqualFold(strings.TrimSpace(selection.ProviderName), name), "current", statusWord(ready, "ready", "auth"))
		panel.Items = append(panel.Items,
			InteractivePanelItem{
				Key:            "set-" + name,
				Section:        "Providers",
				Label:          "Set " + name,
				Detail:         "Use provider " + name + " for runtime",
				Status:         setStatus,
				ApplyInput:     "/provider set " + name,
				ApplyMode:      PanelApplySubmit,
				PreviewIntents: providerSetIntents(name, ready),
				Preview:        previewLines("/provider set " + name + " updates active provider and keeps model in sync when possible."),
			},
			InteractivePanelItem{
				Key:            "models-" + name,
				Section:        "Providers",
				Label:          "List " + name + " models",
				Detail:         "Show models available from " + name,
				Status:         "models",
				ApplyInput:     "/provider models " + name,
				ApplyMode:      PanelApplySubmit,
				PreviewIntents: providerModelsIntents(name),
				Preview:        previewLines("/provider models " + name + " lists model ids for that provider."),
			},
		)
	}

	return panel
}
