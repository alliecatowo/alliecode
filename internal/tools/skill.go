package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/alliecatowo/alliecode/internal/plugins"
	skillspkg "github.com/alliecatowo/alliecode/internal/skills"
	"github.com/alliecatowo/alliecode/internal/types"
)

type SkillTool struct{}

type skillInput struct {
	Skill  string `json:"skill"`
	Args   string `json:"args,omitempty"`
	Action string `json:"action,omitempty"`
}

type skillOutput struct {
	Success             bool                                `json:"success"`
	Action              string                              `json:"action"`
	Skill               string                              `json:"skill,omitempty"`
	Skills              []string                            `json:"skills,omitempty"`
	PluginOrigins       map[string]string                   `json:"plugin_origins"`
	ConflictDiagnostics []skillspkg.SkillConflictDiagnostic `json:"conflict_diagnostics"`
	Usage               []skillUsageRecord                  `json:"usage,omitempty"`
	Result              string                              `json:"result,omitempty"`
	Error               string                              `json:"error,omitempty"`
}

func (t *SkillTool) Name() string { return "skill" }

func (t *SkillTool) Description() string {
	return "Runs local skills and exposes a deterministic skill catalog."
}

func (t *SkillTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"skill":  {Type: "string", Description: "Skill name for invoke action."},
			"args":   {Type: "string", Description: "Optional arguments passed to the skill."},
			"action": {Type: "string", Description: "list, invoke, or history", Enum: []string{"list", "invoke", "history"}},
		},
	}
}

func (t *SkillTool) Execute(_ context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	var in skillInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("invalid input: %v", err), IsError: true}, nil
	}

	action := strings.TrimSpace(strings.ToLower(in.Action))
	if action == "" {
		if strings.TrimSpace(in.Skill) == "" {
			action = "list"
		} else {
			action = "invoke"
		}
	}

	switch action {
	case "list":
		skills, pluginOrigins, conflicts := listSkills(toolCtx.WorkingDir)
		out := skillOutput{Success: true, Action: "list", Skills: skills, PluginOrigins: pluginOrigins, ConflictDiagnostics: conflicts}
		return skillJSONResult(out, false)
	case "history":
		orchestrationState.mu.Lock()
		usage := make([]skillUsageRecord, 0, len(orchestrationState.skillUsage))
		for _, rec := range orchestrationState.skillUsage {
			usage = append(usage, rec)
		}
		orchestrationState.mu.Unlock()
		sort.Slice(usage, func(i, j int) bool { return usage[i].Name < usage[j].Name })
		out := skillOutput{Success: true, Action: "history", Usage: usage}
		return skillJSONResult(out, false)
	case "invoke":
		skillName := strings.TrimSpace(in.Skill)
		if skillName == "" {
			return skillJSONResult(skillOutput{Success: false, Action: "invoke", Error: "skill is required for invoke"}, true)
		}
		content, source, ok := resolveSkillContent(toolCtx.WorkingDir, skillName)
		if !ok {
			return skillJSONResult(skillOutput{Success: false, Action: "invoke", Skill: skillName, Error: "skill not found"}, true)
		}
		orchestrationState.mu.Lock()
		rec := orchestrationState.skillUsage[skillName]
		rec.Name = skillName
		rec.Invocations++
		rec.LastArgs = strings.TrimSpace(in.Args)
		rec.LastUsedAt = time.Now().UTC().Format(time.RFC3339)
		orchestrationState.skillUsage[skillName] = rec
		orchestrationState.mu.Unlock()

		result := fmt.Sprintf("Loaded skill %q from %s", skillName, source)
		if strings.TrimSpace(in.Args) != "" {
			result += fmt.Sprintf(" with args: %s", strings.TrimSpace(in.Args))
		}
		if strings.TrimSpace(content) != "" {
			result += "\n\n" + truncateAtBoundary(content, 1200)
		}
		out := skillOutput{Success: true, Action: "invoke", Skill: skillName, Result: result}
		return skillJSONResult(out, false)
	default:
		return skillJSONResult(skillOutput{Success: false, Action: action, Error: "action must be one of: list, invoke, history"}, true)
	}
}

func listSkills(workingDir string) ([]string, map[string]string, []skillspkg.SkillConflictDiagnostic) {
	candidates := []string{}
	sourcesByName := map[string]string{}
	for _, root := range []string{filepath.Join(workingDir, ".claude", "skills"), filepath.Join(userHomeDir(), ".claude", "skills")} {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			skillPath := filepath.Join(root, entry.Name(), "SKILL.md")
			if _, err := os.Stat(skillPath); err == nil {
				name := entry.Name()
				candidates = append(candidates, name)
				if _, exists := sourcesByName[name]; !exists {
					sourcesByName[name] = skillPath
				}
			}
		}
	}

	pluginOrigins := map[string]string{}
	conflicts := make([]skillspkg.SkillConflictDiagnostic, 0)
	seen := make(map[string]struct{}, len(candidates))
	for _, name := range candidates {
		seen[name] = struct{}{}
	}

	for _, cmd := range resolvedPluginCommandsForSkills(workingDir) {
		name := strings.TrimSpace(cmd.Name)
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			conflicts = append(conflicts, skillspkg.SkillConflictDiagnostic{
				Name:           name,
				ExistingSource: sourcesByName[name],
				IncomingSource: fmt.Sprintf("plugin:%s", cmd.PluginID),
				WinnerSource:   sourcesByName[name],
			})
			continue
		}
		seen[name] = struct{}{}
		candidates = append(candidates, name)
		pluginOrigins[name] = cmd.PluginID
	}

	sort.Strings(candidates)
	sort.Slice(conflicts, func(i, j int) bool {
		if strings.ToLower(conflicts[i].Name) != strings.ToLower(conflicts[j].Name) {
			return strings.ToLower(conflicts[i].Name) < strings.ToLower(conflicts[j].Name)
		}
		return conflicts[i].IncomingSource < conflicts[j].IncomingSource
	})
	return uniqueStrings(candidates), pluginOrigins, conflicts
}

func resolvedPluginCommandsForSkills(workingDir string) []plugins.ResolvedCommand {
	installRoot := filepath.Join(workingDir, ".alliecode", "plugins")
	state, err := plugins.NewPolicyStore(filepath.Join(installRoot, "policy.yaml")).Load()
	if err != nil {
		state = plugins.PolicyState{}
	}

	base, err := plugins.LoadRuntime(installRoot, plugins.Policy{})
	if err != nil {
		return nil
	}

	ids := make([]string, 0, len(base.Plugins()))
	for _, plugin := range base.Plugins() {
		ids = append(ids, plugin.Manifest.ID)
	}

	runtime, err := plugins.LoadRuntime(installRoot, state.RuntimePolicy(ids))
	if err != nil {
		return nil
	}

	commands := runtime.Commands()
	out := make([]plugins.ResolvedCommand, len(commands))
	copy(out, commands)
	return out
}

func resolveSkillContent(workingDir, skillName string) (string, string, bool) {
	paths := []string{
		filepath.Join(workingDir, ".claude", "skills", skillName, "SKILL.md"),
		filepath.Join(userHomeDir(), ".claude", "skills", skillName, "SKILL.md"),
	}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err == nil {
			return string(data), p, true
		}
	}
	return "", "", false
}

func userHomeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return h
}

func uniqueStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, 0, len(in))
	seen := map[string]struct{}{}
	for _, item := range in {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func truncateAtBoundary(s string, max int) string {
	if len(s) <= max {
		return s
	}
	trimmed := strings.TrimSpace(s[:max])
	return trimmed + "..."
}

func skillJSONResult(out skillOutput, isError bool) (types.ToolResult, error) {
	b, err := json.Marshal(out)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}
	return types.ToolResult{Content: string(b), IsError: isError}, nil
}

func (t *SkillTool) IsReadOnly(input types.ToolInput) bool {
	var in skillInput
	if err := json.Unmarshal(input, &in); err != nil {
		return false
	}
	action := strings.ToLower(strings.TrimSpace(in.Action))
	if action == "" {
		return strings.TrimSpace(in.Skill) == ""
	}
	return action == "list" || action == "history"
}

func (t *SkillTool) IsDestructive(input types.ToolInput) bool { return false }

func (t *SkillTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *SkillTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return types.PermissionAllowed
}
