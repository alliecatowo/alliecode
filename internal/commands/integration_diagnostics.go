package commands

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/alliecatowo/alliecode/internal/mcp"
	pluginspkg "github.com/alliecatowo/alliecode/internal/plugins"
	skillspkg "github.com/alliecatowo/alliecode/internal/skills"
)

func renderMCPDiagnostics(state *RuntimeState, manager *mcp.Manager) Result {
	if state == nil {
		state = &RuntimeState{}
	}
	servers := 0
	connected := 0
	authFailed := 0
	pending := 0
	if manager != nil {
		statuses := manager.ServerStatuses()
		servers = len(statuses)
		for _, status := range statuses {
			s := strings.ToLower(string(status.ConnectionState))
			if s == "connected" {
				connected++
			}
			if s == "failed" {
				authFailed++
			}
			if s == "pending" {
				pending++
			}
		}
	} else {
		servers = len(state.MCPConnections)
		for _, ok := range state.MCPConnections {
			if ok {
				connected++
			}
		}
	}
	quickFix := "/mcp doctor"
	if servers == 0 {
		quickFix = "/mcp add <name> stdio <command>"
	} else if connected == 0 {
		quickFix = "/mcp connect <name>"
	}
	lines := []string{
		"MCP_DIAGNOSTICS",
		fmt.Sprintf("manager=%t", manager != nil),
		fmt.Sprintf("servers=%d", servers),
		fmt.Sprintf("connected=%d", connected),
		fmt.Sprintf("pending=%d", pending),
		fmt.Sprintf("failed=%d", authFailed),
		fmt.Sprintf("doctor_runs=%d", state.MCPDoctorCount),
		fmt.Sprintf("repair_runs=%d", state.MCPRepairCount),
		fmt.Sprintf("last_quick_fix=%s", normalizeToken(state.MCPLastQuickFix)),
		fmt.Sprintf("quick_fix=%s", normalizeToken(quickFix)),
	}
	return resultWithIntents(strings.Join(lines, "\n"), mcpDiagnosticsIntents(servers, connected, pending, authFailed, state.MCPDoctorCount, state.MCPRepairCount, quickFix)...)
}

func repairMCPState(state *RuntimeState, manager *mcp.Manager, mode string) Result {
	if state == nil {
		state = &RuntimeState{}
	}
	changed := false
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		mode = "auto"
	}
	quickFix := "/mcp diagnostics"
	switch mode {
	case "auto", "reconnect":
		names := make([]string, 0, len(state.MCPConnections))
		for name := range state.MCPConnections {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			if !state.MCPConnections[name] {
				state.MCPConnections[name] = true
				changed = true
			}
		}
		quickFix = "/mcp status"
	case "cache":
		state.MCPConnections = map[string]bool{}
		changed = true
		quickFix = "/mcp list"
	case "auth":
		quickFix = "/mcp auth-status"
	default:
		quickFix = "/mcp repair auto"
	}
	state.MCPRepairCount++
	state.MCPLastQuickFix = quickFix
	message := fmt.Sprintf("MCP_REPAIR\nmode=%s\nchanged=%t\nquick_fix=%s\nrepair_runs=%d", normalizeToken(mode), changed, normalizeToken(quickFix), state.MCPRepairCount)
	return resultWithIntents(message, mcpRepairIntents(mode, changed, quickFix, state.MCPRepairCount)...)
}

func renderPluginDiagnostics(state *RuntimeState, result pluginspkg.ServiceListResult, conflicts []skillspkg.SkillConflictDiagnostic) Result {
	if state == nil {
		state = &RuntimeState{}
	}
	quickFix := "/plugin doctor"
	if len(state.PluginsInstalled) == 0 {
		quickFix = "/plugin install <plugin-id>"
	} else if state.PluginReloadPending {
		quickFix = "/reload-plugins"
	}
	lines := []string{
		"PLUGIN_DIAGNOSTICS",
		fmt.Sprintf("installed=%d", len(state.PluginsInstalled)),
		fmt.Sprintf("enabled=%d", len(state.PluginsEnabled)),
		fmt.Sprintf("marketplaces=%d", len(state.PluginMarketplaces)),
		fmt.Sprintf("pending_reload=%t", state.PluginReloadPending),
		fmt.Sprintf("mutations=%d", state.PluginMutations),
		fmt.Sprintf("reload_count=%d", state.PluginReloadCount),
		fmt.Sprintf("doctor_runs=%d", state.PluginDoctorCount),
		fmt.Sprintf("diagnostic_groups=%d", len(result.Diagnostics)),
		fmt.Sprintf("conflicts=%d", len(conflicts)),
		fmt.Sprintf("quick_fix=%s", normalizeToken(quickFix)),
	}
	return resultWithIntents(strings.Join(lines, "\n"), pluginDiagnosticsIntents(state, len(result.Diagnostics), len(conflicts), quickFix)...)
}

func repairPluginState(state *RuntimeState, mode string) Result {
	if state == nil {
		state = &RuntimeState{}
	}
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		mode = "auto"
	}
	changed := false
	quickFix := "/plugin diagnostics"
	switch mode {
	case "auto", "reload":
		if state.PluginReloadPending {
			state.PluginReloadPending = false
			state.PluginReloadCount++
			changed = true
		}
		quickFix = "/plugin status"
	case "enable-all":
		before := len(state.PluginsEnabled)
		state.PluginsEnabled = uniqueSortedStrings(append([]string(nil), state.PluginsInstalled...))
		changed = len(state.PluginsEnabled) != before
		state.PluginReloadPending = true
		quickFix = "/reload-plugins"
	case "sync-marketplaces":
		if len(state.PluginMarketplaces) == 0 {
			state.PluginMarketplaces = []string{"https://plugins.alliecode.dev"}
			changed = true
		}
		quickFix = "/plugin marketplace"
	default:
		quickFix = "/plugin repair auto"
	}
	if changed {
		state.PluginMutations++
	}
	message := fmt.Sprintf("PLUGIN_REPAIR\nmode=%s\nchanged=%t\npending_reload=%t\nquick_fix=%s\nmutations=%d", normalizeToken(mode), changed, state.PluginReloadPending, normalizeToken(quickFix), state.PluginMutations)
	return resultWithIntents(message, pluginRepairIntents(mode, changed, state.PluginReloadPending, quickFix, state.PluginMutations)...)
}

type skillsSnapshot struct {
	Names         []string
	Sources       map[string]string
	Origins       map[string]string
	Enabled       map[string]bool
	Conflicts     []skillspkg.SkillConflictDiagnostic
	ManagerDirs   []string
	PluginOrigins map[string]string
	Metadata      map[string]skillspkg.SkillMetadata
	PluginOverlay map[string][]string
}

func emptySkillsSnapshot(current []string) skillsSnapshot {
	names := uniqueSortedStrings(append([]string(nil), current...))
	sources := make(map[string]string, len(names))
	origins := make(map[string]string, len(names))
	enabled := make(map[string]bool, len(names))
	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		sources[trimmed] = "state"
		origins[trimmed] = "state"
		enabled[trimmed] = true
	}
	return skillsSnapshot{Names: names, Sources: sources, Origins: origins, Enabled: enabled}
}

func cloneSkillStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		k := strings.TrimSpace(key)
		if k == "" {
			continue
		}
		out[k] = strings.TrimSpace(value)
	}
	return out
}

func cloneSkillBoolMap(input map[string]bool) map[string]bool {
	if len(input) == 0 {
		return map[string]bool{}
	}
	out := make(map[string]bool, len(input))
	for key, value := range input {
		k := strings.TrimSpace(key)
		if k == "" {
			continue
		}
		out[k] = value
	}
	return out
}

func syncedSkillsFromRuntime(current []string) (skillsSnapshot, string) {
	snapshot := emptySkillsSnapshot(current)
	workingDir, _ := os.Getwd()
	manager := skillspkg.NewManager(skillspkg.DefaultSkillDirs(workingDir))
	snapshot.ManagerDirs = append([]string(nil), manager.SkillDirs()...)
	if err := manager.Load(); err != nil {
		return snapshot, "state"
	}
	pluginSvc, _, err := pluginServiceFromWorkingDir()
	if err != nil {
		names := append([]string(nil), snapshot.Names...)
		for _, skill := range manager.List() {
			if skill == nil {
				continue
			}
			name := strings.TrimSpace(skill.Name)
			if name == "" {
				continue
			}
			names = append(names, name)
			snapshot.Sources[name] = strings.TrimSpace(skill.Source)
			snapshot.Origins[name] = strings.TrimSpace(skill.FilePath)
			snapshot.Enabled[name] = skill.Enabled
		}
		snapshot.Names = uniqueSortedStrings(names)
		return snapshot, "skills-files"
	}
	pluginCommands, err := resolvedPluginCommandsForPluginCommand(pluginSvc)
	if err != nil {
		names := append([]string(nil), snapshot.Names...)
		for _, skill := range manager.List() {
			if skill == nil {
				continue
			}
			name := strings.TrimSpace(skill.Name)
			if name == "" {
				continue
			}
			names = append(names, name)
			snapshot.Sources[name] = strings.TrimSpace(skill.Source)
			snapshot.Origins[name] = strings.TrimSpace(skill.FilePath)
			snapshot.Enabled[name] = skill.Enabled
		}
		snapshot.Names = uniqueSortedStrings(names)
		return snapshot, "skills-files"
	}
	detailed := skillspkg.MergeSkillsWithPluginCommandsDetailed(manager.List(), pluginCommands, skillspkg.PreferPluginCommands)
	names := append([]string(nil), snapshot.Names...)
	for _, skill := range detailed.Skills {
		if skill == nil {
			continue
		}
		name := strings.TrimSpace(skill.Name)
		if name == "" {
			continue
		}
		names = append(names, name)
		snapshot.Sources[name] = strings.TrimSpace(skill.Source)
		snapshot.Origins[name] = strings.TrimSpace(skill.FilePath)
		snapshot.Enabled[name] = skill.Enabled
	}
	snapshot.Names = uniqueSortedStrings(names)
	snapshot.Conflicts = append([]skillspkg.SkillConflictDiagnostic(nil), detailed.Conflicts...)
	snapshot.PluginOrigins = cloneSkillStringMap(detailed.PluginOrigins)
	snapshot.Metadata = make(map[string]skillspkg.SkillMetadata, len(detailed.Metadata))
	for key, meta := range detailed.Metadata {
		k := strings.TrimSpace(key)
		if k == "" {
			continue
		}
		snapshot.Metadata[k] = meta
	}
	snapshot.PluginOverlay = make(map[string][]string, len(detailed.PluginOverlay))
	for key, values := range detailed.PluginOverlay {
		k := strings.TrimSpace(key)
		if k == "" {
			continue
		}
		snapshot.PluginOverlay[k] = append([]string(nil), values...)
	}
	return snapshot, "skills+plugins"
}
