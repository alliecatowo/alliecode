package commands

import (
	"fmt"
	"strings"

	"github.com/alliecatowo/alliecode/internal/types"
)

func rewindStatusIntents(count int, lastTarget string) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Rewind status", "Current rewind request state.",
		field("Count", fmt.Sprintf("%d", count)),
		field("Last target", defaultDash(lastTarget)),
	)}
}

func rewindRequestIntents(target string, count int) []types.RenderIntent {
	return []types.RenderIntent{
		summaryCardIntent("Rewind requested", "Queued rewind request for the current session.",
			field("Target", defaultDash(target)),
			field("Count", fmt.Sprintf("%d", count)),
			field("Requested", "yes"),
		),
		actionHintsIntent("Actions", hint("Review status", "/rewind status")),
	}
}

func tagStatusIntents(tag string, updates int) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Tag status", "Current session tag and update counter.",
		field("Current", defaultDash(tag)),
		field("Updates", fmt.Sprintf("%d", updates)),
	)}
}

func tagMutationIntents(title, tag string, updates int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent(title, "Session tag mutation result.",
		detailRow("Tag", defaultDash(tag), "info", "Affected session tag."),
		detailRow("Updates", fmt.Sprintf("%d", updates), "info", "Total tag updates."),
	)}
}

func remoteEnvStatusIntents(env string, updates int) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Remote environment", "Current default remote execution environment.",
		field("Environment", defaultDash(env)),
		field("Updates", fmt.Sprintf("%d", updates)),
	)}
}

func remoteEnvListIntents() []types.RenderIntent {
	return []types.RenderIntent{optionListIntent("Remote environments", "Supported remote execution presets.",
		option("default", "Balanced baseline environment", "available", "/remote-env set default", false),
		option("code-review", "Optimized for review and diagnostics", "available", "/remote-env set code-review", false),
		option("hardened-linux", "Restricted hardened Linux profile", "available", "/remote-env set hardened-linux", false),
	)}
}

func remoteEnvSetIntents(env string, updates int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Remote environment updated", "Applied remote environment selection.",
		detailRow("Environment", defaultDash(env), "set", "Active remote environment preset."),
		detailRow("Updates", fmt.Sprintf("%d", updates), "info", "Total environment updates."),
	)}
}

func securityReviewStatusIntents(count int, lastTarget string) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Security review", "Security review request status.",
		field("Count", fmt.Sprintf("%d", count)),
		field("Last target", defaultDash(lastTarget)),
	)}
}

func securityReviewRequestIntents(target string, count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Security review requested", "Queued focused security review.",
		detailRow("Target", defaultDash(target), "queued", "Requested review target."),
		detailRow("Count", fmt.Sprintf("%d", count), "info", "Total security review requests."),
		detailRow("Mode", "focused", "info", "Review execution mode."),
	)}
}

func addDirIntents(path string) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Workspace directory added", "Added workspace directory to current session context.",
		field("Path", defaultDash(path)),
	)}
}

func agentsListIntents() []types.RenderIntent {
	return []types.RenderIntent{tableIntent("Agents", "Available local agent profiles.", []string{"Name", "Status"},
		tableRow("local", "available"),
	)}
}

func agentsCreateIntents(name string) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Agent created", "Agent profile creation result.",
		detailRow("Name", defaultDash(name), "created", "Created agent profile name."),
		detailRow("Status", "created", "ok", "Creation state."),
	)}
}

func agentsStatusIntents(configured bool) []types.RenderIntent {
	status := "none"
	if configured {
		status = "configured"
	}
	return []types.RenderIntent{summaryCardIntent("Agent status", "Current agent configuration status.",
		field("Status", status),
	)}
}

func vimStatusIntents(enabled bool) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Vim mode", "Current vim keybinding mode.", field("Enabled", boolState(enabled, "yes", "no")))}
}

func vimSetIntents(enabled bool) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Vim mode updated", "Applied vim mode mutation.",
		detailRow("Enabled", boolState(enabled, "yes", "no"), statusWord(enabled, "enabled", "disabled"), "Current vim mode state."),
	)}
}

func voiceStatusIntents(enabled bool) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Voice mode", "Local voice-mode placeholder controls.",
		field("Mode", "local-placeholder"),
		field("Enabled", boolState(enabled, "yes", "no")),
	)}
}

func voiceSetIntents(enabled bool) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Voice mode updated", "Applied voice mode mutation.",
		detailRow("Mode", "local-placeholder", "info", "Voice integration mode."),
		detailRow("Enabled", boolState(enabled, "yes", "no"), statusWord(enabled, "enabled", "disabled"), "Current voice state."),
	)}
}

func buddyHelpIntents() []types.RenderIntent {
	return []types.RenderIntent{optionListIntent("Buddy commands", "Companion controls.",
		option("status", "Show buddy state", "command", "/buddy status", false),
		option("hatch", "Hatch buddy egg", "command", "/buddy hatch", false),
		option("pet", "Pet your buddy", "command", "/buddy pet", false),
		option("mute", "Mute buddy reactions", "command", "/buddy mute", false),
		option("unmute", "Unmute buddy reactions", "command", "/buddy unmute", false),
	)}
}

func buddyStatusIntents(state, summary string, hatched, muted bool, petCount int) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Buddy status", strings.TrimSpace(summary),
		field("State", defaultDash(state)),
		field("Hatched", boolState(hatched, "yes", "no")),
		field("Muted", boolState(muted, "yes", "no")),
		field("Pet count", fmt.Sprintf("%d", petCount)),
	)}
}

func buddyActionIntents(title, status string, hatched, muted bool, petCount int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent(title, "Buddy action result.",
		detailRow("Status", defaultDash(status), "info", "Command result status."),
		detailRow("Hatched", boolState(hatched, "yes", "no"), "info", "Hatch state."),
		detailRow("Muted", boolState(muted, "yes", "no"), "info", "Mute state."),
		detailRow("Pet count", fmt.Sprintf("%d", petCount), "info", "Number of pets."),
	)}
}

func statuslineStatusIntents(state *RuntimeState, runtime types.AgentRuntimeSnapshot) []types.RenderIntent {
	if state == nil {
		state = &RuntimeState{}
	}
	return []types.RenderIntent{summaryCardIntent("Statusline", "Statusline runtime snapshot.",
		field("Count", fmt.Sprintf("%d", state.StatuslineCount)),
		field("Last prompt", defaultDash(state.LastStatusline)),
		field("Provider", defaultDash(runtime.ProviderName)),
		field("Model", defaultDash(runtime.ModelRef)),
		field("Logged in", boolState(runtime.LoggedIn, "yes", "no")),
		field("Provider ready", boolState(runtime.ProviderReady, "yes", "no")),
		field("Last render ms", fmt.Sprintf("%d", state.StatuslineLastRenderMS)),
		field("Tool calls", fmt.Sprintf("%d", state.StatuslineToolCalls)),
		field("Permission events", fmt.Sprintf("%d", state.StatuslinePermissions)),
		field("Transitions", fmt.Sprintf("%d", state.StatuslineTransitions)),
		field("Turns", fmt.Sprintf("%d", runtime.Turns)),
		field("Tool inflight", fmt.Sprintf("%d", runtime.ToolInflight)),
		field("Tasks total", fmt.Sprintf("%d", runtime.TasksTotal)),
		field("Tasks running", fmt.Sprintf("%d", runtime.TasksRunning)),
		field("Tasks completed", fmt.Sprintf("%d", runtime.TasksCompleted)),
		field("Teams total", fmt.Sprintf("%d", runtime.TeamsTotal)),
		field("Teams active", fmt.Sprintf("%d", runtime.TeamsActive)),
		field("Last stop reason", defaultDash(string(runtime.LastStopReason))),
	)}
}

func statuslineSetupIntents(prompt string, count int) []types.RenderIntent {
	return []types.RenderIntent{
		summaryCardIntent("Statusline setup", "Queued statusline setup prompt.",
			field("Subagent type", "statusline-setup"),
			field("Prompt", defaultDash(prompt)),
			field("Count", fmt.Sprintf("%d", count)),
		),
		actionHintsIntent("Actions", hint("Check status", "/statusline status")),
	}
}

func ideStatusIntents(editor, detected, source string, openCount, configCount int, hints []string) []types.RenderIntent {
	rows := make([]types.RenderOption, 0, len(hints))
	for _, h := range hints {
		rows = append(rows, option(h, "Launch hint", "hint", "", false))
	}
	return []types.RenderIntent{
		summaryCardIntent("IDE status", "Editor detection and open-hint status.",
			field("Editor", defaultDash(editor)),
			field("Detected", defaultDash(detected)),
			field("Source", defaultDash(source)),
			field("Open count", fmt.Sprintf("%d", openCount)),
			field("Config count", fmt.Sprintf("%d", configCount)),
		),
		optionListIntent("IDE hints", "Editor open command hints.", rows...),
	}
}

func ideDetectIntents(editor, source string, hints []string) []types.RenderIntent {
	opts := make([]types.RenderOption, 0, len(hints))
	for _, h := range hints {
		opts = append(opts, option(h, "Detected launch hint", "hint", "", false))
	}
	return []types.RenderIntent{
		summaryCardIntent("IDE detect", "Detected editor from environment.", field("Editor", defaultDash(editor)), field("Source", defaultDash(source))),
		optionListIntent("Hints", "Suggested editor commands.", opts...),
	}
}

func ideSetEditorIntents(editor string, configCount int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("IDE editor set", "Configured preferred IDE editor.",
		detailRow("Editor", defaultDash(editor), "set", "Configured editor name."),
		detailRow("Config count", fmt.Sprintf("%d", configCount), "info", "Editor config updates."),
	)}
}

func ideOpenHintIntents(editor, target, command string, openCount int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("IDE open hint", "Suggested command to open editor target.",
		detailRow("Editor", defaultDash(editor), "info", "Resolved editor."),
		detailRow("Target", defaultDash(target), "info", "Requested open target."),
		detailRow("Command", defaultDash(command), "hint", "Suggested shell command."),
		detailRow("Open count", fmt.Sprintf("%d", openCount), "info", "Open hint invocations."),
	)}
}

func keybindingsOpenDisabledIntents() []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Keybindings", "Keybinding customization is currently disabled in preview mode.", field("Enabled", "no"))}
}

func keybindingsOpenIntents(path string, exists bool) []types.RenderIntent {
	title := "Keybindings opened"
	if !exists {
		title = "Keybindings created"
	}
	status := "opened"
	if !exists {
		status = "created"
	}
	return []types.RenderIntent{detailRowsIntent(title, "Opened keybindings file in editor.",
		detailRow("Path", defaultDash(path), status, "Keybindings file path."),
	)}
}

func keybindingsStatusIntents(enabled bool, path string, exists bool, opens int) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Keybindings status", "Keybinding configuration file status.",
		field("Enabled", boolState(enabled, "yes", "no")),
		field("Path", defaultDash(path)),
		field("Exists", boolState(exists, "yes", "no")),
		field("Opens", fmt.Sprintf("%d", opens)),
	)}
}

func keybindingsPathIntents(path string) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Keybindings path", "Resolved keybindings file path.", field("Path", defaultDash(path)))}
}

func keybindingsSetIntents(enabled bool) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Keybindings updated", "Keybindings enablement changed.",
		detailRow("Enabled", boolState(enabled, "yes", "no"), statusWord(enabled, "enabled", "disabled"), "Current keybindings setting."),
	)}
}
