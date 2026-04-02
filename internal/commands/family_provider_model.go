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
	if len(inv.Args) > 0 {
		sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
		switch sub {
		case "loops", "loop":
			return renderModelCorrectiveLoops(cmdCtx.State), nil
		case "doctor":
			providerName := strings.TrimSpace(cmdCtx.State.ProviderName)
			modelRaw := strings.TrimSpace(cmdCtx.State.Model)
			ready := providerName != "" && modelRaw != ""
			quickFix := "/provider set ollama"
			if providerName != "" {
				quickFix = "/model " + providerName + "/<model>"
			}
			if modelRaw != "" {
				quickFix = "/status"
			}
			return Result{Handled: true, Message: fmt.Sprintf("MODEL_DOCTOR\nprovider=%s\nmodel=%s\nready=%t\nquick_fix=%s", normalizeToken(providerName), normalizeToken(modelRaw), ready, normalizeToken(quickFix))}, nil
		case "list":
			providerName := strings.TrimSpace(cmdCtx.State.ProviderName)
			if len(inv.Args) == 2 {
				providerName = strings.ToLower(strings.TrimSpace(inv.Args[1]))
			}
			if providerName == "" {
				return Result{}, fmt.Errorf("usage: /model [provider/model|model|list [provider]|doctor|repair [provider/model]]")
			}
			models := providers.ListModelsByProvider(providerName)
			lines := []string{"MODEL_LIST", fmt.Sprintf("provider=%s", normalizeToken(providerName)), fmt.Sprintf("count=%d", len(models))}
			for i, model := range models {
				lines = append(lines, fmt.Sprintf("model.%d.id=%s", i+1, normalizeToken(model.Model)))
				lines = append(lines, fmt.Sprintf("model.%d.context_window=%d", i+1, model.ContextWindow))
			}
			return Result{Handled: true, Message: strings.Join(lines, "\n")}, nil
		case "repair", "fix":
			target := ""
			if len(inv.Args) > 1 {
				target = strings.TrimSpace(strings.Join(inv.Args[1:], " "))
			}
			if target == "" {
				providerName := strings.TrimSpace(cmdCtx.State.ProviderName)
				if providerName == "" {
					providerName = "ollama"
				}
				models := providers.ListModelsByProvider(providerName)
				if len(models) > 0 {
					target = providerName + "/" + models[0].Model
				}
			}
			if target == "" {
				return Result{}, fmt.Errorf("usage: /model [provider/model|model|list [provider]|doctor|repair [provider/model]]")
			}
			providerName, modelName, err := parseModelSelection(target, cmdCtx.State.ProviderName)
			if err != nil {
				return Result{}, err
			}
			if _, ok := providers.LookupModelMetadata(providerName, modelName); !ok {
				return Result{}, fmt.Errorf("unknown model %q for provider %q", modelName, providerName)
			}
			cmdCtx.State.ProviderName = providerName
			cmdCtx.State.Model = providerName + "/" + modelName
			cmdCtx.State.ProviderReady = providerName == "ollama" || cmdCtx.State.LoggedIn
			return Result{Handled: true, Message: fmt.Sprintf("MODEL_REPAIR\nprovider=%s\nmodel=%s\nprovider_ready=%t", normalizeToken(providerName), normalizeToken(cmdCtx.State.Model), cmdCtx.State.ProviderReady)}, nil
		}
	}

	if len(inv.Args) == 0 {
		current := cmdCtx.State.Model
		if current == "" {
			current = "unknown"
		}
		summary := "-"
		if providerName, modelName, ok := resolveModelForStatus(current, cmdCtx.State.ProviderName); ok {
			if capabilitySummary, found := providers.LookupModelCapabilitySummary(providerName, modelName); found {
				summary = capabilitySummary
			}
		}
		providerHint := strings.TrimSpace(cmdCtx.State.ProviderName)
		if providerHint == "" {
			providerHint = "<provider>"
		}
		return Result{Handled: true, Message: fmt.Sprintf("Current model: %s\nCapabilities: %s\nQuick fix: /model %s/<model>\nNext: use /model [provider/model|model] to switch, or /provider status to verify provider.", current, summary, providerHint)}, nil
	}

	next := strings.TrimSpace(inv.Args[0])
	if next == "" {
		return Result{}, fmt.Errorf("model cannot be empty")
	}
	providerName, modelName, err := parseModelSelection(next, cmdCtx.State.ProviderName)
	if err != nil {
		return Result{}, err
	}
	if _, ok := providers.LookupModelMetadata(providerName, modelName); !ok {
		if strings.Contains(next, "/") {
			return Result{}, fmt.Errorf("unknown model %q for provider %q", modelName, providerName)
		}
		return Result{}, fmt.Errorf("unknown model %q for configured provider %q", modelName, providerName)
	}
	cmdCtx.State.Model = next
	capabilitySummary := "-"
	if summary, ok := providers.LookupModelCapabilitySummary(providerName, modelName); ok {
		capabilitySummary = summary
	}
	return Result{Handled: true, Message: fmt.Sprintf("Model set to %s\nCapabilities: %s\nQuick fix: /provider status\nNext: run /status to confirm runtime readiness.", next, capabilitySummary)}, nil
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
		providerName := strings.TrimSpace(cmdCtx.State.ProviderName)
		if providerName == "" {
			providerName = "-"
		}
		modelHint := strings.TrimSpace(cmdCtx.State.Model)
		if modelHint == "" {
			modelHint = "<provider/model>"
		}
		return Result{Handled: true, Message: fmt.Sprintf("PROVIDER_STATUS\nprovider=%s\nprovider_ready=%t\nquick_fix_model=/model_%s\nnext=use_/provider_set_<name>_or_/login_provider_<name>", normalizeToken(providerName), cmdCtx.State.ProviderReady, normalizeToken(modelHint))}, nil
	}
	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "list") {
		return Result{Handled: true, Message: providerListMessage()}, nil
	}
	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "doctor") {
		providerName := strings.TrimSpace(cmdCtx.State.ProviderName)
		if providerName == "" {
			providerName = "-"
		}
		modelRaw := strings.TrimSpace(cmdCtx.State.Model)
		modelOK := false
		if p, m, ok := resolveModelForStatus(modelRaw, cmdCtx.State.ProviderName); ok {
			_, modelOK = providers.LookupModelMetadata(p, m)
		}
		quickFix := "/provider set ollama"
		if providerName != "-" {
			quickFix = "/model " + strings.TrimSpace(cmdCtx.State.ProviderName) + "/<model>"
		}
		return Result{Handled: true, Message: fmt.Sprintf("PROVIDER_DOCTOR\nprovider=%s\nprovider_ready=%t\nmodel=%s\nmodel_valid=%t\nquick_fix=%s", normalizeToken(providerName), cmdCtx.State.ProviderReady, normalizeToken(modelRaw), modelOK, normalizeToken(quickFix))}, nil
	}
	if len(inv.Args) == 1 && (strings.EqualFold(strings.TrimSpace(inv.Args[0]), "loops") || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "loop")) {
		return renderProviderCorrectiveLoops(cmdCtx.State), nil
	}
	if len(inv.Args) >= 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "models") {
		providerName := strings.TrimSpace(cmdCtx.State.ProviderName)
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
		return Result{Handled: true, Message: strings.Join(lines, "\n")}, nil
	}

	setProvider := ""
	if len(inv.Args) >= 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "repair") {
		if len(inv.Args) > 2 {
			return Result{}, fmt.Errorf("%s", usage)
		}
		if len(inv.Args) == 2 {
			setProvider = strings.ToLower(strings.TrimSpace(inv.Args[1]))
		} else if strings.TrimSpace(cmdCtx.State.ProviderName) == "" {
			setProvider = "ollama"
		} else {
			setProvider = strings.ToLower(strings.TrimSpace(cmdCtx.State.ProviderName))
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
		return Result{}, fmt.Errorf("unknown provider %q", setProvider)
	}
	cmdCtx.State.ProviderName = setProvider
	cmdCtx.State.ProviderReady = setProvider == "ollama"
	resultType := "PROVIDER_SET"
	if len(inv.Args) >= 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "repair") {
		resultType = "PROVIDER_REPAIR"
	}
	return Result{Handled: true, Message: fmt.Sprintf("%s\nprovider=%s\nprovider_ready=%t\nquick_fix_model=/model_%s/<model>\nnext=run_/model_to_pick_a_model_for_%s", resultType, setProvider, cmdCtx.State.ProviderReady, setProvider, setProvider)}, nil
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
	return Result{Handled: true, Message: strings.Join(lines, "\n")}
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
	return Result{Handled: true, Message: strings.Join(lines, "\n")}
}
