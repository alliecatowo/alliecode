package commands

import (
	"fmt"

	"github.com/alliecatowo/alliecode/internal/types"
)

func resumeStatusIntents(requested bool, count int, lastTarget string) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Resume status", "Current resume request state.",
		field("Requested", boolState(requested, "yes", "no")),
		field("Count", fmt.Sprintf("%d", count)),
		field("Last target", defaultDash(lastTarget)),
	)}
}

func resumeRequestIntents(target string, count int) []types.RenderIntent {
	return []types.RenderIntent{
		summaryCardIntent("Resume requested", "Queued session resume target.",
			field("Target", defaultDash(target)),
			field("Requested", "yes"),
			field("Count", fmt.Sprintf("%d", count)),
		),
		actionHintsIntent("Actions", hint("Inspect status", "/resume status")),
	}
}

func branchStatusIntents(active string, count, created, switches int) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Branch status", "Current branch and mutation counters.",
		field("Active", defaultDash(active)),
		field("Branches", fmt.Sprintf("%d", count)),
		field("Created", fmt.Sprintf("%d", created)),
		field("Switches", fmt.Sprintf("%d", switches)),
	)}
}

func branchListIntents(active string, branches []string) []types.RenderIntent {
	rows := make([]types.RenderTableRow, 0, len(branches))
	for _, branch := range branches {
		rows = append(rows, tableRow(branch, statusWord(branch == active, "active", "tracked")))
	}
	return []types.RenderIntent{tableIntent("Branches", "Known local branch inventory.", []string{"Branch", "State"}, rows...)}
}

func branchMutationIntents(title, active, previous string, changed bool, count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent(title, "Branch mutation result.",
		detailRow("Active", defaultDash(active), statusWord(changed, "updated", "unchanged"), "Current active branch."),
		detailRow("Previous", defaultDash(previous), "info", "Previously active branch."),
		detailRow("Switches", fmt.Sprintf("%d", count), "info", "Total switch operations."),
	)}
}

func diffStatusIntents(mode string, entries, updates, clears int) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Diff status", "Current diff-mode state and counters.",
		field("Mode", defaultDash(mode)),
		field("Entries", fmt.Sprintf("%d", entries)),
		field("Updates", fmt.Sprintf("%d", updates)),
		field("Clears", fmt.Sprintf("%d", clears)),
	)}
}

func diffModeIntents(mode string) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Diff mode", "Updated active diff mode.", field("Mode", defaultDash(mode)))}
}

func diffMutationIntents(path string, added, removed, modified, entries, updates int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Diff entry upsert", "Added or updated deterministic diff row.",
		detailRow("Path", defaultDash(path), "tracked", "Path key for this diff row."),
		detailRow("Added", fmt.Sprintf("%d", added), "info", "Added line count."),
		detailRow("Removed", fmt.Sprintf("%d", removed), "info", "Removed line count."),
		detailRow("Modified", fmt.Sprintf("%d", modified), "info", "Modified line count."),
		detailRow("Entries", fmt.Sprintf("%d", entries), "info", "Total tracked diff rows."),
		detailRow("Updates", fmt.Sprintf("%d", updates), "info", "Total diff updates."),
	)}
}

func diffClearIntents(updates, clears int) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Diff cleared", "Cleared tracked diff rows.", field("Entries", "0"), field("Updates", fmt.Sprintf("%d", updates)), field("Clears", fmt.Sprintf("%d", clears)))}
}

func costBreakdownIntents(input, output, cacheRead, cacheWrite int64) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Cost breakdown", "Token usage counters for this session.",
		field("Input tokens", fmt.Sprintf("%d", input)),
		field("Output tokens", fmt.Sprintf("%d", output)),
		field("Cache read", fmt.Sprintf("%d", cacheRead)),
		field("Cache write", fmt.Sprintf("%d", cacheWrite)),
	)}
}

func usageSnapshotIntents(model, permission string, compactRequested bool) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Usage snapshot", "Current session usage selectors.",
		field("Model", defaultDash(model)),
		field("Permission", defaultDash(permission)),
		field("Compact requested", boolState(compactRequested, "yes", "no")),
	)}
}

func contextStateIntents(state *RuntimeState) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Context state", "Current local context control flags.",
		detailRow("Compact requested", boolState(state.CompactRequested, "yes", "no"), "info", "Whether compaction is queued."),
		detailRow("Resume requested", boolState(state.ResumeRequested, "yes", "no"), "info", "Whether resume is queued."),
		detailRow("Last resume target", defaultDash(state.LastResumeTarget), "info", "Most recent resume target."),
		detailRow("Last compact target", defaultDash(state.LastCompactTarget), "info", "Most recent compact target."),
		detailRow("Compact mode", defaultDash(state.CompactMode), "info", "Active compaction mode."),
		detailRow("Permission mode", modeString(state.PermissionMode), "info", "Current permission mode."),
		detailRow("Model", defaultDash(state.Model), "info", "Current model id."),
		detailRow("Provider", defaultDash(state.ProviderName), "info", "Current provider id."),
	)}
}

func contextClearedIntents() []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Context cleared", "Reset local context control flags.", field("Compact requested", "no"), field("Resume requested", "no"), field("Last resume target", "-"), field("Last compact target", "-"))}
}

func clearStatusIntents(state *RuntimeState) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Clear status", "Clear counters and scope state.",
		field("Count", fmt.Sprintf("%d", state.ClearCount)),
		field("Context clears", fmt.Sprintf("%d", state.ClearContextCount)),
		field("Diff clears", fmt.Sprintf("%d", state.ClearDiffCount)),
		field("Last scope", defaultDash(state.LastClearScope)),
	)}
}

func clearResultIntents(scope string, clearDisplay, clearContext, clearDiff bool, count, contextClears, diffClears int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Clear result", "Applied selected clear scope.",
		detailRow("Scope", defaultDash(scope), "info", "Requested clear scope."),
		detailRow("Display", boolState(clearDisplay, "yes", "no"), statusWord(clearDisplay, "cleared", "untouched"), "Display transcript clearing."),
		detailRow("Context", boolState(clearContext, "yes", "no"), statusWord(clearContext, "cleared", "untouched"), "Context flag reset."),
		detailRow("Diff", boolState(clearDiff, "yes", "no"), statusWord(clearDiff, "cleared", "untouched"), "Diff row reset."),
		detailRow("Count", fmt.Sprintf("%d", count), "info", "Display clear counter."),
		detailRow("Context clears", fmt.Sprintf("%d", contextClears), "info", "Context clear counter."),
		detailRow("Diff clears", fmt.Sprintf("%d", diffClears), "info", "Diff clear counter."),
	)}
}

func filesStatusIntents(count, projectPaths, adds, removes, clears int, lastAction, lastPath string) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Files status", "Context-file inventory and mutation counters.",
		field("Entries", fmt.Sprintf("%d", count)),
		field("Project paths", fmt.Sprintf("%d", projectPaths)),
		field("Adds", fmt.Sprintf("%d", adds)),
		field("Removes", fmt.Sprintf("%d", removes)),
		field("Clears", fmt.Sprintf("%d", clears)),
		field("Last action", defaultDash(lastAction)),
		field("Last path", defaultDash(lastPath)),
	)}
}

func filesMutationIntents(title, path string, changed bool, count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent(title, "Single context-file mutation result.",
		detailRow("Path", defaultDash(path), "info", "Path provided to /files."),
		detailRow("Changed", boolState(changed, "yes", "no"), statusWord(changed, "updated", "noop"), "Whether state changed."),
		detailRow("Count", fmt.Sprintf("%d", count), "info", "Current context-file count."),
	)}
}

func filesListIntents(files []string) []types.RenderIntent {
	rows := make([]types.RenderTableRow, 0, len(files))
	for _, file := range files {
		rows = append(rows, tableRow(file))
	}
	return []types.RenderIntent{tableIntent("Files", "Current deterministic context-file list.", []string{"Path"}, rows...)}
}

func themeStatusIntents(theme string) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Theme status", "Current output theme mode.", field("Theme", defaultDash(theme)))}
}

func themeListIntents() []types.RenderIntent {
	return []types.RenderIntent{optionListIntent("Themes", "Supported theme modes.", option("dark", "Dark theme", "available", "/theme set dark", false), option("light", "Light theme", "available", "/theme set light", false), option("system", "Use terminal/system preference", "available", "/theme set system", false))}
}

func themeMutationIntents(title, theme string, setCount int) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent(title, "Theme mutation result.", field("Theme", defaultDash(theme)), field("Set count", fmt.Sprintf("%d", setCount)))}
}

func themePreviewIntents(theme string) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Theme preview", "Previewed theme without applying it.", detailRow("Theme", defaultDash(theme), "preview", "No persisted change applied."))}
}

func outputStyleStatusIntents(style string, setCount int) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Output style", "Current output formatting style.", field("Style", defaultDash(style)), field("Set count", fmt.Sprintf("%d", setCount)))}
}

func outputStyleListIntents() []types.RenderIntent {
	return []types.RenderIntent{optionListIntent("Output styles", "Supported output style modes.", option("default", "Balanced formatting", "available", "/output-style set default", false), option("concise", "Short responses", "available", "/output-style set concise", false), option("explanatory", "Detailed responses", "available", "/output-style set explanatory", false), option("json", "Machine-oriented responses", "available", "/output-style set json", false))}
}

func memoryStatusIntents(state *RuntimeState) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Memory status", "Deterministic memory inventory counters.",
		field("Entries", fmt.Sprintf("%d", len(state.MemoryEntries))),
		field("Writes", fmt.Sprintf("%d", state.MemoryWrites)),
		field("Removes", fmt.Sprintf("%d", state.MemoryRemoves)),
		field("Clears", fmt.Sprintf("%d", state.MemoryClears)),
		field("Last", defaultDash(state.LastMemory)),
	)}
}

func memoryListIntents(entries []string) []types.RenderIntent {
	items := make([]types.RenderChecklistItem, 0, len(entries))
	for _, entry := range entries {
		items = append(items, checklistItem(entry, true, "stored"))
	}
	return []types.RenderIntent{checklistIntent("Memory entries", "Stored deterministic memory rows.", items...)}
}

func memoryMutationIntents(title, entry string, count, updates int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent(title, "Memory mutation result.",
		detailRow("Entry", defaultDash(entry), "info", "Affected memory row."),
		detailRow("Count", fmt.Sprintf("%d", count), "info", "Current memory count."),
		detailRow("Updates", fmt.Sprintf("%d", updates), "info", "Mutation counter for this operation."),
	)}
}

func privacyStatusIntents(telemetry, training bool, updates int) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Privacy settings", "Telemetry and training preference state.",
		field("Telemetry", boolState(telemetry, "on", "off")),
		field("Training", boolState(training, "on", "off")),
		field("Updates", fmt.Sprintf("%d", updates)),
	)}
}

func privacyFieldsIntents() []types.RenderIntent {
	return []types.RenderIntent{optionListIntent("Privacy fields", "Configurable privacy controls.", option("telemetry", "Diagnostic telemetry collection", "field", "/privacy-settings set telemetry <on|off>", false), option("training", "Model training feedback sharing", "field", "/privacy-settings set training <on|off>", false))}
}

func privacySetIntents(field string, enabled, updated bool, updates int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Privacy update", "Single privacy field mutation result.",
		detailRow("Field", defaultDash(field), "info", "Updated field name."),
		detailRow("Enabled", boolState(enabled, "yes", "no"), statusWord(enabled, "on", "off"), "New field state."),
		detailRow("Updated", boolState(updated, "yes", "no"), statusWord(updated, "changed", "noop"), "Whether value changed."),
		detailRow("Updates", fmt.Sprintf("%d", updates), "info", "Total update count."),
	)}
}

func upgradeStatusIntents(requested bool, count int, lastPlan string) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Upgrade status", "Upgrade guidance request state.",
		field("Requested", boolState(requested, "yes", "no")),
		field("Count", fmt.Sprintf("%d", count)),
		field("Last plan", defaultDash(lastPlan)),
	)}
}

func upgradeRequestIntents(plan string, count int) []types.RenderIntent {
	return []types.RenderIntent{
		summaryCardIntent("Upgrade requested", "Requested provider upgrade guidance.", field("Plan", defaultDash(plan)), field("Requested", "yes"), field("Count", fmt.Sprintf("%d", count))),
		actionHintsIntent("Actions", hint("Next", "complete upgrade in provider portal")),
	}
}

func planStatusIntents(enabled bool, enableCount, openCount int, lastDescription string) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Plan status", "Plan-mode runtime state.",
		field("Enabled", boolState(enabled, "yes", "no")),
		field("Enable count", fmt.Sprintf("%d", enableCount)),
		field("Open count", fmt.Sprintf("%d", openCount)),
		field("Last description", defaultDash(lastDescription)),
	)}
}

func planMutationIntents(title, description string, queryHint bool, openCount int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent(title, "Plan-mode action result.",
		detailRow("Description", defaultDash(description), "info", "Plan query or saved description."),
		detailRow("Query hint", boolState(queryHint, "yes", "no"), statusWord(queryHint, "captured", "none"), "Whether input was interpreted as a plan request."),
		detailRow("Open count", fmt.Sprintf("%d", openCount), "info", "How often plan file was opened."),
	)}
}

func exitRequestIntents(count int, command string) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Exit requested", "Current session exit was requested.", field("Requested", "yes"), field("Count", fmt.Sprintf("%d", count)), field("Command", defaultDash(command)))}
}

func copyResultIntents(text string, count int, last string) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Copy result", "Clipboard copy command payload summary.",
		detailRow("Text", defaultDash(text), "info", "Normalized copied text."),
		detailRow("Count", fmt.Sprintf("%d", count), "info", "Copy command invocation count."),
		detailRow("Last text", defaultDash(last), "info", "Last copied text payload."),
	)}
}

func versionInfoIntents(runtime, version string) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Version info", "Runtime version metadata.", field("Runtime", defaultDash(runtime)), field("Version", defaultDash(version)))}
}
