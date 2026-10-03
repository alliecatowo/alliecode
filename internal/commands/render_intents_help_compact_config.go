package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alliecatowo/alliecode/internal/types"
)

func helpIntents(query string, suggestions []Suggestion) []types.RenderIntent {
	trimmedQuery := strings.TrimSpace(query)
	if len(suggestions) == 0 {
		summary := "Slash command catalog is empty."
		if trimmedQuery != "" {
			summary = fmt.Sprintf("No slash commands matched %q.", trimmedQuery)
		}
		return []types.RenderIntent{
			summaryCardIntent("Help", summary,
				field("Query", defaultDash(trimmedQuery)),
				field("Matches", "0"),
			),
		}
	}

	options := make([]types.RenderOption, 0, len(suggestions))
	for _, suggestion := range suggestions {
		detail := strings.TrimSpace(suggestion.Description)
		if usage := strings.TrimSpace(suggestion.Usage); usage != "" {
			detail = strings.TrimSpace(detail + " | " + usage)
		}
		status := strings.TrimSpace(suggestion.Category)
		if status == "" {
			status = strings.TrimSpace(suggestion.Group)
		}
		options = append(options, option("/"+suggestion.Name, detail, status, suggestion.ArgumentHint, false))
	}

	rows := []types.RenderDetailRow{
		detailRow("Query", defaultDash(trimmedQuery), "info", "Slash command filter text."),
		detailRow("Matches", fmt.Sprintf("%d", len(suggestions)), "info", "Commands matching the current query."),
	}
	return []types.RenderIntent{
		detailRowsIntent("Help search", "Slash command discovery metadata.", rows...),
		optionListIntent("Matching commands", "Structured command list for palette and timeline rendering.", options...),
	}
}

func compactIntents(mode string, requested bool, count int, lastTarget string) []types.RenderIntent {
	actions := []types.RenderAction{
		action("Compact now", "/compact now", "Request a compaction on the next turn.", statusWord(requested, "pending", "available")),
		action("Auto mode", "/compact auto", "Re-enable automatic compaction.", statusWord(mode == "auto", "current", "mode")),
		action("Disable auto mode", "/compact off", "Stop automatic compaction requests.", statusWord(mode == "off", "current", "mode")),
	}
	return []types.RenderIntent{
		summaryCardIntent("Compaction", "Conversation compaction mode and request state.",
			field("Mode", defaultDash(mode)),
			field("Requested", boolState(requested, "yes", "no")),
			field("Count", fmt.Sprintf("%d", count)),
			field("Last target", defaultDash(lastTarget)),
		),
		actionListIntent("Compaction actions", "Common follow-up actions for compaction control.", actions...),
	}
}

func configPanelIntents() []types.RenderIntent {
	return []types.RenderIntent{
		summaryCardIntent("Config panel", "Open the interactive config panel instead of dumping raw settings.", field("Opened", "yes"), field("Mode", "interactive")),
		actionListIntent("Config actions", "Suggested next config actions.",
			action("Show config", "/config show", "Inspect effective key/value pairs.", "browse"),
			action("Run doctor", "/config doctor", "Inspect missing required settings.", "diagnostics"),
			action("Repair", "/config repair", "Apply deterministic config defaults.", "repair"),
		),
	}
}

func configShowIntents(values map[string]string) []types.RenderIntent {
	if len(values) == 0 {
		return []types.RenderIntent{summaryCardIntent("Config", "No config values are currently set.", field("Entries", "0"))}
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	options := make([]types.RenderOption, 0, len(keys))
	for _, key := range keys {
		options = append(options, option(key, values[key], configKeyStatus(key), "/config get "+key, false))
	}
	return []types.RenderIntent{
		summaryCardIntent("Config", "Effective config values for the current runtime.", field("Entries", fmt.Sprintf("%d", len(keys)))),
		optionListIntent("Config entries", "Structured config inventory.", options...),
	}
}

func configGetIntents(key string, found bool, value string) []types.RenderIntent {
	status := "missing"
	if found {
		status = "found"
	}
	return []types.RenderIntent{
		detailRowsIntent("Config key", "Single config lookup result.",
			detailRow("Key", key, "info", "Requested config key."),
			detailRow("Found", boolState(found, "yes", "no"), status, "Whether the key exists in the effective config map."),
			detailRow("Value", defaultDash(value), status, "Normalized config value."),
		),
	}
}

func configRepairIntents(profile string, changed bool, repairs int, values map[string]string) []types.RenderIntent {
	return []types.RenderIntent{
		summaryCardIntent("Config repair", "Applied deterministic config defaults and profile overrides.",
			field("Profile", defaultDash(profile)),
			field("Changed", boolState(changed, "yes", "no")),
			field("Repairs", fmt.Sprintf("%d", repairs)),
		),
		detailRowsIntent("Resulting defaults", "Config keys touched or validated by repair.",
			detailRow("settings.output-style", values["settings.output-style"], configKeyStatus("settings.output-style"), "Render style used for command output."),
			detailRow("settings.output-format", values["settings.output-format"], configKeyStatus("settings.output-format"), "Output encoding / transport payload style."),
			detailRow("settings.transport", values["settings.transport"], configKeyStatus("settings.transport"), "Local vs remote transport selection."),
			detailRow("permissions.mode", values["permissions.mode"], configKeyStatus("permissions.mode"), "Permission default selected by safe/strict profiles."),
		),
	}
}

func configMutationIntents(title, key, value string, changed bool) []types.RenderIntent {
	status := "unchanged"
	if changed {
		status = "updated"
	}
	return []types.RenderIntent{
		detailRowsIntent(title, "Single config key mutation result.",
			detailRow("Key", key, "info", "Mutated config key."),
			detailRow("Value", defaultDash(value), status, "New effective value."),
			detailRow("Changed", boolState(changed, "yes", "no"), status, "Whether runtime state changed."),
		),
	}
}

func configStatusIntents(state *RuntimeState, missing int) []types.RenderIntent {
	count := 0
	doctorRuns := 0
	repairs := 0
	lastProfile := "-"
	if state != nil {
		count = len(state.ConfigValues)
		doctorRuns = state.ConfigDoctorCount
		repairs = state.ConfigRepairCount
		lastProfile = defaultDash(state.ConfigLastRepairProfile)
	}
	return []types.RenderIntent{
		summaryCardIntent("Config status", "Required settings coverage and repair activity.",
			field("Entries", fmt.Sprintf("%d", count)),
			field("Missing required", fmt.Sprintf("%d", missing)),
			field("Repairs", fmt.Sprintf("%d", repairs)),
			field("Doctor runs", fmt.Sprintf("%d", doctorRuns)),
			field("Last profile", lastProfile),
		),
		actionListIntent("Config actions", "Suggested follow-up actions for config state.",
			action("Repair config", "/config repair", "Apply runtime defaults for required config keys.", statusWord(missing == 0, "optional", "recommended")),
			action("Run doctor", "/config doctor", "Inspect missing required settings and suggested fixes.", "diagnostics"),
		),
	}
}

func configDoctorIntents(state *RuntimeState, missing int, quickFix string) []types.RenderIntent {
	status := "healthy"
	if missing > 0 {
		status = "needs attention"
	}
	lastProfile := "-"
	if state != nil {
		lastProfile = defaultDash(state.ConfigLastRepairProfile)
	}
	return []types.RenderIntent{
		diagnosticsIntent("Config doctor", status, "Required settings health and deterministic repair guidance.", []types.RenderDiagnostic{
			diagnostic("Missing required", statusWord(missing == 0, "ok", "warn"), fmt.Sprintf("%d", missing)),
			diagnostic("Last profile", "info", lastProfile),
		}, nil),
		actionListIntent("Config repair actions", "Recommended next config commands.",
			action("Quick fix", quickFix, "Apply the recommended repair profile.", statusWord(missing == 0, "optional", "recommended")),
			action("Show config", "/config show", "Inspect the full effective config map.", "browse"),
		),
	}
}

func configKeyStatus(key string) string {
	key = strings.TrimSpace(strings.ToLower(key))
	switch {
	case strings.HasPrefix(key, "settings."):
		return "setting"
	case strings.HasPrefix(key, "permissions."):
		return "permission"
	default:
		return "config"
	}
}
