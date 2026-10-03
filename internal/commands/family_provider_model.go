package commands

import (
	"fmt"
	"strings"

	"github.com/alliecatowo/alliecode/internal/providers"
)

func executeModelCommand(cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	const modelUsage = "usage: /model [provider/model|model|list [provider|all]|doctor|repair [provider/model]]"
	if len(inv.Args) > 0 {
		sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
		switch sub {
		case "loops", "loop":
			return renderModelCorrectiveLoops(cmdCtx.State), nil
		case "doctor":
			syncProviderReadiness(cmdCtx.State)
			selection := RuntimeSelectionTruth(cmdCtx.State)
			providerName := strings.TrimSpace(selection.ProviderName)
			modelRaw := strings.TrimSpace(selection.ModelName)
			ready := providerName != "" && modelRaw != "" && selection.ProviderReady
			quickFix := modelRecoveryHint(providerName, modelRaw, selection.ProviderReady)
			message := fmt.Sprintf("MODEL_DOCTOR\nprovider=%s\nmodel=%s\nready=%t\nquick_fix=%s", normalizeToken(providerName), normalizeToken(modelRaw), ready, normalizeToken(quickFix))
			return resultWithIntents(message, modelDoctorIntents(providerName, modelRaw, ready, quickFix)...), nil
		case "list":
			if len(inv.Args) > 2 {
				return Result{}, fmt.Errorf("%s", modelUsage)
			}
			if len(inv.Args) == 2 && equalFoldTrimmed(inv.Args[1], "all") {
				message := renderModelListAllMessage()
				return resultWithIntents(message, modelListAllIntents()...), nil
			}
			providerName := strings.TrimSpace(RuntimeSelectionTruth(cmdCtx.State).ProviderName)
			if len(inv.Args) == 2 {
				providerName = strings.ToLower(strings.TrimSpace(inv.Args[1]))
			}
			if providerName == "" {
				message := renderModelListByProviderMessage(cmdCtx.State)
				return resultWithIntents(message, modelListProvidersIntents(cmdCtx.State)...), nil
			}
			if !isSupportedProviderName(providerName) {
				return Result{}, fmt.Errorf("%s", unknownProviderGuidance(providerName))
			}
			message := renderModelListSingleProviderMessage(cmdCtx.State, providerName)
			return resultWithIntents(message, modelListProviderIntents(providerName, ProviderReadyForState(providerName, cmdCtx.State))...), nil
		case "repair", "fix":
			target := ""
			if len(inv.Args) > 1 {
				target = strings.TrimSpace(strings.Join(inv.Args[1:], " "))
			}
			if target == "" {
				providerName := strings.TrimSpace(RuntimeSelectionTruth(cmdCtx.State).ProviderName)
				if providerName == "" {
					providerName = "ollama"
				}
				models := providers.ListModelsByProvider(providerName)
				if len(models) > 0 {
					target = providerName + "/" + models[0].Model
				}
			}
			if target == "" {
				return Result{}, fmt.Errorf("%s", modelUsage)
			}
			providerName, modelName, err := resolveModelSelectionForCommand(target, RuntimeSelectionTruth(cmdCtx.State).ProviderName)
			if err != nil {
				return Result{}, err
			}
			if _, ok := providers.LookupModelMetadata(providerName, modelName); !ok {
				return Result{}, unknownModelForProviderError(providerName, modelName)
			}
			if err := ApplyProviderModelSelection(cmdCtx.State, ProviderModelSelection{ProviderName: providerName, ModelName: modelName}); err != nil {
				return Result{}, err
			}
			syncProviderReadiness(cmdCtx.State)
			selection := RuntimeSelectionTruth(cmdCtx.State)
			message := fmt.Sprintf("MODEL_REPAIR\nprovider=%s\nmodel=%s\nprovider_ready=%t", normalizeToken(providerName), normalizeToken(modelName), selection.ProviderReady)
			return resultWithIntents(message, modelRepairIntents(providerName, modelName, selection.ProviderReady)...), nil
		}
	}

	if len(inv.Args) == 0 {
		syncProviderReadiness(cmdCtx.State)
		selection := RuntimeSelectionTruth(cmdCtx.State)
		current := selection.ModelName
		if current == "" {
			current = "unknown"
		}
		summary := "-"
		if providerName, modelName, ok := resolveModelForStatus(current, selection.ProviderName); ok {
			if capabilitySummary, found := providers.LookupModelCapabilitySummary(providerName, modelName); found {
				summary = capabilitySummary
			}
		}
		providerHint := strings.TrimSpace(selection.ProviderName)
		if providerHint == "" {
			providerHint = "<provider>"
		}
		if strings.TrimSpace(selection.ModelRef) == "" {
			message := fmt.Sprintf("Current model: %s\nCapabilities: %s\nQuick fix: /model %s/<model>\nNext: use /model [provider/model|model] to switch, or /provider status to verify provider.", current, summary, providerHint)
			return resultWithIntents(message, modelStatusIntents(selection.ProviderName, current, summary, "/model "+providerHint+"/<model>", selection.ProviderReady)...), nil
		}
		providerName := strings.TrimSpace(selection.ProviderName)
		if providerName == "" {
			providerName = "-"
		}
		message := fmt.Sprintf("MODEL_STATUS\nprovider=%s\nprovider_ready=%t\nmodel=%s\ncapabilities=%s\nquick_fix=/model_%s/<model>\nnext=use_/model_list_to_browse_or_/model_<provider>/<model>_to_set", normalizeToken(providerName), selection.ProviderReady, normalizeToken(current), normalizeToken(summary), normalizeToken(providerHint))
		return resultWithIntents(message, modelStatusIntents(providerName, current, summary, "/model "+providerHint+"/<model>", selection.ProviderReady)...), nil
	}

	next := strings.TrimSpace(inv.Args[0])
	if next == "" {
		providerHint := strings.TrimSpace(RuntimeSelectionTruth(cmdCtx.State).ProviderName)
		if providerHint == "" {
			providerHint = "<provider>"
		}
		return Result{}, fmt.Errorf("model cannot be empty (next: run /model %s/<model>)", providerHint)
	}
	providerName, modelName, err := resolveModelSelectionForCommand(next, RuntimeSelectionTruth(cmdCtx.State).ProviderName)
	if err != nil {
		return Result{}, err
	}
	if _, ok := providers.LookupModelMetadata(providerName, modelName); !ok {
		return Result{}, unknownModelForProviderError(providerName, modelName)
	}
	if err := ApplyProviderModelSelection(cmdCtx.State, ProviderModelSelection{ProviderName: providerName, ModelName: modelName}); err != nil {
		return Result{}, err
	}
	syncProviderReadiness(cmdCtx.State)
	selection := RuntimeSelectionTruth(cmdCtx.State)
	capabilitySummary := "-"
	if summary, ok := providers.LookupModelCapabilitySummary(providerName, modelName); ok {
		capabilitySummary = summary
	}
	message := fmt.Sprintf("Model set to %s\nCapabilities: %s\nQuick fix: /provider status\nNext: run /status to confirm runtime readiness.", modelName, capabilitySummary)
	return resultWithIntents(message, modelRepairIntents(providerName, modelName, selection.ProviderReady)...), nil
}

func executeProviderCommand(cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	usage := "usage: /provider [status|list|set <name>|doctor|repair [name]|models [name]|<name>]"
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("%s", usage)
		}
		syncProviderReadiness(cmdCtx.State)
		selection := RuntimeSelectionTruth(cmdCtx.State)
		providerName := strings.TrimSpace(selection.ProviderName)
		if providerName == "" {
			providerName = "-"
		}
		modelHint := strings.TrimSpace(selection.ModelName)
		if modelHint == "" {
			modelHint = "<provider/model>"
		}
		next := "use_/provider_set_<name>_or_/login_provider_<name>"
		if selection.ProviderReady {
			next = "use_/model_<provider>/<model>_to_keep_provider_model_valid"
		}
		message := fmt.Sprintf("PROVIDER_STATUS\nprovider=%s\nprovider_ready=%t\nquick_fix_model=/model_%s\nquick_fix_auth=%s\nnext=%s", normalizeToken(providerName), selection.ProviderReady, normalizeToken(modelHint), normalizeToken(providerRecoveryHint(selection.ProviderName)), normalizeToken(next))
		return resultWithIntents(message, providerStatusIntents(providerName, modelHint, selection.ProviderReady)...), nil
	}
	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "list") {
		message := providerListMessage()
		return resultWithIntents(message, modelListProvidersIntents(cmdCtx.State)...), nil
	}
	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "doctor") {
		syncProviderReadiness(cmdCtx.State)
		selection := RuntimeSelectionTruth(cmdCtx.State)
		providerName := strings.TrimSpace(selection.ProviderName)
		if providerName == "" {
			providerName = "-"
		}
		modelRaw := strings.TrimSpace(selection.ModelName)
		modelOK := false
		if p, m, ok := resolveModelForStatus(modelRaw, selection.ProviderName); ok {
			_, modelOK = providers.LookupModelMetadata(p, m)
		}
		quickFix := "/provider set ollama"
		if providerName != "-" {
			quickFix = "/model " + strings.TrimSpace(selection.ProviderName) + "/<model>"
		}
		message := fmt.Sprintf("PROVIDER_DOCTOR\nprovider=%s\nprovider_ready=%t\nmodel=%s\nmodel_valid=%t\nquick_fix=%s", normalizeToken(providerName), selection.ProviderReady, normalizeToken(modelRaw), modelOK, normalizeToken(quickFix))
		return resultWithIntents(message, providerDoctorIntents(providerName, modelRaw, selection.ProviderReady, modelOK, quickFix)...), nil
	}
	if len(inv.Args) == 1 && (strings.EqualFold(strings.TrimSpace(inv.Args[0]), "loops") || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "loop")) {
		return renderProviderCorrectiveLoops(cmdCtx.State), nil
	}
	if len(inv.Args) >= 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "models") {
		providerName := strings.TrimSpace(RuntimeSelectionTruth(cmdCtx.State).ProviderName)
		if len(inv.Args) == 2 {
			providerName = strings.ToLower(strings.TrimSpace(inv.Args[1]))
		}
		if providerName == "" {
			return Result{}, fmt.Errorf("%s", usage)
		}
		models := providers.ListModelsByProvider(providerName)
		lines := []string{"PROVIDER_MODELS", fmt.Sprintf("provider=%s", normalizeToken(providerName)), fmt.Sprintf("count=%d", len(models))}
		for i, m := range models {
			lines = append(lines, fmt.Sprintf("model.%d=%s", i+1, normalizeToken(m.Model)))
		}
		message := strings.Join(lines, "\n")
		return resultWithIntents(message, providerModelsIntents(providerName)...), nil
	}

	setProvider := ""
	if len(inv.Args) >= 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "repair") {
		if len(inv.Args) > 2 {
			return Result{}, fmt.Errorf("%s", usage)
		}
		if len(inv.Args) == 2 {
			setProvider = strings.ToLower(strings.TrimSpace(inv.Args[1]))
		} else if strings.TrimSpace(RuntimeSelectionTruth(cmdCtx.State).ProviderName) == "" {
			setProvider = "ollama"
		} else {
			setProvider = strings.ToLower(strings.TrimSpace(RuntimeSelectionTruth(cmdCtx.State).ProviderName))
		}
	} else if len(inv.Args) == 2 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "set") {
		setProvider = strings.ToLower(strings.TrimSpace(inv.Args[1]))
	} else if len(inv.Args) == 1 {
		setProvider = strings.ToLower(strings.TrimSpace(inv.Args[0]))
	} else {
		return Result{}, fmt.Errorf("%s", usage)
	}
	if setProvider == "" {
		return Result{}, fmt.Errorf("%s", usage)
	}
	if !isSupportedProviderName(setProvider) {
		return Result{}, fmt.Errorf("%s", unknownProviderGuidance(setProvider))
	}
	selection, err := ReconcileProviderSelection(setProvider, RuntimeSelectionTruth(cmdCtx.State).ModelName)
	if err != nil {
		return Result{}, err
	}
	if err := ApplyProviderModelSelection(cmdCtx.State, selection); err != nil {
		return Result{}, err
	}
	syncProviderReadiness(cmdCtx.State)
	selectionView := RuntimeSelectionTruth(cmdCtx.State)
	resultType := "PROVIDER_SET"
	if len(inv.Args) >= 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "repair") {
		resultType = "PROVIDER_REPAIR"
	}
	message := fmt.Sprintf("%s\nprovider=%s\nmodel=%s\nprovider_ready=%t\nquick_fix_model=/model_%s/<model>\nquick_fix_auth=%s\nnext=run_/status_to_confirm_runtime_for_%s", resultType, selection.ProviderName, normalizeToken(selection.ModelName), selectionView.ProviderReady, selection.ProviderName, normalizeToken(providerRecoveryHint(selection.ProviderName)), selection.ProviderName)
	return resultWithIntents(message, providerSetIntents(selection.ProviderName, selectionView.ProviderReady)...), nil
}

func resolveModelSelectionForCommand(raw, defaultProvider string) (string, string, error) {
	providerName, modelName, err := parseModelSelection(raw, defaultProvider)
	if err != nil {
		return "", "", err
	}
	providerName = strings.ToLower(strings.TrimSpace(providerName))
	modelName = strings.TrimSpace(modelName)
	if !isSupportedProviderName(providerName) {
		return "", "", fmt.Errorf("%s", unknownProviderGuidance(providerName))
	}
	selection, err := ResolveProviderModelSelection(providerName, modelName)
	if err != nil {
		return "", "", err
	}
	return selection.ProviderName, selection.ModelName, nil
}

func friendlyModelAlias(providerName, modelName string) (string, bool) {
	provider := strings.ToLower(strings.TrimSpace(providerName))
	model := strings.ToLower(strings.TrimSpace(modelName))
	if provider != "anthropic" {
		return "", false
	}
	switch normalizeAnthropicAlias(model) {
	case "claudeopus46", "opus46", "claude46opus", "opusclaude46", "opus46latest", "claudeopus4", "opus4", "claudeopus", "opus":
		return "claude-opus-4-20250514", true
	case "claudesonnet46", "sonnet46", "claude46sonnet", "sonnetclaude46", "sonnet46latest", "claudesonnet4", "sonnet4", "claudesonnet", "sonnet":
		return "claude-sonnet-4-20250514", true
	case "claudehaiku35", "haiku35", "claude35haiku", "haikuclaude35", "haiku35latest", "claudehaiku", "haiku":
		return "claude-haiku-3-5-20241022", true
	default:
		return "", false
	}
}

func unknownModelForProviderError(providerName, modelName string) error {
	if alias, ok := friendlyModelAlias(providerName, modelName); ok {
		return fmt.Errorf("unknown model %q for provider %q; try /model %s/%s and then /login provider %s", modelName, providerName, providerName, alias, providerName)
	}
	if strings.EqualFold(strings.TrimSpace(providerName), "anthropic") {
		return fmt.Errorf("unknown model %q for provider %q (run /model list %s; then /model %s/opus, /model %s/sonnet, or /model %s/haiku; and /login provider %s if auth is missing)", modelName, providerName, providerName, providerName, providerName, providerName, providerName)
	}
	return fmt.Errorf("unknown model %q for provider %q (run /model list %s, then /model %s/<model>)", modelName, providerName, providerName, providerName)
}

func normalizeAnthropicAlias(model string) string {
	model = strings.ToLower(strings.TrimSpace(model))
	replacer := strings.NewReplacer("-", "", ".", "", "_", "", " ", "")
	return replacer.Replace(model)
}

func renderModelListSingleProviderMessage(state *RuntimeState, providerName string) string {
	models := providers.ListModelsByProvider(providerName)
	ready := ProviderReadyForState(providerName, state)
	lines := []string{
		"MODEL_LIST",
		fmt.Sprintf("provider=%s", normalizeToken(providerName)),
		fmt.Sprintf("provider_ready=%t", ready),
		fmt.Sprintf("count=%d", len(models)),
	}
	for i, model := range models {
		lines = append(lines, fmt.Sprintf("model.%d.id=%s", i+1, normalizeToken(model.Model)))
		lines = append(lines, fmt.Sprintf("model.%d.context_window=%d", i+1, model.ContextWindow))
	}
	return strings.Join(lines, "\n")
}

func renderModelListByProviderMessage(state *RuntimeState) string {
	providersList := providerList()
	lines := []string{"MODEL_LIST", "scope=providers", fmt.Sprintf("provider_count=%d", len(providersList))}
	totalModels := 0
	for i, providerName := range providersList {
		models := providers.ListModelsByProvider(providerName)
		totalModels += len(models)
		idx := i + 1
		ready := ProviderReadyForState(providerName, state)
		lines = append(lines, fmt.Sprintf("provider.%d.name=%s", idx, normalizeToken(providerName)))
		lines = append(lines, fmt.Sprintf("provider.%d.ready=%t", idx, ready))
		lines = append(lines, fmt.Sprintf("provider.%d.count=%d", idx, len(models)))
		for j, model := range models {
			lines = append(lines, fmt.Sprintf("provider.%d.model.%d.id=%s", idx, j+1, normalizeToken(model.Model)))
		}
	}
	lines = append(lines, fmt.Sprintf("total_models=%d", totalModels))
	lines = append(lines, "next=use_/model_<provider>/<model>_to_select")
	return strings.Join(lines, "\n")
}

func renderModelListAllMessage() string {
	providersList := providerList()
	lines := []string{"MODEL_LIST_ALL", fmt.Sprintf("provider_count=%d", len(providersList))}
	entry := 0
	for i, providerName := range providersList {
		models := providers.ListModelsByProvider(providerName)
		lines = append(lines, fmt.Sprintf("provider.%d=%s", i+1, normalizeToken(providerName)))
		for _, model := range models {
			entry++
			lines = append(lines, fmt.Sprintf("entry.%d.provider=%s", entry, normalizeToken(providerName)))
			lines = append(lines, fmt.Sprintf("entry.%d.model=%s", entry, normalizeToken(model.Model)))
			lines = append(lines, fmt.Sprintf("entry.%d.context_window=%d", entry, model.ContextWindow))
		}
	}
	lines = append(lines, fmt.Sprintf("count=%d", entry))
	return strings.Join(lines, "\n")
}

func renderModelCorrectiveLoops(state *RuntimeState) Result {
	lines := []string{"MODEL_LOOPS", fmt.Sprintf("count=%d", len(correctiveLoopsForState(state)))}
	idx := 0
	for _, loop := range correctiveLoopsForState(state) {
		if loop.Area != "model" && loop.Area != "provider" {
			continue
		}
		idx++
		lines = append(lines, fmt.Sprintf("loop.%d.area=%s", idx, normalizeToken(loop.Area)))
		lines = append(lines, fmt.Sprintf("loop.%d.state=%s", idx, normalizeToken(loop.State)))
		lines = append(lines, fmt.Sprintf("loop.%d.action=%s", idx, normalizeToken(loop.Action)))
		lines = append(lines, fmt.Sprintf("loop.%d.next=%s", idx, normalizeToken(loop.Next)))
	}
	lines[1] = fmt.Sprintf("count=%d", idx)
	return resultWithIntents(strings.Join(lines, "\n"), correctiveLoopIntents("Model corrective loops", "Model remediation flow checkpoints.", correctiveLoopsForState(state))...)
}

func renderProviderCorrectiveLoops(state *RuntimeState) Result {
	lines := []string{"PROVIDER_LOOPS", fmt.Sprintf("count=%d", len(correctiveLoopsForState(state)))}
	idx := 0
	for _, loop := range correctiveLoopsForState(state) {
		if loop.Area != "provider" && loop.Area != "permissions" && loop.Area != "settings" {
			continue
		}
		idx++
		lines = append(lines, fmt.Sprintf("loop.%d.area=%s", idx, normalizeToken(loop.Area)))
		lines = append(lines, fmt.Sprintf("loop.%d.state=%s", idx, normalizeToken(loop.State)))
		lines = append(lines, fmt.Sprintf("loop.%d.action=%s", idx, normalizeToken(loop.Action)))
		lines = append(lines, fmt.Sprintf("loop.%d.next=%s", idx, normalizeToken(loop.Next)))
	}
	lines[1] = fmt.Sprintf("count=%d", idx)
	return resultWithIntents(strings.Join(lines, "\n"), correctiveLoopIntents("Provider corrective loops", "Provider remediation flow checkpoints.", correctiveLoopsForState(state))...)
}
