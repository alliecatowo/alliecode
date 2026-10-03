package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alliecatowo/alliecode/internal/mcp"
	"github.com/alliecatowo/alliecode/internal/permissions"
	pluginspkg "github.com/alliecatowo/alliecode/internal/plugins"
	"github.com/alliecatowo/alliecode/internal/providers"
	"github.com/alliecatowo/alliecode/internal/remote"
	skillspkg "github.com/alliecatowo/alliecode/internal/skills"
	"github.com/alliecatowo/alliecode/internal/types"
)

// HelpCommand prints available slash commands.
type HelpCommand struct {
	registry *Registry
}

func NewHelpCommand(registry *Registry) *HelpCommand {
	return &HelpCommand{registry: registry}
}

func (c *HelpCommand) Name() string        { return "help" }
func (c *HelpCommand) Aliases() []string   { return []string{"?"} }
func (c *HelpCommand) Description() string { return "Show available slash commands" }
func (c *HelpCommand) Usage() string       { return "/help [query]" }

func (c *HelpCommand) Execute(_ context.Context, _ Context, inv Invocation) (Result, error) {
	query := strings.TrimSpace(strings.Join(inv.Args, " "))
	query = strings.TrimPrefix(query, "/")
	query = strings.ToLower(query)
	suggestions := c.registry.Suggestions(query)
	if len(suggestions) == 0 {
		if strings.TrimSpace(query) == "" {
			return resultWithIntents("HELP\ncount=0", helpIntents(query, nil)...), nil
		}
		return resultWithIntents(fmt.Sprintf("HELP\nquery=%s\ncount=0", normalizeToken(query)), helpIntents(query, nil)...), nil
	}

	var b strings.Builder
	b.WriteString("HELP\n")
	b.WriteString(fmt.Sprintf("count=%d\n", len(suggestions)))
	if strings.TrimSpace(query) != "" {
		b.WriteString(fmt.Sprintf("query=%s\n", normalizeToken(query)))
	}
	for i, suggestion := range suggestions {
		idx := i + 1
		b.WriteString(fmt.Sprintf("entry.%d.name=/%s\n", idx, normalizeToken(suggestion.Name)))
		b.WriteString(fmt.Sprintf("entry.%d.description=%s\n", idx, normalizeToken(suggestion.Description)))
		b.WriteString(fmt.Sprintf("entry.%d.category=%s\n", idx, normalizeToken(suggestion.Category)))
		b.WriteString(fmt.Sprintf("entry.%d.group=%s\n", idx, normalizeToken(suggestion.Group)))
		b.WriteString(fmt.Sprintf("entry.%d.palette_group=%s\n", idx, normalizeToken(suggestion.PaletteGroup)))
		b.WriteString(fmt.Sprintf("entry.%d.usage=%s\n", idx, normalizeToken(suggestion.Usage)))
		b.WriteString(fmt.Sprintf("entry.%d.argument_hint=%s\n", idx, normalizeToken(suggestion.ArgumentHint)))
		b.WriteString(fmt.Sprintf("entry.%d.help_hint=%s\n", idx, normalizeToken(suggestion.HelpHint)))
		b.WriteString(fmt.Sprintf("entry.%d.alias_count=%d\n", idx, len(suggestion.Aliases)))
		for aliasIdx, alias := range suggestion.Aliases {
			b.WriteString(fmt.Sprintf("entry.%d.alias.%d=/%s\n", idx, aliasIdx+1, normalizeToken(alias)))
		}
		b.WriteString(fmt.Sprintf("entry.%d.keyword_count=%d\n", idx, len(suggestion.Keywords)))
		for keywordIdx, keyword := range suggestion.Keywords {
			b.WriteString(fmt.Sprintf("entry.%d.keyword.%d=%s\n", idx, keywordIdx+1, normalizeToken(keyword)))
		}
		b.WriteString(fmt.Sprintf("entry.%d.example_count=%d\n", idx, len(suggestion.Examples)))
		for exampleIdx, example := range suggestion.Examples {
			b.WriteString(fmt.Sprintf("entry.%d.example.%d=%s\n", idx, exampleIdx+1, normalizeToken(example)))
		}
		b.WriteString(fmt.Sprintf("entry.%d.context_count=%d\n", idx, len(suggestion.Contexts)))
		for contextIdx, contextTag := range suggestion.Contexts {
			b.WriteString(fmt.Sprintf("entry.%d.context.%d=%s\n", idx, contextIdx+1, normalizeToken(contextTag)))
		}
		b.WriteString(fmt.Sprintf("entry.%d.diagnostic_count=%d\n", idx, len(suggestion.Diagnostics)))
		for diagIdx, diagnostic := range suggestion.Diagnostics {
			b.WriteString(fmt.Sprintf("entry.%d.diagnostic.%d=%s\n", idx, diagIdx+1, normalizeToken(diagnostic)))
		}
		b.WriteString(fmt.Sprintf("entry.%d.shortcut_count=%d\n", idx, len(suggestion.Shortcuts)))
		for shortcutIdx, shortcut := range suggestion.Shortcuts {
			b.WriteString(fmt.Sprintf("entry.%d.shortcut.%d=%s\n", idx, shortcutIdx+1, normalizeToken(shortcut)))
		}
		b.WriteString(fmt.Sprintf("entry.%d.match_reason=%s\n", idx, normalizeToken(suggestion.MatchReason)))
	}
	return resultWithIntents(strings.TrimSpace(b.String()), helpIntents(query, suggestions)...), nil
}

// ModelCommand gets or sets the active model.
type ModelCommand struct{}

func NewModelCommand() *ModelCommand      { return &ModelCommand{} }
func (c *ModelCommand) Name() string      { return "model" }
func (c *ModelCommand) Aliases() []string { return []string{"m"} }
func (c *ModelCommand) Description() string {
	return "Get or set active model for this session"
}
func (c *ModelCommand) Usage() string {
	return "/model [provider/model|model|list [provider|all]|doctor|repair [provider/model]]"
}

func (c *ModelCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	return executeModelCommand(cmdCtx, inv)
}

// ProviderCommand gets or sets the active provider for this session.
type ProviderCommand struct{}

func NewProviderCommand() *ProviderCommand     { return &ProviderCommand{} }
func (c *ProviderCommand) Name() string        { return "provider" }
func (c *ProviderCommand) Aliases() []string   { return nil }
func (c *ProviderCommand) Description() string { return "Get or set active provider for this session" }
func (c *ProviderCommand) Usage() string       { return "/provider [status|list|set <name>|<name>]" }

func (c *ProviderCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	return executeProviderCommand(cmdCtx, inv)
}

// CompactCommand requests chat compaction.
type CompactCommand struct{}

func NewCompactCommand() *CompactCommand    { return &CompactCommand{} }
func (c *CompactCommand) Name() string      { return "compact" }
func (c *CompactCommand) Aliases() []string { return nil }
func (c *CompactCommand) Description() string {
	return "Compact conversation context on next model turn"
}
func (c *CompactCommand) Usage() string { return "/compact [now|auto|off|status]" }

func (c *CompactCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) > 1 {
		return Result{}, fmt.Errorf("usage: /compact [now|auto|off|status]")
	}
	if cmdCtx.State.CompactMode == "" {
		cmdCtx.State.CompactMode = "auto"
	}

	mode := "now"
	if len(inv.Args) == 1 {
		mode = strings.ToLower(strings.TrimSpace(inv.Args[0]))
	}

	switch mode {
	case "now":
		cmdCtx.State.CompactRequested = true
		cmdCtx.State.CompactCount++
		cmdCtx.State.LastCompactTarget = "now"
		message := fmt.Sprintf("COMPACT_REQUEST\nmode=%s\nrequested=true\ncount=%d\nlast_target=now", cmdCtx.State.CompactMode, cmdCtx.State.CompactCount)
		return resultWithIntents(message, compactIntents(cmdCtx.State.CompactMode, true, cmdCtx.State.CompactCount, "now")...), nil
	case "auto":
		cmdCtx.State.CompactMode = "auto"
		cmdCtx.State.CompactRequested = false
		message := fmt.Sprintf("COMPACT_MODE\nmode=auto\nrequested=%t\ncount=%d", cmdCtx.State.CompactRequested, cmdCtx.State.CompactCount)
		return resultWithIntents(message, compactIntents(cmdCtx.State.CompactMode, cmdCtx.State.CompactRequested, cmdCtx.State.CompactCount, cmdCtx.State.LastCompactTarget)...), nil
	case "off":
		cmdCtx.State.CompactMode = "off"
		cmdCtx.State.CompactRequested = false
		message := fmt.Sprintf("COMPACT_MODE\nmode=off\nrequested=%t\ncount=%d", cmdCtx.State.CompactRequested, cmdCtx.State.CompactCount)
		return resultWithIntents(message, compactIntents(cmdCtx.State.CompactMode, cmdCtx.State.CompactRequested, cmdCtx.State.CompactCount, cmdCtx.State.LastCompactTarget)...), nil
	case "status":
		lastTarget := normalizeToken(cmdCtx.State.LastCompactTarget)
		message := fmt.Sprintf("COMPACT_STATUS\nmode=%s\nrequested=%t\ncount=%d\nlast_target=%s", cmdCtx.State.CompactMode, cmdCtx.State.CompactRequested, cmdCtx.State.CompactCount, lastTarget)
		return resultWithIntents(message, compactIntents(cmdCtx.State.CompactMode, cmdCtx.State.CompactRequested, cmdCtx.State.CompactCount, lastTarget)...), nil
	default:
		return Result{}, fmt.Errorf("usage: /compact [now|auto|off|status]")
	}
}

// PermissionsCommand gets or sets permission mode.
type PermissionsCommand struct{}

func NewPermissionsCommand() *PermissionsCommand { return &PermissionsCommand{} }
func (c *PermissionsCommand) Name() string       { return "permissions" }
func (c *PermissionsCommand) Aliases() []string  { return []string{"permission", "allowed-tools"} }
func (c *PermissionsCommand) Description() string {
	return "Get or set permission mode, inspect rules and denials"
}
func (c *PermissionsCommand) Usage() string {
	return permissionsUsage
}

const permissionsUsage = "usage: /permissions [status|get|list|summary|rules [list|add <rule>|remove <rule>]|denials|retry-denials|set <plan|default|auto|bypass>|plan|default|auto|bypass]"

func (c *PermissionsCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}

	if len(inv.Args) == 0 {
		return renderPermissionsSummary(cmdCtx.State), nil
	}

	if len(inv.Args) == 1 && (strings.EqualFold(strings.TrimSpace(inv.Args[0]), "get") || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "mode")) {
		message := fmt.Sprintf("PERMISSIONS_MODE\nmode=%s", modeString(cmdCtx.State.PermissionMode))
		return resultWithIntents(message, permissionsModeIntents(modeString(cmdCtx.State.PermissionMode))...), nil
	}

	if len(inv.Args) == 1 && (strings.EqualFold(strings.TrimSpace(inv.Args[0]), "summary") || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status")) {
		return renderPermissionsSummary(cmdCtx.State), nil
	}

	if len(inv.Args) == 1 && (strings.EqualFold(strings.TrimSpace(inv.Args[0]), "doctor") || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "loops")) {
		return renderPermissionsLoops(cmdCtx.State), nil
	}

	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "list") {
		message := "PERMISSIONS_MODES\ncount=4\nmode.1=plan\nmode.2=default\nmode.3=auto\nmode.4=bypass"
		return resultWithIntents(message, permissionsModesIntents([]string{"plan", "default", "auto", "bypass"})...), nil
	}

	if len(inv.Args) >= 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "rules") {
		res, ok, err := c.handleRulesSubcommand(cmdCtx.State, inv.Args[1:])
		if ok || err != nil {
			return res, err
		}
	}

	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "rules") {
		return renderPermissionRules(cmdCtx.State.PermissionRules), nil
	}

	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "denials") {
		denials := append([]PermissionDenial(nil), cmdCtx.State.PermissionDenials...)
		sort.Slice(denials, func(i, j int) bool {
			if denials[i].ID != denials[j].ID {
				return denials[i].ID < denials[j].ID
			}
			if denials[i].Command != denials[j].Command {
				return denials[i].Command < denials[j].Command
			}
			return denials[i].Reason < denials[j].Reason
		})
		return renderPermissionDenials(denials), nil
	}

	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "retry-denials") {
		commands := permissionRetryCommands(cmdCtx.State.PermissionDenials)
		lines := []string{"PERMISSIONS_RETRY", fmt.Sprintf("count=%d", len(commands))}
		for i, command := range commands {
			lines = append(lines, fmt.Sprintf("command.%d=%s", i+1, normalizeToken(command)))
		}
		message := strings.Join(lines, "\n")
		return resultWithIntents(message, permissionRetryIntents(commands)...), nil
	}

	args := inv.Args
	if strings.EqualFold(strings.TrimSpace(args[0]), "set") {
		args = args[1:]
	}
	if len(args) != 1 {
		return Result{}, fmt.Errorf(permissionsUsage)
	}

	mode, ok := parseMode(args[0])
	if !ok {
		return Result{}, fmt.Errorf("unknown permission mode %q (expected: plan|default|auto|bypass)", args[0])
	}
	cmdCtx.State.PermissionMode = mode
	message := fmt.Sprintf("PERMISSIONS_SET\nmode=%s", modeString(mode))
	return resultWithIntents(message, permissionsModeIntents(modeString(mode))...), nil
}

func (c *PermissionsCommand) handleRulesSubcommand(state *RuntimeState, args []string) (Result, bool, error) {
	if len(args) == 0 || strings.EqualFold(strings.TrimSpace(args[0]), "list") {
		if len(args) > 1 {
			return Result{}, true, fmt.Errorf(permissionsUsage)
		}
		return renderPermissionRules(state.PermissionRules), true, nil
	}

	sub := strings.ToLower(strings.TrimSpace(args[0]))
	rule := strings.TrimSpace(strings.Join(args[1:], " "))
	switch sub {
	case "add":
		if rule == "" {
			return Result{}, true, fmt.Errorf(permissionsUsage)
		}
		state.PermissionRules = uniqueSortedStrings(append(state.PermissionRules, rule))
		message := fmt.Sprintf("PERMISSIONS_RULES_ADD\nrule=%s\ncount=%d", normalizeToken(rule), len(state.PermissionRules))
		return resultWithIntents(message, permissionRulesIntents(state.PermissionRules)...), true, nil
	case "remove":
		if rule == "" {
			return Result{}, true, fmt.Errorf(permissionsUsage)
		}
		removed := false
		next := make([]string, 0, len(state.PermissionRules))
		for _, existing := range state.PermissionRules {
			if strings.TrimSpace(existing) == rule {
				removed = true
				continue
			}
			next = append(next, existing)
		}
		state.PermissionRules = uniqueSortedStrings(next)
		message := fmt.Sprintf("PERMISSIONS_RULES_REMOVE\nrule=%s\nremoved=%t\ncount=%d", normalizeToken(rule), removed, len(state.PermissionRules))
		return resultWithIntents(message, permissionRulesIntents(state.PermissionRules)...), true, nil
	default:
		return Result{}, true, fmt.Errorf(permissionsUsage)
	}
}

func renderPermissionRules(rules []string) Result {
	views := permissionRulesFromStrings(rules)
	if len(views) == 0 {
		return resultWithIntents("PERMISSIONS_RULES\ncount=0", permissionRulesIntents(nil)...)
	}
	lines := []string{"PERMISSIONS_RULES", fmt.Sprintf("count=%d", len(views))}
	for i, rule := range views {
		idx := i + 1
		lines = append(lines, fmt.Sprintf("rule.%d.source=%s", idx, normalizeToken(rule.Source)))
		lines = append(lines, fmt.Sprintf("rule.%d.value=%s", idx, normalizeToken(rule.Value)))
	}
	return resultWithIntents(strings.Join(lines, "\n"), permissionRulesIntents(rules)...)
}

type permissionRuleView struct {
	Source string
	Value  string
}

func renderPermissionsSummary(state *RuntimeState) Result {
	rules := permissionRulesFromStrings(state.PermissionRules)
	sourceCounts := map[string]int{"policy": 0, "user": 0, "project": 0, "session": 0}
	for _, rule := range rules {
		sourceCounts[rule.Source]++
	}
	denialGroups := groupPermissionDenials(state.PermissionDenials)
	mode := modeString(state.PermissionMode)
	lines := []string{
		"PERMISSIONS_SUMMARY",
		fmt.Sprintf("mode=%s", mode),
		fmt.Sprintf("mode.aliases=%s", permissionModeAliases(mode)),
		fmt.Sprintf("rules.count=%d", len(rules)),
		fmt.Sprintf("rules.precedence=%s", permissions.RuleSourcePrecedence),
		fmt.Sprintf("rules.policy=%d", sourceCounts["policy"]),
		fmt.Sprintf("rules.user=%d", sourceCounts["user"]),
		fmt.Sprintf("rules.project=%d", sourceCounts["project"]),
		fmt.Sprintf("rules.session=%d", sourceCounts["session"]),
		fmt.Sprintf("denials.count=%d", len(state.PermissionDenials)),
		fmt.Sprintf("denials.groups=%d", len(denialGroups)),
	}
	return resultWithIntents(strings.Join(lines, "\n"), permissionsSummaryIntents(state, denialGroupReasons(denialGroups))...)
}

func renderPermissionsLoops(state *RuntimeState) Result {
	if state == nil {
		state = &RuntimeState{}
	}
	loops := correctiveLoopsForState(state)
	lines := []string{"PERMISSIONS_LOOPS", "count=0"}
	idx := 0
	for _, loop := range loops {
		if loop.Area != "permissions" && loop.Area != "settings" {
			continue
		}
		idx++
		lines = append(lines, fmt.Sprintf("loop.%d.area=%s", idx, normalizeToken(loop.Area)))
		lines = append(lines, fmt.Sprintf("loop.%d.state=%s", idx, normalizeToken(loop.State)))
		lines = append(lines, fmt.Sprintf("loop.%d.action=%s", idx, normalizeToken(loop.Action)))
		lines = append(lines, fmt.Sprintf("loop.%d.next=%s", idx, normalizeToken(loop.Next)))
	}
	lines[1] = fmt.Sprintf("count=%d", idx)
	return resultWithIntents(strings.Join(lines, "\n"), correctiveLoopIntents("Permission loops", "Permission and settings corrective loops.", loops)...)
}

func permissionModeAliases(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "auto":
		return "accept_edits,accept-edits,acceptedits"
	case "bypass":
		return "bypass_permissions,bypasspermissions"
	default:
		return "-"
	}
}

type denialGroup struct {
	Reason  string
	Denials []PermissionDenial
}

func renderPermissionDenials(denials []PermissionDenial) Result {
	if len(denials) == 0 {
		return resultWithIntents("PERMISSIONS_DENIALS\ncount=0\ngroup_count=0", permissionDenialsIntents(nil, nil)...)
	}
	groups := groupPermissionDenials(denials)
	lines := []string{"PERMISSIONS_DENIALS", fmt.Sprintf("count=%d", len(denials)), fmt.Sprintf("group_count=%d", len(groups))}
	flatIndex := 1
	for i, group := range groups {
		gidx := i + 1
		lines = append(lines, fmt.Sprintf("group.%d.reason=%s", gidx, normalizeToken(group.Reason)))
		lines = append(lines, fmt.Sprintf("group.%d.count=%d", gidx, len(group.Denials)))
		for _, denial := range group.Denials {
			lines = append(lines, fmt.Sprintf("denial.%d.id=%s", flatIndex, normalizeToken(denial.ID)))
			lines = append(lines, fmt.Sprintf("denial.%d.command=%s", flatIndex, normalizeToken(denial.Command)))
			lines = append(lines, fmt.Sprintf("denial.%d.reason=%s", flatIndex, normalizeToken(denial.Reason)))
			flatIndex++
		}
	}
	return resultWithIntents(strings.Join(lines, "\n"), permissionDenialsIntents(denials, denialGroupReasons(groups))...)
}

func denialGroupReasons(groups []denialGroup) []string {
	out := make([]string, 0, len(groups))
	for _, group := range groups {
		out = append(out, fmt.Sprintf("%s (%d)", group.Reason, len(group.Denials)))
	}
	return out
}

func groupPermissionDenials(denials []PermissionDenial) []denialGroup {
	if len(denials) == 0 {
		return nil
	}
	byReason := make(map[string][]PermissionDenial)
	for _, denial := range denials {
		reason := strings.TrimSpace(denial.Reason)
		if reason == "" {
			reason = "unknown"
		}
		byReason[reason] = append(byReason[reason], denial)
	}
	reasons := make([]string, 0, len(byReason))
	for reason := range byReason {
		reasons = append(reasons, reason)
	}
	sort.Strings(reasons)
	out := make([]denialGroup, 0, len(reasons))
	for _, reason := range reasons {
		entries := append([]PermissionDenial(nil), byReason[reason]...)
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].Command != entries[j].Command {
				return entries[i].Command < entries[j].Command
			}
			return entries[i].ID < entries[j].ID
		})
		out = append(out, denialGroup{Reason: reason, Denials: entries})
	}
	return out
}

func permissionRetryCommands(denials []PermissionDenial) []string {
	if len(denials) == 0 {
		return nil
	}
	commands := make([]string, 0, len(denials))
	for _, denial := range denials {
		command := strings.TrimSpace(denial.Command)
		if command == "" {
			continue
		}
		commands = append(commands, command)
	}
	return uniqueSortedStrings(commands)
}

func permissionRulesFromStrings(rules []string) []permissionRuleView {
	if len(rules) == 0 {
		return nil
	}
	out := make([]permissionRuleView, 0, len(rules))
	for _, raw := range rules {
		source, value := parsePermissionRuleSource(raw)
		if strings.TrimSpace(value) == "" {
			continue
		}
		out = append(out, permissionRuleView{Source: source, Value: value})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Source != out[j].Source {
			return out[i].Source < out[j].Source
		}
		return out[i].Value < out[j].Value
	})
	return out
}

func parsePermissionRuleSource(raw string) (string, string) {
	raw = strings.TrimSpace(raw)
	parts := strings.SplitN(raw, ":", 2)
	if len(parts) != 2 {
		return "session", raw
	}
	source := strings.ToLower(strings.TrimSpace(parts[0]))
	value := strings.TrimSpace(parts[1])
	switch source {
	case "policy", "user", "project", "session":
		return source, value
	default:
		return "session", raw
	}
}

// ResumeCommand requests resuming a previous session.
type ResumeCommand struct{}

func NewResumeCommand() *ResumeCommand       { return &ResumeCommand{} }
func (c *ResumeCommand) Name() string        { return "resume" }
func (c *ResumeCommand) Aliases() []string   { return []string{"continue"} }
func (c *ResumeCommand) Description() string { return "Resume most recent session context" }
func (c *ResumeCommand) Usage() string       { return "/resume [latest|status|<target>]" }

func (c *ResumeCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) > 1 {
		return Result{}, fmt.Errorf("usage: /resume [latest|status|<target>]")
	}
	if len(inv.Args) == 1 {
		arg := strings.TrimSpace(inv.Args[0])
		if strings.EqualFold(arg, "status") {
			message := fmt.Sprintf("RESUME_STATUS\nrequested=%t\ncount=%d\nlast_target=%s", cmdCtx.State.ResumeRequested, cmdCtx.State.ResumeCount, normalizeToken(cmdCtx.State.LastResumeTarget))
			return resultWithIntents(message, resumeStatusIntents(cmdCtx.State.ResumeRequested, cmdCtx.State.ResumeCount, cmdCtx.State.LastResumeTarget)...), nil
		}
		if arg == "" {
			return Result{}, fmt.Errorf("usage: /resume [latest|status|<target>]")
		}
		cmdCtx.State.LastResumeTarget = arg
	} else {
		cmdCtx.State.LastResumeTarget = "latest"
	}
	cmdCtx.State.ResumeRequested = true
	cmdCtx.State.ResumeCount++
	message := fmt.Sprintf("RESUME_REQUEST\ntarget=%s\nrequested=true\ncount=%d\nlast_target=%s", normalizeToken(cmdCtx.State.LastResumeTarget), cmdCtx.State.ResumeCount, normalizeToken(cmdCtx.State.LastResumeTarget))
	return resultWithIntents(message, resumeRequestIntents(cmdCtx.State.LastResumeTarget, cmdCtx.State.ResumeCount)...), nil
}

// BranchCommand reports branch command availability.
type BranchCommand struct{}

func NewBranchCommand() *BranchCommand       { return &BranchCommand{} }
func (c *BranchCommand) Name() string        { return "branch" }
func (c *BranchCommand) Aliases() []string   { return nil }
func (c *BranchCommand) Description() string { return "Show current branch details" }
func (c *BranchCommand) Usage() string       { return "/branch [status|list|create [name]|switch <name>]" }

func (c *BranchCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	ensureBranchState(cmdCtx.State)

	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /branch [status|list|create [name]|switch <name>]")
		}
		message := fmt.Sprintf("BRANCH_STATUS\nactive=%s\ncount=%d\ncreated=%d\nswitches=%d", normalizeToken(cmdCtx.State.ActiveBranch), len(cmdCtx.State.Branches), cmdCtx.State.BranchCount, cmdCtx.State.BranchSwitchCount)
		return resultWithIntents(message, branchStatusIntents(cmdCtx.State.ActiveBranch, len(cmdCtx.State.Branches), cmdCtx.State.BranchCount, cmdCtx.State.BranchSwitchCount)...), nil
	}

	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "list":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /branch [status|list|create [name]|switch <name>]")
		}
		branches := append([]string(nil), cmdCtx.State.Branches...)
		sort.Strings(branches)
		lines := []string{"BRANCH_LIST", fmt.Sprintf("count=%d", len(branches)), fmt.Sprintf("active=%s", normalizeToken(cmdCtx.State.ActiveBranch))}
		for i, name := range branches {
			lines = append(lines, fmt.Sprintf("branch.%d=%s", i+1, normalizeToken(name)))
		}
		message := strings.Join(lines, "\n")
		return resultWithIntents(message, branchListIntents(cmdCtx.State.ActiveBranch, branches)...), nil
	case "create":
		if len(inv.Args) > 2 {
			return Result{}, fmt.Errorf("usage: /branch [status|list|create [name]|switch <name>]")
		}
		name := ""
		if len(inv.Args) == 2 {
			name = strings.TrimSpace(inv.Args[1])
		}
		if name == "" {
			name = fmt.Sprintf("branch-%d", cmdCtx.State.BranchCount+1)
		}
		cmdCtx.State.Branches = uniqueSortedStrings(append(cmdCtx.State.Branches, name))
		if strings.TrimSpace(cmdCtx.State.ActiveBranch) != strings.TrimSpace(name) {
			cmdCtx.State.BranchSwitchCount++
		}
		cmdCtx.State.ActiveBranch = name
		cmdCtx.State.BranchCount++
		message := fmt.Sprintf("BRANCH_CREATE\nname=%s\nactive=%s\ncount=%d\ncreated=%d\nswitched=true\nswitches=%d", normalizeToken(name), normalizeToken(cmdCtx.State.ActiveBranch), len(cmdCtx.State.Branches), cmdCtx.State.BranchCount, cmdCtx.State.BranchSwitchCount)
		return resultWithIntents(message, branchMutationIntents("Branch created", name, "-", true, cmdCtx.State.BranchSwitchCount)...), nil
	case "switch":
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf("usage: /branch [status|list|create [name]|switch <name>]")
		}
		name := strings.TrimSpace(inv.Args[1])
		if name == "" {
			return Result{}, fmt.Errorf("usage: /branch [status|list|create [name]|switch <name>]")
		}
		if !containsString(cmdCtx.State.Branches, name) {
			return Result{}, fmt.Errorf("branch not found: %s", name)
		}
		previous := cmdCtx.State.ActiveBranch
		changed := strings.TrimSpace(previous) != strings.TrimSpace(name)
		cmdCtx.State.ActiveBranch = name
		if changed {
			cmdCtx.State.BranchSwitchCount++
		}
		message := fmt.Sprintf("BRANCH_SWITCH\nactive=%s\nprevious=%s\nchanged=%t\nswitches=%d", normalizeToken(name), normalizeToken(previous), changed, cmdCtx.State.BranchSwitchCount)
		return resultWithIntents(message, branchMutationIntents("Branch switched", name, previous, changed, cmdCtx.State.BranchSwitchCount)...), nil
	default:
		return Result{}, fmt.Errorf("usage: /branch [status|list|create [name]|switch <name>]")
	}
}

// DiffCommand reports diff command availability.
type DiffCommand struct{}

func NewDiffCommand() *DiffCommand         { return &DiffCommand{} }
func (c *DiffCommand) Name() string        { return "diff" }
func (c *DiffCommand) Aliases() []string   { return nil }
func (c *DiffCommand) Description() string { return "Show current working tree diff" }
func (c *DiffCommand) Usage() string {
	return "/diff [status|list|mode <working|staged|all>|add <path> <added> <removed> [modified]|clear]"
}

func (c *DiffCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if cmdCtx.State.DiffMode == "" {
		cmdCtx.State.DiffMode = "working"
	}

	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "list") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return renderDiffList(cmdCtx.State), nil
	}

	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "status":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		message := fmt.Sprintf("DIFF_STATUS\nmode=%s\nentries=%d\nupdates=%d\nclears=%d", normalizeToken(cmdCtx.State.DiffMode), len(cmdCtx.State.DiffEntries), cmdCtx.State.DiffCount, cmdCtx.State.DiffClearCount)
		return resultWithIntents(message, diffStatusIntents(cmdCtx.State.DiffMode, len(cmdCtx.State.DiffEntries), cmdCtx.State.DiffCount, cmdCtx.State.DiffClearCount)...), nil
	case "mode":
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		mode := strings.ToLower(strings.TrimSpace(inv.Args[1]))
		switch mode {
		case "working", "staged", "all":
			cmdCtx.State.DiffMode = mode
			message := fmt.Sprintf("DIFF_MODE\nmode=%s", mode)
			return resultWithIntents(message, diffModeIntents(mode)...), nil
		default:
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
	case "add":
		if len(inv.Args) < 4 || len(inv.Args) > 5 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		path := strings.TrimSpace(inv.Args[1])
		if path == "" {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		added, err := parseNonNegativeInt(inv.Args[2])
		if err != nil {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		removed, err := parseNonNegativeInt(inv.Args[3])
		if err != nil {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		modified := 0
		if len(inv.Args) == 5 {
			modified, err = parseNonNegativeInt(inv.Args[4])
			if err != nil {
				return Result{}, fmt.Errorf("usage: %s", c.Usage())
			}
		}
		upsertDiffEntry(cmdCtx.State, DiffEntry{Path: path, Added: added, Removed: removed, Modified: modified})
		cmdCtx.State.DiffCount++
		message := fmt.Sprintf("DIFF_ADD\npath=%s\nadded=%d\nremoved=%d\nmodified=%d\nentries=%d\nupdates=%d", normalizeToken(path), added, removed, modified, len(cmdCtx.State.DiffEntries), cmdCtx.State.DiffCount)
		return resultWithIntents(message, diffMutationIntents(path, added, removed, modified, len(cmdCtx.State.DiffEntries), cmdCtx.State.DiffCount)...), nil
	case "clear":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.DiffEntries = nil
		cmdCtx.State.DiffClearCount++
		message := fmt.Sprintf("DIFF_CLEAR\nentries=0\nupdates=%d\nclears=%d", cmdCtx.State.DiffCount, cmdCtx.State.DiffClearCount)
		return resultWithIntents(message, diffClearIntents(cmdCtx.State.DiffCount, cmdCtx.State.DiffClearCount)...), nil
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
}

// CostCommand reports cost command availability.
type CostCommand struct{}

func NewCostCommand() *CostCommand         { return &CostCommand{} }
func (c *CostCommand) Name() string        { return "cost" }
func (c *CostCommand) Aliases() []string   { return nil }
func (c *CostCommand) Description() string { return "Show token/cost usage summary" }
func (c *CostCommand) Usage() string       { return "/cost" }

func (c *CostCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if len(inv.Args) > 0 {
		return Result{}, fmt.Errorf("usage: /cost")
	}
	state := cmdCtx.State
	if state == nil {
		state = &RuntimeState{}
	}
	lines := []string{
		"COST_BREAKDOWN",
		fmt.Sprintf("input_tokens=%d", state.CostInputTokens),
		fmt.Sprintf("output_tokens=%d", state.CostOutputTokens),
		fmt.Sprintf("cache_read_tokens=%d", state.CostCacheRead),
		fmt.Sprintf("cache_write_tokens=%d", state.CostCacheWrite),
	}
	message := strings.Join(lines, "\n")
	return resultWithIntents(message, costBreakdownIntents(state.CostInputTokens, state.CostOutputTokens, state.CostCacheRead, state.CostCacheWrite)...), nil
}

// DoctorCommand reports doctor command availability.
type DoctorCommand struct{}

func NewDoctorCommand() *DoctorCommand       { return &DoctorCommand{} }
func (c *DoctorCommand) Name() string        { return "doctor" }
func (c *DoctorCommand) Aliases() []string   { return nil }
func (c *DoctorCommand) Description() string { return "Run runtime diagnostics" }
func (c *DoctorCommand) Usage() string       { return "/doctor [human|json|fix]" }

func (c *DoctorCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	format := "human"
	if len(inv.Args) == 1 {
		switch strings.ToLower(strings.TrimSpace(inv.Args[0])) {
		case "json", "--json", "machine":
			format = "json"
		case "fix", "repair":
			format = "fix"
		case "human", "status", "summary", "diagnostics":
			format = "human"
		default:
			return Result{}, fmt.Errorf("usage: /doctor [human|json|fix]")
		}
	} else if len(inv.Args) > 1 {
		return Result{}, fmt.Errorf("usage: /doctor [human|json|fix]")
	}

	if cmdCtx.State != nil {
		cmdCtx.State.DoctorRunCount++
	}
	res, err := c.renderDoctor(cmdCtx.State, format)
	if err != nil {
		return Result{}, err
	}
	if cmdCtx.State != nil {
		status, sections := doctorChecks(cmdCtx.State)
		cmdCtx.State.DoctorLastStatus = status
		cmdCtx.State.DoctorLastWarnSections = countWarnDoctorSections(sections)
		if format == "fix" {
			cmdCtx.State.DoctorFixCount++
			cmdCtx.State.DoctorLastQuickFixes = countDoctorFixes(res.Message)
		}
	}
	return res, nil
}

func (c *DoctorCommand) renderDoctor(state *RuntimeState, format string) (Result, error) {
	status, sections := doctorChecks(state)
	checks := flattenDoctorChecks(sections)
	if format == "fix" {
		return renderDoctorFixPlan(state, status, sections), nil
	}
	if format == "json" {
		payload, err := json.Marshal(struct {
			Status   string          `json:"status"`
			Sections []doctorSection `json:"sections"`
			Checks   []doctorCheck   `json:"checks"`
		}{
			Status:   status,
			Sections: sections,
			Checks:   checks,
		})
		if err != nil {
			return Result{}, fmt.Errorf("encode doctor report: %w", err)
		}
		return resultWithIntents(string(payload), legacyOutputIntents(string(payload))...), nil
	}

	lines := []string{
		"DOCTOR_REPORT",
		"format=human",
		fmt.Sprintf("status=%s", status),
		fmt.Sprintf("section_count=%d", len(sections)),
		fmt.Sprintf("check_count=%d", len(checks)),
	}
	for i, section := range sections {
		sidx := i + 1
		lines = append(lines, fmt.Sprintf("section.%d.name=%s", sidx, normalizeToken(section.Name)))
		lines = append(lines, fmt.Sprintf("section.%d.status=%s", sidx, normalizeToken(section.Status)))
		lines = append(lines, fmt.Sprintf("section.%d.check_count=%d", sidx, len(section.Checks)))
		for j, check := range section.Checks {
			cidx := j + 1
			lines = append(lines, fmt.Sprintf("section.%d.check.%d.id=%s", sidx, cidx, normalizeToken(check.ID)))
			lines = append(lines, fmt.Sprintf("section.%d.check.%d.status=%s", sidx, cidx, normalizeToken(check.Status)))
			lines = append(lines, fmt.Sprintf("section.%d.check.%d.detail=%s", sidx, cidx, normalizeToken(check.Detail)))
		}
	}
	return resultWithIntents(strings.Join(lines, "\n"), doctorReportIntents(status, sections)...), nil
}

func renderDoctorFixPlan(state *RuntimeState, status string, sections []doctorSection) Result {
	quickFixes := []string{}
	providerName := ""
	modelName := ""
	if state != nil {
		selection := RuntimeSelectionTruth(state)
		providerName = strings.TrimSpace(selection.ProviderName)
		modelName = strings.TrimSpace(selection.ModelName)
	}
	if providerName == "" {
		quickFixes = append(quickFixes, "/provider set ollama")
		quickFixes = append(quickFixes, "/provider doctor")
	}
	if modelName == "" {
		if providerName == "" {
			quickFixes = append(quickFixes, "/model ollama/llama3")
		} else {
			quickFixes = append(quickFixes, "/model "+providerName+"/<model>")
		}
	}
	quickFixes = append(quickFixes, "/sandbox check", "/config show")
	quickFixes = uniqueSortedStrings(quickFixes)

	warnSections := 0
	for _, section := range sections {
		if section.Status == "warn" {
			warnSections++
		}
	}
	lines := []string{
		"DOCTOR_FIX",
		fmt.Sprintf("status=%s", status),
		fmt.Sprintf("section_count=%d", len(sections)),
		fmt.Sprintf("warn_sections=%d", warnSections),
		fmt.Sprintf("fixes=%d", len(quickFixes)),
	}
	for i, fix := range quickFixes {
		lines = append(lines, fmt.Sprintf("fix.%d=%s", i+1, normalizeToken(fix)))
	}
	for i, loop := range correctiveLoopsForState(state) {
		idx := i + 1
		lines = append(lines, fmt.Sprintf("loop.%d.area=%s", idx, normalizeToken(loop.Area)))
		lines = append(lines, fmt.Sprintf("loop.%d.state=%s", idx, normalizeToken(loop.State)))
		lines = append(lines, fmt.Sprintf("loop.%d.action=%s", idx, normalizeToken(loop.Action)))
		lines = append(lines, fmt.Sprintf("loop.%d.next=%s", idx, normalizeToken(loop.Next)))
	}
	return resultWithIntents(strings.Join(lines, "\n"), doctorFixPlanIntents(status, warnSections, quickFixes, correctiveLoopsForState(state))...)
}

func countWarnDoctorSections(sections []doctorSection) int {
	count := 0
	for _, section := range sections {
		if strings.EqualFold(section.Status, "warn") {
			count++
		}
	}
	return count
}

func countDoctorFixes(message string) int {
	if strings.TrimSpace(message) == "" {
		return 0
	}
	count := 0
	for _, line := range strings.Split(message, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "fix.") {
			count++
		}
	}
	return count
}

type doctorCheck struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

type doctorSection struct {
	Name   string        `json:"name"`
	Status string        `json:"status"`
	Checks []doctorCheck `json:"checks"`
}

func doctorChecks(state *RuntimeState) (string, []doctorSection) {
	if state == nil {
		return "warn", []doctorSection{{
			Name:   "runtime",
			Status: "warn",
			Checks: []doctorCheck{{ID: "runtime_state", Status: "warn", Detail: "runtime state unavailable"}},
		}}
	}

	runtimeChecks := []doctorCheck{{ID: "runtime_state", Status: "ok", Detail: "runtime state available"}}
	selection := RuntimeSelectionTruth(state)
	transport := strings.TrimSpace(state.TransportMode)
	if transport == "" {
		transport = "local"
	}
	runtimeChecks = append(runtimeChecks, doctorCheck{ID: "transport", Status: "ok", Detail: fmt.Sprintf("mode=%s", transport)})

	providerChecks := make([]doctorCheck, 0, 3)
	providerConfiguredStatus := "warn"
	providerConfiguredDetail := "provider not configured"
	if strings.TrimSpace(selection.ProviderName) != "" {
		providerConfiguredStatus = "ok"
		providerConfiguredDetail = "provider configured"
	}
	providerChecks = append(providerChecks, doctorCheck{ID: "provider_configured", Status: providerConfiguredStatus, Detail: providerConfiguredDetail})

	providerReadyStatus := "warn"
	providerReadyDetail := "provider unavailable"
	if strings.TrimSpace(selection.ProviderName) == "" {
		providerReadyDetail = "provider unavailable (not configured)"
	} else if selection.ProviderReady {
		providerReadyStatus = "ok"
		providerReadyDetail = "provider available"
	}
	providerChecks = append(providerChecks, doctorCheck{ID: "provider_ready", Status: providerReadyStatus, Detail: providerReadyDetail})

	modelStatus := "warn"
	modelDetail := "model not selected"
	if strings.TrimSpace(selection.ModelName) != "" {
		modelStatus = "ok"
		modelDetail = "model selected"
	}
	providerChecks = append(providerChecks, doctorCheck{ID: "model", Status: modelStatus, Detail: modelDetail})

	configChecks := make([]doctorCheck, 0, 2)
	configStatus := "warn"
	configDetail := "layered config path unresolved"
	if strings.TrimSpace(state.ConfigPath) != "" {
		configStatus = "ok"
		configDetail = fmt.Sprintf("layered config path: %s", state.ConfigPath)
	}
	configChecks = append(configChecks, doctorCheck{ID: "config_path", Status: configStatus, Detail: configDetail})
	workspaceStatus := "warn"
	workspaceDetail := "workspace directories unavailable"
	if len(state.WorkspaceDirs) > 0 {
		workspaceStatus = "ok"
		workspaceDetail = fmt.Sprintf("workspace directories configured: %d", len(state.WorkspaceDirs))
	}
	configChecks = append(configChecks, doctorCheck{ID: "workspace_dirs", Status: workspaceStatus, Detail: workspaceDetail})

	sessionChecks := make([]doctorCheck, 0, 2)
	sessionChecks = append(sessionChecks, doctorCheck{ID: "permission_mode", Status: "ok", Detail: fmt.Sprintf("mode=%s", modeString(state.PermissionMode))})
	compactMode := strings.TrimSpace(state.CompactMode)
	if compactMode == "" {
		compactMode = "auto"
	}
	sessionChecks = append(sessionChecks, doctorCheck{ID: "compact_mode", Status: "ok", Detail: fmt.Sprintf("mode=%s", compactMode)})
	sessionChecks = append(sessionChecks, doctorCheck{ID: "auth", Status: ternaryStatus(selection.LoggedIn), Detail: fmt.Sprintf("logged_in=%t provider=%s", selection.LoggedIn, normalizeToken(selection.ProviderName))})

	integrationChecks := []doctorCheck{
		{ID: "terminal_setup", Status: ternaryStatus(state.TerminalConfigured), Detail: fmt.Sprintf("configured=%t profile=%s", state.TerminalConfigured, normalizeToken(state.TerminalProfile))},
		{ID: "github_app", Status: ternaryStatus(state.GitHubAppInstalls > 0), Detail: fmt.Sprintf("installs=%d repo=%s", state.GitHubAppInstalls, normalizeToken(state.LastGitHubRepo))},
		{ID: "slack_app", Status: ternaryStatus(state.SlackAppInstalls > 0), Detail: fmt.Sprintf("installs=%d", state.SlackAppInstalls)},
		{ID: "release_notes", Status: ternaryStatus(state.ReleaseNotesSeen > 0), Detail: fmt.Sprintf("seen=%d last=%s", state.ReleaseNotesSeen, normalizeToken(state.LastReleaseVersion))},
	}

	sections := []doctorSection{
		{Name: "runtime", Status: doctorSectionStatus(runtimeChecks), Checks: runtimeChecks},
		{Name: "provider", Status: doctorSectionStatus(providerChecks), Checks: providerChecks},
		{Name: "config", Status: doctorSectionStatus(configChecks), Checks: configChecks},
		{Name: "session", Status: doctorSectionStatus(sessionChecks), Checks: sessionChecks},
		{Name: "integrations", Status: doctorSectionStatus(integrationChecks), Checks: integrationChecks},
	}

	return doctorOverallStatus(sections), sections
}

func flattenDoctorChecks(sections []doctorSection) []doctorCheck {
	total := 0
	for _, section := range sections {
		total += len(section.Checks)
	}
	out := make([]doctorCheck, 0, total)
	for _, section := range sections {
		out = append(out, section.Checks...)
	}
	return out
}

func doctorSectionStatus(checks []doctorCheck) string {
	for _, check := range checks {
		if check.Status == "warn" {
			return "warn"
		}
	}
	return "ok"
}

func doctorOverallStatus(sections []doctorSection) string {
	for _, section := range sections {
		if section.Status == "warn" {
			return "warn"
		}
	}
	return "ok"
}

func ternaryStatus(ok bool) string {
	if ok {
		return "ok"
	}
	return "warn"
}

// ConfigCommand provides stable config command parsing and messages.
type ConfigCommand struct{}

func NewConfigCommand() *ConfigCommand       { return &ConfigCommand{} }
func (c *ConfigCommand) Name() string        { return "config" }
func (c *ConfigCommand) Aliases() []string   { return []string{"settings"} }
func (c *ConfigCommand) Description() string { return "Show or update configuration values" }
func (c *ConfigCommand) Usage() string {
	return "/config [show|status|doctor|panel|get <key>|set <key> <value>|unset <key>|repair [runtime|safe|strict]]"
}

const configUsage = "usage: /config [show|status|doctor|panel|get <key>|set <key> <value>|unset <key>|repair [runtime|safe|strict]]"

func (c *ConfigCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if len(inv.Args) > 0 && strings.EqualFold(inv.Args[0], "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf(configUsage)
		}
		return renderConfigStatus(cmdCtx.State), nil
	}

	if len(inv.Args) > 0 && strings.EqualFold(inv.Args[0], "doctor") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf(configUsage)
		}
		if cmdCtx.State != nil {
			cmdCtx.State.ConfigDoctorCount++
		}
		return renderConfigDoctor(cmdCtx.State), nil
	}

	if len(inv.Args) > 0 && strings.EqualFold(inv.Args[0], "panel") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf(configUsage)
		}
		return resultWithIntents("CONFIG_PANEL\nopened=true\nmode=interactive", configPanelIntents()...), nil
	}

	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "show") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf(configUsage)
		}
		if cmdCtx.State == nil || len(cmdCtx.State.ConfigValues) == 0 {
			return resultWithIntents("CONFIG_SHOW\ncount=0", configShowIntents(nil)...), nil
		}
		keys := make([]string, 0, len(cmdCtx.State.ConfigValues))
		for key := range cmdCtx.State.ConfigValues {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		lines := []string{"CONFIG_SHOW", fmt.Sprintf("count=%d", len(keys))}
		for i, key := range keys {
			idx := i + 1
			lines = append(lines, fmt.Sprintf("entry.%d.key=%s", idx, normalizeToken(key)))
			lines = append(lines, fmt.Sprintf("entry.%d.value=%s", idx, normalizeToken(cmdCtx.State.ConfigValues[key])))
		}
		return resultWithIntents(strings.Join(lines, "\n"), configShowIntents(cmdCtx.State.ConfigValues)...), nil
	}

	if strings.EqualFold(inv.Args[0], "get") {
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf(configUsage)
		}
		key := strings.TrimSpace(inv.Args[1])
		if key == "" {
			return Result{}, fmt.Errorf(configUsage)
		}
		if cmdCtx.State == nil || cmdCtx.State.ConfigValues == nil {
			message := fmt.Sprintf("CONFIG_GET\nkey=%s\nfound=false", normalizeToken(key))
			return resultWithIntents(message, configGetIntents(key, false, "")...), nil
		}
		value, ok := cmdCtx.State.ConfigValues[key]
		if !ok {
			message := fmt.Sprintf("CONFIG_GET\nkey=%s\nfound=false", normalizeToken(key))
			return resultWithIntents(message, configGetIntents(key, false, "")...), nil
		}
		message := fmt.Sprintf("CONFIG_GET\nkey=%s\nfound=true\nvalue=%s", normalizeToken(key), normalizeToken(value))
		return resultWithIntents(message, configGetIntents(key, true, value)...), nil
	}

	if strings.EqualFold(inv.Args[0], "repair") {
		if len(inv.Args) > 2 {
			return Result{}, fmt.Errorf(configUsage)
		}
		profile := "runtime"
		if len(inv.Args) == 2 {
			profile = strings.ToLower(strings.TrimSpace(inv.Args[1]))
			if profile == "" {
				return Result{}, fmt.Errorf(configUsage)
			}
			if profile != "runtime" && profile != "safe" && profile != "strict" {
				return Result{}, fmt.Errorf(configUsage)
			}
		}
		if cmdCtx.State.ConfigValues == nil {
			cmdCtx.State.ConfigValues = make(map[string]string)
		}
		changed := false
		if strings.TrimSpace(cmdCtx.State.ConfigValues["settings.output-style"]) == "" {
			cmdCtx.State.ConfigValues["settings.output-style"] = "human"
			changed = true
		}
		if strings.TrimSpace(cmdCtx.State.ConfigValues["settings.output-format"]) == "" {
			cmdCtx.State.ConfigValues["settings.output-format"] = "text"
			changed = true
		}
		if strings.TrimSpace(cmdCtx.State.ConfigValues["settings.transport"]) == "" {
			cmdCtx.State.ConfigValues["settings.transport"] = "local"
			changed = true
		}
		if profile == "safe" {
			if strings.TrimSpace(cmdCtx.State.ConfigValues["permissions.mode"]) == "" {
				cmdCtx.State.ConfigValues["permissions.mode"] = "default"
				changed = true
			}
		}
		if profile == "strict" {
			if strings.TrimSpace(cmdCtx.State.ConfigValues["permissions.mode"]) == "" {
				cmdCtx.State.ConfigValues["permissions.mode"] = "plan"
				changed = true
			}
			if strings.TrimSpace(cmdCtx.State.ConfigValues["settings.output-style"]) == "human" {
				cmdCtx.State.ConfigValues["settings.output-style"] = "compact"
				changed = true
			}
		}
		cmdCtx.State.ConfigRepairCount++
		cmdCtx.State.ConfigLastRepairProfile = profile
		message := fmt.Sprintf("CONFIG_REPAIR\nchanged=%t\noutput_style=%s\noutput_format=%s\ntransport=%s\nprofile=%s\nrepairs=%d", changed, normalizeToken(cmdCtx.State.ConfigValues["settings.output-style"]), normalizeToken(cmdCtx.State.ConfigValues["settings.output-format"]), normalizeToken(cmdCtx.State.ConfigValues["settings.transport"]), normalizeToken(profile), cmdCtx.State.ConfigRepairCount)
		return resultWithIntents(message, configRepairIntents(profile, changed, cmdCtx.State.ConfigRepairCount, cmdCtx.State.ConfigValues)...), nil
	}

	if strings.EqualFold(inv.Args[0], "set") {
		if len(inv.Args) < 3 {
			return Result{}, fmt.Errorf(configUsage)
		}
		if cmdCtx.State == nil {
			return Result{}, fmt.Errorf("missing command runtime state")
		}
		if cmdCtx.State.ConfigValues == nil {
			cmdCtx.State.ConfigValues = make(map[string]string)
		}
		key := strings.TrimSpace(inv.Args[1])
		if key == "" {
			return Result{}, fmt.Errorf(configUsage)
		}
		value := strings.Join(inv.Args[2:], " ")
		previous, existed := cmdCtx.State.ConfigValues[key]
		cmdCtx.State.ConfigValues[key] = value
		message := fmt.Sprintf("CONFIG_SET\nkey=%s\nupdated=%t\nvalue=%s", normalizeToken(key), existed, normalizeToken(value))
		return resultWithIntents(message, configMutationIntents("Config updated", key, value, !existed || previous != value)...), nil
	}

	if strings.EqualFold(inv.Args[0], "unset") {
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf(configUsage)
		}
		key := strings.TrimSpace(inv.Args[1])
		if key == "" {
			return Result{}, fmt.Errorf(configUsage)
		}
		if cmdCtx.State == nil || cmdCtx.State.ConfigValues == nil {
			message := fmt.Sprintf("CONFIG_UNSET\nkey=%s\nremoved=false", normalizeToken(key))
			return resultWithIntents(message, configMutationIntents("Config removed", key, "", false)...), nil
		}
		_, ok := cmdCtx.State.ConfigValues[key]
		delete(cmdCtx.State.ConfigValues, key)
		message := fmt.Sprintf("CONFIG_UNSET\nkey=%s\nremoved=%t", normalizeToken(key), ok)
		return resultWithIntents(message, configMutationIntents("Config removed", key, "", ok)...), nil
	}

	return Result{}, fmt.Errorf(configUsage)
}

func renderConfigStatus(state *RuntimeState) Result {
	if state == nil {
		message := "CONFIG_STATUS\ncount=0\nmissing_required=3\nrepairs=0\ndoctor_runs=0\nlast_profile=-"
		return resultWithIntents(message, configStatusIntents(nil, 3)...)
	}
	required := []string{"settings.output-style", "settings.output-format", "settings.transport"}
	missing := 0
	for _, key := range required {
		if strings.TrimSpace(state.ConfigValues[key]) == "" {
			missing++
		}
	}
	message := fmt.Sprintf("CONFIG_STATUS\ncount=%d\nmissing_required=%d\nrepairs=%d\ndoctor_runs=%d\nlast_profile=%s", len(state.ConfigValues), missing, state.ConfigRepairCount, state.ConfigDoctorCount, normalizeToken(state.ConfigLastRepairProfile))
	return resultWithIntents(message, configStatusIntents(state, missing)...)
}

func renderConfigDoctor(state *RuntimeState) Result {
	status := renderConfigStatus(state)
	quickFix := "/config repair"
	if state != nil && strings.TrimSpace(state.ConfigLastRepairProfile) != "" {
		quickFix = "/config repair " + strings.TrimSpace(state.ConfigLastRepairProfile)
	}
	lines := []string{
		"CONFIG_DOCTOR",
		strings.TrimSpace(strings.TrimPrefix(status.Message, "CONFIG_STATUS\n")),
		fmt.Sprintf("quick_fix=%s", normalizeToken(quickFix)),
	}
	missing := 3
	if state != nil {
		required := []string{"settings.output-style", "settings.output-format", "settings.transport"}
		missing = 0
		for _, key := range required {
			if strings.TrimSpace(state.ConfigValues[key]) == "" {
				missing++
			}
		}
	}
	return resultWithIntents(strings.Join(lines, "\n"), configDoctorIntents(state, missing, quickFix)...)
}

// InitCommand provides stable init command parsing and messages.
type InitCommand struct{}

func NewInitCommand() *InitCommand         { return &InitCommand{} }
func (c *InitCommand) Name() string        { return "init" }
func (c *InitCommand) Aliases() []string   { return nil }
func (c *InitCommand) Description() string { return "Initialize project configuration" }
func (c *InitCommand) Usage() string       { return "/init [path]" }

func (c *InitCommand) Execute(_ context.Context, _ Context, inv Invocation) (Result, error) {
	if len(inv.Args) > 1 {
		return Result{}, fmt.Errorf("usage: /init [path]")
	}
	if len(inv.Args) == 0 {
		return resultWithIntents("Initialization requested for current workspace.", legacyOutputIntents("Initialization requested for current workspace.")...), nil
	}
	return resultWithIntents(fmt.Sprintf("Initialization requested for %s.", inv.Args[0]), legacyOutputIntents(fmt.Sprintf("Initialization requested for %s.", inv.Args[0]))...), nil
}

// CopyCommand provides stable copy command parsing and messages.
type CopyCommand struct{}

func NewCopyCommand() *CopyCommand         { return &CopyCommand{} }
func (c *CopyCommand) Name() string        { return "copy" }
func (c *CopyCommand) Aliases() []string   { return nil }
func (c *CopyCommand) Description() string { return "Copy text or selection to clipboard" }
func (c *CopyCommand) Usage() string       { return "/copy <text>" }

func (c *CopyCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if len(inv.Args) == 0 {
		return Result{}, fmt.Errorf("usage: /copy <text>")
	}
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	text := strings.Join(inv.Args, " ")
	normalized := normalizeToken(text)
	cmdCtx.State.CopyCount++
	cmdCtx.State.LastCopiedText = text
	message := fmt.Sprintf("COPY_RESULT\ntext=%s\ncount=%d\nlast_text=%s", normalized, cmdCtx.State.CopyCount, normalizeToken(cmdCtx.State.LastCopiedText))
	return resultWithIntents(message, copyResultIntents(text, cmdCtx.State.CopyCount, cmdCtx.State.LastCopiedText)...), nil
}

// VersionCommand reports slash command runtime version.
type VersionCommand struct{}

func NewVersionCommand() *VersionCommand      { return &VersionCommand{} }
func (c *VersionCommand) Name() string        { return "version" }
func (c *VersionCommand) Aliases() []string   { return nil }
func (c *VersionCommand) Description() string { return "Show CLI version information" }
func (c *VersionCommand) Usage() string       { return "/version" }

func (c *VersionCommand) Execute(_ context.Context, _ Context, inv Invocation) (Result, error) {
	if len(inv.Args) > 0 {
		return Result{}, fmt.Errorf("usage: /version")
	}
	message := "VERSION_INFO\nruntime=parity\nversion=v1"
	return resultWithIntents(message, versionInfoIntents("parity", "v1")...), nil
}

// UsageCommand reports session usage-oriented state.
type UsageCommand struct{}

func NewUsageCommand() *UsageCommand      { return &UsageCommand{} }
func (c *UsageCommand) Name() string      { return "usage" }
func (c *UsageCommand) Aliases() []string { return nil }
func (c *UsageCommand) Description() string {
	return "Show current session usage snapshot"
}
func (c *UsageCommand) Usage() string { return "/usage" }

func (c *UsageCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if len(inv.Args) > 0 {
		return Result{}, fmt.Errorf("usage: /usage")
	}
	if cmdCtx.State == nil {
		message := "USAGE_SNAPSHOT\nmodel=unknown\npermission=default\ncompact_requested=false"
		return resultWithIntents(message, usageSnapshotIntents("unknown", "default", false)...), nil
	}
	model := cmdCtx.State.Model
	if strings.TrimSpace(model) == "" {
		model = "unknown"
	}
	lines := []string{
		"USAGE_SNAPSHOT",
		fmt.Sprintf("model=%s", normalizeToken(model)),
		fmt.Sprintf("permission=%s", modeString(cmdCtx.State.PermissionMode)),
		fmt.Sprintf("compact_requested=%t", cmdCtx.State.CompactRequested),
	}
	message := strings.Join(lines, "\n")
	return resultWithIntents(message, usageSnapshotIntents(model, modeString(cmdCtx.State.PermissionMode), cmdCtx.State.CompactRequested)...), nil
}

// ContextCommand reports and mutates local context flags.
type ContextCommand struct{}

func NewContextCommand() *ContextCommand      { return &ContextCommand{} }
func (c *ContextCommand) Name() string        { return "context" }
func (c *ContextCommand) Aliases() []string   { return nil }
func (c *ContextCommand) Description() string { return "Show or clear local context flags" }
func (c *ContextCommand) Usage() string       { return "/context [show|clear]" }

func (c *ContextCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}

	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "show") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /context [show|clear]")
		}
		lines := []string{
			"CONTEXT_STATE",
			fmt.Sprintf("compact_requested=%t", cmdCtx.State.CompactRequested),
			fmt.Sprintf("resume_requested=%t", cmdCtx.State.ResumeRequested),
			fmt.Sprintf("last_resume_target=%s", normalizeToken(cmdCtx.State.LastResumeTarget)),
			fmt.Sprintf("last_compact_target=%s", normalizeToken(cmdCtx.State.LastCompactTarget)),
			fmt.Sprintf("compact_mode=%s", normalizeToken(cmdCtx.State.CompactMode)),
			fmt.Sprintf("permission_mode=%s", modeString(cmdCtx.State.PermissionMode)),
			fmt.Sprintf("model=%s", normalizeToken(cmdCtx.State.Model)),
			fmt.Sprintf("provider=%s", normalizeToken(cmdCtx.State.ProviderName)),
			fmt.Sprintf("workspace_dir_count=%d", len(cmdCtx.State.WorkspaceDirs)),
			fmt.Sprintf("branch_active=%s", normalizeToken(cmdCtx.State.ActiveBranch)),
			fmt.Sprintf("diff_entries=%d", len(cmdCtx.State.DiffEntries)),
		}
		message := strings.Join(lines, "\n")
		return resultWithIntents(message, contextStateIntents(cmdCtx.State)...), nil
	}

	if strings.EqualFold(inv.Args[0], "clear") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /context [show|clear]")
		}
		cmdCtx.State.LastResumeTarget = ""
		cmdCtx.State.LastCompactTarget = ""
		cmdCtx.State.CompactRequested = false
		cmdCtx.State.ResumeRequested = false
		message := "CONTEXT_CLEAR\ncompact_requested=false\nresume_requested=false\nlast_resume_target=-\nlast_compact_target=-"
		return resultWithIntents(message, contextClearedIntents()...), nil
	}

	return Result{}, fmt.Errorf("usage: /context [show|clear]")
}

// ExitCommand provides deterministic exit intent.
type ExitCommand struct{}

func NewExitCommand() *ExitCommand         { return &ExitCommand{} }
func (c *ExitCommand) Name() string        { return "exit" }
func (c *ExitCommand) Aliases() []string   { return []string{"quit"} }
func (c *ExitCommand) Description() string { return "Exit current session" }
func (c *ExitCommand) Usage() string       { return "/exit" }

func (c *ExitCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) > 0 {
		return Result{}, fmt.Errorf("usage: /exit")
	}
	cmdCtx.State.ExitRequested = true
	cmdCtx.State.ExitCount++
	cmdCtx.State.LastExitCommand = inv.Name
	message := fmt.Sprintf("EXIT_REQUEST\nrequested=true\ncount=%d\ncommand=%s", cmdCtx.State.ExitCount, normalizeToken(cmdCtx.State.LastExitCommand))
	return resultWithIntents(message, exitRequestIntents(cmdCtx.State.ExitCount, cmdCtx.State.LastExitCommand)...), nil
}

// PlanCommand provides deterministic plan-mode behavior.
type PlanCommand struct{}

func NewPlanCommand() *PlanCommand       { return &PlanCommand{} }
func (c *PlanCommand) Name() string      { return "plan" }
func (c *PlanCommand) Aliases() []string { return nil }
func (c *PlanCommand) Description() string {
	return "Enable plan mode or view current plan"
}
func (c *PlanCommand) Usage() string { return "/plan [open|status|<description>]" }

func (c *PlanCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		message := fmt.Sprintf("PLAN_STATUS\nenabled=%t\nenable_count=%d\nopen_count=%d\nlast_description=%s", cmdCtx.State.PlanModeEnabled, cmdCtx.State.PlanEnableCount, cmdCtx.State.PlanOpenCount, normalizeToken(cmdCtx.State.LastPlanDescription))
		return resultWithIntents(message, planStatusIntents(cmdCtx.State.PlanModeEnabled, cmdCtx.State.PlanEnableCount, cmdCtx.State.PlanOpenCount, cmdCtx.State.LastPlanDescription)...), nil
	}

	raw := strings.TrimSpace(strings.Join(inv.Args, " "))
	if !cmdCtx.State.PlanModeEnabled {
		cmdCtx.State.PlanModeEnabled = true
		cmdCtx.State.PlanEnableCount++
		if raw != "" && !strings.EqualFold(raw, "open") {
			cmdCtx.State.LastPlanDescription = raw
			message := fmt.Sprintf("PLAN_ENABLE\nenabled=true\nenable_count=%d\ndescription=%s\nquery_hint=true", cmdCtx.State.PlanEnableCount, normalizeToken(raw))
			return resultWithIntents(message, planMutationIntents("Plan enabled", raw, true, cmdCtx.State.PlanOpenCount)...), nil
		}
		message := fmt.Sprintf("PLAN_ENABLE\nenabled=true\nenable_count=%d\ndescription=-\nquery_hint=false", cmdCtx.State.PlanEnableCount)
		return resultWithIntents(message, planMutationIntents("Plan enabled", "", false, cmdCtx.State.PlanOpenCount)...), nil
	}

	if len(inv.Args) > 0 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "open") {
		cmdCtx.State.PlanOpenCount++
		message := fmt.Sprintf("PLAN_OPEN\npath=.claude/plan.md\nopen_count=%d", cmdCtx.State.PlanOpenCount)
		return resultWithIntents(message, planMutationIntents("Plan opened", cmdCtx.State.LastPlanDescription, false, cmdCtx.State.PlanOpenCount)...), nil
	}

	message := fmt.Sprintf("PLAN_CURRENT\nenabled=true\npath=.claude/plan.md\nlast_description=%s", normalizeToken(cmdCtx.State.LastPlanDescription))
	return resultWithIntents(message, planMutationIntents("Plan current", cmdCtx.State.LastPlanDescription, false, cmdCtx.State.PlanOpenCount)...), nil
}

// ReviewCommand provides deterministic review intent behavior.
type ReviewCommand struct{}

func NewReviewCommand() *ReviewCommand       { return &ReviewCommand{} }
func (c *ReviewCommand) Name() string        { return "review" }
func (c *ReviewCommand) Aliases() []string   { return nil }
func (c *ReviewCommand) Description() string { return "Review a pull request" }
func (c *ReviewCommand) Usage() string {
	return "/review [status [<pr-or-target>]|focus <risk|perf|tests|security>|<pr-or-target>]"
}

func (c *ReviewCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) > 0 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		targetRaw := strings.TrimSpace(strings.Join(inv.Args[1:], " "))
		if len(inv.Args) > 1 && targetRaw == "" {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		if targetRaw == "" {
			targetRaw = cmdCtx.State.LastReviewTarget
		}
		last := classifyReviewTarget(cmdCtx.State.LastReviewTarget)
		target := classifyReviewTarget(targetRaw)
		sections := buildReviewChecklistSections(target.Mode)
		lines := []string{
			"REVIEW_STATUS",
			fmt.Sprintf("count=%d", cmdCtx.State.ReviewCount),
			fmt.Sprintf("last_target=%s", normalizeToken(cmdCtx.State.LastReviewTarget)),
			fmt.Sprintf("last_mode=%s", normalizeToken(last.Mode)),
			fmt.Sprintf("mode=%s", normalizeToken(target.Mode)),
			fmt.Sprintf("target=%s", normalizeToken(target.Normalized)),
			fmt.Sprintf("section_count=%d", len(sections)),
		}
		for i, section := range sections {
			sidx := i + 1
			lines = append(lines, fmt.Sprintf("section.%d.title=%s", sidx, normalizeToken(section.Title)))
			lines = append(lines, fmt.Sprintf("section.%d.item_count=%d", sidx, len(section.Items)))
			for j, item := range section.Items {
				lines = append(lines, fmt.Sprintf("section.%d.item.%d=%s", sidx, j+1, normalizeToken(item)))
			}
		}
		return resultWithIntents(strings.Join(lines, "\n"), legacyOutputIntents(strings.Join(lines, "\n"))...), nil
	}
	if len(inv.Args) == 2 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "focus") {
		focus := strings.ToLower(strings.TrimSpace(inv.Args[1]))
		switch focus {
		case "risk", "perf", "tests", "security":
			cmdCtx.State.LastReviewTarget = "focus:" + focus
			return resultWithIntents(fmt.Sprintf("REVIEW_FOCUS\nfocus=%s\nnext=/review_status", focus), legacyOutputIntents(fmt.Sprintf("REVIEW_FOCUS\nfocus=%s\nnext=/review_status", focus))...), nil
		default:
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
	}

	rawTarget := strings.TrimSpace(strings.Join(inv.Args, " "))
	if len(inv.Args) > 0 && rawTarget == "" {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	target := classifyReviewTarget(rawTarget)
	if target.Normalized != "-" {
		cmdCtx.State.LastReviewTarget = target.Normalized
	}
	cmdCtx.State.ReviewCount++
	checklist := buildReviewChecklist(target.Mode)
	lines := []string{
		"REVIEW_REQUEST",
		fmt.Sprintf("mode=%s", normalizeToken(target.Mode)),
		fmt.Sprintf("target=%s", normalizeToken(target.Normalized)),
		fmt.Sprintf("count=%d", cmdCtx.State.ReviewCount),
		fmt.Sprintf("checklist.count=%d", len(checklist)),
	}
	for i, item := range checklist {
		lines = append(lines, fmt.Sprintf("checklist.%d=%s", i+1, normalizeToken(item)))
	}
	return resultWithIntents(strings.Join(lines, "\n"), legacyOutputIntents(strings.Join(lines, "\n"))...), nil
}

type reviewTarget struct {
	Mode       string
	Normalized string
}

type reviewChecklistSection struct {
	Title string
	Items []string
}

func classifyReviewTarget(raw string) reviewTarget {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return reviewTarget{Mode: "list", Normalized: "-"}
	}
	if strings.HasPrefix(raw, "#") {
		raw = strings.TrimPrefix(raw, "#")
	}
	if _, err := strconv.Atoi(raw); err == nil {
		return reviewTarget{Mode: "pr", Normalized: raw}
	}
	if isHexID(raw) {
		return reviewTarget{Mode: "commit", Normalized: raw}
	}
	if strings.Contains(raw, "...") {
		return reviewTarget{Mode: "range", Normalized: raw}
	}
	if strings.Contains(raw, "/") && !strings.Contains(raw, ".") {
		return reviewTarget{Mode: "branch", Normalized: raw}
	}
	if strings.Contains(raw, ".") {
		return reviewTarget{Mode: "file", Normalized: raw}
	}
	return reviewTarget{Mode: "target", Normalized: raw}
}

func buildReviewChecklist(mode string) []string {
	sections := buildReviewChecklistSections(mode)
	total := 0
	for _, section := range sections {
		total += len(section.Items)
	}
	items := make([]string, 0, total)
	for _, section := range sections {
		items = append(items, section.Items...)
	}
	return items
}

func buildReviewChecklistSections(mode string) []reviewChecklistSection {
	sections := []reviewChecklistSection{
		{
			Title: "Intent",
			Items: []string{
				"Summarize intent and changed scope",
				"Call out assumptions and non-goals",
			},
		},
		{
			Title: "Correctness",
			Items: []string{
				"Validate correctness and edge cases",
				"Verify test coverage and missing cases",
			},
		},
		{
			Title: "Risk",
			Items: []string{"Assess security and data handling"},
		},
	}
	modeItem := "Capture follow-up questions for clarification"
	switch mode {
	case "pr":
		modeItem = "Check merge readiness and unresolved feedback"
	case "range":
		modeItem = "Look for regressions across commit boundaries"
	case "file":
		modeItem = "Confirm file-level conventions and ownership"
	case "branch":
		modeItem = "Review branch diff against intended base"
	case "commit":
		modeItem = "Inspect commit message and atomicity"
	}
	sections = append(sections, reviewChecklistSection{Title: "Mode-specific", Items: []string{modeItem}})
	return sections
}

func parseNumstatTotals(raw string) (added int, removed int) {
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 2 {
			continue
		}
		a := strings.TrimSpace(parts[0])
		r := strings.TrimSpace(parts[1])
		if a != "-" {
			if av, err := strconv.Atoi(a); err == nil && av > 0 {
				added += av
			}
		}
		if r != "-" {
			if rv, err := strconv.Atoi(r); err == nil && rv > 0 {
				removed += rv
			}
		}
	}
	return added, removed
}

type gitWorkingTreeSummary struct {
	Available bool
	Branch    string
	Upstream  string
	Tracked   bool
	Staged    int
	Unstaged  int
	Untracked int
	Files     int
	Added     int
	Removed   int
	Modified  int
	Remote    bool
	Err       string
}

func readGitWorkingTreeSummary(ctx context.Context) gitWorkingTreeSummary {
	if ctx == nil {
		ctx = context.Background()
	}
	summary := gitWorkingTreeSummary{}
	inside, err := runGit(ctx, "rev-parse", "--is-inside-work-tree")
	if err != nil || strings.TrimSpace(inside) != "true" {
		summary.Err = "not_a_git_repository"
		return summary
	}
	summary.Available = true

	if branch, err := runGit(ctx, "branch", "--show-current"); err == nil {
		summary.Branch = strings.TrimSpace(branch)
	}
	if upstream, err := runGit(ctx, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}"); err == nil {
		summary.Upstream = strings.TrimSpace(upstream)
		summary.Tracked = summary.Upstream != ""
	}
	if remotes, err := runGit(ctx, "remote"); err == nil {
		for _, remote := range strings.Split(remotes, "\n") {
			if strings.TrimSpace(remote) == "origin" {
				summary.Remote = true
				break
			}
		}
	}

	porcelain, err := runGit(ctx, "status", "--porcelain=v1", "--branch")
	if err != nil {
		summary.Err = "git_status_failed"
		return summary
	}
	for _, line := range strings.Split(porcelain, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, "##") {
			if summary.Branch == "" {
				branch := strings.TrimSpace(strings.TrimPrefix(line, "##"))
				if idx := strings.Index(branch, "..."); idx >= 0 {
					summary.Branch = strings.TrimSpace(branch[:idx])
				} else {
					summary.Branch = branch
				}
			}
			continue
		}
		if len(line) < 3 {
			continue
		}
		summary.Files++
		x := line[0]
		y := line[1]
		if x == '?' && y == '?' {
			summary.Untracked++
			continue
		}
		if x != ' ' {
			summary.Staged++
		}
		if y != ' ' {
			summary.Unstaged++
		}
		if x == 'M' || y == 'M' || x == 'R' || y == 'R' || x == 'C' || y == 'C' || x == 'T' || y == 'T' {
			summary.Modified++
		}
	}

	if unstagedNumstat, err := runGit(ctx, "diff", "--numstat"); err == nil {
		a, r := parseNumstatTotals(unstagedNumstat)
		summary.Added += a
		summary.Removed += r
	}
	if stagedNumstat, err := runGit(ctx, "diff", "--cached", "--numstat"); err == nil {
		a, r := parseNumstatTotals(stagedNumstat)
		summary.Added += a
		summary.Removed += r
	}
	if summary.Modified == 0 {
		summary.Modified = summary.Staged + summary.Unstaged
	}
	if summary.Branch == "" {
		summary.Branch = "main"
	}
	if summary.Upstream == "" {
		summary.Upstream = "-"
	}
	return summary
}

func runGit(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out)), nil
}

type commitPlanStep struct {
	Status string
	Action string
}

func buildCommitPlanSteps(summary diffSummary) []commitPlanStep {
	if summary.Files == 0 {
		return []commitPlanStep{
			{Status: "blocked", Action: "Create or modify files before committing"},
			{Status: "blocked", Action: "Run /diff list to confirm pending changes"},
		}
	}
	steps := []commitPlanStep{
		{Status: "ready", Action: "Review staged diff for correctness"},
		{Status: "ready", Action: "Use suggested message as commit starting point"},
		{Status: "required", Action: "Run tests before finalizing commit"},
	}
	if summary.Unstaged > 0 || summary.Untracked > 0 {
		steps = append([]commitPlanStep{{Status: "required", Action: "Stage intended files with git add"}}, steps...)
	}
	return steps
}

func isHexID(v string) bool {
	v = strings.TrimSpace(strings.ToLower(v))
	if len(v) < 7 || len(v) > 40 {
		return false
	}
	for _, r := range v {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}

// SessionCommand provides deterministic remote session details.
type SessionCommand struct{}

func NewSessionCommand() *SessionCommand      { return &SessionCommand{} }
func (c *SessionCommand) Name() string        { return "session" }
func (c *SessionCommand) Aliases() []string   { return []string{"remote"} }
func (c *SessionCommand) Description() string { return "Show remote session URL and QR availability" }
func (c *SessionCommand) Usage() string {
	return "/session [status|set-url <url>|host [addr]|connect <addr> <token>|token [show|set <token>|clear]|disconnect]"
}

const sessionUsage = "usage: /session [status|set-url <url>|host [addr]|connect <addr> <token>|token [show|set <token>|clear]|disconnect]"

func (c *SessionCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}

	if len(inv.Args) > 0 {
		sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
		switch sub {
		case "set-url":
			if len(inv.Args) < 2 {
				return Result{}, fmt.Errorf(sessionUsage)
			}
			url := strings.TrimSpace(strings.Join(inv.Args[1:], " "))
			if url == "" {
				return Result{}, fmt.Errorf(sessionUsage)
			}
			cmdCtx.State.RemoteSessionURL = url
			message := fmt.Sprintf("SESSION_URL_SET\nurl=%s", normalizeToken(url))
			return resultWithIntents(message, sessionURLIntents(url)...), nil
		case "host":
			if len(inv.Args) > 2 {
				return Result{}, fmt.Errorf(sessionUsage)
			}
			addr := "127.0.0.1:4317"
			if len(inv.Args) == 2 {
				addr = strings.TrimSpace(inv.Args[1])
				if addr == "" {
					return Result{}, fmt.Errorf(sessionUsage)
				}
			}
			if err := closeSessionRuntime(cmdCtx.State); err != nil {
				return Result{}, err
			}
			nextHostCount := cmdCtx.State.SessionHostCount + 1
			token := fmt.Sprintf("host-%s-%d", strings.ReplaceAll(addr, " ", ""), nextHostCount)
			transport, err := remote.NewTCPTransport(remote.TCPTransportConfig{
				Mode:  remote.TCPTransportModeHost,
				Addr:  addr,
				Token: token,
			})
			if err != nil {
				return Result{}, fmt.Errorf("session host failed: %w", err)
			}
			mgr, err := remote.NewSessionManager(transport, sessionIDOrDefault(cmdCtx.State), remote.DefaultRetryPolicy(), time.Now().UTC())
			if err != nil {
				_ = transport.Shutdown()
				return Result{}, fmt.Errorf("session host failed: %w", err)
			}

			cmdCtx.State.SessionMode = "host"
			cmdCtx.State.SessionHosted = true
			cmdCtx.State.SessionHostAddr = transport.Addr()
			cmdCtx.State.SessionConnected = false
			cmdCtx.State.SessionConnectedAddr = ""
			cmdCtx.State.SessionTokenSource = "host"
			cmdCtx.State.SessionToken = token
			cmdCtx.State.SessionTokenPrefix = sessionTokenPrefix(token)
			cmdCtx.State.SessionHostCount = nextHostCount
			cmdCtx.State.SessionTransport = transport
			cmdCtx.State.SessionTCPTransport = transport
			cmdCtx.State.SessionManager = mgr
			message := fmt.Sprintf("SESSION_HOST\nhosting=true\naddr=%s\ntoken_prefix=%s\nhost_count=%d", normalizeToken(cmdCtx.State.SessionHostAddr), normalizeToken(cmdCtx.State.SessionTokenPrefix), cmdCtx.State.SessionHostCount)
			return resultWithIntents(message, sessionHostIntents(cmdCtx.State.SessionHostAddr, cmdCtx.State.SessionTokenPrefix, cmdCtx.State.SessionHostCount)...), nil
		case "connect":
			if len(inv.Args) != 3 {
				return Result{}, fmt.Errorf(sessionUsage)
			}
			addr := strings.TrimSpace(inv.Args[1])
			token := strings.TrimSpace(inv.Args[2])
			if addr == "" || token == "" {
				return Result{}, fmt.Errorf(sessionUsage)
			}
			if err := closeSessionRuntime(cmdCtx.State); err != nil {
				return Result{}, err
			}
			transport, err := remote.NewTCPTransport(remote.TCPTransportConfig{
				Mode:  remote.TCPTransportModeClient,
				Addr:  addr,
				Token: token,
			})
			if err != nil {
				return Result{}, fmt.Errorf("session connect failed: %w", err)
			}
			mgr, err := remote.NewSessionManager(transport, sessionIDOrDefault(cmdCtx.State), remote.DefaultRetryPolicy(), time.Now().UTC())
			if err != nil {
				_ = transport.Shutdown()
				return Result{}, fmt.Errorf("session connect failed: %w", err)
			}

			connectCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := mgr.Connect(connectCtx); err != nil {
				_ = mgr.Close()
				_ = transport.Shutdown()
				return Result{}, fmt.Errorf("session connect failed: %w", err)
			}

			cmdCtx.State.SessionMode = "client"
			cmdCtx.State.SessionHosted = false
			cmdCtx.State.SessionHostAddr = ""
			cmdCtx.State.SessionConnected = true
			cmdCtx.State.SessionConnectedAddr = addr
			cmdCtx.State.SessionTokenSource = "connect"
			cmdCtx.State.SessionToken = token
			cmdCtx.State.SessionTokenPrefix = sessionTokenPrefix(token)
			cmdCtx.State.SessionConnectCount++
			cmdCtx.State.SessionTransport = transport
			cmdCtx.State.SessionTCPTransport = transport
			cmdCtx.State.SessionManager = mgr
			message := fmt.Sprintf("SESSION_CONNECT\nconnected=true\naddr=%s\ntoken_prefix=%s\nconnect_count=%d", normalizeToken(addr), normalizeToken(cmdCtx.State.SessionTokenPrefix), cmdCtx.State.SessionConnectCount)
			return resultWithIntents(message, sessionConnectIntents(addr, cmdCtx.State.SessionTokenPrefix, cmdCtx.State.SessionConnectCount)...), nil
		case "disconnect":
			if len(inv.Args) != 1 {
				return Result{}, fmt.Errorf(sessionUsage)
			}
			wasConnected := cmdCtx.State.SessionConnected || strings.TrimSpace(cmdCtx.State.SessionConnectedAddr) != ""
			previousAddr := cmdCtx.State.SessionConnectedAddr
			if err := closeSessionRuntime(cmdCtx.State); err != nil {
				return Result{}, err
			}
			cmdCtx.State.SessionMode = ""
			cmdCtx.State.SessionHosted = false
			cmdCtx.State.SessionHostAddr = ""
			cmdCtx.State.SessionConnected = false
			cmdCtx.State.SessionConnectedAddr = ""
			cmdCtx.State.SessionTokenSource = ""
			cmdCtx.State.SessionToken = ""
			cmdCtx.State.SessionTokenPrefix = ""
			cmdCtx.State.SessionDisconnectCount++
			message := fmt.Sprintf("SESSION_DISCONNECT\nconnected=false\nwas_connected=%t\nprevious_addr=%s\ndisconnect_count=%d", wasConnected, normalizeToken(previousAddr), cmdCtx.State.SessionDisconnectCount)
			return resultWithIntents(message, sessionDisconnectIntents(wasConnected, previousAddr, cmdCtx.State.SessionDisconnectCount)...), nil
		case "diagnostics", "doctor":
			if len(inv.Args) != 1 {
				return Result{}, fmt.Errorf(sessionUsage)
			}
			cmdCtx.State.SessionDiagnosticsCount++
			return renderSessionDiagnostics(cmdCtx.State), nil
		case "repair":
			if len(inv.Args) > 2 {
				return Result{}, fmt.Errorf(sessionUsage)
			}
			action := "auto"
			if len(inv.Args) == 2 {
				action = strings.ToLower(strings.TrimSpace(inv.Args[1]))
				if action == "" {
					return Result{}, fmt.Errorf(sessionUsage)
				}
			}
			return repairSessionState(cmdCtx.State, action), nil
		case "status":
			if len(inv.Args) != 1 {
				return Result{}, fmt.Errorf(sessionUsage)
			}
		case "token":
			if len(inv.Args) == 1 || (len(inv.Args) == 2 && strings.EqualFold(strings.TrimSpace(inv.Args[1]), "show")) {
				message := fmt.Sprintf("SESSION_TOKEN\nsource=%s\nprefix=%s\nset=%t", normalizeToken(cmdCtx.State.SessionTokenSource), normalizeToken(cmdCtx.State.SessionTokenPrefix), strings.TrimSpace(cmdCtx.State.SessionToken) != "")
				return resultWithIntents(message, sessionTokenIntents(cmdCtx.State.SessionTokenSource, cmdCtx.State.SessionTokenPrefix, strings.TrimSpace(cmdCtx.State.SessionToken) != "")...), nil
			}
			if len(inv.Args) == 2 && strings.EqualFold(strings.TrimSpace(inv.Args[1]), "clear") {
				cmdCtx.State.SessionTokenSource = ""
				cmdCtx.State.SessionToken = ""
				cmdCtx.State.SessionTokenPrefix = ""
				return resultWithIntents("SESSION_TOKEN_SET\nset=false\nprefix=-", sessionTokenIntents("", "", false)...), nil
			}
			if len(inv.Args) < 3 || !strings.EqualFold(strings.TrimSpace(inv.Args[1]), "set") {
				return Result{}, fmt.Errorf(sessionUsage)
			}
			token := strings.TrimSpace(strings.Join(inv.Args[2:], " "))
			if token == "" {
				return Result{}, fmt.Errorf(sessionUsage)
			}
			cmdCtx.State.SessionTokenSource = "manual"
			cmdCtx.State.SessionToken = token
			cmdCtx.State.SessionTokenPrefix = sessionTokenPrefix(token)
			message := fmt.Sprintf("SESSION_TOKEN_SET\nset=true\nprefix=%s", normalizeToken(cmdCtx.State.SessionTokenPrefix))
			return resultWithIntents(message, sessionTokenIntents(cmdCtx.State.SessionTokenSource, cmdCtx.State.SessionTokenPrefix, true)...), nil
		default:
			return Result{}, fmt.Errorf(sessionUsage)
		}
	}

	cmdCtx.State.SessionViewCount++
	url := strings.TrimSpace(cmdCtx.State.RemoteSessionURL)
	qrAvailable := url != ""
	remoteMode := strings.EqualFold(strings.TrimSpace(cmdCtx.State.TransportMode), "remote")
	health := sessionHealthSummary(cmdCtx.State.SessionManager)
	message := fmt.Sprintf("SESSION_INFO\nremote_mode=%t\nsession_id=%s\nsession_path=%s\nurl=%s\nqr_available=%t\nviews=%d\nhosting=%t\nhost_addr=%s\nconnected=%t\nconnected_addr=%s\ntoken_source=%s\ntoken_prefix=%s\nhost_count=%d\nconnect_count=%d\ndisconnect_count=%d\nmode=%s\nmanager_state=%s\ntransport_state=%s\nreconnecting=%t\nreconnect_reason=%s\nreconnect_error_class=%s\nreconnect_count=%d\nreconnect_detail=%s", remoteMode, normalizeToken(cmdCtx.State.SessionID), normalizeToken(cmdCtx.State.SessionPath), normalizeToken(url), qrAvailable, cmdCtx.State.SessionViewCount, cmdCtx.State.SessionHosted, normalizeToken(cmdCtx.State.SessionHostAddr), cmdCtx.State.SessionConnected, normalizeToken(cmdCtx.State.SessionConnectedAddr), normalizeToken(cmdCtx.State.SessionTokenSource), normalizeToken(cmdCtx.State.SessionTokenPrefix), cmdCtx.State.SessionHostCount, cmdCtx.State.SessionConnectCount, cmdCtx.State.SessionDisconnectCount, normalizeToken(cmdCtx.State.SessionMode), normalizeToken(health.managerState), normalizeToken(health.transportState), health.reconnecting, normalizeToken(health.reconnectReason), normalizeToken(health.reconnectErrorClass), health.reconnectCount, normalizeToken(health.reconnectDetail))
	return resultWithIntents(message, sessionInfoIntents(cmdCtx.State, remoteMode, qrAvailable, health)...), nil
}

func renderSessionDiagnostics(state *RuntimeState) Result {
	if state == nil {
		state = &RuntimeState{}
	}
	health := sessionHealthSummary(state.SessionManager)
	quickFix := "/session host"
	if strings.TrimSpace(state.SessionMode) == "client" && !state.SessionConnected {
		quickFix = "/session connect <addr> <token>"
	}
	if strings.TrimSpace(state.SessionToken) == "" {
		quickFix = "/session token set <token>"
	}
	lines := []string{
		"SESSION_DIAGNOSTICS",
		fmt.Sprintf("mode=%s", normalizeToken(state.SessionMode)),
		fmt.Sprintf("hosting=%t", state.SessionHosted),
		fmt.Sprintf("connected=%t", state.SessionConnected),
		fmt.Sprintf("host_addr=%s", normalizeToken(state.SessionHostAddr)),
		fmt.Sprintf("connected_addr=%s", normalizeToken(state.SessionConnectedAddr)),
		fmt.Sprintf("token_set=%t", strings.TrimSpace(state.SessionToken) != ""),
		fmt.Sprintf("token_source=%s", normalizeToken(state.SessionTokenSource)),
		fmt.Sprintf("manager_state=%s", normalizeToken(health.managerState)),
		fmt.Sprintf("transport_state=%s", normalizeToken(health.transportState)),
		fmt.Sprintf("reconnecting=%t", health.reconnecting),
		fmt.Sprintf("reconnect_count=%d", health.reconnectCount),
		fmt.Sprintf("diagnostics=%d", state.SessionDiagnosticsCount),
		fmt.Sprintf("repairs=%d", state.SessionRepairCount),
		fmt.Sprintf("last_repair=%s", normalizeToken(state.SessionLastRepairAction)),
		fmt.Sprintf("quick_fix=%s", normalizeToken(quickFix)),
	}
	return resultWithIntents(strings.Join(lines, "\n"), sessionDiagnosticsIntents(state, health, quickFix)...)
}

func repairSessionState(state *RuntimeState, action string) Result {
	if state == nil {
		return resultWithIntents("SESSION_REPAIR\naction=none\nchanged=false\nquick_fix=/session host\nrepairs=0", sessionRepairIntents("none", false, "/session host", 0)...)
	}
	changed := false
	applied := action
	if applied == "auto" {
		switch {
		case strings.TrimSpace(state.SessionToken) == "":
			applied = "token"
		case strings.TrimSpace(state.SessionMode) == "client" && !state.SessionConnected:
			applied = "disconnect"
		default:
			applied = "noop"
		}
	}
	quickFix := "/session status"
	switch applied {
	case "token":
		if strings.TrimSpace(state.SessionToken) == "" {
			seed := state.SessionHostCount + state.SessionConnectCount + 1
			token := fmt.Sprintf("repair-token-%d", seed)
			state.SessionToken = token
			state.SessionTokenPrefix = sessionTokenPrefix(token)
			state.SessionTokenSource = "repair"
			changed = true
		}
		quickFix = "/session token show"
	case "disconnect":
		if state.SessionConnected || strings.TrimSpace(state.SessionConnectedAddr) != "" {
			state.SessionConnected = false
			state.SessionConnectedAddr = ""
			changed = true
		}
		quickFix = "/session connect <addr> <token>"
	case "noop":
		quickFix = "/session diagnostics"
	default:
		quickFix = "/session repair"
	}
	state.SessionRepairCount++
	state.SessionLastRepairAction = applied
	message := fmt.Sprintf("SESSION_REPAIR\naction=%s\nchanged=%t\nquick_fix=%s\nrepairs=%d", normalizeToken(applied), changed, normalizeToken(quickFix), state.SessionRepairCount)
	return resultWithIntents(message, sessionRepairIntents(applied, changed, quickFix, state.SessionRepairCount)...)
}

type sessionHealth struct {
	managerState        string
	transportState      string
	reconnecting        bool
	reconnectReason     string
	reconnectErrorClass string
	reconnectCount      int
	reconnectDetail     string
}

func sessionHealthSummary(mgr *remote.SessionManager) sessionHealth {
	if mgr == nil {
		return sessionHealth{}
	}
	summary := mgr.HealthSummary()
	return sessionHealth{
		managerState:        string(summary.State),
		transportState:      string(summary.TransportState),
		reconnecting:        summary.Reconnecting,
		reconnectReason:     string(summary.LastCause.Reason),
		reconnectErrorClass: string(summary.LastCause.ErrorClass),
		reconnectCount:      max(summary.LastCause.ReconnectCount, summary.ReconnectCount),
		reconnectDetail:     summary.LastCause.Detail,
	}
}

func closeSessionRuntime(state *RuntimeState) error {
	if state == nil {
		return nil
	}
	var closeErr error
	if state.SessionManager != nil {
		if err := state.SessionManager.Close(); err != nil {
			closeErr = err
		}
	}
	if state.SessionTCPTransport != nil {
		if err := state.SessionTCPTransport.Shutdown(); err != nil && closeErr == nil {
			closeErr = err
		}
	}
	state.SessionManager = nil
	state.SessionTransport = nil
	state.SessionTCPTransport = nil
	return closeErr
}

func sessionIDOrDefault(state *RuntimeState) string {
	if state == nil {
		return "session-runtime"
	}
	sessionID := strings.TrimSpace(state.SessionID)
	if sessionID == "" {
		return "session-runtime"
	}
	return sessionID
}

// SkillsCommand provides deterministic skills listing and edits.
type SkillsCommand struct{}

func NewSkillsCommand() *SkillsCommand     { return &SkillsCommand{} }
func (c *SkillsCommand) Name() string      { return "skills" }
func (c *SkillsCommand) Aliases() []string { return nil }
func (c *SkillsCommand) Description() string {
	return "Inspect and repair skills from file/plugin sources"
}
func (c *SkillsCommand) Usage() string {
	return "/skills [list|status|doctor|repair [sync|auto|dedupe]|add <name>|remove <name>|sync]"
}

const skillsUsage = "usage: /skills [list|status|doctor|repair [sync|auto|dedupe]|add <name>|remove <name>|sync]"

func normalizedSkillMaps(state *RuntimeState) (map[string]string, map[string]string, map[string]bool) {
	sources := make(map[string]string, len(state.SkillsSources))
	origins := make(map[string]string, len(state.SkillsOrigins))
	enabled := make(map[string]bool, len(state.SkillsEnabled))
	for key, value := range state.SkillsSources {
		k := strings.TrimSpace(key)
		if k == "" {
			continue
		}
		sources[k] = normalizeToken(strings.TrimSpace(value))
	}
	for key, value := range state.SkillsOrigins {
		k := strings.TrimSpace(key)
		if k == "" {
			continue
		}
		origins[k] = normalizeToken(strings.TrimSpace(value))
	}
	for key, value := range state.SkillsEnabled {
		k := strings.TrimSpace(key)
		if k == "" {
			continue
		}
		enabled[k] = value
	}
	for _, raw := range state.Skills {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		if _, ok := sources[name]; !ok {
			sources[name] = "state"
		}
		if _, ok := origins[name]; !ok {
			origins[name] = "state"
		}
		if _, ok := enabled[name]; !ok {
			enabled[name] = true
		}
	}
	return sources, origins, enabled
}

func applySkillsSnapshot(state *RuntimeState, snapshot skillsSnapshot, source string) {
	state.Skills = uniqueSortedStrings(append([]string(nil), snapshot.Names...))
	state.SkillsSources = cloneSkillStringMap(snapshot.Sources)
	state.SkillsOrigins = cloneSkillStringMap(snapshot.Origins)
	state.SkillsEnabled = cloneSkillBoolMap(snapshot.Enabled)
	state.SkillsConflictCount = len(snapshot.Conflicts)
	state.SkillsLastSyncSource = source
}

func (c *SkillsCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "list") || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf(skillsUsage)
		}
		cmdCtx.State.SkillsViewCount++
		skills := uniqueSortedStrings(append([]string(nil), cmdCtx.State.Skills...))
		sources, origins, enabled := normalizedSkillMaps(cmdCtx.State)
		enabledCount := 0
		disabledCount := 0
		lines := []string{
			"SKILLS_LIST",
			fmt.Sprintf("count=%d", len(skills)),
			fmt.Sprintf("views=%d", cmdCtx.State.SkillsViewCount),
			"enabled=0",
			"disabled=0",
			fmt.Sprintf("conflicts=%d", cmdCtx.State.SkillsConflictCount),
			fmt.Sprintf("last_source=%s", normalizeToken(cmdCtx.State.SkillsLastSyncSource)),
		}
		for i, skill := range skills {
			if enabled[skill] {
				enabledCount++
			} else {
				disabledCount++
			}
			lines = append(lines,
				fmt.Sprintf("skill.%d.name=%s", i+1, normalizeToken(skill)),
				fmt.Sprintf("skill.%d.state=%s", i+1, boolState(enabled[skill], "enabled", "disabled")),
				fmt.Sprintf("skill.%d.source=%s", i+1, normalizeToken(sources[skill])),
				fmt.Sprintf("skill.%d.origin=%s", i+1, normalizeToken(origins[skill])),
			)
		}
		lines[3] = fmt.Sprintf("enabled=%d", enabledCount)
		lines[4] = fmt.Sprintf("disabled=%d", disabledCount)
		return resultWithIntents(strings.Join(lines, "\n"), skillsListIntents(skills, cmdCtx.State.SkillsViewCount, sources, origins, enabled, cmdCtx.State.SkillsConflictCount)...), nil
	}

	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "sync") {
		before := strings.Join(uniqueSortedStrings(append([]string(nil), cmdCtx.State.Skills...)), "|")
		snapshot, source := syncedSkillsFromRuntime(cmdCtx.State.Skills)
		applySkillsSnapshot(cmdCtx.State, snapshot, source)
		cmdCtx.State.SkillsSyncCount++
		after := strings.Join(cmdCtx.State.Skills, "|")
		changed := before != after
		message := fmt.Sprintf("SKILLS_SYNC\ncount=%d\nchanged=%t\nsource=%s\nsync_count=%d\nconflicts=%d", len(cmdCtx.State.Skills), changed, normalizeToken(source), cmdCtx.State.SkillsSyncCount, cmdCtx.State.SkillsConflictCount)
		return resultWithIntents(message, skillsRepairIntents("sync", changed, "/skills status", len(cmdCtx.State.Skills))...), nil
	}
	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "doctor") {
		cmdCtx.State.SkillsDoctorCount++
		sources, origins, enabled := normalizedSkillMaps(cmdCtx.State)
		enabledCount := 0
		disabledCount := 0
		for _, name := range cmdCtx.State.Skills {
			trimmed := strings.TrimSpace(name)
			if trimmed == "" {
				continue
			}
			if enabled[trimmed] {
				enabledCount++
			} else {
				disabledCount++
			}
		}
		quickFix := "/skills sync"
		if len(cmdCtx.State.Skills) == 0 {
			quickFix = "/skills add <name>"
		} else if cmdCtx.State.SkillsConflictCount > 0 {
			quickFix = "/skills sync"
		}
		message := fmt.Sprintf("SKILLS_DOCTOR\ncount=%d\nenabled=%d\ndisabled=%d\nviews=%d\nsync_count=%d\ndoctor_count=%d\nconflicts=%d\nlast_source=%s\nquick_fix=%s", len(cmdCtx.State.Skills), enabledCount, disabledCount, cmdCtx.State.SkillsViewCount, cmdCtx.State.SkillsSyncCount, cmdCtx.State.SkillsDoctorCount, cmdCtx.State.SkillsConflictCount, normalizeToken(cmdCtx.State.SkillsLastSyncSource), normalizeToken(quickFix))
		return resultWithIntents(message, skillsDoctorIntents(len(cmdCtx.State.Skills), enabledCount, disabledCount, cmdCtx.State.SkillsConflictCount, cmdCtx.State.SkillsViewCount, cmdCtx.State.SkillsSyncCount, cmdCtx.State.SkillsDoctorCount, cmdCtx.State.SkillsLastSyncSource, quickFix, sources, origins)...), nil
	}
	if len(inv.Args) >= 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "repair") {
		if len(inv.Args) > 2 {
			return Result{}, fmt.Errorf(skillsUsage)
		}
		mode := "sync"
		if len(inv.Args) == 2 {
			mode = strings.ToLower(strings.TrimSpace(inv.Args[1]))
			if mode == "" {
				return Result{}, fmt.Errorf(skillsUsage)
			}
		}
		changed := false
		quickFix := "/skills doctor"
		switch mode {
		case "sync", "auto":
			before := strings.Join(uniqueSortedStrings(append([]string(nil), cmdCtx.State.Skills...)), "|")
			snapshot, source := syncedSkillsFromRuntime(cmdCtx.State.Skills)
			applySkillsSnapshot(cmdCtx.State, snapshot, source)
			after := strings.Join(cmdCtx.State.Skills, "|")
			changed = before != after
			quickFix = "/skills status"
		case "dedupe":
			before := len(cmdCtx.State.Skills)
			cmdCtx.State.Skills = uniqueSortedStrings(cmdCtx.State.Skills)
			changed = len(cmdCtx.State.Skills) != before
			quickFix = "/skills list"
		default:
			quickFix = "/skills repair sync"
		}
		message := fmt.Sprintf("SKILLS_REPAIR\nmode=%s\nchanged=%t\nquick_fix=%s\ncount=%d\nconflicts=%d", normalizeToken(mode), changed, normalizeToken(quickFix), len(cmdCtx.State.Skills), cmdCtx.State.SkillsConflictCount)
		return resultWithIntents(message, skillsRepairIntents(mode, changed, quickFix, len(cmdCtx.State.Skills))...), nil
	}
	if len(inv.Args) < 2 {
		return Result{}, fmt.Errorf(skillsUsage)
	}
	action := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	name := strings.TrimSpace(strings.Join(inv.Args[1:], " "))
	if name == "" {
		return Result{}, fmt.Errorf("usage: /skills [add <name>|remove <name>]")
	}
	switch action {
	case "add":
		before := len(cmdCtx.State.Skills)
		cmdCtx.State.Skills = uniqueSortedStrings(append(cmdCtx.State.Skills, name))
		added := len(cmdCtx.State.Skills) > before
		if cmdCtx.State.SkillsSources == nil {
			cmdCtx.State.SkillsSources = map[string]string{}
		}
		if cmdCtx.State.SkillsOrigins == nil {
			cmdCtx.State.SkillsOrigins = map[string]string{}
		}
		if cmdCtx.State.SkillsEnabled == nil {
			cmdCtx.State.SkillsEnabled = map[string]bool{}
		}
		cmdCtx.State.SkillsSources[name] = "state"
		cmdCtx.State.SkillsOrigins[name] = "manual"
		cmdCtx.State.SkillsEnabled[name] = true
		message := fmt.Sprintf("SKILLS_ADD\nname=%s\nadded=%t\ncount=%d\nsource=state\norigin=manual\nstate=enabled", normalizeToken(name), added, len(cmdCtx.State.Skills))
		return resultWithIntents(message, skillsMutationIntents("Skill added", name, len(cmdCtx.State.Skills), added)...), nil
	case "remove":
		removed := false
		next := make([]string, 0, len(cmdCtx.State.Skills))
		for _, s := range cmdCtx.State.Skills {
			if strings.TrimSpace(s) == name {
				removed = true
				continue
			}
			next = append(next, s)
		}
		cmdCtx.State.Skills = uniqueSortedStrings(next)
		if cmdCtx.State.SkillsSources != nil {
			delete(cmdCtx.State.SkillsSources, name)
		}
		if cmdCtx.State.SkillsOrigins != nil {
			delete(cmdCtx.State.SkillsOrigins, name)
		}
		if cmdCtx.State.SkillsEnabled != nil {
			delete(cmdCtx.State.SkillsEnabled, name)
		}
		message := fmt.Sprintf("SKILLS_REMOVE\nname=%s\nremoved=%t\ncount=%d", normalizeToken(name), removed, len(cmdCtx.State.Skills))
		return resultWithIntents(message, skillsMutationIntents("Skill removed", name, len(cmdCtx.State.Skills), removed)...), nil
	default:
		return Result{}, fmt.Errorf(skillsUsage)
	}
}

// RewindCommand provides deterministic rewind intent.
type RewindCommand struct{}

func NewRewindCommand() *RewindCommand     { return &RewindCommand{} }
func (c *RewindCommand) Name() string      { return "rewind" }
func (c *RewindCommand) Aliases() []string { return []string{"checkpoint"} }
func (c *RewindCommand) Description() string {
	return "Restore code and/or conversation to previous point"
}
func (c *RewindCommand) Usage() string { return "/rewind [status|<target>]" }

func (c *RewindCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		message := fmt.Sprintf("REWIND_STATUS\ncount=%d\nlast_target=%s", cmdCtx.State.RewindCount, normalizeToken(cmdCtx.State.LastRewindTarget))
		return resultWithIntents(message, rewindStatusIntents(cmdCtx.State.RewindCount, cmdCtx.State.LastRewindTarget)...), nil
	}
	target := "latest"
	if len(inv.Args) > 0 {
		target = strings.TrimSpace(strings.Join(inv.Args, " "))
		if target == "" {
			return Result{}, fmt.Errorf("usage: /rewind [status|<target>]")
		}
	}
	cmdCtx.State.RewindCount++
	cmdCtx.State.LastRewindTarget = target
	cmdCtx.State.CompactRequested = false
	cmdCtx.State.ResumeRequested = false
	message := fmt.Sprintf("REWIND_REQUEST\ntarget=%s\ncount=%d\nrequested=true", normalizeToken(target), cmdCtx.State.RewindCount)
	return resultWithIntents(message, rewindRequestIntents(target, cmdCtx.State.RewindCount)...), nil
}

// TagCommand provides deterministic session tag toggling.
type TagCommand struct{}

func NewTagCommand() *TagCommand        { return &TagCommand{} }
func (c *TagCommand) Name() string      { return "tag" }
func (c *TagCommand) Aliases() []string { return nil }
func (c *TagCommand) Description() string {
	return "Toggle a searchable tag on current session"
}
func (c *TagCommand) Usage() string { return "/tag [status|<tag-name>]" }

func (c *TagCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || (len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status")) {
		message := fmt.Sprintf("TAG_STATUS\ncurrent=%s\nupdates=%d", normalizeToken(cmdCtx.State.CurrentTag), cmdCtx.State.TagUpdates)
		return resultWithIntents(message, tagStatusIntents(cmdCtx.State.CurrentTag, cmdCtx.State.TagUpdates)...), nil
	}
	name := strings.TrimSpace(strings.Join(inv.Args, " "))
	if name == "" {
		return Result{}, fmt.Errorf("usage: /tag [status|<tag-name>]")
	}
	if strings.TrimSpace(cmdCtx.State.CurrentTag) == name {
		cmdCtx.State.CurrentTag = ""
		cmdCtx.State.TagUpdates++
		message := fmt.Sprintf("TAG_REMOVE\ntag=%s\nupdates=%d", normalizeToken(name), cmdCtx.State.TagUpdates)
		return resultWithIntents(message, tagMutationIntents("Tag removed", name, cmdCtx.State.TagUpdates)...), nil
	}
	cmdCtx.State.CurrentTag = name
	cmdCtx.State.TagUpdates++
	message := fmt.Sprintf("TAG_SET\ntag=%s\nupdates=%d", normalizeToken(name), cmdCtx.State.TagUpdates)
	return resultWithIntents(message, tagMutationIntents("Tag set", name, cmdCtx.State.TagUpdates)...), nil
}

// RemoteEnvCommand provides deterministic remote environment configuration.
type RemoteEnvCommand struct{}

func NewRemoteEnvCommand() *RemoteEnvCommand    { return &RemoteEnvCommand{} }
func (c *RemoteEnvCommand) Name() string        { return "remote-env" }
func (c *RemoteEnvCommand) Aliases() []string   { return nil }
func (c *RemoteEnvCommand) Description() string { return "Configure default remote environment" }
func (c *RemoteEnvCommand) Usage() string       { return "/remote-env [status|list|set <name>]" }

func (c *RemoteEnvCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /remote-env [status|list|set <name>]")
		}
		env := normalizeToken(cmdCtx.State.RemoteEnvironment)
		if env == "-" {
			env = "default"
		}
		message := fmt.Sprintf("REMOTE_ENV_STATUS\nenvironment=%s\nupdates=%d", env, cmdCtx.State.RemoteEnvUpdates)
		return resultWithIntents(message, remoteEnvStatusIntents(env, cmdCtx.State.RemoteEnvUpdates)...), nil
	}
	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "list") {
		message := "REMOTE_ENV_LIST\ncount=3\nenv.1=default\nenv.2=code-review\nenv.3=hardened-linux"
		return resultWithIntents(message, remoteEnvListIntents()...), nil
	}
	if len(inv.Args) < 2 || !strings.EqualFold(strings.TrimSpace(inv.Args[0]), "set") {
		return Result{}, fmt.Errorf("usage: /remote-env [status|list|set <name>]")
	}
	env := strings.TrimSpace(strings.Join(inv.Args[1:], " "))
	if env == "" {
		return Result{}, fmt.Errorf("usage: /remote-env [status|list|set <name>]")
	}
	cmdCtx.State.RemoteEnvironment = env
	cmdCtx.State.RemoteEnvUpdates++
	message := fmt.Sprintf("REMOTE_ENV_SET\nenvironment=%s\nupdates=%d", normalizeToken(env), cmdCtx.State.RemoteEnvUpdates)
	return resultWithIntents(message, remoteEnvSetIntents(env, cmdCtx.State.RemoteEnvUpdates)...), nil
}

// SecurityReviewCommand provides deterministic security review intent.
type SecurityReviewCommand struct{}

func NewSecurityReviewCommand() *SecurityReviewCommand { return &SecurityReviewCommand{} }
func (c *SecurityReviewCommand) Name() string          { return "security-review" }
func (c *SecurityReviewCommand) Aliases() []string     { return nil }
func (c *SecurityReviewCommand) Description() string {
	return "Complete a security review of pending branch changes"
}
func (c *SecurityReviewCommand) Usage() string { return "/security-review [status|<target>]" }

func (c *SecurityReviewCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		message := fmt.Sprintf("SECURITY_REVIEW_STATUS\ncount=%d\nlast_target=%s", cmdCtx.State.SecurityReviewCount, normalizeToken(cmdCtx.State.LastSecurityTarget))
		return resultWithIntents(message, securityReviewStatusIntents(cmdCtx.State.SecurityReviewCount, cmdCtx.State.LastSecurityTarget)...), nil
	}
	target := "current-branch"
	if len(inv.Args) > 0 {
		target = strings.TrimSpace(strings.Join(inv.Args, " "))
		if target == "" {
			return Result{}, fmt.Errorf("usage: /security-review [status|<target>]")
		}
	}
	cmdCtx.State.SecurityReviewCount++
	cmdCtx.State.LastSecurityTarget = target
	message := fmt.Sprintf("SECURITY_REVIEW_REQUEST\ntarget=%s\ncount=%d\nmode=focused", normalizeToken(target), cmdCtx.State.SecurityReviewCount)
	return resultWithIntents(message, securityReviewRequestIntents(target, cmdCtx.State.SecurityReviewCount)...), nil
}

// AddDirCommand provides stable add-dir parsing and messages.
type AddDirCommand struct{}

func NewAddDirCommand() *AddDirCommand       { return &AddDirCommand{} }
func (c *AddDirCommand) Name() string        { return "add-dir" }
func (c *AddDirCommand) Aliases() []string   { return nil }
func (c *AddDirCommand) Description() string { return "Add a workspace directory to current session" }
func (c *AddDirCommand) Usage() string       { return "/add-dir <path>" }

func (c *AddDirCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if len(inv.Args) != 1 {
		return Result{}, fmt.Errorf("usage: /add-dir <path>")
	}

	raw := strings.TrimSpace(inv.Args[0])
	if raw == "" {
		return Result{}, fmt.Errorf("usage: /add-dir <path>")
	}

	abs, err := filepath.Abs(raw)
	if err != nil {
		return Result{}, fmt.Errorf("resolving path: %w", err)
	}
	clean := filepath.Clean(abs)
	info, err := os.Stat(clean)
	if err != nil {
		if os.IsNotExist(err) {
			return Result{}, fmt.Errorf("directory does not exist: %s", clean)
		}
		return Result{}, fmt.Errorf("checking path: %w", err)
	}
	if !info.IsDir() {
		return Result{}, fmt.Errorf("path is not a directory: %s", clean)
	}

	if cmdCtx.State != nil {
		cmdCtx.State.WorkspaceDirs = append(cmdCtx.State.WorkspaceDirs, clean)
		cmdCtx.State.WorkspaceDirs = uniqueSortedStrings(cmdCtx.State.WorkspaceDirs)
	}
	message := fmt.Sprintf("Workspace directory added: %s", clean)
	return resultWithIntents(message, addDirIntents(clean)...), nil
}

// AgentsCommand provides stable agents subcommand parsing and messages.
type AgentsCommand struct{}

func NewAgentsCommand() *AgentsCommand       { return &AgentsCommand{} }
func (c *AgentsCommand) Name() string        { return "agents" }
func (c *AgentsCommand) Aliases() []string   { return nil }
func (c *AgentsCommand) Description() string { return "Manage agent profiles: list|create|status" }
func (c *AgentsCommand) Usage() string       { return "/agents [list|create <name>|status]" }

func (c *AgentsCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "list") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /agents [list|create <name>|status]")
		}
		message := "AGENTS_LIST\ncount=1\nagent.1.name=local\nagent.1.status=available"
		return resultWithIntents(message, agentsListIntents()...), nil
	}

	switch strings.ToLower(strings.TrimSpace(inv.Args[0])) {
	case "create":
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf("usage: /agents [list|create <name>|status]")
		}
		message := fmt.Sprintf("AGENTS_CREATE\nname=%s\nstatus=created", normalizeToken(inv.Args[1]))
		return resultWithIntents(message, agentsCreateIntents(inv.Args[1])...), nil
	case "status":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /agents [list|create <name>|status]")
		}
		if cmdCtx.State != nil && cmdCtx.State.Agent != nil {
			message := "AGENTS_STATUS\nstatus=configured"
			return resultWithIntents(message, agentsStatusIntents(true)...), nil
		}
		message := "AGENTS_STATUS\nstatus=none"
		return resultWithIntents(message, agentsStatusIntents(false)...), nil
	default:
		return Result{}, fmt.Errorf("usage: /agents [list|create <name>|status]")
	}
}

// ClearCommand provides stable clear command parsing and messages.
type ClearCommand struct{}

func NewClearCommand() *ClearCommand      { return &ClearCommand{} }
func (c *ClearCommand) Name() string      { return "clear" }
func (c *ClearCommand) Aliases() []string { return []string{"reset", "new"} }
func (c *ClearCommand) Description() string {
	return "Clear local conversation display for current session"
}
func (c *ClearCommand) Usage() string { return "/clear [all|context|display|status]" }

func (c *ClearCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) > 1 {
		return Result{}, fmt.Errorf("usage: /clear [all|context|display|status]")
	}
	scope := "all"
	if len(inv.Args) == 1 {
		scope = strings.ToLower(strings.TrimSpace(inv.Args[0]))
	}

	if scope == "status" {
		message := fmt.Sprintf("CLEAR_STATUS\ncount=%d\ncontext_clears=%d\ndiff_clears=%d\nlast_scope=%s\ncompact_requested=%t\nresume_requested=%t\nlast_resume_target=%s", cmdCtx.State.ClearCount, cmdCtx.State.ClearContextCount, cmdCtx.State.ClearDiffCount, normalizeToken(cmdCtx.State.LastClearScope), cmdCtx.State.CompactRequested, cmdCtx.State.ResumeRequested, normalizeToken(cmdCtx.State.LastResumeTarget))
		return resultWithIntents(message, clearStatusIntents(cmdCtx.State)...), nil
	}

	clearDisplay := false
	clearContext := false
	clearDiff := false
	switch scope {
	case "all":
		clearDisplay = true
		clearContext = true
		clearDiff = true
	case "display":
		clearDisplay = true
	case "context":
		clearContext = true
	default:
		return Result{}, fmt.Errorf("usage: /clear [all|context|display|status]")
	}

	if clearContext {
		cmdCtx.State.CompactRequested = false
		cmdCtx.State.ResumeRequested = false
		cmdCtx.State.LastResumeTarget = ""
		cmdCtx.State.LastCompactTarget = ""
		cmdCtx.State.ClearContextCount++
	}
	if clearDiff {
		cmdCtx.State.DiffEntries = nil
		cmdCtx.State.ClearDiffCount++
	}
	if clearDisplay {
		cmdCtx.State.ClearCount++
	}
	cmdCtx.State.LastClearScope = scope

	message := fmt.Sprintf("CLEAR_RESULT\nscope=%s\nclear_display=%t\nclear_context=%t\nclear_diff=%t\ncount=%d\ncontext_clears=%d\ndiff_clears=%d", scope, clearDisplay, clearContext, clearDiff, cmdCtx.State.ClearCount, cmdCtx.State.ClearContextCount, cmdCtx.State.ClearDiffCount)
	return resultWithIntents(message, clearResultIntents(scope, clearDisplay, clearContext, clearDiff, cmdCtx.State.ClearCount, cmdCtx.State.ClearContextCount, cmdCtx.State.ClearDiffCount)...), nil
}

// HistoryCommand provides stable history subcommand parsing and messages.
type HistoryCommand struct{}

func NewHistoryCommand() *HistoryCommand      { return &HistoryCommand{} }
func (c *HistoryCommand) Name() string        { return "history" }
func (c *HistoryCommand) Aliases() []string   { return nil }
func (c *HistoryCommand) Description() string { return "Manage session history: list|show" }
func (c *HistoryCommand) Usage() string {
	return "/history [list [model <model>|text <query>|limit <n>|latest]|latest|show <id>]"
}

const historyUsage = "usage: /history [list [model <model>|text <query>|limit <n>|latest]|latest|show <id>]"

func (c *HistoryCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	entries := defaultHistoryEntries(cmdCtx.State)
	if cmdCtx.State != nil {
		cmdCtx.State.HistoryViews++
	}

	if len(inv.Args) == 1 && strings.EqualFold(inv.Args[0], "status") {
		return renderHistoryStatus(cmdCtx.State, entries), nil
	}
	if len(inv.Args) == 1 && strings.EqualFold(inv.Args[0], "doctor") {
		return renderHistoryDoctor(cmdCtx.State, entries), nil
	}

	if len(inv.Args) == 1 && strings.EqualFold(inv.Args[0], "latest") {
		if len(entries) == 0 {
			return resultWithIntents("HISTORY_LATEST\nfound=false", historyLatestIntents(HistoryEntry{}, false)...), nil
		}
		entry := entries[len(entries)-1]
		message := fmt.Sprintf("HISTORY_LATEST\nfound=true\nid=%s\npath=%s\ncreated=%s\nmodel=%s\nturns=%d\ntitle=%s", normalizeToken(entry.ID), normalizeToken(entry.Path), normalizeToken(entry.CreatedAt), normalizeToken(entry.Model), entry.Turns, normalizeToken(entry.Title))
		return resultWithIntents(message, historyLatestIntents(entry, true)...), nil
	}

	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "list") {
		filterType, filterValue, limit, err := parseHistoryListFilter(inv.Args)
		if err != nil {
			return Result{}, err
		}
		if cmdCtx.State != nil {
			cmdCtx.State.HistoryLastFilterType = filterType
			cmdCtx.State.HistoryLastFilterValue = filterValue
			cmdCtx.State.HistoryLastLimit = limit
		}
		entries = filterHistoryEntries(entries, filterType, filterValue)
		if limit > 0 && len(entries) > limit {
			entries = entries[len(entries)-limit:]
		}
		if len(entries) == 0 {
			message := fmt.Sprintf("HISTORY_LIST\nfilter.type=%s\nfilter.value=%s\nlimit=%d\ncount=0", normalizeToken(filterType), normalizeToken(filterValue), limit)
			return resultWithIntents(message, historyListIntents(filterType, filterValue, limit, nil)...), nil
		}
		lines := []string{
			"HISTORY_LIST",
			fmt.Sprintf("filter.type=%s", normalizeToken(filterType)),
			fmt.Sprintf("filter.value=%s", normalizeToken(filterValue)),
			fmt.Sprintf("limit=%d", limit),
			fmt.Sprintf("count=%d", len(entries)),
		}
		for i, entry := range entries {
			idx := i + 1
			lines = append(lines, fmt.Sprintf("entry.%d.id=%s", idx, normalizeToken(entry.ID)))
			lines = append(lines, fmt.Sprintf("entry.%d.path=%s", idx, normalizeToken(entry.Path)))
			lines = append(lines, fmt.Sprintf("entry.%d.created=%s", idx, normalizeToken(entry.CreatedAt)))
			lines = append(lines, fmt.Sprintf("entry.%d.model=%s", idx, normalizeToken(entry.Model)))
			lines = append(lines, fmt.Sprintf("entry.%d.turns=%d", idx, entry.Turns))
			lines = append(lines, fmt.Sprintf("entry.%d.title=%s", idx, normalizeToken(entry.Title)))
		}
		return resultWithIntents(strings.Join(lines, "\n"), historyListIntents(filterType, filterValue, limit, entries)...), nil
	}

	if strings.EqualFold(inv.Args[0], "show") {
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf(historyUsage)
		}
		id := strings.TrimSpace(inv.Args[1])
		if id == "" {
			return Result{}, fmt.Errorf(historyUsage)
		}
		for _, entry := range entries {
			if entry.ID == id {
				lines := []string{
					"HISTORY_SHOW",
					fmt.Sprintf("id=%s", normalizeToken(entry.ID)),
					"section_count=2",
					"section.1.name=meta",
					fmt.Sprintf("section.1.path=%s", normalizeToken(entry.Path)),
					fmt.Sprintf("section.1.created=%s", normalizeToken(entry.CreatedAt)),
					fmt.Sprintf("section.1.model=%s", normalizeToken(entry.Model)),
					fmt.Sprintf("section.1.turns=%d", entry.Turns),
					"section.2.name=content",
					fmt.Sprintf("section.2.title=%s", normalizeToken(entry.Title)),
					fmt.Sprintf("section.2.summary=%s", normalizeToken(entry.Summary)),
				}
				return resultWithIntents(strings.Join(lines, "\n"), historyShowIntents(entry)...), nil
			}
		}
		return Result{}, fmt.Errorf("history entry not found: %s", id)
	}

	return Result{}, fmt.Errorf(historyUsage)
}

func parseHistoryListFilter(args []string) (string, string, int, error) {
	limit := 0
	if len(args) <= 1 {
		return "none", "-", limit, nil
	}
	segments := args[1:]
	if len(segments) == 1 && strings.EqualFold(strings.TrimSpace(segments[0]), "latest") {
		return "none", "-", 1, nil
	}
	if len(segments) < 2 {
		return "", "", 0, fmt.Errorf(historyUsage)
	}
	filterType := "none"
	filterValue := "-"
	compound := []string{}
	for i := 0; i < len(segments); {
		kind := strings.ToLower(strings.TrimSpace(segments[i]))
		i++
		if kind == "latest" {
			limit = 1
			continue
		}
		if i >= len(segments) {
			return "", "", 0, fmt.Errorf(historyUsage)
		}
		value := strings.TrimSpace(segments[i])
		i++
		if value == "" {
			return "", "", 0, fmt.Errorf(historyUsage)
		}
		switch kind {
		case "model", "text":
			if filterType == "none" {
				filterType = kind
				filterValue = value
			} else {
				compound = append(compound, kind+":"+value)
			}
		case "limit":
			v, err := parseNonNegativeInt(value)
			if err != nil {
				return "", "", 0, fmt.Errorf(historyUsage)
			}
			limit = v
		default:
			return "", "", 0, fmt.Errorf(historyUsage)
		}
	}
	if len(compound) > 0 {
		base := filterType + ":" + filterValue
		all := append([]string{base}, compound...)
		filterType = "compound"
		filterValue = strings.Join(all, ",")
	}
	return filterType, filterValue, limit, nil
}

func filterHistoryEntries(entries []HistoryEntry, filterType, filterValue string) []HistoryEntry {
	if filterType == "compound" {
		rules := strings.Split(filterValue, ",")
		filtered := append([]HistoryEntry(nil), entries...)
		for _, rule := range rules {
			parts := strings.SplitN(strings.TrimSpace(rule), ":", 2)
			if len(parts) != 2 {
				continue
			}
			filtered = filterHistoryEntries(filtered, parts[0], parts[1])
		}
		return filtered
	}
	if filterType == "none" {
		return entries
	}
	needle := strings.ToLower(strings.TrimSpace(filterValue))
	if needle == "" {
		return entries
	}
	filtered := make([]HistoryEntry, 0, len(entries))
	for _, entry := range entries {
		switch filterType {
		case "model":
			if strings.EqualFold(strings.TrimSpace(entry.Model), filterValue) {
				filtered = append(filtered, entry)
			}
		case "text":
			haystack := strings.ToLower(strings.Join([]string{entry.ID, entry.Title, entry.Summary, entry.Model}, " "))
			if strings.Contains(haystack, needle) {
				filtered = append(filtered, entry)
			}
		}
	}
	return filtered
}

func defaultHistoryEntries(state *RuntimeState) []HistoryEntry {
	if state == nil || len(state.HistoryEntries) == 0 {
		return nil
	}
	entries := append([]HistoryEntry(nil), state.HistoryEntries...)
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].CreatedAt == entries[j].CreatedAt {
			return entries[i].ID < entries[j].ID
		}
		return entries[i].CreatedAt < entries[j].CreatedAt
	})
	return entries
}

func renderHistoryStatus(state *RuntimeState, entries []HistoryEntry) Result {
	if state == nil {
		return resultWithIntents("HISTORY_STATUS\ncount=0\nviews=0\nlast_filter_type=-\nlast_filter_value=-\nlast_limit=0\nmodels=0", historyStatusIntents(&RuntimeState{}, 0, 0)...)
	}
	modelSet := make(map[string]struct{})
	for _, entry := range entries {
		model := strings.TrimSpace(entry.Model)
		if model == "" {
			continue
		}
		modelSet[model] = struct{}{}
	}
	message := fmt.Sprintf("HISTORY_STATUS\ncount=%d\nviews=%d\nlast_filter_type=%s\nlast_filter_value=%s\nlast_limit=%d\nmodels=%d", len(entries), state.HistoryViews, normalizeToken(state.HistoryLastFilterType), normalizeToken(state.HistoryLastFilterValue), state.HistoryLastLimit, len(modelSet))
	return resultWithIntents(message, historyStatusIntents(state, len(entries), len(modelSet))...)
}

func renderHistoryDoctor(state *RuntimeState, entries []HistoryEntry) Result {
	status := renderHistoryStatus(state, entries)
	quickFix := "/history list latest"
	if len(entries) == 0 {
		quickFix = "/history list"
	}
	lines := []string{"HISTORY_DOCTOR", strings.TrimSpace(strings.TrimPrefix(status.Message, "HISTORY_STATUS\n")), fmt.Sprintf("quick_fix=%s", normalizeToken(quickFix))}
	return resultWithIntents(strings.Join(lines, "\n"), historyDoctorIntents(len(entries), quickFix)...)
}

func normalizeToken(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "-"
	}
	v = strings.ReplaceAll(v, "\n", " ")
	v = strings.ReplaceAll(v, "\r", " ")
	v = strings.Join(strings.Fields(v), " ")
	return v
}

func sessionTokenPrefix(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}
	token = strings.Join(strings.Fields(token), "")
	runes := []rune(token)
	prefixLen := 6
	if len(runes) < prefixLen {
		prefixLen = len(runes)
	}
	if prefixLen == len(runes) && len(runes) > 1 {
		prefixLen = len(runes) - 1
	}
	if prefixLen < 1 {
		prefixLen = 1
	}
	return string(runes[:prefixLen]) + "..."
}

func uniqueSortedStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	uniq := make(map[string]struct{}, len(in))
	for _, v := range in {
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			continue
		}
		uniq[trimmed] = struct{}{}
	}
	out := make([]string, 0, len(uniq))
	for k := range uniq {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func ensureBranchState(state *RuntimeState) {
	if len(state.Branches) == 0 {
		state.Branches = []string{"main"}
	}
	state.Branches = uniqueSortedStrings(state.Branches)
	if strings.TrimSpace(state.ActiveBranch) == "" || !containsString(state.Branches, state.ActiveBranch) {
		state.ActiveBranch = state.Branches[0]
	}
}

func containsString(values []string, want string) bool {
	for _, v := range values {
		if strings.TrimSpace(v) == strings.TrimSpace(want) {
			return true
		}
	}
	return false
}

func renderDiffList(state *RuntimeState) Result {
	entries := append([]DiffEntry(nil), state.DiffEntries...)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})
	lines := []string{
		"DIFF_LIST",
		fmt.Sprintf("mode=%s", normalizeToken(state.DiffMode)),
		fmt.Sprintf("count=%d", len(entries)),
	}
	for i, entry := range entries {
		idx := i + 1
		lines = append(lines, fmt.Sprintf("entry.%d.path=%s", idx, normalizeToken(entry.Path)))
		lines = append(lines, fmt.Sprintf("entry.%d.added=%d", idx, entry.Added))
		lines = append(lines, fmt.Sprintf("entry.%d.removed=%d", idx, entry.Removed))
		lines = append(lines, fmt.Sprintf("entry.%d.modified=%d", idx, entry.Modified))
	}
	rows := make([]types.RenderTableRow, 0, len(state.PluginMarketplaces))
	for _, value := range state.PluginMarketplaces {
		rows = append(rows, tableRow(value))
	}
	return resultWithIntents(strings.Join(lines, "\n"), tableIntent("Plugin marketplaces", "Configured marketplace endpoints.", []string{"URL"}, rows...))
}

func parseNonNegativeInt(raw string) (int, error) {
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || v < 0 {
		return 0, fmt.Errorf("expected non-negative integer")
	}
	return v, nil
}

func upsertDiffEntry(state *RuntimeState, next DiffEntry) {
	for i, existing := range state.DiffEntries {
		if strings.TrimSpace(existing.Path) == strings.TrimSpace(next.Path) {
			state.DiffEntries[i] = next
			return
		}
	}
	state.DiffEntries = append(state.DiffEntries, next)
}

// MCPCommand provides stable mcp subcommand parsing and messages.
type MCPCommand struct{}

func NewMCPCommand() *MCPCommand        { return &MCPCommand{} }
func (c *MCPCommand) Name() string      { return "mcp" }
func (c *MCPCommand) Aliases() []string { return nil }
func (c *MCPCommand) Description() string {
	return "Manage MCP servers and manager-backed MCP state"
}
func (c *MCPCommand) Usage() string {
	return "/mcp [list|doctor|diagnostics|repair [mode]|enable [server-name]|disable [server-name]|reconnect <server-name>|no-redirect|add <name> [transport] <command-or-url> [args...]|remove <name>|status [name]|list-tools [name]|list-resources [name]|auth-status [name]|connect <name>|disconnect <name>]"
}

const mcpUsage = "usage: /mcp [list|doctor|diagnostics|repair [mode]|enable [server-name]|disable [server-name]|reconnect <server-name>|no-redirect|add <name> [transport] <command-or-url> [args...]|remove <name>|status [name]|list-tools [name]|list-resources [name]|auth-status [name]|connect <name>|disconnect <name>]"

func (c *MCPCommand) Execute(ctx context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}

	if cmdCtx.State.MCPConnections == nil {
		cmdCtx.State.MCPConnections = make(map[string]bool)
	}

	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "list") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf(mcpUsage)
		}
		if cmdCtx.MCPManager != nil {
			statuses := cmdCtx.MCPManager.ServerStatuses()
			lines := []string{"MCP_LIST", fmt.Sprintf("count=%d", len(statuses))}
			for i, status := range statuses {
				idx := i + 1
				lines = append(lines, fmt.Sprintf("server.%d.name=%s", idx, normalizeToken(status.ServerName)))
				lines = append(lines, fmt.Sprintf("server.%d.transport=%s", idx, normalizeToken(string(status.TransportType))))
				lines = append(lines, fmt.Sprintf("server.%d.connection_state=%s", idx, normalizeToken(string(status.ConnectionState))))
				lines = append(lines, fmt.Sprintf("server.%d.auth_status=%s", idx, normalizeToken(string(status.AuthStatus))))
				lines = append(lines, fmt.Sprintf("server.%d.authenticated=%t", idx, status.Authenticated))
			}
			rows := make([]types.RenderTableRow, 0, len(statuses))
			for _, status := range statuses {
				rows = append(rows, tableRow(status.ServerName, string(status.TransportType), string(status.AuthStatus), boolState(status.Authenticated, "yes", "no")))
			}
			return resultWithIntents(strings.Join(lines, "\n"), mcpListIntents(rows)...), nil
		}

		names := make([]string, 0, len(cmdCtx.State.MCPConnections))
		for name := range cmdCtx.State.MCPConnections {
			names = append(names, name)
		}
		sort.Strings(names)
		lines := []string{"MCP_LIST", fmt.Sprintf("count=%d", len(names))}
		for i, name := range names {
			idx := i + 1
			connected := cmdCtx.State.MCPConnections[name]
			status := "disconnected"
			if connected {
				status = "connected"
			}
			lines = append(lines, fmt.Sprintf("server.%d.name=%s", idx, normalizeToken(name)))
			lines = append(lines, fmt.Sprintf("server.%d.status=%s", idx, status))
			lines = append(lines, fmt.Sprintf("server.%d.connected=%t", idx, connected))
		}
		rows := make([]types.RenderTableRow, 0, len(names))
		for _, name := range names {
			rows = append(rows, tableRow(name, boolState(cmdCtx.State.MCPConnections[name], "connected", "disconnected"), "-", boolState(cmdCtx.State.MCPConnections[name], "yes", "no")))
		}
		return resultWithIntents(strings.Join(lines, "\n"), mcpListIntents(rows)...), nil
	}

	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "no-redirect":
		if len(inv.Args) != 1 {
			return Result{}, fmt.Errorf(mcpUsage)
		}
		return resultWithIntents("MCP_SETTINGS\nredirect=false", mcpSettingsIntents(false)...), nil
	case "reconnect":
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf(mcpUsage)
		}
		name := strings.TrimSpace(inv.Args[1])
		if name == "" {
			return Result{}, fmt.Errorf(mcpUsage)
		}
		cmdCtx.State.MCPConnections[name] = true
		cmdCtx.State.MCPLastQuickFix = "/mcp status " + name
		message := fmt.Sprintf("MCP_RECONNECT\nname=%s\nconnected=true\nquick_fix=%s", normalizeToken(name), normalizeToken(cmdCtx.State.MCPLastQuickFix))
		return resultWithIntents(message, mcpReconnectIntents(name, cmdCtx.State.MCPLastQuickFix)...), nil
	case "enable", "disable":
		if len(inv.Args) > 2 {
			return Result{}, fmt.Errorf(mcpUsage)
		}
		target := "all"
		if len(inv.Args) == 2 {
			target = strings.TrimSpace(inv.Args[1])
			if target == "" {
				return Result{}, fmt.Errorf(mcpUsage)
			}
		}
		enable := sub == "enable"
		changed := 0
		if strings.EqualFold(target, "all") {
			names := make([]string, 0, len(cmdCtx.State.MCPConnections))
			for name := range cmdCtx.State.MCPConnections {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				if cmdCtx.State.MCPConnections[name] != enable {
					cmdCtx.State.MCPConnections[name] = enable
					changed++
				}
			}
		} else {
			before := cmdCtx.State.MCPConnections[target]
			cmdCtx.State.MCPConnections[target] = enable
			if before != enable {
				changed = 1
			}
		}
		action := "Enabled"
		if !enable {
			action = "Disabled"
		}
		message := fmt.Sprintf("%s %d MCP server(s)", action, changed)
		payload := fmt.Sprintf("MCP_TOGGLE\naction=%s\ntarget=%s\nchanged=%d\nmessage=%s", normalizeToken(sub), normalizeToken(target), changed, normalizeToken(message))
		return resultWithIntents(payload, mcpToggleIntents(sub, target, changed, message)...), nil
	case "doctor":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf(mcpUsage)
		}
		cmdCtx.State.MCPDoctorCount++
		connected := 0
		if cmdCtx.MCPManager != nil {
			for _, status := range cmdCtx.MCPManager.ServerStatuses() {
				if strings.EqualFold(string(status.ConnectionState), "connected") {
					connected++
				}
			}
		} else {
			for _, ok := range cmdCtx.State.MCPConnections {
				if ok {
					connected++
				}
			}
		}
		fix := "/mcp add <name> stdio <command>"
		if connected > 0 {
			fix = "/mcp list-tools"
		}
		cmdCtx.State.MCPLastQuickFix = fix
		message := fmt.Sprintf("MCP_DOCTOR\nmanager=%t\nservers=%d\nconnected=%d\nquick_fix=%s\ndoctor_runs=%d", cmdCtx.MCPManager != nil, len(cmdCtx.State.MCPConnections), connected, normalizeToken(fix), cmdCtx.State.MCPDoctorCount)
		return resultWithIntents(message, mcpDoctorIntents(cmdCtx.MCPManager != nil, len(cmdCtx.State.MCPConnections), connected, cmdCtx.State.MCPDoctorCount, fix)...), nil
	case "diagnostics":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf(mcpUsage)
		}
		return renderMCPDiagnostics(cmdCtx.State, cmdCtx.MCPManager), nil
	case "repair":
		if len(inv.Args) > 2 {
			return Result{}, fmt.Errorf(mcpUsage)
		}
		mode := "auto"
		if len(inv.Args) == 2 {
			mode = strings.ToLower(strings.TrimSpace(inv.Args[1]))
			if mode == "" {
				return Result{}, fmt.Errorf(mcpUsage)
			}
		}
		return repairMCPState(cmdCtx.State, cmdCtx.MCPManager, mode), nil
	case "add":
		if cmdCtx.MCPManager == nil {
			return Result{}, fmt.Errorf("mcp manager unavailable")
		}
		if len(inv.Args) < 3 {
			return Result{}, fmt.Errorf(mcpUsage)
		}
		name := strings.TrimSpace(inv.Args[1])
		if name == "" {
			return Result{}, fmt.Errorf(mcpUsage)
		}

		transportRaw := strings.ToLower(strings.TrimSpace(inv.Args[2]))
		transport := transportRaw
		targetIdx := 2
		switch transportRaw {
		case "stdio", "sse", "ws", "websocket":
			targetIdx = 3
		default:
			transport = "stdio"
		}
		if len(inv.Args) <= targetIdx {
			return Result{}, fmt.Errorf(mcpUsage)
		}

		target := strings.TrimSpace(inv.Args[targetIdx])
		if target == "" {
			return Result{}, fmt.Errorf(mcpUsage)
		}

		cfg := buildMCPServerConfig(name, transport, target, inv.Args[targetIdx+1:])
		if err := cmdCtx.MCPManager.AddServerConfig(cfg); err != nil {
			return Result{}, err
		}
		message := fmt.Sprintf("MCP_ADD\nname=%s\ntransport=%s\nadded=true", normalizeToken(name), normalizeToken(string(cfg.Transport)))
		return resultWithIntents(message, mcpMutationIntents("MCP server added", name, string(cfg.Transport), true)...), nil
	case "remove":
		if cmdCtx.MCPManager == nil {
			return Result{}, fmt.Errorf("mcp manager unavailable")
		}
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf(mcpUsage)
		}
		name := strings.TrimSpace(inv.Args[1])
		if name == "" {
			return Result{}, fmt.Errorf(mcpUsage)
		}
		removed := cmdCtx.MCPManager.RemoveServerConfig(name)
		message := fmt.Sprintf("MCP_REMOVE\nname=%s\nremoved=%t", normalizeToken(name), removed)
		return resultWithIntents(message, mcpMutationIntents("MCP server removed", name, "-", removed)...), nil
	case "connect":
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf(mcpUsage)
		}
		name := strings.TrimSpace(inv.Args[1])
		if name == "" {
			return Result{}, fmt.Errorf(mcpUsage)
		}
		cmdCtx.State.MCPConnections[name] = true
		message := fmt.Sprintf("MCP_CONNECT\nname=%s\nstatus=connected\nconnected=true", normalizeToken(name))
		return resultWithIntents(message, mcpConnectionIntents("MCP connected", name, "connected", true)...), nil
	case "disconnect":
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf(mcpUsage)
		}
		name := strings.TrimSpace(inv.Args[1])
		if name == "" {
			return Result{}, fmt.Errorf(mcpUsage)
		}
		cmdCtx.State.MCPConnections[name] = false
		message := fmt.Sprintf("MCP_DISCONNECT\nname=%s\nstatus=disconnected\nconnected=false", normalizeToken(name))
		return resultWithIntents(message, mcpConnectionIntents("MCP disconnected", name, "disconnected", false)...), nil
	case "status":
		if cmdCtx.MCPManager != nil {
			if len(inv.Args) > 2 {
				return Result{}, fmt.Errorf(mcpUsage)
			}
			if len(inv.Args) == 1 {
				statuses := cmdCtx.MCPManager.ServerStatuses()
				lines := []string{"MCP_STATUS", "scope=all", fmt.Sprintf("count=%d", len(statuses))}
				for i, status := range statuses {
					idx := i + 1
					lines = append(lines, fmt.Sprintf("server.%d.name=%s", idx, normalizeToken(status.ServerName)))
					lines = append(lines, fmt.Sprintf("server.%d.connection_state=%s", idx, normalizeToken(string(status.ConnectionState))))
					lines = append(lines, fmt.Sprintf("server.%d.transport=%s", idx, normalizeToken(string(status.TransportType))))
					lines = append(lines, fmt.Sprintf("server.%d.auth_status=%s", idx, normalizeToken(string(status.AuthStatus))))
					lines = append(lines, fmt.Sprintf("server.%d.authenticated=%t", idx, status.Authenticated))
				}
				options := make([]types.RenderOption, 0, len(statuses))
				for _, status := range statuses {
					options = append(options, option(status.ServerName, fmt.Sprintf("transport=%s auth=%s", status.TransportType, status.AuthStatus), string(status.ConnectionState), "/mcp status "+status.ServerName, false))
				}
				return resultWithIntents(strings.Join(lines, "\n"), mcpStatusIntents("all", options)...), nil
			}

			name := strings.TrimSpace(inv.Args[1])
			if name == "" {
				return Result{}, fmt.Errorf(mcpUsage)
			}
			status, ok := cmdCtx.MCPManager.ServerStatus(name)
			if !ok {
				return Result{}, fmt.Errorf("mcp server not found: %s", name)
			}
			message := fmt.Sprintf("MCP_STATUS\nname=%s\nconnection_state=%s\ntransport=%s\nauth_status=%s\nauthenticated=%t\ncached_resources=%d\ncached_contents=%d", normalizeToken(name), normalizeToken(string(status.ConnectionState)), normalizeToken(string(status.TransportType)), normalizeToken(string(status.AuthStatus)), status.Authenticated, status.CachedResourceCount, status.CachedContentEntries)
			options := []types.RenderOption{option(name, fmt.Sprintf("transport=%s auth=%s cached=%d/%d", status.TransportType, status.AuthStatus, status.CachedResourceCount, status.CachedContentEntries), string(status.ConnectionState), "/mcp auth-status "+name, true)}
			return resultWithIntents(message, mcpStatusIntents(name, options)...), nil
		}
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf(mcpUsage)
		}
		name := strings.TrimSpace(inv.Args[1])
		if name == "" {
			return Result{}, fmt.Errorf(mcpUsage)
		}
		connected := cmdCtx.State.MCPConnections[name]
		status := "disconnected"
		if connected {
			status = "connected"
		}
		message := fmt.Sprintf("MCP_STATUS\nname=%s\nstatus=%s\nconnected=%t", normalizeToken(name), status, connected)
		return resultWithIntents(message, mcpStatusIntents(name, []types.RenderOption{option(name, "connection tracked in local runtime state", status, toggleCommand(connected, "/mcp disconnect "+name, "/mcp connect "+name), true)})...), nil
	case "list-tools":
		if cmdCtx.MCPManager == nil {
			return Result{}, fmt.Errorf("mcp manager unavailable")
		}
		if len(inv.Args) > 2 {
			return Result{}, fmt.Errorf(mcpUsage)
		}
		serverFilter := ""
		if len(inv.Args) == 2 {
			serverFilter = strings.TrimSpace(inv.Args[1])
			if serverFilter == "" {
				return Result{}, fmt.Errorf(mcpUsage)
			}
		}
		tools := cmdCtx.MCPManager.Tools()
		sort.Slice(tools, func(i, j int) bool {
			if tools[i].ServerName != tools[j].ServerName {
				return tools[i].ServerName < tools[j].ServerName
			}
			if tools[i].Def.Name != tools[j].Def.Name {
				return tools[i].Def.Name < tools[j].Def.Name
			}
			return tools[i].Def.Description < tools[j].Def.Description
		})
		lines := []string{"MCP_TOOLS"}
		count := 0
		for _, tool := range tools {
			if serverFilter != "" && !strings.EqualFold(tool.ServerName, serverFilter) {
				continue
			}
			count++
			lines = append(lines, fmt.Sprintf("tool.%d.server=%s", count, normalizeToken(tool.ServerName)))
			lines = append(lines, fmt.Sprintf("tool.%d.name=%s", count, normalizeToken(tool.Def.Name)))
			lines = append(lines, fmt.Sprintf("tool.%d.description=%s", count, normalizeToken(tool.Def.Description)))
		}
		lines = append([]string{"MCP_TOOLS", fmt.Sprintf("count=%d", count)}, lines[1:]...)
		if serverFilter != "" {
			lines = append([]string{lines[0], lines[1], fmt.Sprintf("server=%s", normalizeToken(serverFilter))}, lines[2:]...)
		}
		options := make([]types.RenderOption, 0, count)
		for _, tool := range tools {
			if serverFilter != "" && !strings.EqualFold(tool.ServerName, serverFilter) {
				continue
			}
			options = append(options, option(tool.Def.Name, tool.Def.Description, tool.ServerName, "/mcp list-tools "+tool.ServerName, false))
		}
		return resultWithIntents(strings.Join(lines, "\n"), mcpToolsIntents(serverFilter, options)...), nil
	case "list-resources":
		if cmdCtx.MCPManager == nil {
			return Result{}, fmt.Errorf("mcp manager unavailable")
		}
		if len(inv.Args) > 2 {
			return Result{}, fmt.Errorf(mcpUsage)
		}
		serverFilter := ""
		if len(inv.Args) == 2 {
			serverFilter = strings.TrimSpace(inv.Args[1])
			if serverFilter == "" {
				return Result{}, fmt.Errorf(mcpUsage)
			}
		}
		resources, err := cmdCtx.MCPManager.ListResources(ctx)
		if err != nil {
			return Result{}, err
		}
		lines := []string{"MCP_RESOURCES"}
		count := 0
		for _, resource := range resources {
			if serverFilter != "" && !strings.EqualFold(resource.ServerName, serverFilter) {
				continue
			}
			count++
			lines = append(lines, fmt.Sprintf("resource.%d.server=%s", count, normalizeToken(resource.ServerName)))
			lines = append(lines, fmt.Sprintf("resource.%d.uri=%s", count, normalizeToken(resource.URI)))
			lines = append(lines, fmt.Sprintf("resource.%d.name=%s", count, normalizeToken(resource.Name)))
			lines = append(lines, fmt.Sprintf("resource.%d.mime_type=%s", count, normalizeToken(resource.MIMEType)))
		}
		lines = append([]string{"MCP_RESOURCES", fmt.Sprintf("count=%d", count)}, lines[1:]...)
		if serverFilter != "" {
			lines = append([]string{lines[0], lines[1], fmt.Sprintf("server=%s", normalizeToken(serverFilter))}, lines[2:]...)
		}
		options := make([]types.RenderOption, 0, count)
		for _, resource := range resources {
			if serverFilter != "" && !strings.EqualFold(resource.ServerName, serverFilter) {
				continue
			}
			options = append(options, option(resource.Name, resource.URI, resource.MIMEType, "/mcp list-resources "+resource.ServerName, false))
		}
		return resultWithIntents(strings.Join(lines, "\n"), mcpResourcesIntents(serverFilter, options)...), nil
	case "auth-status":
		if cmdCtx.MCPManager == nil {
			return Result{}, fmt.Errorf("mcp manager unavailable")
		}
		if len(inv.Args) > 2 {
			return Result{}, fmt.Errorf(mcpUsage)
		}
		if len(inv.Args) == 2 {
			name := strings.TrimSpace(inv.Args[1])
			if name == "" {
				return Result{}, fmt.Errorf(mcpUsage)
			}
			status, ok := cmdCtx.MCPManager.ServerStatus(name)
			if !ok {
				return Result{}, fmt.Errorf("mcp server not found: %s", name)
			}
			message := fmt.Sprintf("MCP_AUTH_STATUS\nname=%s\nauth_status=%s\nauthenticated=%t\nconnection_state=%s", normalizeToken(name), normalizeToken(string(status.AuthStatus)), status.Authenticated, normalizeToken(string(status.ConnectionState)))
			return resultWithIntents(message, mcpAuthStatusIntents([]types.RenderOption{option(name, fmt.Sprintf("connection=%s authenticated=%t", status.ConnectionState, status.Authenticated), string(status.AuthStatus), "/mcp status "+name, true)})...), nil
		}
		statuses := cmdCtx.MCPManager.ServerStatuses()
		lines := []string{"MCP_AUTH_STATUS", fmt.Sprintf("count=%d", len(statuses))}
		for i, status := range statuses {
			idx := i + 1
			lines = append(lines, fmt.Sprintf("server.%d.name=%s", idx, normalizeToken(status.ServerName)))
			lines = append(lines, fmt.Sprintf("server.%d.auth_status=%s", idx, normalizeToken(string(status.AuthStatus))))
			lines = append(lines, fmt.Sprintf("server.%d.authenticated=%t", idx, status.Authenticated))
			lines = append(lines, fmt.Sprintf("server.%d.connection_state=%s", idx, normalizeToken(string(status.ConnectionState))))
		}
		options := make([]types.RenderOption, 0, len(statuses))
		for _, status := range statuses {
			options = append(options, option(status.ServerName, fmt.Sprintf("connection=%s authenticated=%t", status.ConnectionState, status.Authenticated), string(status.AuthStatus), "/mcp auth-status "+status.ServerName, false))
		}
		return resultWithIntents(strings.Join(lines, "\n"), mcpAuthStatusIntents(options)...), nil
	default:
		return Result{}, fmt.Errorf(mcpUsage)
	}
}

func buildMCPServerConfig(name, transport, target string, args []string) mcp.ServerConfig {
	cfg := mcp.ServerConfig{Name: name}
	switch strings.ToLower(strings.TrimSpace(transport)) {
	case "sse":
		cfg.Transport = mcp.TransportSSE
		cfg.URL = target
	case "ws", "websocket":
		cfg.Transport = mcp.TransportWebSocket
		cfg.URL = target
	default:
		cfg.Transport = mcp.TransportStdio
		cfg.Command = target
		cfg.Args = append([]string(nil), args...)
	}
	return cfg
}

// VimCommand provides stable vim subcommand parsing and messages.
type VimCommand struct{}

func NewVimCommand() *VimCommand          { return &VimCommand{} }
func (c *VimCommand) Name() string        { return "vim" }
func (c *VimCommand) Aliases() []string   { return nil }
func (c *VimCommand) Description() string { return "Manage vim mode: enable|disable|status" }
func (c *VimCommand) Usage() string       { return "/vim [enable|disable|status]" }

func (c *VimCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}

	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /vim [enable|disable|status]")
		}
		message := fmt.Sprintf("VIM_STATUS\nenabled=%t", cmdCtx.State.VimEnabled)
		return resultWithIntents(message, vimStatusIntents(cmdCtx.State.VimEnabled)...), nil
	}

	switch strings.ToLower(strings.TrimSpace(inv.Args[0])) {
	case "enable":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /vim [enable|disable|status]")
		}
		cmdCtx.State.VimEnabled = true
		message := "VIM_SET\nenabled=true"
		return resultWithIntents(message, vimSetIntents(true)...), nil
	case "disable":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /vim [enable|disable|status]")
		}
		cmdCtx.State.VimEnabled = false
		message := "VIM_SET\nenabled=false"
		return resultWithIntents(message, vimSetIntents(false)...), nil
	default:
		return Result{}, fmt.Errorf("usage: /vim [enable|disable|status]")
	}
}

// VoiceCommand provides stable voice subcommand parsing and messages.
type VoiceCommand struct{}

func NewVoiceCommand() *VoiceCommand      { return &VoiceCommand{} }
func (c *VoiceCommand) Name() string      { return "voice" }
func (c *VoiceCommand) Aliases() []string { return nil }
func (c *VoiceCommand) Description() string {
	return "Manage local voice mode placeholder: enable|disable|status"
}
func (c *VoiceCommand) Usage() string { return "/voice [enable|disable|status]" }

func (c *VoiceCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}

	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /voice [enable|disable|status]")
		}
		message := fmt.Sprintf("VOICE_STATUS\nmode=local-placeholder\nenabled=%t", cmdCtx.State.VoiceEnabled)
		return resultWithIntents(message, voiceStatusIntents(cmdCtx.State.VoiceEnabled)...), nil
	}

	switch strings.ToLower(strings.TrimSpace(inv.Args[0])) {
	case "enable":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /voice [enable|disable|status]")
		}
		cmdCtx.State.VoiceEnabled = true
		message := "VOICE_SET\nmode=local-placeholder\nenabled=true"
		return resultWithIntents(message, voiceSetIntents(true)...), nil
	case "disable":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /voice [enable|disable|status]")
		}
		cmdCtx.State.VoiceEnabled = false
		message := "VOICE_SET\nmode=local-placeholder\nenabled=false"
		return resultWithIntents(message, voiceSetIntents(false)...), nil
	default:
		return Result{}, fmt.Errorf("usage: /voice [enable|disable|status]")
	}
}

// LoginCommand provides deterministic provider-agnostic login guidance.
type LoginCommand struct{}

func NewLoginCommand() *LoginCommand      { return &LoginCommand{} }
func (c *LoginCommand) Name() string      { return "login" }
func (c *LoginCommand) Aliases() []string { return []string{"signin", "auth"} }
func (c *LoginCommand) Description() string {
	return "Start provider-agnostic login guidance and auth intent"
}
func (c *LoginCommand) Usage() string {
	return "/login [status|provider <name>|account <name>|<provider>]"
}

func (c *LoginCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	return executeLoginCommand(cmdCtx, inv)
}

// LogoutCommand provides deterministic logout behavior.
type LogoutCommand struct{}

func NewLogoutCommand() *LogoutCommand       { return &LogoutCommand{} }
func (c *LogoutCommand) Name() string        { return "logout" }
func (c *LogoutCommand) Aliases() []string   { return []string{"signout"} }
func (c *LogoutCommand) Description() string { return "Clear local auth session state" }
func (c *LogoutCommand) Usage() string       { return "/logout [status]" }

func (c *LogoutCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	return executeLogoutCommand(cmdCtx, inv)
}

// BuddyCommand provides stable buddy companion controls.
type BuddyCommand struct{}

func NewBuddyCommand() *BuddyCommand      { return &BuddyCommand{} }
func (c *BuddyCommand) Name() string      { return "buddy" }
func (c *BuddyCommand) Aliases() []string { return nil }
func (c *BuddyCommand) Description() string {
	return "Manage buddy companion: status|hatch|pet|mute|unmute|help"
}
func (c *BuddyCommand) Usage() string {
	return "/buddy [status|hatch|pet|mute|unmute|help]"
}

func (c *BuddyCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}

	if len(inv.Args) == 0 {
		return c.renderBuddyStatus(cmdCtx.State), nil
	}

	if len(inv.Args) > 1 {
		return Result{}, fmt.Errorf("usage: /buddy [status|hatch|pet|mute|unmute|help]")
	}

	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "status":
		return c.renderBuddyStatus(cmdCtx.State), nil
	case "help":
		message := "BUDDY_HELP\nusage=/buddy [status|hatch|pet|mute|unmute|help]\nsubcommands=status,hatch,pet,mute,unmute,help"
		return resultWithIntents(message, buddyHelpIntents()...), nil
	case "hatch":
		if cmdCtx.State.BuddyHatched {
			message := "BUDDY_HATCH\nstatus=already_hatched\nhatched=true"
			return resultWithIntents(message, buddyActionIntents("Buddy hatch", "already_hatched", true, cmdCtx.State.BuddyMuted, cmdCtx.State.BuddyPetCount)...), nil
		}
		cmdCtx.State.BuddyHatched = true
		cmdCtx.State.BuddyPetCount = 0
		message := "BUDDY_HATCH\nstatus=hatched\nhatched=true"
		return resultWithIntents(message, buddyActionIntents("Buddy hatch", "hatched", true, cmdCtx.State.BuddyMuted, cmdCtx.State.BuddyPetCount)...), nil
	case "pet":
		if !cmdCtx.State.BuddyHatched {
			message := "BUDDY_PET\nstatus=not_hatched\nhatched=false\npet_count=0"
			return resultWithIntents(message, buddyActionIntents("Buddy pet", "not_hatched", false, cmdCtx.State.BuddyMuted, 0)...), nil
		}
		cmdCtx.State.BuddyPetCount++
		if cmdCtx.State.BuddyMuted {
			message := fmt.Sprintf("BUDDY_PET\nstatus=pet_muted\nhatched=true\nmuted=true\npet_count=%d", cmdCtx.State.BuddyPetCount)
			return resultWithIntents(message, buddyActionIntents("Buddy pet", "pet_muted", true, true, cmdCtx.State.BuddyPetCount)...), nil
		}
		message := fmt.Sprintf("BUDDY_PET\nstatus=pet\nhatched=true\nmuted=false\npet_count=%d", cmdCtx.State.BuddyPetCount)
		return resultWithIntents(message, buddyActionIntents("Buddy pet", "pet", true, false, cmdCtx.State.BuddyPetCount)...), nil
	case "mute":
		cmdCtx.State.BuddyMuted = true
		message := "BUDDY_MUTE\nmuted=true"
		return resultWithIntents(message, buddyActionIntents("Buddy mute", "muted", cmdCtx.State.BuddyHatched, true, cmdCtx.State.BuddyPetCount)...), nil
	case "unmute":
		cmdCtx.State.BuddyMuted = false
		message := "BUDDY_UNMUTE\nmuted=false"
		return resultWithIntents(message, buddyActionIntents("Buddy unmute", "unmuted", cmdCtx.State.BuddyHatched, false, cmdCtx.State.BuddyPetCount)...), nil
	default:
		return Result{}, fmt.Errorf("usage: /buddy [status|hatch|pet|mute|unmute|help]")
	}
}

func (c *BuddyCommand) renderBuddyStatus(state *RuntimeState) Result {
	status := "egg"
	summary := "Buddy is waiting to hatch."
	if state.BuddyHatched {
		status = "hatched"
		summary = "Buddy is active and can be petted."
	}
	message := fmt.Sprintf("BUDDY_STATUS\nstate=%s\nhatched=%t\nmuted=%t\npet_count=%d", status, state.BuddyHatched, state.BuddyMuted, state.BuddyPetCount)
	return resultWithIntents(message, buddyStatusIntents(status, summary, state.BuddyHatched, state.BuddyMuted, state.BuddyPetCount)...)
}

// FilesCommand provides deterministic context-file controls.
type FilesCommand struct{}

func NewFilesCommand() *FilesCommand      { return &FilesCommand{} }
func (c *FilesCommand) Name() string      { return "files" }
func (c *FilesCommand) Aliases() []string { return nil }
func (c *FilesCommand) Description() string {
	return "List and manage deterministic context files"
}
func (c *FilesCommand) Usage() string {
	return "/files [list|add <path>|remove <path>|clear|status]"
}

func (c *FilesCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "list") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /files [list|add <path>|remove <path>|clear|status]")
		}
		return renderFilesList(effectiveContextFiles(cmdCtx.State)), nil
	}

	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "status":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /files [list|add <path>|remove <path>|clear|status]")
		}
		message := fmt.Sprintf("FILES_STATUS\ncount=%d\nproject_paths=%d\nadds=%d\nremoves=%d\nclears=%d\nlast_action=%s\nlast_path=%s", len(effectiveContextFiles(cmdCtx.State)), len(uniqueSortedStrings(append([]string(nil), cmdCtx.State.ProjectPaths...))), cmdCtx.State.FilesAdds, cmdCtx.State.FilesRemoves, cmdCtx.State.FilesClears, normalizeToken(cmdCtx.State.LastFileAction), normalizeToken(cmdCtx.State.LastFilePath))
		return resultWithIntents(message, filesStatusIntents(len(effectiveContextFiles(cmdCtx.State)), len(uniqueSortedStrings(append([]string(nil), cmdCtx.State.ProjectPaths...))), cmdCtx.State.FilesAdds, cmdCtx.State.FilesRemoves, cmdCtx.State.FilesClears, cmdCtx.State.LastFileAction, cmdCtx.State.LastFilePath)...), nil
	case "add":
		if len(inv.Args) < 2 {
			return Result{}, fmt.Errorf("usage: /files [list|add <path>|remove <path>|clear|status]")
		}
		path := strings.TrimSpace(strings.Join(inv.Args[1:], " "))
		if path == "" {
			return Result{}, fmt.Errorf("usage: /files [list|add <path>|remove <path>|clear|status]")
		}
		before := len(cmdCtx.State.ContextFiles)
		cmdCtx.State.ContextFiles = uniqueSortedStrings(append(cmdCtx.State.ContextFiles, path))
		added := len(cmdCtx.State.ContextFiles) > before
		if added {
			cmdCtx.State.FilesAdds++
		}
		cmdCtx.State.LastFileAction = "add"
		cmdCtx.State.LastFilePath = path
		message := fmt.Sprintf("FILES_ADD\npath=%s\nadded=%t\ncount=%d", normalizeToken(path), added, len(cmdCtx.State.ContextFiles))
		return resultWithIntents(message, filesMutationIntents("File added", path, added, len(cmdCtx.State.ContextFiles))...), nil
	case "remove":
		if len(inv.Args) < 2 {
			return Result{}, fmt.Errorf("usage: /files [list|add <path>|remove <path>|clear|status]")
		}
		path := strings.TrimSpace(strings.Join(inv.Args[1:], " "))
		if path == "" {
			return Result{}, fmt.Errorf("usage: /files [list|add <path>|remove <path>|clear|status]")
		}
		removed := false
		next := make([]string, 0, len(cmdCtx.State.ContextFiles))
		for _, existing := range cmdCtx.State.ContextFiles {
			if strings.TrimSpace(existing) == path {
				removed = true
				continue
			}
			next = append(next, existing)
		}
		cmdCtx.State.ContextFiles = uniqueSortedStrings(next)
		if removed {
			cmdCtx.State.FilesRemoves++
		}
		cmdCtx.State.LastFileAction = "remove"
		cmdCtx.State.LastFilePath = path
		message := fmt.Sprintf("FILES_REMOVE\npath=%s\nremoved=%t\ncount=%d", normalizeToken(path), removed, len(cmdCtx.State.ContextFiles))
		return resultWithIntents(message, filesMutationIntents("File removed", path, removed, len(cmdCtx.State.ContextFiles))...), nil
	case "clear":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /files [list|add <path>|remove <path>|clear|status]")
		}
		cmdCtx.State.ContextFiles = nil
		cmdCtx.State.FilesClears++
		cmdCtx.State.LastFileAction = "clear"
		cmdCtx.State.LastFilePath = ""
		message := "FILES_CLEAR\ncount=0"
		return resultWithIntents(message, filesMutationIntents("Files cleared", "", true, 0)...), nil
	default:
		return Result{}, fmt.Errorf("usage: /files [list|add <path>|remove <path>|clear|status]")
	}
}

func renderFilesList(files []string) Result {
	files = uniqueSortedStrings(append([]string(nil), files...))
	if len(files) == 0 {
		return resultWithIntents("FILES_LIST\ncount=0", filesListIntents(nil)...)
	}
	lines := []string{"FILES_LIST", fmt.Sprintf("count=%d", len(files))}
	for i, f := range files {
		lines = append(lines, fmt.Sprintf("file.%d=%s", i+1, normalizeToken(f)))
	}
	message := strings.Join(lines, "\n")
	return resultWithIntents(message, filesListIntents(files)...)
}

func effectiveContextFiles(state *RuntimeState) []string {
	if state == nil {
		return nil
	}
	combined := append([]string(nil), state.ProjectPaths...)
	combined = append(combined, state.ContextFiles...)
	return uniqueSortedStrings(combined)
}

// ThemeCommand provides deterministic theme controls.
type ThemeCommand struct{}

func NewThemeCommand() *ThemeCommand      { return &ThemeCommand{} }
func (c *ThemeCommand) Name() string      { return "theme" }
func (c *ThemeCommand) Aliases() []string { return []string{"themes"} }
func (c *ThemeCommand) Description() string {
	return "Show or set CLI output theme"
}
func (c *ThemeCommand) Usage() string {
	return "/theme [get|status|list|set <light|dark|system>|cycle|preview <light|dark|system>]"
}

func (c *ThemeCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "get") || strings.EqualFold(inv.Args[0], "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /theme [get|status|list|set <light|dark|system>]")
		}
		theme := normalizeToken(cmdCtx.State.OutputStyle)
		if theme == "-" {
			theme = "system"
		}
		message := fmt.Sprintf("THEME_STATUS\ntheme=%s", theme)
		return resultWithIntents(message, themeStatusIntents(theme)...), nil
	}

	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "list":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /theme [get|status|list|set <light|dark|system>]")
		}
		message := "THEME_LIST\ncount=3\ntheme.1=dark\ntheme.2=light\ntheme.3=system"
		return resultWithIntents(message, themeListIntents()...), nil
	case "set":
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf("usage: /theme [get|status|list|set <light|dark|system>]")
		}
		next := strings.ToLower(strings.TrimSpace(inv.Args[1]))
		switch next {
		case "dark", "light", "system":
			cmdCtx.State.OutputStyle = next
			cmdCtx.State.ThemeSetCount++
			message := fmt.Sprintf("THEME_SET\ntheme=%s\nset_count=%d", next, cmdCtx.State.ThemeSetCount)
			return resultWithIntents(message, themeMutationIntents("Theme set", next, cmdCtx.State.ThemeSetCount)...), nil
		default:
			return Result{}, fmt.Errorf("usage: /theme [get|status|list|set <light|dark|system>|cycle|preview <light|dark|system>]")
		}
	case "cycle":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /theme [get|status|list|set <light|dark|system>|cycle|preview <light|dark|system>]")
		}
		current := strings.ToLower(strings.TrimSpace(cmdCtx.State.OutputStyle))
		next := "system"
		switch current {
		case "system":
			next = "light"
		case "light":
			next = "dark"
		case "dark":
			next = "system"
		default:
			next = "system"
		}
		cmdCtx.State.OutputStyle = next
		cmdCtx.State.ThemeSetCount++
		message := fmt.Sprintf("THEME_CYCLE\ntheme=%s\nset_count=%d", next, cmdCtx.State.ThemeSetCount)
		return resultWithIntents(message, themeMutationIntents("Theme cycled", next, cmdCtx.State.ThemeSetCount)...), nil
	case "preview":
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf("usage: /theme [get|status|list|set <light|dark|system>|cycle|preview <light|dark|system>]")
		}
		next := strings.ToLower(strings.TrimSpace(inv.Args[1]))
		switch next {
		case "dark", "light", "system":
			message := fmt.Sprintf("THEME_PREVIEW\ntheme=%s\napplied=false", next)
			return resultWithIntents(message, themePreviewIntents(next)...), nil
		default:
			return Result{}, fmt.Errorf("usage: /theme [get|status|list|set <light|dark|system>|cycle|preview <light|dark|system>]")
		}
	default:
		return Result{}, fmt.Errorf("usage: /theme [get|status|list|set <light|dark|system>|cycle|preview <light|dark|system>]")
	}
}

// OutputStyleCommand provides deterministic output-style controls.
type OutputStyleCommand struct{}

func NewOutputStyleCommand() *OutputStyleCommand  { return &OutputStyleCommand{} }
func (c *OutputStyleCommand) Name() string        { return "output-style" }
func (c *OutputStyleCommand) Aliases() []string   { return []string{"style"} }
func (c *OutputStyleCommand) Description() string { return "Show or set output style mode" }
func (c *OutputStyleCommand) Usage() string {
	return "/output-style [status|get|list|set <default|concise|explanatory|json>]"
}
func (c *OutputStyleCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	usage := "usage: /output-style [status|get|list|set <default|concise|explanatory|json>]"
	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "status") || strings.EqualFold(inv.Args[0], "get") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("%s", usage)
		}
		style := normalizeToken(cmdCtx.State.OutputFormat)
		if style == "-" {
			style = "default"
		}
		message := fmt.Sprintf("OUTPUT_STYLE_STATUS\nstyle=%s\nset_count=%d", style, cmdCtx.State.OutputStyleSetCount)
		return resultWithIntents(message, outputStyleStatusIntents(style, cmdCtx.State.OutputStyleSetCount)...), nil
	}
	if len(inv.Args) == 1 && strings.EqualFold(inv.Args[0], "list") {
		message := "OUTPUT_STYLE_LIST\ncount=4\nstyle.1=default\nstyle.2=concise\nstyle.3=explanatory\nstyle.4=json"
		return resultWithIntents(message, outputStyleListIntents()...), nil
	}
	if len(inv.Args) != 2 || !strings.EqualFold(inv.Args[0], "set") {
		return Result{}, fmt.Errorf("%s", usage)
	}
	style := strings.ToLower(strings.TrimSpace(inv.Args[1]))
	switch style {
	case "default", "concise", "explanatory", "json":
		cmdCtx.State.OutputFormat = style
		cmdCtx.State.OutputStyleSetCount++
		message := fmt.Sprintf("OUTPUT_STYLE_SET\nstyle=%s\nset_count=%d", style, cmdCtx.State.OutputStyleSetCount)
		return resultWithIntents(message, outputStyleStatusIntents(style, cmdCtx.State.OutputStyleSetCount)...), nil
	default:
		return Result{}, fmt.Errorf("%s", usage)
	}
}

// StatuslineCommand provides deterministic statusline setup requests.
type StatuslineCommand struct{}

func NewStatuslineCommand() *StatuslineCommand   { return &StatuslineCommand{} }
func (c *StatuslineCommand) Name() string        { return "statusline" }
func (c *StatuslineCommand) Aliases() []string   { return nil }
func (c *StatuslineCommand) Description() string { return "Set up statusline agent prompt" }
func (c *StatuslineCommand) Usage() string       { return "/statusline [status|setup [prompt]]" }

func (c *StatuslineCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 {
		return c.statuslineSetup(cmdCtx.State, "Configure my statusLine from my shell PS1 configuration"), nil
	}

	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "status":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /statusline [status|setup [prompt]]")
		}
		runtime := statusRuntimeSnapshot(cmdCtx.State)
		workingDir := strings.TrimSpace(statusWorkingDir(cmdCtx.State))
		if workingDir == "" {
			workingDir = "."
		}
		message := fmt.Sprintf("STATUSLINE_STATUS\ncount=%d\nlast_prompt=%s\nprovider=%s\nmodel=%s\nmodel_ref=%s\nlogged_in=%t\nprovider_ready=%t\nworking_dir=%s\nworkspace_root=%s\npath_scope=%s\nlast_render_ms=%d\ntool_calls=%d\npermission_events=%d\nstate_transitions=%d\nturns=%d\ntool_inflight=%d\ntasks_total=%d\ntasks_running=%d\ntasks_completed=%d\nteams_total=%d\nteams_active=%d\nlast_stop_reason=%s", cmdCtx.State.StatuslineCount, normalizeToken(cmdCtx.State.LastStatusline), normalizeToken(runtime.ProviderName), normalizeToken(runtime.Model), normalizeToken(runtime.ModelRef), runtime.LoggedIn, runtime.ProviderReady, normalizeToken(workingDir), normalizeToken(workingDir), normalizeToken("workspace"), cmdCtx.State.StatuslineLastRenderMS, cmdCtx.State.StatuslineToolCalls, cmdCtx.State.StatuslinePermissions, cmdCtx.State.StatuslineTransitions, runtime.Turns, runtime.ToolInflight, runtime.TasksTotal, runtime.TasksRunning, runtime.TasksCompleted, runtime.TeamsTotal, runtime.TeamsActive, normalizeToken(string(runtime.LastStopReason)))
		return resultWithIntents(message, statuslineStatusIntents(cmdCtx.State, runtime)...), nil
	case "setup":
		prompt := strings.TrimSpace(strings.Join(inv.Args[1:], " "))
		if prompt == "" {
			prompt = "Configure my statusLine from my shell PS1 configuration"
		}
		return c.statuslineSetup(cmdCtx.State, prompt), nil
	default:
		prompt := strings.TrimSpace(strings.Join(inv.Args, " "))
		if prompt == "" {
			return Result{}, fmt.Errorf("usage: /statusline [status|setup [prompt]]")
		}
		return c.statuslineSetup(cmdCtx.State, prompt), nil
	}
}

func (c *StatuslineCommand) statuslineSetup(state *RuntimeState, prompt string) Result {
	state.StatuslineCount++
	state.LastStatusline = prompt
	message := fmt.Sprintf("STATUSLINE_SETUP\nsubagent_type=statusline-setup\nprompt=%s\ncount=%d", normalizeToken(prompt), state.StatuslineCount)
	return resultWithIntents(message, statuslineSetupIntents(prompt, state.StatuslineCount)...)
}

// IDECommand provides provider-agnostic editor detection and open hints.
type IDECommand struct{}

func NewIDECommand() *IDECommand        { return &IDECommand{} }
func (c *IDECommand) Name() string      { return "ide" }
func (c *IDECommand) Aliases() []string { return []string{"editor"} }
func (c *IDECommand) Description() string {
	return "Detect editor and show IDE launch/open hints"
}
func (c *IDECommand) Usage() string {
	return "/ide [status|detect|open [path]|set-editor <name>|hints]"
}

func (c *IDECommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	usage := "usage: /ide [status|detect|open [path]|set-editor <name>|hints]"

	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("%s", usage)
		}
		editor := normalizeToken(cmdCtx.State.IDEEditor)
		detected := normalizeToken(cmdCtx.State.IDEDetectedEditor)
		source := normalizeToken(cmdCtx.State.IDEDetectSource)
		hints := ideHints(editorFromStateOrDetected(cmdCtx.State))
		lines := []string{
			"IDE_STATUS",
			fmt.Sprintf("editor=%s", editor),
			fmt.Sprintf("detected=%s", detected),
			fmt.Sprintf("source=%s", source),
			fmt.Sprintf("open_count=%d", cmdCtx.State.IDEOpenCount),
			fmt.Sprintf("config_count=%d", cmdCtx.State.IDEConfigCount),
			fmt.Sprintf("hint_count=%d", len(hints)),
		}
		for i, hint := range hints {
			lines = append(lines, fmt.Sprintf("hint.%d=%s", i+1, normalizeToken(hint)))
		}
		message := strings.Join(lines, "\n")
		return resultWithIntents(message, ideStatusIntents(cmdCtx.State.IDEEditor, cmdCtx.State.IDEDetectedEditor, cmdCtx.State.IDEDetectSource, cmdCtx.State.IDEOpenCount, cmdCtx.State.IDEConfigCount, hints)...), nil
	}

	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "detect":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("%s", usage)
		}
		editor, source := detectEditorFromEnvironment()
		cmdCtx.State.IDEDetectedEditor = editor
		cmdCtx.State.IDEDetectSource = source
		hints := ideHints(editor)
		lines := []string{
			"IDE_DETECT",
			fmt.Sprintf("editor=%s", normalizeToken(editor)),
			fmt.Sprintf("source=%s", normalizeToken(source)),
			fmt.Sprintf("hint_count=%d", len(hints)),
		}
		for i, hint := range hints {
			lines = append(lines, fmt.Sprintf("hint.%d=%s", i+1, normalizeToken(hint)))
		}
		message := strings.Join(lines, "\n")
		return resultWithIntents(message, ideDetectIntents(editor, source, hints)...), nil
	case "set-editor":
		if len(inv.Args) < 2 {
			return Result{}, fmt.Errorf("%s", usage)
		}
		editor := normalizeEditorName(strings.Join(inv.Args[1:], " "))
		if strings.TrimSpace(editor) == "" {
			return Result{}, fmt.Errorf("%s", usage)
		}
		cmdCtx.State.IDEEditor = editor
		cmdCtx.State.IDEConfigCount++
		message := fmt.Sprintf("IDE_SET_EDITOR\neditor=%s\nconfig_count=%d", normalizeToken(editor), cmdCtx.State.IDEConfigCount)
		return resultWithIntents(message, ideSetEditorIntents(editor, cmdCtx.State.IDEConfigCount)...), nil
	case "open":
		if len(inv.Args) > 2 {
			return Result{}, fmt.Errorf("%s", usage)
		}
		target := "."
		if len(inv.Args) == 2 {
			target = strings.TrimSpace(inv.Args[1])
			if target == "" {
				return Result{}, fmt.Errorf("%s", usage)
			}
		}
		editor := editorFromStateOrDetected(cmdCtx.State)
		if strings.TrimSpace(editor) == "" {
			editor, cmdCtx.State.IDEDetectSource = detectEditorFromEnvironment()
			cmdCtx.State.IDEDetectedEditor = editor
		}
		cmdCtx.State.IDEOpenCount++
		command := editorOpenHint(editor, target)
		message := fmt.Sprintf("IDE_OPEN_HINT\neditor=%s\ntarget=%s\ncommand=%s\nopen_count=%d", normalizeToken(editor), normalizeToken(target), normalizeToken(command), cmdCtx.State.IDEOpenCount)
		return resultWithIntents(message, ideOpenHintIntents(editor, target, command, cmdCtx.State.IDEOpenCount)...), nil
	case "hints":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("%s", usage)
		}
		editor := editorFromStateOrDetected(cmdCtx.State)
		if strings.TrimSpace(editor) == "" {
			editor, cmdCtx.State.IDEDetectSource = detectEditorFromEnvironment()
			cmdCtx.State.IDEDetectedEditor = editor
		}
		hints := ideHints(editor)
		lines := []string{"IDE_HINTS", fmt.Sprintf("editor=%s", normalizeToken(editor)), fmt.Sprintf("count=%d", len(hints))}
		for i, hint := range hints {
			lines = append(lines, fmt.Sprintf("hint.%d=%s", i+1, normalizeToken(hint)))
		}
		message := strings.Join(lines, "\n")
		return resultWithIntents(message, ideDetectIntents(editor, cmdCtx.State.IDEDetectSource, hints)...), nil
	default:
		return Result{}, fmt.Errorf("%s", usage)
	}
}

// KeybindingsCommand provides deterministic keybindings setup/open behavior.
type KeybindingsCommand struct{}

func NewKeybindingsCommand() *KeybindingsCommand { return &KeybindingsCommand{} }
func (c *KeybindingsCommand) Name() string       { return "keybindings" }
func (c *KeybindingsCommand) Aliases() []string  { return nil }
func (c *KeybindingsCommand) Description() string {
	return "Open or create keybindings configuration file"
}
func (c *KeybindingsCommand) Usage() string { return "/keybindings [status|path|open|enable|disable]" }

func (c *KeybindingsCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "open") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /keybindings [status|path|open|enable|disable]")
		}
		if !cmdCtx.State.KeybindingsEnabled {
			message := "Keybinding customization is not enabled. This feature is currently in preview."
			return resultWithIntents(message, keybindingsOpenDisabledIntents()...), nil
		}
		if strings.TrimSpace(cmdCtx.State.KeybindingsPath) == "" {
			cmdCtx.State.KeybindingsPath = "~/.claude/keybindings.json"
		}
		cmdCtx.State.KeybindingsOpens++
		if cmdCtx.State.KeybindingsExists {
			message := fmt.Sprintf("Opened %s in your editor.", cmdCtx.State.KeybindingsPath)
			return resultWithIntents(message, keybindingsOpenIntents(cmdCtx.State.KeybindingsPath, true)...), nil
		}
		cmdCtx.State.KeybindingsExists = true
		message := fmt.Sprintf("Created %s with template. Opened in your editor.", cmdCtx.State.KeybindingsPath)
		return resultWithIntents(message, keybindingsOpenIntents(cmdCtx.State.KeybindingsPath, false)...), nil
	}

	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "status":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /keybindings [status|path|open|enable|disable]")
		}
		path := normalizeToken(cmdCtx.State.KeybindingsPath)
		if path == "-" {
			path = "~/.claude/keybindings.json"
		}
		message := fmt.Sprintf("KEYBINDINGS_STATUS\nenabled=%t\npath=%s\nexists=%t\nopens=%d", cmdCtx.State.KeybindingsEnabled, path, cmdCtx.State.KeybindingsExists, cmdCtx.State.KeybindingsOpens)
		return resultWithIntents(message, keybindingsStatusIntents(cmdCtx.State.KeybindingsEnabled, path, cmdCtx.State.KeybindingsExists, cmdCtx.State.KeybindingsOpens)...), nil
	case "path":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /keybindings [status|path|open|enable|disable]")
		}
		path := normalizeToken(cmdCtx.State.KeybindingsPath)
		if path == "-" {
			path = "~/.claude/keybindings.json"
		}
		message := fmt.Sprintf("KEYBINDINGS_PATH\npath=%s", path)
		return resultWithIntents(message, keybindingsPathIntents(path)...), nil
	case "enable":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /keybindings [status|path|open|enable|disable]")
		}
		cmdCtx.State.KeybindingsEnabled = true
		message := "KEYBINDINGS_SET\nenabled=true"
		return resultWithIntents(message, keybindingsSetIntents(true)...), nil
	case "disable":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /keybindings [status|path|open|enable|disable]")
		}
		cmdCtx.State.KeybindingsEnabled = false
		message := "KEYBINDINGS_SET\nenabled=false"
		return resultWithIntents(message, keybindingsSetIntents(false)...), nil
	default:
		return Result{}, fmt.Errorf("usage: /keybindings [status|path|open|enable|disable]")
	}
}

// StatusCommand reports deterministic runtime status.
type StatusCommand struct{}

func NewStatusCommand() *StatusCommand       { return &StatusCommand{} }
func (c *StatusCommand) Name() string        { return "status" }
func (c *StatusCommand) Aliases() []string   { return []string{"state"} }
func (c *StatusCommand) Description() string { return "Show deterministic runtime status" }
func (c *StatusCommand) Usage() string       { return "/status" }

func (c *StatusCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if len(inv.Args) == 1 {
		sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
		if sub == "diagnostics" || sub == "doctor" || sub == "checks" {
			if cmdCtx.State != nil {
				cmdCtx.State.StatusDiagnosticsCount++
			}
			return renderStatusDiagnostics(cmdCtx.State), nil
		}
	}
	if len(inv.Args) > 0 {
		return Result{}, fmt.Errorf("usage: /status")
	}
	state := cmdCtx.State
	if state == nil {
		state = &RuntimeState{}
	}
	syncProviderReadiness(state)
	state.StatusViewCount++
	selection := RuntimeSelectionTruth(state)
	model := normalizeToken(selection.ModelName)
	if model == "-" {
		model = "unknown"
	}
	compactMode := normalizeToken(state.CompactMode)
	if compactMode == "-" {
		compactMode = "auto"
	}
	runtime := statusRuntimeSnapshot(state)
	workingDir := strings.TrimSpace(statusWorkingDir(state))
	if workingDir == "" {
		workingDir = "."
	}
	lines := []string{
		"STATUS_REPORT",
		fmt.Sprintf("model=%s", model),
		fmt.Sprintf("model_capabilities=%s", statusModelCapabilitySummary(state)),
		fmt.Sprintf("provider=%s", normalizeToken(selection.ProviderName)),
		fmt.Sprintf("logged_in=%t", selection.LoggedIn),
		fmt.Sprintf("account=%s", normalizeToken(state.AuthAccount)),
		fmt.Sprintf("provider_ready=%t", selection.ProviderReady),
		fmt.Sprintf("permission_mode=%s", modeString(state.PermissionMode)),
		fmt.Sprintf("compact_mode=%s", compactMode),
		fmt.Sprintf("compact_requested=%t", state.CompactRequested),
		fmt.Sprintf("resume_requested=%t", state.ResumeRequested),
		fmt.Sprintf("output_style=%s", normalizeToken(state.OutputFormat)),
		fmt.Sprintf("theme=%s", normalizeToken(state.OutputStyle)),
		fmt.Sprintf("session_id=%s", normalizeToken(state.SessionID)),
		fmt.Sprintf("session_path=%s", normalizeToken(state.SessionPath)),
		fmt.Sprintf("context_files=%d", len(effectiveContextFiles(state))),
		fmt.Sprintf("project_paths=%d", len(uniqueSortedStrings(append([]string(nil), state.ProjectPaths...)))),
		fmt.Sprintf("history_entries=%d", len(defaultHistoryEntries(state))),
		fmt.Sprintf("memory_entries=%d", len(state.MemoryEntries)),
		fmt.Sprintf("working_dir=%s", normalizeToken(workingDir)),
		fmt.Sprintf("workspace_root=%s", normalizeToken(workingDir)),
		fmt.Sprintf("path_scope=%s", normalizeToken("workspace")),
	}
	lines = append(lines,
		fmt.Sprintf("turns=%d", runtime.Turns),
		fmt.Sprintf("tool_inflight=%d", runtime.ToolInflight),
		fmt.Sprintf("tasks_total=%d", runtime.TasksTotal),
		fmt.Sprintf("tasks_running=%d", runtime.TasksRunning),
		fmt.Sprintf("tasks_completed=%d", runtime.TasksCompleted),
		fmt.Sprintf("teams_total=%d", runtime.TeamsTotal),
		fmt.Sprintf("teams_active=%d", runtime.TeamsActive),
		fmt.Sprintf("last_stop_reason=%s", normalizeToken(string(runtime.LastStopReason))),
	)
	if next := strings.TrimSpace(RuntimeSelectionNextAction(state)); next != "" {
		lines = append(lines, fmt.Sprintf("next=%s", normalizeToken(next)))
	}
	return resultWithIntents(strings.Join(lines, "\n"), statusReportIntents(state, runtime)...), nil
}

func renderStatusDiagnostics(state *RuntimeState) Result {
	if state == nil {
		state = &RuntimeState{}
	}
	selection := RuntimeSelectionTruth(state)
	loops := correctiveLoopsForState(state)
	lines := []string{
		"STATUS_DIAGNOSTICS",
		fmt.Sprintf("views=%d", state.StatusViewCount),
		fmt.Sprintf("diagnostics=%d", state.StatusDiagnosticsCount),
		fmt.Sprintf("provider=%s", normalizeToken(selection.ProviderName)),
		fmt.Sprintf("model=%s", normalizeToken(selection.ModelName)),
		fmt.Sprintf("permission_mode=%s", modeString(state.PermissionMode)),
		fmt.Sprintf("loop_count=%d", len(loops)),
	}
	for i, loop := range loops {
		idx := i + 1
		lines = append(lines, fmt.Sprintf("loop.%d.area=%s", idx, normalizeToken(loop.Area)))
		lines = append(lines, fmt.Sprintf("loop.%d.state=%s", idx, normalizeToken(loop.State)))
		lines = append(lines, fmt.Sprintf("loop.%d.action=%s", idx, normalizeToken(loop.Action)))
		lines = append(lines, fmt.Sprintf("loop.%d.next=%s", idx, normalizeToken(loop.Next)))
	}
	return resultWithIntents(strings.Join(lines, "\n"), statusDiagnosticsIntents(state, loops)...)
}

func statusRuntimeSnapshot(state *RuntimeState) types.AgentRuntimeSnapshot {
	if state == nil {
		return types.AgentRuntimeSnapshot{}
	}
	runtime := state.Runtime
	selection := RuntimeSelectionTruth(state)
	runtime.ProviderName = selection.ProviderName
	runtime.Model = selection.ModelName
	runtime.ModelRef = selection.ModelRef
	runtime.LoggedIn = selection.LoggedIn
	runtime.ProviderReady = selection.ProviderReady
	if state.Agent != nil {
		agentRuntime := state.Agent.RuntimeSnapshot()
		if strings.TrimSpace(agentRuntime.ProviderName) != "" {
			runtime.ProviderName = agentRuntime.ProviderName
		}
		if strings.TrimSpace(agentRuntime.Model) != "" {
			runtime.Model = agentRuntime.Model
		}
		if strings.TrimSpace(agentRuntime.ModelRef) != "" {
			runtime.ModelRef = agentRuntime.ModelRef
		}
		if agentRuntime.Turns > runtime.Turns {
			runtime.Turns = agentRuntime.Turns
		}
		runtime.ToolInflight = agentRuntime.ToolInflight
		if agentRuntime.TasksTotal > runtime.TasksTotal {
			runtime.TasksTotal = agentRuntime.TasksTotal
		}
		if agentRuntime.TasksRunning > runtime.TasksRunning {
			runtime.TasksRunning = agentRuntime.TasksRunning
		}
		if agentRuntime.TasksCompleted > runtime.TasksCompleted {
			runtime.TasksCompleted = agentRuntime.TasksCompleted
		}
		if agentRuntime.TeamsTotal > runtime.TeamsTotal {
			runtime.TeamsTotal = agentRuntime.TeamsTotal
		}
		if agentRuntime.TeamsActive > runtime.TeamsActive {
			runtime.TeamsActive = agentRuntime.TeamsActive
		}
		if agentRuntime.LastStopReason != "" {
			runtime.LastStopReason = agentRuntime.LastStopReason
		}
	}
	if runtime.TasksTotal == 0 {
		runtime.TasksTotal = len(state.Tasks)
	}
	if runtime.TasksCompleted == 0 {
		runtime.TasksCompleted = state.TasksCompleted
	}
	return runtime
}

func statusWorkingDir(state *RuntimeState) string {
	if state == nil {
		return "."
	}
	if state.Agent != nil {
		if wd := strings.TrimSpace(state.Agent.WorkingDir()); wd != "" {
			return wd
		}
	}
	return "."
}

func statusModelCapabilitySummary(state *RuntimeState) string {
	if state == nil {
		return "-"
	}
	selection := RuntimeSelectionTruth(state)
	provider, model, ok := resolveModelForStatus(selection.ModelName, selection.ProviderName)
	if !ok {
		return "-"
	}
	summary, found := providers.LookupModelCapabilitySummary(provider, model)
	if !found {
		return "-"
	}
	return summary
}

// StatsCommand reports deterministic command counters.
type StatsCommand struct{}

func NewStatsCommand() *StatsCommand      { return &StatsCommand{} }
func (c *StatsCommand) Name() string      { return "stats" }
func (c *StatsCommand) Aliases() []string { return []string{"stat"} }
func (c *StatsCommand) Description() string {
	return "Show deterministic slash command counters"
}
func (c *StatsCommand) Usage() string { return "/stats" }

func (c *StatsCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if len(inv.Args) > 0 {
		return Result{}, fmt.Errorf("usage: /stats")
	}
	state := cmdCtx.State
	if state == nil {
		state = &RuntimeState{}
	}
	lines := []string{
		"STATS_REPORT",
		fmt.Sprintf("compact_count=%d", state.CompactCount),
		fmt.Sprintf("resume_count=%d", state.ResumeCount),
		fmt.Sprintf("copy_count=%d", state.CopyCount),
		fmt.Sprintf("clear_count=%d", state.ClearCount),
		fmt.Sprintf("branch_creates=%d", state.BranchCount),
		fmt.Sprintf("diff_updates=%d", state.DiffCount),
		fmt.Sprintf("memory_writes=%d", state.MemoryWrites),
		fmt.Sprintf("tasks_completed=%d", state.TasksCompleted),
	}
	return resultWithIntents(strings.Join(lines, "\n"), legacyOutputIntents(strings.Join(lines, "\n"))...), nil
}

// MemoryCommand provides deterministic local memory controls.
type MemoryCommand struct{}

func NewMemoryCommand() *MemoryCommand       { return &MemoryCommand{} }
func (c *MemoryCommand) Name() string        { return "memory" }
func (c *MemoryCommand) Aliases() []string   { return []string{"mem"} }
func (c *MemoryCommand) Description() string { return "Manage deterministic local memory entries" }
func (c *MemoryCommand) Usage() string {
	return "/memory [status|list|add <text>|remove <index>|clear]"
}

func (c *MemoryCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /memory [status|list|add <text>|remove <index>|clear]")
		}
		message := fmt.Sprintf("MEMORY_STATUS\ncount=%d\nwrites=%d\nremoves=%d\nclears=%d\nlast=%s", len(cmdCtx.State.MemoryEntries), cmdCtx.State.MemoryWrites, cmdCtx.State.MemoryRemoves, cmdCtx.State.MemoryClears, normalizeToken(cmdCtx.State.LastMemory))
		return resultWithIntents(message, memoryStatusIntents(cmdCtx.State)...), nil
	}

	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "list":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /memory [status|list|add <text>|remove <index>|clear]")
		}
		lines := []string{"MEMORY_LIST", fmt.Sprintf("count=%d", len(cmdCtx.State.MemoryEntries))}
		for i, entry := range cmdCtx.State.MemoryEntries {
			lines = append(lines, fmt.Sprintf("entry.%d=%s", i+1, normalizeToken(entry)))
		}
		message := strings.Join(lines, "\n")
		return resultWithIntents(message, memoryListIntents(cmdCtx.State.MemoryEntries)...), nil
	case "add":
		if len(inv.Args) < 2 {
			return Result{}, fmt.Errorf("usage: /memory [status|list|add <text>|remove <index>|clear]")
		}
		entry := strings.TrimSpace(strings.Join(inv.Args[1:], " "))
		if entry == "" {
			return Result{}, fmt.Errorf("usage: /memory [status|list|add <text>|remove <index>|clear]")
		}
		cmdCtx.State.MemoryEntries = append(cmdCtx.State.MemoryEntries, entry)
		cmdCtx.State.MemoryWrites++
		cmdCtx.State.LastMemory = entry
		message := fmt.Sprintf("MEMORY_ADD\nentry=%s\ncount=%d\nwrites=%d", normalizeToken(entry), len(cmdCtx.State.MemoryEntries), cmdCtx.State.MemoryWrites)
		return resultWithIntents(message, memoryMutationIntents("Memory added", entry, len(cmdCtx.State.MemoryEntries), cmdCtx.State.MemoryWrites)...), nil
	case "remove":
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf("usage: /memory [status|list|add <text>|remove <index>|clear]")
		}
		idx, err := parseNonNegativeInt(inv.Args[1])
		if err != nil || idx == 0 || idx > len(cmdCtx.State.MemoryEntries) {
			return Result{}, fmt.Errorf("usage: /memory [status|list|add <text>|remove <index>|clear]")
		}
		removed := cmdCtx.State.MemoryEntries[idx-1]
		cmdCtx.State.MemoryEntries = append(cmdCtx.State.MemoryEntries[:idx-1], cmdCtx.State.MemoryEntries[idx:]...)
		cmdCtx.State.MemoryRemoves++
		if len(cmdCtx.State.MemoryEntries) == 0 {
			cmdCtx.State.LastMemory = ""
		}
		message := fmt.Sprintf("MEMORY_REMOVE\nindex=%d\nentry=%s\ncount=%d\nremoves=%d", idx, normalizeToken(removed), len(cmdCtx.State.MemoryEntries), cmdCtx.State.MemoryRemoves)
		return resultWithIntents(message, memoryMutationIntents("Memory removed", removed, len(cmdCtx.State.MemoryEntries), cmdCtx.State.MemoryRemoves)...), nil
	case "clear":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /memory [status|list|add <text>|remove <index>|clear]")
		}
		cmdCtx.State.MemoryEntries = nil
		cmdCtx.State.LastMemory = ""
		cmdCtx.State.MemoryClears++
		message := fmt.Sprintf("MEMORY_CLEAR\ncount=0\nclears=%d", cmdCtx.State.MemoryClears)
		return resultWithIntents(message, memoryMutationIntents("Memory cleared", "", 0, cmdCtx.State.MemoryClears)...), nil
	default:
		return Result{}, fmt.Errorf("usage: /memory [status|list|add <text>|remove <index>|clear]")
	}
}

// PrivacySettingsCommand provides deterministic privacy controls.
type PrivacySettingsCommand struct{}

func NewPrivacySettingsCommand() *PrivacySettingsCommand { return &PrivacySettingsCommand{} }
func (c *PrivacySettingsCommand) Name() string           { return "privacy-settings" }
func (c *PrivacySettingsCommand) Aliases() []string      { return []string{"privacy"} }
func (c *PrivacySettingsCommand) Description() string {
	return "Show or set telemetry and training privacy flags"
}
func (c *PrivacySettingsCommand) Usage() string {
	return "/privacy-settings [show|status|list|set telemetry <on|off>|set training <on|off>]"
}

func (c *PrivacySettingsCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "show") || strings.EqualFold(inv.Args[0], "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /privacy-settings [show|status|list|set telemetry <on|off>|set training <on|off>]")
		}
		message := fmt.Sprintf("PRIVACY_SETTINGS\ntelemetry=%t\ntraining=%t\nupdates=%d", cmdCtx.State.PrivacyTelemetry, cmdCtx.State.PrivacyTraining, cmdCtx.State.PrivacyUpdates)
		return resultWithIntents(message, privacyStatusIntents(cmdCtx.State.PrivacyTelemetry, cmdCtx.State.PrivacyTraining, cmdCtx.State.PrivacyUpdates)...), nil
	}
	if len(inv.Args) == 1 && strings.EqualFold(inv.Args[0], "list") {
		message := "PRIVACY_FIELDS\ncount=2\nfield.1=telemetry\nfield.2=training"
		return resultWithIntents(message, privacyFieldsIntents()...), nil
	}
	if len(inv.Args) != 3 || !strings.EqualFold(inv.Args[0], "set") {
		return Result{}, fmt.Errorf("usage: /privacy-settings [show|status|list|set telemetry <on|off>|set training <on|off>]")
	}
	field := strings.ToLower(strings.TrimSpace(inv.Args[1]))
	value := strings.ToLower(strings.TrimSpace(inv.Args[2]))
	updated := false
	var enabled bool
	switch value {
	case "on", "true", "enabled":
		enabled = true
	case "off", "false", "disabled":
		enabled = false
	default:
		return Result{}, fmt.Errorf("usage: /privacy-settings [show|set telemetry <on|off>|set training <on|off>]")
	}
	switch field {
	case "telemetry":
		updated = cmdCtx.State.PrivacyTelemetry != enabled
		cmdCtx.State.PrivacyTelemetry = enabled
	case "training":
		updated = cmdCtx.State.PrivacyTraining != enabled
		cmdCtx.State.PrivacyTraining = enabled
	default:
		return Result{}, fmt.Errorf("usage: /privacy-settings [show|status|list|set telemetry <on|off>|set training <on|off>]")
	}
	if updated {
		cmdCtx.State.PrivacyUpdates++
	}
	message := fmt.Sprintf("PRIVACY_SET\nfield=%s\nenabled=%t\nupdated=%t\nupdates=%d", field, enabled, updated, cmdCtx.State.PrivacyUpdates)
	return resultWithIntents(message, privacySetIntents(field, enabled, updated, cmdCtx.State.PrivacyUpdates)...), nil
}

// UpgradeCommand provides deterministic upgrade intent behavior.
type UpgradeCommand struct{}

func NewUpgradeCommand() *UpgradeCommand      { return &UpgradeCommand{} }
func (c *UpgradeCommand) Name() string        { return "upgrade" }
func (c *UpgradeCommand) Aliases() []string   { return []string{"plan-upgrade"} }
func (c *UpgradeCommand) Description() string { return "Request account plan upgrade guidance" }
func (c *UpgradeCommand) Usage() string       { return "/upgrade [status|max|pro|team|enterprise]" }

func (c *UpgradeCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 1 && strings.EqualFold(inv.Args[0], "status") {
		message := fmt.Sprintf("UPGRADE_STATUS\nrequested=%t\ncount=%d\nlast_plan=%s", cmdCtx.State.UpgradeRequested, cmdCtx.State.UpgradeCount, normalizeToken(cmdCtx.State.LastUpgradePlan))
		return resultWithIntents(message, upgradeStatusIntents(cmdCtx.State.UpgradeRequested, cmdCtx.State.UpgradeCount, cmdCtx.State.LastUpgradePlan)...), nil
	}
	if len(inv.Args) > 1 {
		return Result{}, fmt.Errorf("usage: /upgrade [status|max|pro|team|enterprise]")
	}
	plan := "max"
	if len(inv.Args) == 1 {
		plan = strings.ToLower(strings.TrimSpace(inv.Args[0]))
	}
	switch plan {
	case "max", "pro", "team", "enterprise":
		cmdCtx.State.UpgradeRequested = true
		cmdCtx.State.UpgradeCount++
		cmdCtx.State.LastUpgradePlan = plan
		message := fmt.Sprintf("UPGRADE_REQUEST\nplan=%s\nrequested=true\ncount=%d\nnext=complete_upgrade_in_provider_portal", plan, cmdCtx.State.UpgradeCount)
		return resultWithIntents(message, upgradeRequestIntents(plan, cmdCtx.State.UpgradeCount)...), nil
	default:
		return Result{}, fmt.Errorf("usage: /upgrade [status|max|pro|team|enterprise]")
	}
}

// TerminalSetupCommand provides deterministic setup tracking.
type TerminalSetupCommand struct{}

func NewTerminalSetupCommand() *TerminalSetupCommand { return &TerminalSetupCommand{} }
func (c *TerminalSetupCommand) Name() string         { return "terminal-setup" }
func (c *TerminalSetupCommand) Aliases() []string    { return []string{"terminalSetup"} }
func (c *TerminalSetupCommand) Description() string {
	return "Configure terminal newline keybinding guidance"
}
func (c *TerminalSetupCommand) Usage() string {
	return "/terminal-setup [status|detect|apply [profile]]"
}
func (c *TerminalSetupCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "detect") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /terminal-setup [status|detect|apply [profile]]")
		}
		profile, source, supported, hints := detectTerminalProfile()
		cmdCtx.State.TerminalDetectedProfile = profile
		cmdCtx.State.TerminalDetectSource = source
		cmdCtx.State.TerminalSetupHints = append([]string(nil), hints...)
		lines := []string{
			"TERMINAL_SETUP_DETECT",
			fmt.Sprintf("profile=%s", normalizeToken(profile)),
			fmt.Sprintf("source=%s", normalizeToken(source)),
			fmt.Sprintf("supported=%t", supported),
			fmt.Sprintf("recommended=%t", supported),
			fmt.Sprintf("hint_count=%d", len(hints)),
		}
		for i, hint := range hints {
			lines = append(lines, fmt.Sprintf("hint.%d=%s", i+1, normalizeToken(hint)))
		}
		return resultWithIntents(strings.Join(lines, "\n"), terminalSetupDetectIntents(profile, source, supported, hints)...), nil
	}
	if len(inv.Args) == 1 && strings.EqualFold(inv.Args[0], "status") {
		lines := []string{
			"TERMINAL_SETUP_STATUS",
			fmt.Sprintf("configured=%t", cmdCtx.State.TerminalConfigured),
			fmt.Sprintf("count=%d", cmdCtx.State.TerminalSetupCount),
			fmt.Sprintf("profile=%s", normalizeToken(cmdCtx.State.TerminalProfile)),
			fmt.Sprintf("detected_profile=%s", normalizeToken(cmdCtx.State.TerminalDetectedProfile)),
			fmt.Sprintf("detect_source=%s", normalizeToken(cmdCtx.State.TerminalDetectSource)),
			fmt.Sprintf("hint_count=%d", len(cmdCtx.State.TerminalSetupHints)),
		}
		for i, hint := range cmdCtx.State.TerminalSetupHints {
			lines = append(lines, fmt.Sprintf("hint.%d=%s", i+1, normalizeToken(hint)))
		}
		return resultWithIntents(strings.Join(lines, "\n"), terminalSetupStatusIntents(cmdCtx.State)...), nil
	}
	if !strings.EqualFold(inv.Args[0], "apply") || len(inv.Args) > 2 {
		return Result{}, fmt.Errorf("usage: /terminal-setup [status|detect|apply [profile]]")
	}
	profile := "vscode"
	if len(inv.Args) == 2 {
		profile = strings.ToLower(strings.TrimSpace(inv.Args[1]))
		if profile == "" {
			return Result{}, fmt.Errorf("usage: /terminal-setup [status|detect|apply [profile]]")
		}
	}
	if len(inv.Args) == 1 {
		detected, source, _, hints := detectTerminalProfile()
		if strings.TrimSpace(detected) != "" {
			profile = detected
		}
		cmdCtx.State.TerminalDetectedProfile = detected
		cmdCtx.State.TerminalDetectSource = source
		cmdCtx.State.TerminalSetupHints = append([]string(nil), hints...)
	}
	cmdCtx.State.TerminalConfigured = true
	cmdCtx.State.TerminalSetupCount++
	cmdCtx.State.TerminalProfile = profile
	return resultWithIntents(fmt.Sprintf("TERMINAL_SETUP_APPLY\nprofile=%s\nconfigured=true\ncount=%d", normalizeToken(profile), cmdCtx.State.TerminalSetupCount), terminalSetupApplyIntents(profile, cmdCtx.State.TerminalSetupCount)...), nil
}

func detectTerminalProfile() (profile string, source string, supported bool, hints []string) {
	termProgram := strings.ToLower(strings.TrimSpace(os.Getenv("TERM_PROGRAM")))
	term := strings.ToLower(strings.TrimSpace(os.Getenv("TERM")))

	switch {
	case strings.Contains(termProgram, "ghostty") || strings.Contains(term, "ghostty"):
		return "ghostty", terminalSource(termProgram, term), true, []string{
			"Ghostty: map Shift+Enter (or Alt+Enter) to send newline in terminal input",
			"Use /terminal-setup apply ghostty after updating keybinds",
		}
	case strings.Contains(termProgram, "iterm"):
		return "iterm", terminalSource(termProgram, term), true, []string{
			"iTerm2: assign Send Hex Code 0x0a to Shift+Enter for newline",
			"Restart terminal tab/session after profile changes",
		}
	case strings.Contains(termProgram, "vscode"):
		return "vscode", terminalSource(termProgram, term), true, []string{
			"VS Code: keep shell keybindings for Shift+Enter newline behavior",
			"Use /terminal-setup apply vscode to record active profile",
		}
	default:
		return "generic", terminalSource(termProgram, term), false, []string{
			"Detect profile manually with /terminal-setup apply <profile>",
			"Known profiles include ghostty, iterm, vscode, and zed",
		}
	}
}

func terminalSource(termProgram, term string) string {
	if strings.TrimSpace(termProgram) != "" {
		return "term_program"
	}
	if strings.TrimSpace(term) != "" {
		return "term"
	}
	return "default"
}

func detectEditorFromEnvironment() (editor string, source string) {
	if raw := strings.TrimSpace(os.Getenv("VISUAL")); raw != "" {
		return normalizeEditorName(raw), "visual"
	}
	if raw := strings.TrimSpace(os.Getenv("EDITOR")); raw != "" {
		return normalizeEditorName(raw), "editor"
	}
	termProgram := strings.ToLower(strings.TrimSpace(os.Getenv("TERM_PROGRAM")))
	switch {
	case strings.Contains(termProgram, "vscode"):
		return "vscode", "term_program"
	case strings.Contains(termProgram, "iterm"):
		return "cursor", "term_program"
	default:
		return "vscode", "default"
	}
}

func editorFromStateOrDetected(state *RuntimeState) string {
	if state == nil {
		return ""
	}
	if strings.TrimSpace(state.IDEEditor) != "" {
		return normalizeEditorName(state.IDEEditor)
	}
	if strings.TrimSpace(state.IDEDetectedEditor) != "" {
		return normalizeEditorName(state.IDEDetectedEditor)
	}
	return ""
}

func normalizeEditorName(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return ""
	}
	name := strings.ToLower(filepath.Base(parts[0]))
	name = strings.TrimSuffix(name, ".exe")
	switch name {
	case "code", "code-insiders", "vscode":
		return "vscode"
	case "cursor":
		return "cursor"
	case "windsurf":
		return "windsurf"
	case "zed":
		return "zed"
	case "idea", "goland", "pycharm", "webstorm", "clion":
		return "jetbrains"
	case "nvim", "vim":
		return "vim"
	default:
		return name
	}
}

func editorOpenHint(editor, target string) string {
	editor = normalizeEditorName(editor)
	target = strings.TrimSpace(target)
	if target == "" {
		target = "."
	}
	switch editor {
	case "vscode":
		return fmt.Sprintf("code %s", target)
	case "cursor":
		return fmt.Sprintf("cursor %s", target)
	case "windsurf":
		return fmt.Sprintf("windsurf %s", target)
	case "zed":
		return fmt.Sprintf("zed %s", target)
	case "jetbrains":
		return fmt.Sprintf("idea %s", target)
	case "vim":
		return fmt.Sprintf("nvim %s", target)
	default:
		return fmt.Sprintf("%s %s", normalizeToken(editor), target)
	}
}

func ideHints(editor string) []string {
	editor = normalizeEditorName(editor)
	switch editor {
	case "vscode", "cursor", "windsurf":
		return []string{
			"Use /ide open to print a launch command for current workspace",
			"Install Claude Code extension/plugin in the detected editor",
		}
	case "jetbrains":
		return []string{
			"Ensure JetBrains plugin is installed and IDE is restarted",
			"Use /ide open and launch via IDE CLI launcher",
		}
	default:
		return []string{
			"Use /ide set-editor <name> to pin a preferred editor",
			"Use /ide open [path] for provider-agnostic launch hints",
		}
	}
}

// ReleaseNotesCommand provides deterministic release notes views.
type ReleaseNotesCommand struct{}

func NewReleaseNotesCommand() *ReleaseNotesCommand { return &ReleaseNotesCommand{} }
func (c *ReleaseNotesCommand) Name() string        { return "release-notes" }
func (c *ReleaseNotesCommand) Aliases() []string   { return []string{"changelog"} }
func (c *ReleaseNotesCommand) Description() string {
	return "Show deterministic release note snapshots"
}
func (c *ReleaseNotesCommand) Usage() string { return "/release-notes [latest|list|status]" }
func (c *ReleaseNotesCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "latest") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /release-notes [latest|list|status]")
		}
		cmdCtx.State.ReleaseNotesSeen++
		cmdCtx.State.LastReleaseVersion = "v1.0.0"
		return resultWithIntents(fmt.Sprintf("RELEASE_NOTES_LATEST\nversion=v1.0.0\nentry_count=2\nentry.1=Massive command parity improvements\nentry.2=Deterministic slash command output contracts\nseen=%d", cmdCtx.State.ReleaseNotesSeen), releaseNotesLatestIntents("v1.0.0", cmdCtx.State.ReleaseNotesSeen)...), nil
	}
	if len(inv.Args) == 1 && strings.EqualFold(inv.Args[0], "list") {
		cmdCtx.State.ReleaseNotesSeen++
		return resultWithIntents(fmt.Sprintf("RELEASE_NOTES_LIST\ncount=3\nversion.1=v1.0.0\nversion.2=v0.9.0\nversion.3=v0.8.0\nseen=%d", cmdCtx.State.ReleaseNotesSeen), releaseNotesListIntents([]string{"v1.0.0", "v0.9.0", "v0.8.0"}, cmdCtx.State.ReleaseNotesSeen)...), nil
	}
	if len(inv.Args) == 1 && strings.EqualFold(inv.Args[0], "status") {
		return resultWithIntents(fmt.Sprintf("RELEASE_NOTES_STATUS\nseen=%d\nlast_version=%s", cmdCtx.State.ReleaseNotesSeen, normalizeToken(cmdCtx.State.LastReleaseVersion)), releaseNotesStatusIntents(cmdCtx.State.ReleaseNotesSeen, cmdCtx.State.LastReleaseVersion)...), nil
	}
	return Result{}, fmt.Errorf("usage: /release-notes [latest|list|status]")
}

// InstallGitHubAppCommand provides deterministic install tracking.
type InstallGitHubAppCommand struct{}

func NewInstallGitHubAppCommand() *InstallGitHubAppCommand { return &InstallGitHubAppCommand{} }
func (c *InstallGitHubAppCommand) Name() string            { return "install-github-app" }
func (c *InstallGitHubAppCommand) Aliases() []string       { return []string{"github-app"} }
func (c *InstallGitHubAppCommand) Description() string     { return "Track GitHub app install flow intent" }
func (c *InstallGitHubAppCommand) Usage() string {
	return "/install-github-app [status|repo <owner/repo>|start [owner/repo]]"
}
func (c *InstallGitHubAppCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "start") {
		if len(inv.Args) > 2 {
			return Result{}, fmt.Errorf("usage: /install-github-app [status|repo <owner/repo>|start [owner/repo]]")
		}
		repo := strings.TrimSpace(cmdCtx.State.LastGitHubRepo)
		if len(inv.Args) == 2 {
			repo = strings.TrimSpace(inv.Args[1])
		}
		if repo == "" {
			repo = "owner/repo"
		}
		cmdCtx.State.LastGitHubRepo = repo
		cmdCtx.State.GitHubAppInstalls++
		return resultWithIntents(fmt.Sprintf("INSTALL_GITHUB_APP_START\nrepo=%s\ncount=%d\nnext=run_gh_auth_and_repo_setup", normalizeToken(repo), cmdCtx.State.GitHubAppInstalls), installFlowIntents("Install GitHub app", detailRow("Repo", defaultDash(repo), "info", "Target repository."), detailRow("Count", fmt.Sprintf("%d", cmdCtx.State.GitHubAppInstalls), "info", "Install attempts."), detailRow("Next", "run_gh_auth_and_repo_setup", "hint", "Suggested next step."))...), nil
	}
	if len(inv.Args) == 1 && strings.EqualFold(inv.Args[0], "status") {
		return resultWithIntents(fmt.Sprintf("INSTALL_GITHUB_APP_STATUS\ncount=%d\nlast_repo=%s", cmdCtx.State.GitHubAppInstalls, normalizeToken(cmdCtx.State.LastGitHubRepo)), installFlowIntents("Install GitHub app status", detailRow("Count", fmt.Sprintf("%d", cmdCtx.State.GitHubAppInstalls), "info", "Install attempts."), detailRow("Last repo", defaultDash(cmdCtx.State.LastGitHubRepo), "info", "Most recent target repository."))...), nil
	}
	if len(inv.Args) == 2 && strings.EqualFold(inv.Args[0], "repo") {
		repo := strings.TrimSpace(inv.Args[1])
		if repo == "" {
			return Result{}, fmt.Errorf("usage: /install-github-app [status|repo <owner/repo>|start [owner/repo]]")
		}
		cmdCtx.State.LastGitHubRepo = repo
		return resultWithIntents(fmt.Sprintf("INSTALL_GITHUB_APP_REPO\nrepo=%s", normalizeToken(repo)), installFlowIntents("Install GitHub app repo", detailRow("Repo", defaultDash(repo), "set", "Configured repository target."))...), nil
	}
	return Result{}, fmt.Errorf("usage: /install-github-app [status|repo <owner/repo>|start [owner/repo]]")
}

// InstallSlackAppCommand provides deterministic install tracking.
type InstallSlackAppCommand struct{}

func NewInstallSlackAppCommand() *InstallSlackAppCommand { return &InstallSlackAppCommand{} }
func (c *InstallSlackAppCommand) Name() string           { return "install-slack-app" }
func (c *InstallSlackAppCommand) Aliases() []string      { return []string{"slack-app"} }
func (c *InstallSlackAppCommand) Description() string    { return "Track Slack app install flow intent" }
func (c *InstallSlackAppCommand) Usage() string          { return "/install-slack-app [status|start]" }
func (c *InstallSlackAppCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "start") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /install-slack-app [status|start]")
		}
		cmdCtx.State.SlackAppInstalls++
		return resultWithIntents(fmt.Sprintf("INSTALL_SLACK_APP_START\ncount=%d\nnext=open_slack_marketplace_link", cmdCtx.State.SlackAppInstalls), installFlowIntents("Install Slack app", detailRow("Count", fmt.Sprintf("%d", cmdCtx.State.SlackAppInstalls), "info", "Install attempts."), detailRow("Next", "open_slack_marketplace_link", "hint", "Suggested next step."))...), nil
	}
	if len(inv.Args) == 1 && strings.EqualFold(inv.Args[0], "status") {
		return resultWithIntents(fmt.Sprintf("INSTALL_SLACK_APP_STATUS\ncount=%d", cmdCtx.State.SlackAppInstalls), installFlowIntents("Install Slack app status", detailRow("Count", fmt.Sprintf("%d", cmdCtx.State.SlackAppInstalls), "info", "Install attempts."))...), nil
	}
	return Result{}, fmt.Errorf("usage: /install-slack-app [status|start]")
}

// FeedbackCommand provides deterministic feedback capture.
type FeedbackCommand struct{}

func NewFeedbackCommand() *FeedbackCommand     { return &FeedbackCommand{} }
func (c *FeedbackCommand) Name() string        { return "feedback" }
func (c *FeedbackCommand) Aliases() []string   { return []string{"bug"} }
func (c *FeedbackCommand) Description() string { return "Submit deterministic feedback payload" }
func (c *FeedbackCommand) Usage() string       { return "/feedback [status|submit <message>]" }
func (c *FeedbackCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /feedback [status|submit <message>]")
		}
		return resultWithIntents(fmt.Sprintf("FEEDBACK_STATUS\ncount=%d\nlast=%s", cmdCtx.State.FeedbackCount, normalizeToken(cmdCtx.State.LastFeedback)), feedbackStatusIntents(cmdCtx.State.FeedbackCount, cmdCtx.State.LastFeedback)...), nil
	}
	if !strings.EqualFold(inv.Args[0], "submit") || len(inv.Args) < 2 {
		return Result{}, fmt.Errorf("usage: /feedback [status|submit <message>]")
	}
	message := strings.TrimSpace(strings.Join(inv.Args[1:], " "))
	if message == "" {
		return Result{}, fmt.Errorf("usage: /feedback [status|submit <message>]")
	}
	cmdCtx.State.LastFeedback = message
	cmdCtx.State.FeedbackCount++
	return resultWithIntents(fmt.Sprintf("FEEDBACK_SUBMIT\nmessage=%s\ncount=%d", normalizeToken(message), cmdCtx.State.FeedbackCount), feedbackSubmitIntents(message, cmdCtx.State.FeedbackCount)...), nil
}

// HooksCommand provides deterministic hook controls.
type HooksCommand struct{}

func NewHooksCommand() *HooksCommand      { return &HooksCommand{} }
func (c *HooksCommand) Name() string      { return "hooks" }
func (c *HooksCommand) Aliases() []string { return nil }
func (c *HooksCommand) Description() string {
	return "Show or set pre/post hook toggles"
}
func (c *HooksCommand) Usage() string {
	return "/hooks [status|enable <pre|post|all>|disable <pre|post|all>]"
}

func (c *HooksCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /hooks [status|enable <pre|post|all>|disable <pre|post|all>]")
		}
		return resultWithIntents(fmt.Sprintf("HOOKS_STATUS\npre=%t\npost=%t", cmdCtx.State.HookPreEnabled, cmdCtx.State.HookPostEnabled), hooksStatusIntents(cmdCtx.State.HookPreEnabled, cmdCtx.State.HookPostEnabled)...), nil
	}
	if len(inv.Args) != 2 {
		return Result{}, fmt.Errorf("usage: /hooks [status|enable <pre|post|all>|disable <pre|post|all>]")
	}
	action := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	target := strings.ToLower(strings.TrimSpace(inv.Args[1]))
	enabled := false
	switch action {
	case "enable":
		enabled = true
	case "disable":
		enabled = false
	default:
		return Result{}, fmt.Errorf("usage: /hooks [status|enable <pre|post|all>|disable <pre|post|all>]")
	}
	switch target {
	case "pre":
		cmdCtx.State.HookPreEnabled = enabled
	case "post":
		cmdCtx.State.HookPostEnabled = enabled
	case "all":
		cmdCtx.State.HookPreEnabled = enabled
		cmdCtx.State.HookPostEnabled = enabled
	default:
		return Result{}, fmt.Errorf("usage: /hooks [status|enable <pre|post|all>|disable <pre|post|all>]")
	}
	return resultWithIntents(fmt.Sprintf("HOOKS_SET\naction=%s\ntarget=%s\npre=%t\npost=%t", action, target, cmdCtx.State.HookPreEnabled, cmdCtx.State.HookPostEnabled), hooksSetIntents(action, target, cmdCtx.State.HookPreEnabled, cmdCtx.State.HookPostEnabled)...), nil
}

// SandboxCommand provides deterministic sandbox mode controls.
type SandboxCommand struct{}

func NewSandboxCommand() *SandboxCommand      { return &SandboxCommand{} }
func (c *SandboxCommand) Name() string        { return "sandbox" }
func (c *SandboxCommand) Aliases() []string   { return nil }
func (c *SandboxCommand) Description() string { return "Show or set sandbox mode" }
func (c *SandboxCommand) Usage() string {
	return sandboxUsage
}

const sandboxUsage = "usage: /sandbox [status|check|repair|set <read-only|workspace-write|danger-full-access>|mode <read-only|workspace-write|danger-full-access>|lock [status|on|off]|exclude [list|add <pattern>|remove <pattern>|clear]|exclude <pattern>]"

var sandboxLookPath = exec.LookPath

func (c *SandboxCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	settings, err := loadSandboxSettings(cmdCtx.State)
	if err != nil {
		return Result{}, fmt.Errorf("load sandbox settings: %w", err)
	}
	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf(sandboxUsage)
		}
		return renderSandboxStatus(settings, cmdCtx.State), nil
	}

	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "check":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf(sandboxUsage)
		}
		return renderSandboxCheck(settings), nil
	case "repair":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf(sandboxUsage)
		}
		before := permissions.EffectiveSandboxPolicySummary(settings)
		settings = permissions.NormalizeSandboxSettings(settings)
		if strings.TrimSpace(settings.Mode) == "danger-full-access" {
			settings.Mode = "workspace-write"
		}
		if err := saveSandboxSettings(cmdCtx.State, settings); err != nil {
			return Result{}, fmt.Errorf("persist sandbox settings: %w", err)
		}
		after := permissions.EffectiveSandboxPolicySummary(settings)
		changed := before.Mode != after.Mode || before.WorkspaceLocked != after.WorkspaceLocked || before.ExcludedCount != after.ExcludedCount
		return resultWithIntents(fmt.Sprintf("SANDBOX_REPAIR\nchanged=%t\nmode=%s\nworkspace_locked=%t\nexcluded_count=%d", changed, normalizeToken(after.Mode), after.WorkspaceLocked, after.ExcludedCount), legacyOutputIntents(fmt.Sprintf("SANDBOX_REPAIR\nchanged=%t\nmode=%s\nworkspace_locked=%t\nexcluded_count=%d", changed, normalizeToken(after.Mode), after.WorkspaceLocked, after.ExcludedCount))...), nil
	case "set", "mode":
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf(sandboxUsage)
		}
		mode := strings.ToLower(strings.TrimSpace(inv.Args[1]))
		if mode != "read-only" && mode != "workspace-write" && mode != "danger-full-access" {
			return Result{}, fmt.Errorf(sandboxUsage)
		}
		settings.Mode = mode
		if err := saveSandboxSettings(cmdCtx.State, settings); err != nil {
			return Result{}, fmt.Errorf("persist sandbox settings: %w", err)
		}
		return resultWithIntents(fmt.Sprintf("SANDBOX_SET\nmode=%s", mode), legacyOutputIntents(fmt.Sprintf("SANDBOX_SET\nmode=%s", mode))...), nil
	case "lock":
		if len(inv.Args) == 1 || (len(inv.Args) == 2 && strings.EqualFold(strings.TrimSpace(inv.Args[1]), "status")) {
			return resultWithIntents(fmt.Sprintf("SANDBOX_LOCK\nworkspace_locked=%t", settings.WorkspaceLocked), legacyOutputIntents(fmt.Sprintf("SANDBOX_LOCK\nworkspace_locked=%t", settings.WorkspaceLocked))...), nil
		}
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf(sandboxUsage)
		}
		value := strings.ToLower(strings.TrimSpace(inv.Args[1]))
		switch value {
		case "on", "true", "1":
			settings.WorkspaceLocked = true
		case "off", "false", "0":
			settings.WorkspaceLocked = false
		default:
			return Result{}, fmt.Errorf(sandboxUsage)
		}
		if err := saveSandboxSettings(cmdCtx.State, settings); err != nil {
			return Result{}, fmt.Errorf("persist sandbox settings: %w", err)
		}
		return resultWithIntents(fmt.Sprintf("SANDBOX_LOCK_SET\nworkspace_locked=%t", settings.WorkspaceLocked), legacyOutputIntents(fmt.Sprintf("SANDBOX_LOCK_SET\nworkspace_locked=%t", settings.WorkspaceLocked))...), nil
	case "exclude":
		return c.executeSandboxExclude(cmdCtx.State, settings, inv.Args[1:])
	default:
		return Result{}, fmt.Errorf(sandboxUsage)
	}
}

func (c *SandboxCommand) executeSandboxExclude(state *RuntimeState, settings permissions.SandboxSettings, args []string) (Result, error) {
	if len(args) == 0 || strings.EqualFold(strings.TrimSpace(args[0]), "list") {
		if len(args) > 1 {
			return Result{}, fmt.Errorf(sandboxUsage)
		}
		return renderSandboxExcludeList(settings.ExcludedCommands), nil
	}

	sub := strings.ToLower(strings.TrimSpace(args[0]))
	switch sub {
	case "add":
		pattern := strings.TrimSpace(strings.Join(args[1:], " "))
		if pattern == "" {
			return Result{}, fmt.Errorf(sandboxUsage)
		}
		pattern = trimQuoted(pattern)
		before := len(settings.ExcludedCommands)
		settings.ExcludedCommands = uniqueSortedStrings(append(settings.ExcludedCommands, pattern))
		added := len(settings.ExcludedCommands) > before
		if err := saveSandboxSettings(state, settings); err != nil {
			return Result{}, fmt.Errorf("persist sandbox settings: %w", err)
		}
		return resultWithIntents(fmt.Sprintf("SANDBOX_EXCLUDE_ADD\npattern=%s\nadded=%t\ncount=%d", normalizeToken(pattern), added, len(settings.ExcludedCommands)), legacyOutputIntents(fmt.Sprintf("SANDBOX_EXCLUDE_ADD\npattern=%s\nadded=%t\ncount=%d", normalizeToken(pattern), added, len(settings.ExcludedCommands)))...), nil
	case "remove":
		pattern := strings.TrimSpace(strings.Join(args[1:], " "))
		if pattern == "" {
			return Result{}, fmt.Errorf(sandboxUsage)
		}
		pattern = trimQuoted(pattern)
		removed := false
		next := make([]string, 0, len(settings.ExcludedCommands))
		for _, existing := range settings.ExcludedCommands {
			if strings.TrimSpace(existing) == pattern {
				removed = true
				continue
			}
			next = append(next, existing)
		}
		settings.ExcludedCommands = uniqueSortedStrings(next)
		if err := saveSandboxSettings(state, settings); err != nil {
			return Result{}, fmt.Errorf("persist sandbox settings: %w", err)
		}
		return resultWithIntents(fmt.Sprintf("SANDBOX_EXCLUDE_REMOVE\npattern=%s\nremoved=%t\ncount=%d", normalizeToken(pattern), removed, len(settings.ExcludedCommands)), legacyOutputIntents(fmt.Sprintf("SANDBOX_EXCLUDE_REMOVE\npattern=%s\nremoved=%t\ncount=%d", normalizeToken(pattern), removed, len(settings.ExcludedCommands)))...), nil
	case "clear":
		removed := len(settings.ExcludedCommands)
		settings.ExcludedCommands = nil
		if err := saveSandboxSettings(state, settings); err != nil {
			return Result{}, fmt.Errorf("persist sandbox settings: %w", err)
		}
		return resultWithIntents(fmt.Sprintf("SANDBOX_EXCLUDE_CLEAR\nremoved=%d\ncount=0", removed), legacyOutputIntents(fmt.Sprintf("SANDBOX_EXCLUDE_CLEAR\nremoved=%d\ncount=0", removed))...), nil
	default:
		pattern := strings.TrimSpace(strings.Join(args, " "))
		if pattern == "" {
			return Result{}, fmt.Errorf(sandboxUsage)
		}
		pattern = trimQuoted(pattern)
		before := len(settings.ExcludedCommands)
		settings.ExcludedCommands = uniqueSortedStrings(append(settings.ExcludedCommands, pattern))
		added := len(settings.ExcludedCommands) > before
		if err := saveSandboxSettings(state, settings); err != nil {
			return Result{}, fmt.Errorf("persist sandbox settings: %w", err)
		}
		return resultWithIntents(fmt.Sprintf("SANDBOX_EXCLUDE_ADD\npattern=%s\nadded=%t\ncount=%d", normalizeToken(pattern), added, len(settings.ExcludedCommands)), legacyOutputIntents(fmt.Sprintf("SANDBOX_EXCLUDE_ADD\npattern=%s\nadded=%t\ncount=%d", normalizeToken(pattern), added, len(settings.ExcludedCommands)))...), nil
	}
}

func renderSandboxStatus(settings permissions.SandboxSettings, state *RuntimeState) Result {
	policy := permissions.EffectiveSandboxPolicySummary(settings)
	mode := normalizeToken(policy.Mode)
	diagnostics := gatherSandboxDiagnostics(settings)
	lines := []string{
		"SANDBOX_STATUS",
		fmt.Sprintf("mode=%s", mode),
		fmt.Sprintf("workspace_locked=%t", policy.WorkspaceLocked),
		fmt.Sprintf("excluded_count=%d", policy.ExcludedCount),
		fmt.Sprintf("availability.shell=%s", diagnostics.ShellStatus),
		fmt.Sprintf("availability.shell.path=%s", normalizeToken(diagnostics.ShellPath)),
		fmt.Sprintf("dependency.status=%s", diagnostics.DependencyStatus),
		fmt.Sprintf("dependency.available=%d", diagnostics.DependencyAvailable),
		fmt.Sprintf("dependency.missing=%d", diagnostics.DependencyMissing),
		fmt.Sprintf("policy.effective.mode=%s", mode),
		fmt.Sprintf("policy.effective.alias=%s", policy.ModeAlias),
		fmt.Sprintf("policy.effective.workspace_locked=%t", policy.WorkspaceLocked),
		fmt.Sprintf("policy.effective.excluded_count=%d", policy.ExcludedCount),
		fmt.Sprintf("scope.working_dir=%s", normalizeToken(statusWorkingDir(state))),
		fmt.Sprintf("scope.workspace_root=%s", normalizeToken(statusWorkingDir(state))),
		fmt.Sprintf("scope.path_scope=%s", normalizeToken("workspace")),
		fmt.Sprintf("scope.restrictions=%s", normalizeToken(sandboxScopeRestrictionLabel(policy))),
	}
	return resultWithIntents(strings.Join(lines, "\n"), legacyOutputIntents(strings.Join(lines, "\n"))...)
}

type sandboxDiagnostics struct {
	ShellStatus         string
	ShellPath           string
	DependencyStatus    string
	DependencyAvailable int
	DependencyMissing   int
	Checks              []sandboxPreflightCheck
}

type sandboxPreflightCheck struct {
	ID     string
	Status string
	Detail string
}

func sandboxScopeRestrictionLabel(policy permissions.SandboxPolicySummary) string {
	if strings.EqualFold(strings.TrimSpace(policy.Mode), "read-only") {
		return "read_only_filesystem"
	}
	if policy.WorkspaceLocked {
		return "workspace_locked"
	}
	if strings.EqualFold(strings.TrimSpace(policy.Mode), "danger-full-access") {
		return "full_access"
	}
	return "workspace_write"
}

func gatherSandboxDiagnostics(settings permissions.SandboxSettings) sandboxDiagnostics {
	bashPath, bashErr := sandboxLookPath("bash")
	shPath, shErr := sandboxLookPath("sh")
	gitPath, gitErr := sandboxLookPath("git")

	shellStatus := "missing"
	shellPath := "-"
	shellDetail := "no shell executable found (sh/bash)"
	if shErr == nil {
		shellStatus = "available"
		shellPath = shPath
		shellDetail = fmt.Sprintf("sh available at %s", shPath)
	} else if bashErr == nil {
		shellStatus = "available"
		shellPath = bashPath
		shellDetail = fmt.Sprintf("bash available at %s", bashPath)
	}

	gitStatus := "warn"
	gitDetail := "git binary not found"
	if gitErr == nil {
		gitStatus = "ok"
		gitDetail = fmt.Sprintf("git available at %s", gitPath)
	}

	policySummary := permissions.EffectiveSandboxPolicySummary(settings)
	policyStatus := "ok"
	policyDetail := fmt.Sprintf("mode=%s alias=%s", normalizeToken(policySummary.Mode), policySummary.ModeAlias)
	if strings.EqualFold(strings.TrimSpace(settings.Mode), "danger-full-access") {
		policyStatus = "warn"
		policyDetail = "mode=danger-full-access alias=permissive"
	}

	lockStatus := "ok"
	lockDetail := "workspace lock disabled"
	if settings.WorkspaceLocked {
		lockDetail = "workspace lock enabled"
	}

	excludeStatus := "ok"
	excludeDetail := fmt.Sprintf("excluded command patterns=%d", len(settings.ExcludedCommands))

	checks := []sandboxPreflightCheck{
		{ID: "shell_available", Status: ternaryStatus(shellStatus == "available"), Detail: shellDetail},
		{ID: "dependency_git", Status: gitStatus, Detail: gitDetail},
		{ID: "policy_mode", Status: policyStatus, Detail: policyDetail},
		{ID: "policy_workspace_lock", Status: lockStatus, Detail: lockDetail},
		{ID: "policy_excludes", Status: excludeStatus, Detail: excludeDetail},
	}

	available := 0
	missing := 0
	for _, check := range checks {
		if check.Status == "ok" {
			available++
		} else {
			missing++
		}
	}

	dependencyStatus := "healthy"
	if missing > 0 {
		dependencyStatus = "degraded"
	}

	return sandboxDiagnostics{
		ShellStatus:         shellStatus,
		ShellPath:           shellPath,
		DependencyStatus:    dependencyStatus,
		DependencyAvailable: available,
		DependencyMissing:   missing,
		Checks:              checks,
	}
}

func renderSandboxCheck(settings permissions.SandboxSettings) Result {
	diagnostics := gatherSandboxDiagnostics(settings)
	status := "ok"
	for _, check := range diagnostics.Checks {
		if check.Status != "ok" {
			status = "warn"
			break
		}
	}
	lines := []string{
		"SANDBOX_CHECK",
		fmt.Sprintf("status=%s", status),
		fmt.Sprintf("check_count=%d", len(diagnostics.Checks)),
	}
	for i, check := range diagnostics.Checks {
		idx := i + 1
		lines = append(lines, fmt.Sprintf("check.%d.id=%s", idx, normalizeToken(check.ID)))
		lines = append(lines, fmt.Sprintf("check.%d.status=%s", idx, normalizeToken(check.Status)))
		lines = append(lines, fmt.Sprintf("check.%d.detail=%s", idx, normalizeToken(check.Detail)))
	}
	return resultWithIntents(strings.Join(lines, "\n"), legacyOutputIntents(strings.Join(lines, "\n"))...)
}

func renderSandboxExcludeList(patterns []string) Result {
	patterns = append([]string(nil), patterns...)
	sort.Strings(patterns)
	if len(patterns) == 0 {
		return resultWithIntents("SANDBOX_EXCLUDE_LIST\ncount=0", legacyOutputIntents("SANDBOX_EXCLUDE_LIST\ncount=0")...)
	}
	lines := []string{"SANDBOX_EXCLUDE_LIST", fmt.Sprintf("count=%d", len(patterns))}
	for i, pattern := range patterns {
		lines = append(lines, fmt.Sprintf("pattern.%d=%s", i+1, normalizeToken(pattern)))
	}
	return resultWithIntents(strings.Join(lines, "\n"), legacyOutputIntents(strings.Join(lines, "\n"))...)
}

func loadSandboxSettings(state *RuntimeState) (permissions.SandboxSettings, error) {
	defaults := permissions.DefaultSandboxSettings()
	settings := permissions.SandboxSettings{
		Mode:             state.SandboxMode,
		WorkspaceLocked:  state.SandboxWorkspaceLocked,
		ExcludedCommands: append([]string(nil), state.SandboxExcludedCommands...),
	}
	if strings.TrimSpace(settings.Mode) == "" {
		settings.Mode = defaults.Mode
	}
	if strings.TrimSpace(state.SandboxSettingsPath) != "" {
		store := permissions.NewSandboxSettingsStore(state.SandboxSettingsPath)
		loaded, err := store.Load()
		if err != nil {
			return permissions.SandboxSettings{}, err
		}
		settings = loaded
	}
	settings = permissions.NormalizeSandboxSettings(settings)
	state.SandboxMode = settings.Mode
	state.SandboxWorkspaceLocked = settings.WorkspaceLocked
	state.SandboxExcludedCommands = append([]string(nil), settings.ExcludedCommands...)
	return settings, nil
}

func saveSandboxSettings(state *RuntimeState, settings permissions.SandboxSettings) error {
	settings = permissions.NormalizeSandboxSettings(settings)
	state.SandboxMode = settings.Mode
	state.SandboxWorkspaceLocked = settings.WorkspaceLocked
	state.SandboxExcludedCommands = append([]string(nil), settings.ExcludedCommands...)
	if strings.TrimSpace(state.SandboxSettingsPath) == "" {
		return nil
	}
	store := permissions.NewSandboxSettingsStore(state.SandboxSettingsPath)
	return store.Save(settings)
}

func trimQuoted(raw string) string {
	raw = strings.TrimSpace(raw)
	if len(raw) >= 2 {
		if (raw[0] == '"' && raw[len(raw)-1] == '"') || (raw[0] == '\'' && raw[len(raw)-1] == '\'') {
			return strings.TrimSpace(raw[1 : len(raw)-1])
		}
	}
	return raw
}

// TasksCommand provides deterministic task list controls.
type TasksCommand struct{}

func NewTasksCommand() *TasksCommand      { return &TasksCommand{} }
func (c *TasksCommand) Name() string      { return "tasks" }
func (c *TasksCommand) Aliases() []string { return []string{"todo", "bashes"} }
func (c *TasksCommand) Description() string {
	return "Manage deterministic local tasks list"
}
func (c *TasksCommand) Usage() string {
	return "/tasks [list|add <text>|done <index>|clear]"
}

func (c *TasksCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "list") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /tasks [list|add <text>|done <index>|clear]")
		}
		lines := []string{"TASKS_LIST", fmt.Sprintf("count=%d", len(cmdCtx.State.Tasks)), fmt.Sprintf("completed=%d", cmdCtx.State.TasksCompleted)}
		for i, task := range cmdCtx.State.Tasks {
			lines = append(lines, fmt.Sprintf("task.%d=%s", i+1, normalizeToken(task)))
		}
		return resultWithIntents(strings.Join(lines, "\n"), tasksListStateIntents(cmdCtx.State.Tasks, cmdCtx.State.TasksCompleted)...), nil
	}
	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "add":
		if len(inv.Args) < 2 {
			return Result{}, fmt.Errorf("usage: /tasks [list|add <text>|done <index>|clear]")
		}
		task := strings.TrimSpace(strings.Join(inv.Args[1:], " "))
		if task == "" {
			return Result{}, fmt.Errorf("usage: /tasks [list|add <text>|done <index>|clear]")
		}
		cmdCtx.State.Tasks = append(cmdCtx.State.Tasks, task)
		message := fmt.Sprintf("TASKS_ADD\ntask=%s\ncount=%d", normalizeToken(task), len(cmdCtx.State.Tasks))
		return resultWithIntents(message, tasksMutationIntents("Task added", task, len(cmdCtx.State.Tasks), cmdCtx.State.TasksCompleted)...), nil
	case "done":
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf("usage: /tasks [list|add <text>|done <index>|clear]")
		}
		idx, err := parseNonNegativeInt(inv.Args[1])
		if err != nil || idx == 0 || idx > len(cmdCtx.State.Tasks) {
			return Result{}, fmt.Errorf("usage: /tasks [list|add <text>|done <index>|clear]")
		}
		task := cmdCtx.State.Tasks[idx-1]
		cmdCtx.State.Tasks = append(cmdCtx.State.Tasks[:idx-1], cmdCtx.State.Tasks[idx:]...)
		cmdCtx.State.TasksCompleted++
		message := fmt.Sprintf("TASKS_DONE\nindex=%d\ntask=%s\ncount=%d\ncompleted=%d", idx, normalizeToken(task), len(cmdCtx.State.Tasks), cmdCtx.State.TasksCompleted)
		return resultWithIntents(message, tasksMutationIntents("Task completed", task, len(cmdCtx.State.Tasks), cmdCtx.State.TasksCompleted)...), nil
	case "clear":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /tasks [list|add <text>|done <index>|clear]")
		}
		cmdCtx.State.Tasks = nil
		message := fmt.Sprintf("TASKS_CLEAR\ncount=0\ncompleted=%d", cmdCtx.State.TasksCompleted)
		return resultWithIntents(message, tasksMutationIntents("Tasks cleared", "", 0, cmdCtx.State.TasksCompleted)...), nil
	default:
		return Result{}, fmt.Errorf("usage: /tasks [list|add <text>|done <index>|clear]")
	}
}

// AdvisorCommand configures a deterministic advisor model setting.
type AdvisorCommand struct{}

func NewAdvisorCommand() *AdvisorCommand      { return &AdvisorCommand{} }
func (c *AdvisorCommand) Name() string        { return "advisor" }
func (c *AdvisorCommand) Aliases() []string   { return nil }
func (c *AdvisorCommand) Description() string { return "Configure advisor model for this session" }
func (c *AdvisorCommand) Usage() string       { return "/advisor [status|unset|off|<model>]" }

func (c *AdvisorCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		active := strings.TrimSpace(cmdCtx.State.AdvisorModel) != ""
		return resultWithIntents(fmt.Sprintf("ADVISOR_STATUS\nmodel=%s\nactive=%t\nupdates=%d", normalizeToken(cmdCtx.State.AdvisorModel), active, cmdCtx.State.AdvisorUpdates), advisorStatusIntents(cmdCtx.State.AdvisorModel, active, cmdCtx.State.AdvisorUpdates)...), nil
	}
	if len(inv.Args) != 1 {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	arg := strings.TrimSpace(inv.Args[0])
	if arg == "" {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	if strings.EqualFold(arg, "unset") || strings.EqualFold(arg, "off") {
		prev := cmdCtx.State.AdvisorModel
		cmdCtx.State.AdvisorModel = ""
		cmdCtx.State.AdvisorUpdates++
		return resultWithIntents(fmt.Sprintf("ADVISOR_UNSET\nprevious=%s\nactive=false\nupdates=%d", normalizeToken(prev), cmdCtx.State.AdvisorUpdates), advisorSetIntents("", false, cmdCtx.State.AdvisorUpdates)...), nil
	}
	cmdCtx.State.AdvisorModel = arg
	cmdCtx.State.AdvisorUpdates++
	return resultWithIntents(fmt.Sprintf("ADVISOR_SET\nmodel=%s\nactive=true\nupdates=%d", normalizeToken(arg), cmdCtx.State.AdvisorUpdates), advisorSetIntents(arg, true, cmdCtx.State.AdvisorUpdates)...), nil
}

// BtwCommand tracks side-question usage in a deterministic way.
type BtwCommand struct{}

func NewBtwCommand() *BtwCommand        { return &BtwCommand{} }
func (c *BtwCommand) Name() string      { return "btw" }
func (c *BtwCommand) Aliases() []string { return nil }
func (c *BtwCommand) Description() string {
	return "Ask a side question without mutating primary context"
}
func (c *BtwCommand) Usage() string { return "/btw <question>" }

func (c *BtwCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	question := strings.TrimSpace(strings.Join(inv.Args, " "))
	if question == "" {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	cmdCtx.State.BtwUseCount++
	cmdCtx.State.LastBtwQuestion = question
	return resultWithIntents(fmt.Sprintf("BTW_RESULT\nquestion=%s\ncount=%d\nstatus=queued", normalizeToken(question), cmdCtx.State.BtwUseCount), btwIntents(question, cmdCtx.State.BtwUseCount)...), nil
}

// ChromeCommand controls deterministic Claude-in-Chrome state.
type ChromeCommand struct{}

func NewChromeCommand() *ChromeCommand       { return &ChromeCommand{} }
func (c *ChromeCommand) Name() string        { return "chrome" }
func (c *ChromeCommand) Aliases() []string   { return nil }
func (c *ChromeCommand) Description() string { return "Manage Claude in Chrome settings" }
func (c *ChromeCommand) Usage() string {
	return "/chrome [status|install-extension|reconnect|manage-permissions|toggle-default [on|off]]"
}

func (c *ChromeCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return resultWithIntents(fmt.Sprintf("CHROME_STATUS\ndefault_enabled=%t\nextension_installed=%t\nconnected=%t\nactions=%d", cmdCtx.State.ChromeDefault, cmdCtx.State.ChromeExtension, cmdCtx.State.ChromeConnected, cmdCtx.State.ChromeActions), chromeStatusIntents(cmdCtx.State.ChromeDefault, cmdCtx.State.ChromeExtension, cmdCtx.State.ChromeConnected, cmdCtx.State.ChromeActions)...), nil
	}

	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "install-extension":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.ChromeExtension = true
		cmdCtx.State.ChromeActions++
		return resultWithIntents(fmt.Sprintf("CHROME_EXTENSION\ninstalled=true\nactions=%d", cmdCtx.State.ChromeActions), chromeActionIntents("Chrome extension", cmdCtx.State.ChromeDefault, cmdCtx.State.ChromeExtension, cmdCtx.State.ChromeConnected, cmdCtx.State.ChromeActions)...), nil
	case "reconnect":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.ChromeConnected = cmdCtx.State.ChromeExtension
		cmdCtx.State.ChromeActions++
		return resultWithIntents(fmt.Sprintf("CHROME_RECONNECT\nconnected=%t\nextension_installed=%t\nactions=%d", cmdCtx.State.ChromeConnected, cmdCtx.State.ChromeExtension, cmdCtx.State.ChromeActions), chromeActionIntents("Chrome reconnect", cmdCtx.State.ChromeDefault, cmdCtx.State.ChromeExtension, cmdCtx.State.ChromeConnected, cmdCtx.State.ChromeActions)...), nil
	case "manage-permissions":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.ChromeActions++
		return resultWithIntents(fmt.Sprintf("CHROME_PERMISSIONS\nopened=true\nactions=%d", cmdCtx.State.ChromeActions), chromeActionIntents("Chrome permissions", cmdCtx.State.ChromeDefault, cmdCtx.State.ChromeExtension, cmdCtx.State.ChromeConnected, cmdCtx.State.ChromeActions)...), nil
	case "toggle-default":
		if len(inv.Args) > 2 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		if len(inv.Args) == 2 {
			next := strings.ToLower(strings.TrimSpace(inv.Args[1]))
			switch next {
			case "on", "true", "1":
				cmdCtx.State.ChromeDefault = true
			case "off", "false", "0":
				cmdCtx.State.ChromeDefault = false
			default:
				return Result{}, fmt.Errorf("usage: %s", c.Usage())
			}
		} else {
			cmdCtx.State.ChromeDefault = !cmdCtx.State.ChromeDefault
		}
		cmdCtx.State.ChromeActions++
		return resultWithIntents(fmt.Sprintf("CHROME_DEFAULT\nenabled=%t\nactions=%d", cmdCtx.State.ChromeDefault, cmdCtx.State.ChromeActions), chromeActionIntents("Chrome default", cmdCtx.State.ChromeDefault, cmdCtx.State.ChromeExtension, cmdCtx.State.ChromeConnected, cmdCtx.State.ChromeActions)...), nil
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
}

// ColorCommand configures deterministic session color state.
type ColorCommand struct{}

func NewColorCommand() *ColorCommand        { return &ColorCommand{} }
func (c *ColorCommand) Name() string        { return "color" }
func (c *ColorCommand) Aliases() []string   { return nil }
func (c *ColorCommand) Description() string { return "Set session color" }
func (c *ColorCommand) Usage() string       { return "/color <color|default>" }

func (c *ColorCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 {
		return resultWithIntents(fmt.Sprintf("COLOR_STATUS\ncolor=%s\nset_count=%d", normalizeToken(cmdCtx.State.SessionColor), cmdCtx.State.ColorSetCount), colorStatusIntents(cmdCtx.State.SessionColor, cmdCtx.State.ColorSetCount)...), nil
	}
	if len(inv.Args) != 1 {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	arg := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	if arg == "" {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	if arg == "default" || arg == "reset" || arg == "none" || arg == "gray" || arg == "grey" {
		cmdCtx.State.SessionColor = ""
		cmdCtx.State.ColorSetCount++
		return resultWithIntents(fmt.Sprintf("COLOR_RESET\ncolor=-\nset_count=%d", cmdCtx.State.ColorSetCount), colorSetIntents("", cmdCtx.State.ColorSetCount)...), nil
	}
	if !isSupportedColor(arg) {
		return Result{}, fmt.Errorf("invalid color %q", arg)
	}
	cmdCtx.State.SessionColor = arg
	cmdCtx.State.ColorSetCount++
	return resultWithIntents(fmt.Sprintf("COLOR_SET\ncolor=%s\nset_count=%d", normalizeToken(cmdCtx.State.SessionColor), cmdCtx.State.ColorSetCount), colorSetIntents(cmdCtx.State.SessionColor, cmdCtx.State.ColorSetCount)...), nil
}

// DesktopCommand tracks desktop handoff requests.
type DesktopCommand struct{}

func NewDesktopCommand() *DesktopCommand      { return &DesktopCommand{} }
func (c *DesktopCommand) Name() string        { return "desktop" }
func (c *DesktopCommand) Aliases() []string   { return []string{"app"} }
func (c *DesktopCommand) Description() string { return "Continue current session in desktop app" }
func (c *DesktopCommand) Usage() string       { return "/desktop [status|open]" }

func (c *DesktopCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "open") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.DesktopHandoffCount++
		cmdCtx.State.LastDesktopTarget = "claude-desktop"
		return resultWithIntents(fmt.Sprintf("DESKTOP_HANDOFF\ntarget=%s\ncount=%d", cmdCtx.State.LastDesktopTarget, cmdCtx.State.DesktopHandoffCount), desktopHandoffIntents(cmdCtx.State.LastDesktopTarget, cmdCtx.State.DesktopHandoffCount)...), nil
	}
	if len(inv.Args) == 1 && strings.EqualFold(inv.Args[0], "status") {
		return resultWithIntents(fmt.Sprintf("DESKTOP_STATUS\ncount=%d\nlast_target=%s", cmdCtx.State.DesktopHandoffCount, normalizeToken(cmdCtx.State.LastDesktopTarget)), desktopStatusIntents(cmdCtx.State.DesktopHandoffCount, cmdCtx.State.LastDesktopTarget)...), nil
	}
	return Result{}, fmt.Errorf("usage: %s", c.Usage())
}

// MobileCommand controls deterministic mobile QR flow state.
type MobileCommand struct{}

func NewMobileCommand() *MobileCommand       { return &MobileCommand{} }
func (c *MobileCommand) Name() string        { return "mobile" }
func (c *MobileCommand) Aliases() []string   { return []string{"ios", "android"} }
func (c *MobileCommand) Description() string { return "Show mobile app QR code state" }
func (c *MobileCommand) Usage() string       { return "/mobile [ios|android|status]" }

func (c *MobileCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	platform := ""
	if strings.EqualFold(inv.Name, "ios") || strings.EqualFold(inv.Name, "android") {
		platform = strings.ToLower(inv.Name)
	}

	if len(inv.Args) == 0 {
		if platform == "" {
			platform = "ios"
		}
		cmdCtx.State.MobilePlatform = platform
		cmdCtx.State.MobileQRCount++
		return resultWithIntents(fmt.Sprintf("MOBILE_QR\nplatform=%s\ncount=%d", cmdCtx.State.MobilePlatform, cmdCtx.State.MobileQRCount), mobileQRIntents(cmdCtx.State.MobilePlatform, cmdCtx.State.MobileQRCount)...), nil
	}
	if len(inv.Args) != 1 {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	arg := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch arg {
	case "status":
		return resultWithIntents(fmt.Sprintf("MOBILE_STATUS\nplatform=%s\ncount=%d", normalizeToken(cmdCtx.State.MobilePlatform), cmdCtx.State.MobileQRCount), mobileStatusIntents(cmdCtx.State.MobilePlatform, cmdCtx.State.MobileQRCount)...), nil
	case "ios", "android":
		cmdCtx.State.MobilePlatform = arg
		cmdCtx.State.MobileQRCount++
		return resultWithIntents(fmt.Sprintf("MOBILE_QR\nplatform=%s\ncount=%d", cmdCtx.State.MobilePlatform, cmdCtx.State.MobileQRCount), mobileQRIntents(cmdCtx.State.MobilePlatform, cmdCtx.State.MobileQRCount)...), nil
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
}

// FastCommand toggles deterministic fast-mode runtime flag.
type FastCommand struct{}

func NewFastCommand() *FastCommand         { return &FastCommand{} }
func (c *FastCommand) Name() string        { return "fast" }
func (c *FastCommand) Aliases() []string   { return nil }
func (c *FastCommand) Description() string { return "Toggle fast mode" }
func (c *FastCommand) Usage() string       { return "/fast [on|off|status]" }

func (c *FastCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return resultWithIntents(fmt.Sprintf("FAST_STATUS\nenabled=%t\ntoggles=%d", cmdCtx.State.FastMode, cmdCtx.State.FastToggleCount), fastStatusIntents(cmdCtx.State.FastMode, cmdCtx.State.FastToggleCount)...), nil
	}
	if len(inv.Args) != 1 {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	switch strings.ToLower(strings.TrimSpace(inv.Args[0])) {
	case "on":
		cmdCtx.State.FastMode = true
		cmdCtx.State.FastToggleCount++
		return resultWithIntents(fmt.Sprintf("FAST_SET\nenabled=true\ntoggles=%d", cmdCtx.State.FastToggleCount), fastSetIntents(true, cmdCtx.State.FastToggleCount)...), nil
	case "off":
		cmdCtx.State.FastMode = false
		cmdCtx.State.FastToggleCount++
		return resultWithIntents(fmt.Sprintf("FAST_SET\nenabled=false\ntoggles=%d", cmdCtx.State.FastToggleCount), fastSetIntents(false, cmdCtx.State.FastToggleCount)...), nil
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
}

// EffortCommand sets or reports deterministic effort preference.
type EffortCommand struct{}

func NewEffortCommand() *EffortCommand       { return &EffortCommand{} }
func (c *EffortCommand) Name() string        { return "effort" }
func (c *EffortCommand) Aliases() []string   { return nil }
func (c *EffortCommand) Description() string { return "Set effort level" }
func (c *EffortCommand) Usage() string       { return "/effort [low|medium|high|max|auto|status]" }

func (c *EffortCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "status") || strings.EqualFold(inv.Args[0], "current") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		value := cmdCtx.State.EffortLevel
		if strings.TrimSpace(value) == "" {
			value = "auto"
		}
		return resultWithIntents(fmt.Sprintf("EFFORT_STATUS\nvalue=%s\nupdates=%d", normalizeToken(value), cmdCtx.State.EffortSetCount), effortStatusIntents(value, cmdCtx.State.EffortSetCount)...), nil
	}
	if len(inv.Args) != 1 {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	arg := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	if !isValidEffort(arg) {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	if arg == "auto" || arg == "unset" {
		cmdCtx.State.EffortLevel = ""
		cmdCtx.State.EffortSetCount++
		return resultWithIntents(fmt.Sprintf("EFFORT_SET\nvalue=auto\nupdates=%d", cmdCtx.State.EffortSetCount), effortSetIntents("auto", cmdCtx.State.EffortSetCount)...), nil
	}
	cmdCtx.State.EffortLevel = arg
	cmdCtx.State.EffortSetCount++
	return resultWithIntents(fmt.Sprintf("EFFORT_SET\nvalue=%s\nupdates=%d", normalizeToken(arg), cmdCtx.State.EffortSetCount), effortSetIntents(arg, cmdCtx.State.EffortSetCount)...), nil
}

// PluginCommand manages deterministic plugin inventory and enabled state.
type PluginCommand struct{}

func NewPluginCommand() *PluginCommand       { return &PluginCommand{} }
func (c *PluginCommand) Name() string        { return "plugin" }
func (c *PluginCommand) Aliases() []string   { return []string{"plugins", "marketplace"} }
func (c *PluginCommand) Description() string { return "Manage local plugin inventory" }
func (c *PluginCommand) Usage() string {
	return "/plugin [list|status|doctor|install <name[@version]>|update <name[@version]>|remove <name>|enable <name>|disable <name>|marketplace [list]]"
}

func (c *PluginCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	svc, workingDir, err := pluginServiceFromWorkingDir()
	if err != nil {
		return Result{}, err
	}
	result, listErr := svc.ListWithDiagnostics()
	if listErr != nil {
		result = pluginspkg.ServiceListResult{}
	}
	pluginCommands, _ := resolvedPluginCommandsForPluginCommand(svc)
	conflicts := pluginSkillConflictsForWorkingDir(workingDir, pluginCommands)
	syncPluginRuntimeState(cmdCtx.State, result)

	if len(inv.Args) == 0 {
		if strings.EqualFold(inv.Name, "marketplace") {
			return renderPluginMarketplaces(cmdCtx.State), nil
		}
		return renderPluginList(cmdCtx.State, result, conflicts), nil
	}

	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "list":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return renderPluginList(cmdCtx.State, result, conflicts), nil
	case "status":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		lines := []string{
			"PLUGIN_STATUS",
			fmt.Sprintf("installed=%d", len(cmdCtx.State.PluginsInstalled)),
			fmt.Sprintf("enabled=%d", len(cmdCtx.State.PluginsEnabled)),
			fmt.Sprintf("pending_reload=%t", cmdCtx.State.PluginReloadPending),
			fmt.Sprintf("mutations=%d", cmdCtx.State.PluginMutations),
			fmt.Sprintf("diagnostic_groups=%d", len(result.Diagnostics)),
			fmt.Sprintf("conflicts=%d", len(conflicts)),
		}
		return resultWithIntents(strings.Join(lines, "\n"), pluginStatusIntents(cmdCtx.State, len(result.Diagnostics), len(conflicts))...), nil
	case "doctor":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.PluginDoctorCount++
		quickFix := "/plugin install <plugin-id>"
		if len(cmdCtx.State.PluginsInstalled) > 0 && cmdCtx.State.PluginReloadPending {
			quickFix = "/reload-plugins"
		}
		lines := []string{
			"PLUGIN_DOCTOR",
			fmt.Sprintf("installed=%d", len(cmdCtx.State.PluginsInstalled)),
			fmt.Sprintf("enabled=%d", len(cmdCtx.State.PluginsEnabled)),
			fmt.Sprintf("pending_reload=%t", cmdCtx.State.PluginReloadPending),
			fmt.Sprintf("diagnostic_groups=%d", len(result.Diagnostics)),
			fmt.Sprintf("conflicts=%d", len(conflicts)),
			fmt.Sprintf("quick_fix=%s", normalizeToken(quickFix)),
		}
		lines = append(lines, fmt.Sprintf("doctor_runs=%d", cmdCtx.State.PluginDoctorCount))
		return resultWithIntents(strings.Join(lines, "\n"), pluginDiagnosticsIntents(cmdCtx.State, len(result.Diagnostics), len(conflicts), quickFix)...), nil
	case "diagnostics":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return renderPluginDiagnostics(cmdCtx.State, result, conflicts), nil
	case "repair":
		if len(inv.Args) > 2 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		mode := "auto"
		if len(inv.Args) == 2 {
			mode = strings.ToLower(strings.TrimSpace(inv.Args[1]))
			if mode == "" {
				return Result{}, fmt.Errorf("usage: %s", c.Usage())
			}
		}
		return repairPluginState(cmdCtx.State, mode), nil
	case "install":
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		name := strings.TrimSpace(inv.Args[1])
		if name == "" {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		if err := svc.Install(name); err != nil {
			return Result{}, err
		}
		updated, err := svc.ListWithDiagnostics()
		if err != nil {
			return Result{}, err
		}
		syncPluginRuntimeState(cmdCtx.State, updated)
		cmdCtx.State.PluginReloadPending = true
		cmdCtx.State.PluginMutations++
		message := fmt.Sprintf("PLUGIN_INSTALL\nname=%s\ninstalled=%d\nenabled=%d\npending_reload=true\nmutations=%d", normalizeToken(name), len(cmdCtx.State.PluginsInstalled), len(cmdCtx.State.PluginsEnabled), cmdCtx.State.PluginMutations)
		return resultWithIntents(message, pluginMutationIntents("Plugin installed", name, cmdCtx.State.PluginMutations)...), nil
	case "update":
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		name := strings.TrimSpace(inv.Args[1])
		if name == "" {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		if err := svc.Update(name); err != nil {
			return Result{}, err
		}
		updated, err := svc.ListWithDiagnostics()
		if err != nil {
			return Result{}, err
		}
		syncPluginRuntimeState(cmdCtx.State, updated)
		cmdCtx.State.PluginReloadPending = true
		cmdCtx.State.PluginMutations++
		message := fmt.Sprintf("PLUGIN_UPDATE\nname=%s\ninstalled=%d\nenabled=%d\npending_reload=true\nmutations=%d", normalizeToken(name), len(cmdCtx.State.PluginsInstalled), len(cmdCtx.State.PluginsEnabled), cmdCtx.State.PluginMutations)
		return resultWithIntents(message, pluginMutationIntents("Plugin updated", name, cmdCtx.State.PluginMutations)...), nil
	case "remove", "uninstall":
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		name := strings.TrimSpace(inv.Args[1])
		if name == "" {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		installedBefore := containsString(cmdCtx.State.PluginsInstalled, splitPluginRefName(name))
		enabledBefore := containsString(cmdCtx.State.PluginsEnabled, splitPluginRefName(name))
		if err := svc.Remove(name); err != nil {
			return Result{}, err
		}
		updated, err := svc.ListWithDiagnostics()
		if err != nil {
			return Result{}, err
		}
		syncPluginRuntimeState(cmdCtx.State, updated)
		cmdCtx.State.PluginReloadPending = true
		cmdCtx.State.PluginMutations++
		message := fmt.Sprintf("PLUGIN_REMOVE\nname=%s\nremoved_installed=%t\nremoved_enabled=%t\npending_reload=true\nmutations=%d", normalizeToken(name), installedBefore, enabledBefore, cmdCtx.State.PluginMutations)
		return resultWithIntents(message, pluginMutationIntents("Plugin removed", name, cmdCtx.State.PluginMutations)...), nil
	case "enable":
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		name := strings.TrimSpace(inv.Args[1])
		if name == "" {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		if err := svc.Enable(name); err != nil {
			return Result{}, err
		}
		updated, err := svc.ListWithDiagnostics()
		if err != nil {
			return Result{}, err
		}
		syncPluginRuntimeState(cmdCtx.State, updated)
		cmdCtx.State.PluginReloadPending = true
		cmdCtx.State.PluginMutations++
		message := fmt.Sprintf("PLUGIN_ENABLE\nname=%s\nenabled=%t\npending_reload=true\nmutations=%d", normalizeToken(name), true, cmdCtx.State.PluginMutations)
		return resultWithIntents(message, pluginMutationIntents("Plugin enabled", name, cmdCtx.State.PluginMutations)...), nil
	case "disable":
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		name := strings.TrimSpace(inv.Args[1])
		if name == "" {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		removed := containsString(cmdCtx.State.PluginsEnabled, splitPluginRefName(name))
		if err := svc.Disable(name); err != nil {
			return Result{}, err
		}
		updated, err := svc.ListWithDiagnostics()
		if err != nil {
			return Result{}, err
		}
		syncPluginRuntimeState(cmdCtx.State, updated)
		cmdCtx.State.PluginReloadPending = true
		cmdCtx.State.PluginMutations++
		message := fmt.Sprintf("PLUGIN_DISABLE\nname=%s\nremoved=%t\npending_reload=true\nmutations=%d", normalizeToken(name), removed, cmdCtx.State.PluginMutations)
		return resultWithIntents(message, pluginMutationIntents("Plugin disabled", name, cmdCtx.State.PluginMutations)...), nil
	case "marketplace":
		return c.handleMarketplace(cmdCtx.State, inv.Args[1:])
	default:
		if strings.EqualFold(inv.Name, "marketplace") {
			return c.handleMarketplace(cmdCtx.State, inv.Args)
		}
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
}

func (c *PluginCommand) handleMarketplace(state *RuntimeState, args []string) (Result, error) {
	if len(args) == 0 || strings.EqualFold(strings.TrimSpace(args[0]), "list") {
		if len(args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return renderPluginMarketplaces(state), nil
	}
	return Result{}, fmt.Errorf("usage: %s", c.Usage())
}

// ReloadPluginsCommand applies pending plugin changes.
type ReloadPluginsCommand struct{}

func NewReloadPluginsCommand() *ReloadPluginsCommand { return &ReloadPluginsCommand{} }
func (c *ReloadPluginsCommand) Name() string         { return "reload-plugins" }
func (c *ReloadPluginsCommand) Aliases() []string    { return nil }
func (c *ReloadPluginsCommand) Description() string  { return "Apply pending plugin changes" }
func (c *ReloadPluginsCommand) Usage() string        { return "/reload-plugins" }
func (c *ReloadPluginsCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) > 0 {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	normalizePluginState(cmdCtx.State)
	pendingBefore := cmdCtx.State.PluginReloadPending
	cmdCtx.State.PluginReloadPending = false
	cmdCtx.State.PluginReloadCount++
	return resultWithIntents(fmt.Sprintf("RELOAD_PLUGINS_RESULT\npending_before=%t\ninstalled=%d\nenabled=%d\nmarketplaces=%d\nreload_count=%d", pendingBefore, len(cmdCtx.State.PluginsInstalled), len(cmdCtx.State.PluginsEnabled), len(cmdCtx.State.PluginMarketplaces), cmdCtx.State.PluginReloadCount), reloadPluginsIntents(pendingBefore, len(cmdCtx.State.PluginsInstalled), len(cmdCtx.State.PluginsEnabled), len(cmdCtx.State.PluginMarketplaces), cmdCtx.State.PluginReloadCount)...), nil
}

// ExportCommand tracks deterministic export destination data.
type ExportCommand struct{}

func NewExportCommand() *ExportCommand       { return &ExportCommand{} }
func (c *ExportCommand) Name() string        { return "export" }
func (c *ExportCommand) Aliases() []string   { return nil }
func (c *ExportCommand) Description() string { return "Export current conversation transcript" }
func (c *ExportCommand) Usage() string       { return "/export [filename]" }

func (c *ExportCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) > 1 {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	filename := ""
	if len(inv.Args) == 1 {
		filename = strings.TrimSpace(inv.Args[0])
	}
	if filename == "" {
		filename = fmt.Sprintf("conversation-%03d.txt", cmdCtx.State.ExportsCount+1)
	}
	if !strings.HasSuffix(strings.ToLower(filename), ".txt") {
		filename = filename + ".txt"
	}
	cmdCtx.State.ExportsCount++
	cmdCtx.State.LastExportPath = filename
	cmdCtx.State.LastExportFormat = "txt"
	return resultWithIntents(fmt.Sprintf("EXPORT_RESULT\npath=%s\nformat=txt\ncount=%d", normalizeToken(cmdCtx.State.LastExportPath), cmdCtx.State.ExportsCount), exportResultIntents(cmdCtx.State.LastExportPath, cmdCtx.State.ExportsCount)...), nil
}

// ExtraUsageCommand models extra-usage toggles and request flow.
type ExtraUsageCommand struct{}

func NewExtraUsageCommand() *ExtraUsageCommand   { return &ExtraUsageCommand{} }
func (c *ExtraUsageCommand) Name() string        { return "extra-usage" }
func (c *ExtraUsageCommand) Aliases() []string   { return nil }
func (c *ExtraUsageCommand) Description() string { return "Manage extra usage when limits are reached" }
func (c *ExtraUsageCommand) Usage() string       { return "/extra-usage [status|enable|disable|request]" }
func (c *ExtraUsageCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return resultWithIntents(fmt.Sprintf("EXTRA_USAGE_STATUS\nenabled=%t\nrequests=%d", cmdCtx.State.ExtraUsageEnabled, cmdCtx.State.ExtraUsageRequests), extraUsageStatusIntents(cmdCtx.State.ExtraUsageEnabled, cmdCtx.State.ExtraUsageRequests)...), nil
	}
	if len(inv.Args) != 1 {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "enable":
		cmdCtx.State.ExtraUsageEnabled = true
		return resultWithIntents(fmt.Sprintf("EXTRA_USAGE_SET\nenabled=true\nrequests=%d", cmdCtx.State.ExtraUsageRequests), extraUsageSetIntents(true, cmdCtx.State.ExtraUsageRequests)...), nil
	case "disable":
		cmdCtx.State.ExtraUsageEnabled = false
		return resultWithIntents(fmt.Sprintf("EXTRA_USAGE_SET\nenabled=false\nrequests=%d", cmdCtx.State.ExtraUsageRequests), extraUsageSetIntents(false, cmdCtx.State.ExtraUsageRequests)...), nil
	case "request":
		cmdCtx.State.ExtraUsageRequests++
		return resultWithIntents(fmt.Sprintf("EXTRA_USAGE_REQUEST\nrequests=%d\nenabled=%t", cmdCtx.State.ExtraUsageRequests, cmdCtx.State.ExtraUsageEnabled), extraUsageRequestIntents(cmdCtx.State.ExtraUsageEnabled, cmdCtx.State.ExtraUsageRequests)...), nil
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
}

// RateLimitOptionsCommand provides deterministic action selection output.
type RateLimitOptionsCommand struct{}

func NewRateLimitOptionsCommand() *RateLimitOptionsCommand { return &RateLimitOptionsCommand{} }
func (c *RateLimitOptionsCommand) Name() string            { return "rate-limit-options" }
func (c *RateLimitOptionsCommand) Aliases() []string       { return nil }
func (c *RateLimitOptionsCommand) Description() string {
	return "Show deterministic options for rate-limit events"
}
func (c *RateLimitOptionsCommand) Usage() string {
	return "/rate-limit-options [status|upgrade|extra-usage|cancel]"
}

func (c *RateLimitOptionsCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.RateLimitPrompts++
		options := []string{"cancel", "extra-usage", "upgrade"}
		lines := []string{"RATE_LIMIT_OPTIONS", fmt.Sprintf("count=%d", len(options)), fmt.Sprintf("prompts=%d", cmdCtx.State.RateLimitPrompts), fmt.Sprintf("last_action=%s", normalizeToken(cmdCtx.State.LastRateLimitAction))}
		for i, option := range options {
			lines = append(lines, fmt.Sprintf("option.%d=%s", i+1, option))
		}
		return resultWithIntents(strings.Join(lines, "\n"), rateLimitOptionsIntents(options, cmdCtx.State.RateLimitPrompts, cmdCtx.State.LastRateLimitAction)...), nil
	}
	if len(inv.Args) != 1 {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	action := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	if action != "upgrade" && action != "extra-usage" && action != "cancel" {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	cmdCtx.State.LastRateLimitAction = action
	return resultWithIntents(fmt.Sprintf("RATE_LIMIT_ACTION\naction=%s\nprompts=%d", action, cmdCtx.State.RateLimitPrompts), rateLimitActionIntents(action, cmdCtx.State.RateLimitPrompts)...), nil
}

// PRCommentsCommand tracks deterministic pull request comment fetch requests.
type PRCommentsCommand struct{}

func NewPRCommentsCommand() *PRCommentsCommand   { return &PRCommentsCommand{} }
func (c *PRCommentsCommand) Name() string        { return "pr-comments" }
func (c *PRCommentsCommand) Aliases() []string   { return nil }
func (c *PRCommentsCommand) Description() string { return "Fetch pull-request comments metadata" }
func (c *PRCommentsCommand) Usage() string {
	return "/pr-comments [status|<pr-ref>|file=<path>|state]"
}
func (c *PRCommentsCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return resultWithIntents(fmt.Sprintf("PR_COMMENTS_STATUS\nfetches=%d\nlast_ref=%s", cmdCtx.State.PRCommentsFetches, normalizeToken(cmdCtx.State.LastPRCommentsRef)), prCommentsStatusIntents(cmdCtx.State.PRCommentsFetches, cmdCtx.State.LastPRCommentsRef)...), nil
	}
	if len(inv.Args) > 1 {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	ref, sourceName, threads, err := resolvePRCommentThreads(cmdCtx.State, strings.TrimSpace(inv.Args[0]))
	if err != nil {
		return Result{}, err
	}
	cmdCtx.State.PRCommentsFetches++
	cmdCtx.State.LastPRCommentsRef = ref
	lines := []string{
		"PR_COMMENTS_SUMMARY",
		fmt.Sprintf("ref=%s", normalizeToken(ref)),
		fmt.Sprintf("source=%s", normalizeToken(sourceName)),
		fmt.Sprintf("fetches=%d", cmdCtx.State.PRCommentsFetches),
		fmt.Sprintf("threads=%d", len(threads)),
	}
	for i, thread := range threads {
		idx := i + 1
		lines = append(lines, fmt.Sprintf("thread.%d.path=%s", idx, normalizeToken(thread.Path)))
		lines = append(lines, fmt.Sprintf("thread.%d.line=%d", idx, thread.Line))
		lines = append(lines, fmt.Sprintf("thread.%d.comments=%d", idx, len(thread.Comments)))
		for j, comment := range thread.Comments {
			commentIdx := j + 1
			lines = append(lines, fmt.Sprintf("thread.%d.comment.%d.author=%s", idx, commentIdx, normalizeToken(comment.Author)))
			lines = append(lines, fmt.Sprintf("thread.%d.comment.%d.depth=%d", idx, commentIdx, comment.Depth))
			lines = append(lines, fmt.Sprintf("thread.%d.comment.%d.body=%s", idx, commentIdx, normalizeToken(comment.Body)))
		}
	}
	return resultWithIntents(strings.Join(lines, "\n"), prCommentsSummaryIntents(ref, sourceName, cmdCtx.State.PRCommentsFetches, threads)...), nil
}

type threadedComment struct {
	Author  string `json:"author"`
	Body    string `json:"body"`
	ReplyTo string `json:"reply_to"`
	Depth   int    `json:"-"`
}

type commentThread struct {
	Path     string            `json:"path"`
	Line     int               `json:"line"`
	Comments []threadedComment `json:"comments"`
}

type commentsPayload struct {
	Ref     string          `json:"ref"`
	Threads []commentThread `json:"threads"`
}

func resolvePRCommentThreads(state *RuntimeState, arg string) (string, string, []commentThread, error) {
	if arg == "" {
		return "current", "state", nil, nil
	}
	if strings.EqualFold(arg, "state") {
		payload, ok, err := loadPRCommentsPayloadFromState(state)
		if err != nil {
			return "", "", nil, err
		}
		if !ok {
			return "state", "state", nil, nil
		}
		ref := normalizeToken(payload.Ref)
		if ref == "-" {
			ref = "state"
		}
		threads := normalizeCommentThreads(payload.Threads)
		return ref, "state", threads, nil
	}
	if strings.HasPrefix(arg, "file=") {
		path := strings.TrimSpace(strings.TrimPrefix(arg, "file="))
		if path == "" {
			return "", "", nil, fmt.Errorf("usage: /pr-comments [status|<pr-ref>|file=<path>|state]")
		}
		payload, err := loadPRCommentsPayloadFromFile(path)
		if err != nil {
			return "", "", nil, err
		}
		ref := normalizeToken(payload.Ref)
		if ref == "-" {
			ref = filepath.Base(path)
		}
		threads := normalizeCommentThreads(payload.Threads)
		return ref, path, threads, nil
	}
	if payload, ok, err := tryLoadCommentsPayloadFromPath(arg); err != nil {
		return "", "", nil, err
	} else if ok {
		ref := normalizeToken(payload.Ref)
		if ref == "-" {
			ref = filepath.Base(arg)
		}
		threads := normalizeCommentThreads(payload.Threads)
		return ref, arg, threads, nil
	}
	return arg, "ref", nil, nil
}

func tryLoadCommentsPayloadFromPath(path string) (commentsPayload, bool, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return commentsPayload{}, false, nil
	}
	info, err := os.Stat(trimmed)
	if err != nil {
		if os.IsNotExist(err) {
			return commentsPayload{}, false, nil
		}
		return commentsPayload{}, false, fmt.Errorf("checking comments source: %w", err)
	}
	if info.IsDir() {
		return commentsPayload{}, false, fmt.Errorf("comments source is a directory: %s", trimmed)
	}
	payload, err := loadPRCommentsPayloadFromFile(trimmed)
	if err != nil {
		return commentsPayload{}, false, err
	}
	return payload, true, nil
}

func loadPRCommentsPayloadFromFile(path string) (commentsPayload, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return commentsPayload{}, fmt.Errorf("reading comments source: %w", err)
	}
	return decodeCommentsPayload(raw)
}

func loadPRCommentsPayloadFromState(state *RuntimeState) (commentsPayload, bool, error) {
	if state == nil || state.ConfigValues == nil {
		return commentsPayload{}, false, nil
	}
	if raw := strings.TrimSpace(state.ConfigValues["pr-comments.mock"]); raw != "" {
		payload, err := decodeCommentsPayload([]byte(raw))
		if err != nil {
			return commentsPayload{}, false, fmt.Errorf("invalid pr-comments.mock payload: %w", err)
		}
		return payload, true, nil
	}
	if path := strings.TrimSpace(state.ConfigValues["pr-comments.file"]); path != "" {
		payload, err := loadPRCommentsPayloadFromFile(path)
		if err != nil {
			return commentsPayload{}, false, err
		}
		return payload, true, nil
	}
	return commentsPayload{}, false, nil
}

func decodeCommentsPayload(raw []byte) (commentsPayload, error) {
	var payload commentsPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return commentsPayload{}, fmt.Errorf("parsing comments payload: %w", err)
	}
	return payload, nil
}

func normalizeCommentThreads(threads []commentThread) []commentThread {
	out := append([]commentThread(nil), threads...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Path != out[j].Path {
			return out[i].Path < out[j].Path
		}
		if out[i].Line != out[j].Line {
			return out[i].Line < out[j].Line
		}
		return len(out[i].Comments) < len(out[j].Comments)
	})
	for i := range out {
		commentDepths := make(map[string]int)
		for idx := range out[i].Comments {
			depth := 0
			replyTo := strings.TrimSpace(out[i].Comments[idx].ReplyTo)
			if replyTo != "" {
				depth = commentDepths[replyTo] + 1
			}
			out[i].Comments[idx].Depth = depth
			commentDepths[strconv.Itoa(idx+1)] = depth
		}
	}
	return out
}

// WebSetupCommand models /web-setup connectivity state.
type WebSetupCommand struct{}

func NewWebSetupCommand() *WebSetupCommand     { return &WebSetupCommand{} }
func (c *WebSetupCommand) Name() string        { return "web-setup" }
func (c *WebSetupCommand) Aliases() []string   { return nil }
func (c *WebSetupCommand) Description() string { return "Connect local session to Claude web" }
func (c *WebSetupCommand) Usage() string       { return "/web-setup [status|connect|disconnect]" }
func (c *WebSetupCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(inv.Args[0], "connect") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.WebSetupConnected = true
		cmdCtx.State.WebSetupCount++
		if strings.TrimSpace(cmdCtx.State.RemoteSessionURL) == "" {
			cmdCtx.State.RemoteSessionURL = "https://claude.ai/code"
		}
		return resultWithIntents(fmt.Sprintf("WEB_SETUP_CONNECT\nconnected=true\nurl=%s\ncount=%d", normalizeToken(cmdCtx.State.RemoteSessionURL), cmdCtx.State.WebSetupCount), webSetupConnectIntents(true, cmdCtx.State.RemoteSessionURL, cmdCtx.State.WebSetupCount)...), nil
	}
	if len(inv.Args) != 1 {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "status":
		return resultWithIntents(fmt.Sprintf("WEB_SETUP_STATUS\nconnected=%t\nurl=%s\ncount=%d", cmdCtx.State.WebSetupConnected, normalizeToken(cmdCtx.State.RemoteSessionURL), cmdCtx.State.WebSetupCount), webSetupStatusIntents(cmdCtx.State.WebSetupConnected, cmdCtx.State.RemoteSessionURL, cmdCtx.State.WebSetupCount)...), nil
	case "disconnect":
		cmdCtx.State.WebSetupConnected = false
		cmdCtx.State.WebSetupCount++
		return resultWithIntents(fmt.Sprintf("WEB_SETUP_CONNECT\nconnected=false\nurl=%s\ncount=%d", normalizeToken(cmdCtx.State.RemoteSessionURL), cmdCtx.State.WebSetupCount), webSetupConnectIntents(false, cmdCtx.State.RemoteSessionURL, cmdCtx.State.WebSetupCount)...), nil
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
}

// BridgeKickCommand models deterministic bridge fault injection intent.
type BridgeKickCommand struct{}

func NewBridgeKickCommand() *BridgeKickCommand   { return &BridgeKickCommand{} }
func (c *BridgeKickCommand) Name() string        { return "bridge-kick" }
func (c *BridgeKickCommand) Aliases() []string   { return nil }
func (c *BridgeKickCommand) Description() string { return "Inject deterministic bridge failure states" }
func (c *BridgeKickCommand) Usage() string {
	return "/bridge-kick [status|close <code>|poll <status>|reconnect]"
}
func (c *BridgeKickCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return resultWithIntents(fmt.Sprintf("BRIDGE_KICK_STATUS\ncount=%d\nlast_action=%s\nlast_code=%d", cmdCtx.State.BridgeKickCount, normalizeToken(cmdCtx.State.BridgeKickLastAction), cmdCtx.State.BridgeKickLastCode), bridgeKickStatusIntents(cmdCtx.State.BridgeKickCount, cmdCtx.State.BridgeKickLastAction, cmdCtx.State.BridgeKickLastCode)...), nil
	}
	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "close", "poll":
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		code, err := parseNonNegativeInt(inv.Args[1])
		if err != nil {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.BridgeKickCount++
		cmdCtx.State.BridgeKickLastAction = sub
		cmdCtx.State.BridgeKickLastCode = code
		return resultWithIntents(fmt.Sprintf("BRIDGE_KICK_APPLY\naction=%s\ncode=%d\ncount=%d", sub, code, cmdCtx.State.BridgeKickCount), bridgeKickApplyIntents(sub, code, cmdCtx.State.BridgeKickCount)...), nil
	case "reconnect":
		if len(inv.Args) != 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.BridgeKickCount++
		cmdCtx.State.BridgeKickLastAction = "reconnect"
		cmdCtx.State.BridgeKickLastCode = 0
		return resultWithIntents(fmt.Sprintf("BRIDGE_KICK_APPLY\naction=reconnect\ncode=0\ncount=%d", cmdCtx.State.BridgeKickCount), bridgeKickApplyIntents("reconnect", 0, cmdCtx.State.BridgeKickCount)...), nil
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
}

// BriefCommand toggles deterministic brief-only mode.
type BriefCommand struct{}

func NewBriefCommand() *BriefCommand        { return &BriefCommand{} }
func (c *BriefCommand) Name() string        { return "brief" }
func (c *BriefCommand) Aliases() []string   { return nil }
func (c *BriefCommand) Description() string { return "Toggle brief-only mode" }
func (c *BriefCommand) Usage() string       { return "/brief [on|off|toggle|status]" }
func (c *BriefCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) > 1 {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	mode := "toggle"
	if len(inv.Args) == 1 {
		mode = strings.ToLower(strings.TrimSpace(inv.Args[0]))
	}
	switch mode {
	case "status":
		return resultWithIntents(fmt.Sprintf("BRIEF_STATUS\nenabled=%t\ntoggles=%d", cmdCtx.State.BriefOnly, cmdCtx.State.BriefToggleCount), briefStatusIntents(cmdCtx.State.BriefOnly, cmdCtx.State.BriefToggleCount)...), nil
	case "toggle":
		cmdCtx.State.BriefOnly = !cmdCtx.State.BriefOnly
		cmdCtx.State.BriefToggleCount++
	case "on":
		if !cmdCtx.State.BriefOnly {
			cmdCtx.State.BriefToggleCount++
		}
		cmdCtx.State.BriefOnly = true
	case "off":
		if cmdCtx.State.BriefOnly {
			cmdCtx.State.BriefToggleCount++
		}
		cmdCtx.State.BriefOnly = false
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	return resultWithIntents(fmt.Sprintf("BRIEF_SET\nenabled=%t\ntoggles=%d", cmdCtx.State.BriefOnly, cmdCtx.State.BriefToggleCount), briefSetIntents(cmdCtx.State.BriefOnly, cmdCtx.State.BriefToggleCount)...), nil
}

// CommitCommand records deterministic local commit intent.
type CommitCommand struct{}

func NewCommitCommand() *CommitCommand       { return &CommitCommand{} }
func (c *CommitCommand) Name() string        { return "commit" }
func (c *CommitCommand) Aliases() []string   { return nil }
func (c *CommitCommand) Description() string { return "Build commit plan from workspace git summary" }
func (c *CommitCommand) Usage() string       { return "/commit [status|suggest|check|<message>]" }
func (c *CommitCommand) Execute(ctx context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	gitSummary := readGitWorkingTreeSummary(ctx)
	summary := summarizeDiffEntries(cmdCtx.State.DiffEntries)
	branch := activeBranch(cmdCtx.State)
	source := "state"
	if gitSummary.Available {
		source = "git"
		summary = diffSummary{
			Files:     gitSummary.Files,
			Staged:    gitSummary.Staged,
			Unstaged:  gitSummary.Unstaged,
			Untracked: gitSummary.Untracked,
			Added:     gitSummary.Added,
			Removed:   gitSummary.Removed,
			Modified:  gitSummary.Modified,
			FirstPath: "workspace",
		}
		branch = gitSummary.Branch
	}
	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		lines := []string{
			"COMMIT_STATUS",
			fmt.Sprintf("count=%d", cmdCtx.State.CommitCount),
			fmt.Sprintf("last_message=%s", normalizeToken(cmdCtx.State.LastCommitMessage)),
			fmt.Sprintf("source=%s", source),
			fmt.Sprintf("branch=%s", normalizeToken(branch)),
			fmt.Sprintf("files=%d", summary.Files),
			fmt.Sprintf("staged=%d", summary.Staged),
			fmt.Sprintf("unstaged=%d", summary.Unstaged),
			fmt.Sprintf("untracked=%d", summary.Untracked),
		}
		if gitSummary.Err != "" {
			lines = append(lines, fmt.Sprintf("status=%s", normalizeToken(gitSummary.Err)))
		}
		return resultWithIntents(strings.Join(lines, "\n"), legacyOutputIntents(strings.Join(lines, "\n"))...), nil
	}
	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "suggest") {
		message := suggestCommitMessage(summary)
		return resultWithIntents(fmt.Sprintf("COMMIT_SUGGEST\nsource=%s\nbranch=%s\nfiles=%d\nsuggested_message=%s", source, normalizeToken(branch), summary.Files, normalizeToken(message)), legacyOutputIntents(fmt.Sprintf("COMMIT_SUGGEST\nsource=%s\nbranch=%s\nfiles=%d\nsuggested_message=%s", source, normalizeToken(branch), summary.Files, normalizeToken(message)))...), nil
	}
	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "check") {
		ready := summary.Files > 0 && summary.Staged > 0
		reason := "ready"
		if summary.Files == 0 {
			reason = "no_changes"
		} else if summary.Staged == 0 {
			reason = "not_staged"
		}
		return resultWithIntents(fmt.Sprintf("COMMIT_CHECK\nready=%t\nreason=%s\nfiles=%d\nstaged=%d\nunstaged=%d\nuntracked=%d", ready, reason, summary.Files, summary.Staged, summary.Unstaged, summary.Untracked), legacyOutputIntents(fmt.Sprintf("COMMIT_CHECK\nready=%t\nreason=%s\nfiles=%d\nstaged=%d\nunstaged=%d\nuntracked=%d", ready, reason, summary.Files, summary.Staged, summary.Unstaged, summary.Untracked))...), nil
	}

	message := strings.TrimSpace(strings.Join(inv.Args, " "))
	if message == "" {
		message = suggestCommitMessage(summary)
	}
	cmdCtx.State.LastCommitMessage = message
	if summary.Files > 0 {
		cmdCtx.State.CommitCount++
	}
	steps := buildCommitPlanSteps(summary)
	lines := []string{
		"COMMIT_PLAN",
		fmt.Sprintf("ready=%t", summary.Files > 0),
		fmt.Sprintf("source=%s", source),
		fmt.Sprintf("branch=%s", normalizeToken(branch)),
		fmt.Sprintf("files=%d", summary.Files),
		fmt.Sprintf("staged=%d", summary.Staged),
		fmt.Sprintf("unstaged=%d", summary.Unstaged),
		fmt.Sprintf("untracked=%d", summary.Untracked),
		fmt.Sprintf("added=%d", summary.Added),
		fmt.Sprintf("removed=%d", summary.Removed),
		fmt.Sprintf("modified=%d", summary.Modified),
		fmt.Sprintf("suggested_message=%s", normalizeToken(message)),
		fmt.Sprintf("steps=%d", len(steps)),
	}
	for i, step := range steps {
		lines = append(lines, fmt.Sprintf("step.%d.status=%s", i+1, normalizeToken(step.Status)))
		lines = append(lines, fmt.Sprintf("step.%d.action=%s", i+1, normalizeToken(step.Action)))
	}
	return resultWithIntents(strings.Join(lines, "\n"), legacyOutputIntents(strings.Join(lines, "\n"))...), nil
}

// CommitPushPRCommand records deterministic commit+PR flow intent.
type CommitPushPRCommand struct{}

func NewCommitPushPRCommand() *CommitPushPRCommand { return &CommitPushPRCommand{} }
func (c *CommitPushPRCommand) Name() string        { return "commit-push-pr" }
func (c *CommitPushPRCommand) Aliases() []string   { return nil }
func (c *CommitPushPRCommand) Description() string {
	return "Validate and stage commit/push/pr execution plan"
}
func (c *CommitPushPRCommand) Usage() string {
	return "/commit-push-pr [status|doctor|--force] [--base <branch>] [<title>]"
}
func (c *CommitPushPRCommand) Execute(ctx context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		branch := activeBranch(cmdCtx.State)
		upstream, tracked := detectBranchTracking(cmdCtx.State, branch)
		remoteReady := tracked
		source := "state"
		gitSummary := readGitWorkingTreeSummary(ctx)
		if gitSummary.Available {
			source = "git"
			branch = gitSummary.Branch
			upstream = gitSummary.Upstream
			tracked = gitSummary.Tracked
			remoteReady = gitSummary.Remote && gitSummary.Tracked
		}
		lines := []string{
			"COMMIT_PUSH_PR_STATUS",
			fmt.Sprintf("count=%d", cmdCtx.State.CommitPushPRCount),
			fmt.Sprintf("last_url=%s", normalizeToken(cmdCtx.State.LastPRURL)),
			fmt.Sprintf("source=%s", source),
			fmt.Sprintf("branch=%s", normalizeToken(branch)),
			fmt.Sprintf("tracked=%t", tracked),
			fmt.Sprintf("upstream=%s", normalizeToken(upstream)),
			fmt.Sprintf("remote_ready=%t", remoteReady),
		}
		if gitSummary.Err != "" {
			lines = append(lines, fmt.Sprintf("status=%s", normalizeToken(gitSummary.Err)))
		}
		return resultWithIntents(strings.Join(lines, "\n"), legacyOutputIntents(strings.Join(lines, "\n"))...), nil
	}
	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "doctor") {
		branch := activeBranch(cmdCtx.State)
		upstream, tracked := detectBranchTracking(cmdCtx.State, branch)
		summary := summarizeDiffEntries(cmdCtx.State.DiffEntries)
		quickFix := "/commit check"
		if tracked && summary.Staged > 0 {
			quickFix = "/commit-push-pr --force"
		}
		return resultWithIntents(fmt.Sprintf("COMMIT_PUSH_PR_DOCTOR\nbranch=%s\ntracked=%t\nupstream=%s\nfiles=%d\nstaged=%d\nquick_fix=%s", normalizeToken(branch), tracked, normalizeToken(upstream), summary.Files, summary.Staged, normalizeToken(quickFix)), legacyOutputIntents(fmt.Sprintf("COMMIT_PUSH_PR_DOCTOR\nbranch=%s\ntracked=%t\nupstream=%s\nfiles=%d\nstaged=%d\nquick_fix=%s", normalizeToken(branch), tracked, normalizeToken(upstream), summary.Files, summary.Staged, normalizeToken(quickFix)))...), nil
	}
	opts, err := parseCommitPushPROptions(inv.Args)
	if err != nil {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	branch := activeBranch(cmdCtx.State)
	upstream, tracked := detectBranchTracking(cmdCtx.State, branch)
	summary := summarizeDiffEntries(cmdCtx.State.DiffEntries)
	validate := "ok"
	if summary.Files == 0 {
		validate = "blocked:no_changes"
	}
	stage := "ok"
	if summary.Staged == 0 && summary.Files > 0 {
		stage = "blocked:not_staged"
	}
	track := "ok"
	if !tracked {
		track = "blocked:missing_upstream"
	}
	commitStep, pushStep, prStep := "blocked", "blocked", "blocked"
	if validate == "ok" && stage == "ok" && track == "ok" {
		if opts.Force {
			commitStep, pushStep, prStep = "executed", "executed", "executed"
		} else {
			commitStep, pushStep, prStep = "dry-run", "dry-run", "dry-run"
		}
	}
	title := normalizeToken(opts.Title)
	if title == "-" {
		title = suggestCommitMessage(summary)
	}
	url := "-"
	if opts.Force && prStep == "executed" {
		cmdCtx.State.CommitPushPRCount++
		url = fmt.Sprintf("https://example.invalid/%s/pull/%d", strings.ReplaceAll(branch, "/", "-"), cmdCtx.State.CommitPushPRCount)
		cmdCtx.State.LastPRURL = url
	} else if strings.TrimSpace(cmdCtx.State.LastPRURL) != "" {
		url = cmdCtx.State.LastPRURL
	}
	mode := "dry-run"
	if opts.Force {
		mode = "execute"
	}
	lines := []string{
		"COMMIT_PUSH_PR_PLAN",
		fmt.Sprintf("mode=%s", mode),
		fmt.Sprintf("branch=%s", normalizeToken(branch)),
		fmt.Sprintf("base=%s", normalizeToken(opts.Base)),
		fmt.Sprintf("tracked=%t", tracked),
		fmt.Sprintf("upstream=%s", normalizeToken(upstream)),
		fmt.Sprintf("title=%s", normalizeToken(title)),
		fmt.Sprintf("step.1.validate_git_state=%s", validate),
		fmt.Sprintf("step.2.validate_staged_changes=%s", stage),
		fmt.Sprintf("step.3.validate_branch_tracking=%s", track),
		fmt.Sprintf("step.4.commit=%s", commitStep),
		fmt.Sprintf("step.5.push=%s", pushStep),
		fmt.Sprintf("step.6.open_pr=%s", prStep),
		fmt.Sprintf("url=%s", normalizeToken(url)),
	}
	return resultWithIntents(strings.Join(lines, "\n"), legacyOutputIntents(strings.Join(lines, "\n"))...), nil
}

type diffSummary struct {
	Files     int
	Staged    int
	Unstaged  int
	Untracked int
	Added     int
	Removed   int
	Modified  int
	FirstPath string
}

func summarizeDiffEntries(entries []DiffEntry) diffSummary {
	summary := diffSummary{}
	sorted := append([]DiffEntry(nil), entries...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Path < sorted[j].Path
	})
	for _, entry := range sorted {
		if strings.TrimSpace(entry.Path) == "" {
			continue
		}
		if summary.FirstPath == "" {
			summary.FirstPath = entry.Path
		}
		summary.Files++
		summary.Added += entry.Added
		summary.Removed += entry.Removed
		summary.Modified += entry.Modified
		switch {
		case entry.Untracked:
			summary.Untracked++
		case entry.Staged:
			summary.Staged++
		default:
			summary.Unstaged++
		}
	}
	return summary
}

func suggestCommitMessage(summary diffSummary) string {
	if summary.Files == 0 {
		return "-"
	}
	prefix := "chore"
	if summary.Untracked > 0 {
		prefix = "feat"
	} else if summary.Removed > summary.Added+summary.Modified {
		prefix = "refactor"
	} else if summary.Modified > 0 {
		prefix = "fix"
	}
	scope := summary.FirstPath
	if scope == "" {
		scope = "workspace"
	}
	if strings.Contains(scope, "/") {
		scope = strings.Split(scope, "/")[0]
	}
	return fmt.Sprintf("%s: update %s changes", prefix, scope)
}

func activeBranch(state *RuntimeState) string {
	if state == nil {
		return "main"
	}
	ensureBranchState(state)
	return state.ActiveBranch
}

type commitPushPROptions struct {
	Force bool
	Base  string
	Title string
}

func parseCommitPushPROptions(args []string) (commitPushPROptions, error) {
	opts := commitPushPROptions{Base: "main"}
	var titleParts []string
	for i := 0; i < len(args); i++ {
		arg := strings.TrimSpace(args[i])
		if arg == "" {
			continue
		}
		switch arg {
		case "--force":
			opts.Force = true
		case "--base":
			i++
			if i >= len(args) {
				return commitPushPROptions{}, fmt.Errorf("missing base branch")
			}
			opts.Base = strings.TrimSpace(args[i])
			if opts.Base == "" {
				return commitPushPROptions{}, fmt.Errorf("missing base branch")
			}
		default:
			if strings.HasPrefix(arg, "--") {
				return commitPushPROptions{}, fmt.Errorf("unknown option: %s", arg)
			}
			titleParts = append(titleParts, arg)
		}
	}
	opts.Title = strings.TrimSpace(strings.Join(titleParts, " "))
	return opts, nil
}

func detectBranchTracking(state *RuntimeState, branch string) (string, bool) {
	branch = strings.TrimSpace(branch)
	if state != nil && state.ConfigValues != nil {
		if upstream := strings.TrimSpace(state.ConfigValues["git.upstream."+branch]); upstream != "" {
			return upstream, true
		}
		if upstream := strings.TrimSpace(state.ConfigValues["git.upstream"]); upstream != "" {
			return upstream, true
		}
	}
	if branch == "main" || branch == "master" {
		return "origin/" + branch, true
	}
	return "-", false
}

// InitVerifiersCommand records deterministic verifier scaffolding intent.
type InitVerifiersCommand struct{}

func NewInitVerifiersCommand() *InitVerifiersCommand { return &InitVerifiersCommand{} }
func (c *InitVerifiersCommand) Name() string         { return "init-verifiers" }
func (c *InitVerifiersCommand) Aliases() []string    { return nil }
func (c *InitVerifiersCommand) Description() string {
	return "Create deterministic verifier scaffold metadata"
}
func (c *InitVerifiersCommand) Usage() string { return "/init-verifiers [status|<verifier-name>]" }
func (c *InitVerifiersCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		return resultWithIntents(fmt.Sprintf("INIT_VERIFIERS_STATUS\ncount=%d\nlast_name=%s", cmdCtx.State.InitVerifiersCount, normalizeToken(cmdCtx.State.LastVerifierName)), initVerifiersStatusIntents(cmdCtx.State.InitVerifiersCount, cmdCtx.State.LastVerifierName)...), nil
	}
	name := strings.TrimSpace(strings.Join(inv.Args, " "))
	if name == "" {
		name = "verifier-default"
	}
	cmdCtx.State.InitVerifiersCount++
	cmdCtx.State.LastVerifierName = name
	return resultWithIntents(fmt.Sprintf("INIT_VERIFIERS_CREATED\nname=%s\ncount=%d", normalizeToken(name), cmdCtx.State.InitVerifiersCount), initVerifiersCreatedIntents(name, cmdCtx.State.InitVerifiersCount)...), nil
}

// InsightsCommand records deterministic analytics-report generation intent.
type InsightsCommand struct{}

func NewInsightsCommand() *InsightsCommand   { return &InsightsCommand{} }
func (c *InsightsCommand) Name() string      { return "insights" }
func (c *InsightsCommand) Aliases() []string { return nil }
func (c *InsightsCommand) Description() string {
	return "Generate deterministic session insights summary"
}
func (c *InsightsCommand) Usage() string { return "/insights [status|report [scope]]" }
func (c *InsightsCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 {
		inv.Args = []string{"report"}
	}
	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		return resultWithIntents(fmt.Sprintf("INSIGHTS_STATUS\ncount=%d\nlast_scope=%s", cmdCtx.State.InsightsCount, normalizeToken(cmdCtx.State.LastInsightsScope)), insightsStatusIntents(cmdCtx.State.InsightsCount, cmdCtx.State.LastInsightsScope)...), nil
	}
	if !strings.EqualFold(strings.TrimSpace(inv.Args[0]), "report") || len(inv.Args) > 2 {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	scope := "workspace"
	if len(inv.Args) == 2 {
		scope = strings.TrimSpace(inv.Args[1])
		if scope == "" {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
	}
	cmdCtx.State.InsightsCount++
	cmdCtx.State.LastInsightsScope = scope
	return resultWithIntents(fmt.Sprintf("INSIGHTS_REPORT\nscope=%s\ncount=%d", normalizeToken(scope), cmdCtx.State.InsightsCount), insightsReportIntents(scope, cmdCtx.State.InsightsCount)...), nil
}

// PassesCommand records deterministic referral-pass interactions.
type PassesCommand struct{}

func NewPassesCommand() *PassesCommand       { return &PassesCommand{} }
func (c *PassesCommand) Name() string        { return "passes" }
func (c *PassesCommand) Aliases() []string   { return nil }
func (c *PassesCommand) Description() string { return "Track deterministic guest pass availability" }
func (c *PassesCommand) Usage() string       { return "/passes [status|claim [count]]" }
func (c *PassesCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if cmdCtx.State.PassesRemaining <= 0 {
		cmdCtx.State.PassesRemaining = 5
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return resultWithIntents(fmt.Sprintf("PASSES_STATUS\nvisits=%d\nremaining=%d", cmdCtx.State.PassesVisitCount, cmdCtx.State.PassesRemaining), passesStatusIntents(cmdCtx.State.PassesVisitCount, cmdCtx.State.PassesRemaining)...), nil
	}
	if !strings.EqualFold(strings.TrimSpace(inv.Args[0]), "claim") || len(inv.Args) > 2 {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	claimCount := 1
	if len(inv.Args) == 2 {
		value, err := parseNonNegativeInt(inv.Args[1])
		if err != nil || value == 0 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		claimCount = value
	}
	if claimCount > cmdCtx.State.PassesRemaining {
		claimCount = cmdCtx.State.PassesRemaining
	}
	cmdCtx.State.PassesRemaining -= claimCount
	cmdCtx.State.PassesVisitCount++
	return resultWithIntents(fmt.Sprintf("PASSES_CLAIM\nclaimed=%d\nremaining=%d\nvisits=%d", claimCount, cmdCtx.State.PassesRemaining, cmdCtx.State.PassesVisitCount), passesClaimIntents(claimCount, cmdCtx.State.PassesRemaining, cmdCtx.State.PassesVisitCount)...), nil
}

// RenameCommand stores deterministic conversation title state.
type RenameCommand struct{}

func NewRenameCommand() *RenameCommand       { return &RenameCommand{} }
func (c *RenameCommand) Name() string        { return "rename" }
func (c *RenameCommand) Aliases() []string   { return nil }
func (c *RenameCommand) Description() string { return "Rename current conversation deterministically" }
func (c *RenameCommand) Usage() string       { return "/rename [status|<name>]" }
func (c *RenameCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		updated := countHistoryTitleMatches(cmdCtx.State, cmdCtx.State.SessionTitle)
		return resultWithIntents(fmt.Sprintf("RENAME_STATUS\ncount=%d\nname=%s\nhistory_matches=%d", cmdCtx.State.RenameCount, normalizeToken(cmdCtx.State.SessionTitle), updated), renameStatusIntents(cmdCtx.State.RenameCount, cmdCtx.State.SessionTitle, updated)...), nil
	}
	name := strings.TrimSpace(strings.Join(inv.Args, " "))
	if name == "" {
		name = "Untitled Session"
	}
	cmdCtx.State.RenameCount++
	cmdCtx.State.SessionTitle = name
	updated := persistSessionTitleToHistory(cmdCtx.State, name)
	return resultWithIntents(fmt.Sprintf("RENAME_SET\nname=%s\ncount=%d\nhistory_updated=%d", normalizeToken(name), cmdCtx.State.RenameCount, updated), renameSetIntents(name, cmdCtx.State.RenameCount, updated)...), nil
}

func persistSessionTitleToHistory(state *RuntimeState, title string) int {
	if state == nil || len(state.HistoryEntries) == 0 {
		return 0
	}
	updated := 0
	for i := range state.HistoryEntries {
		entry := &state.HistoryEntries[i]
		if strings.TrimSpace(state.SessionID) != "" && strings.TrimSpace(entry.ID) == strings.TrimSpace(state.SessionID) {
			entry.Title = title
			updated++
			continue
		}
		if strings.TrimSpace(state.SessionPath) != "" && strings.TrimSpace(entry.Path) == strings.TrimSpace(state.SessionPath) {
			entry.Title = title
			updated++
		}
	}
	if updated > 0 {
		return updated
	}
	latestIdx := 0
	for i := range state.HistoryEntries {
		if state.HistoryEntries[i].CreatedAt > state.HistoryEntries[latestIdx].CreatedAt {
			latestIdx = i
		}
	}
	state.HistoryEntries[latestIdx].Title = title
	return 1
}

func countHistoryTitleMatches(state *RuntimeState, title string) int {
	if state == nil {
		return 0
	}
	count := 0
	for _, entry := range state.HistoryEntries {
		if strings.TrimSpace(entry.Title) == strings.TrimSpace(title) && strings.TrimSpace(title) != "" {
			count++
		}
	}
	return count
}

// StickersCommand records deterministic sticker-order intent.
type StickersCommand struct{}

func NewStickersCommand() *StickersCommand     { return &StickersCommand{} }
func (c *StickersCommand) Name() string        { return "stickers" }
func (c *StickersCommand) Aliases() []string   { return nil }
func (c *StickersCommand) Description() string { return "Order Claude Code stickers" }
func (c *StickersCommand) Usage() string       { return "/stickers [status|order]" }
func (c *StickersCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "order") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.StickersCount++
		return resultWithIntents(fmt.Sprintf("STICKERS_ORDER\ncount=%d\nurl=https://www.stickermule.com/claudecode", cmdCtx.State.StickersCount), stickersOrderIntents(cmdCtx.State.StickersCount, "https://www.stickermule.com/claudecode")...), nil
	}
	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		return resultWithIntents(fmt.Sprintf("STICKERS_STATUS\ncount=%d", cmdCtx.State.StickersCount), stickersStatusIntents(cmdCtx.State.StickersCount)...), nil
	}
	return Result{}, fmt.Errorf("usage: %s", c.Usage())
}

// ThinkbackCommand records deterministic think-back lifecycle intent.
type ThinkbackCommand struct{}

func NewThinkbackCommand() *ThinkbackCommand    { return &ThinkbackCommand{} }
func (c *ThinkbackCommand) Name() string        { return "think-back" }
func (c *ThinkbackCommand) Aliases() []string   { return []string{"thinkback"} }
func (c *ThinkbackCommand) Description() string { return "Generate year-in-review metadata" }
func (c *ThinkbackCommand) Usage() string       { return "/think-back [status|generate|play]" }
func (c *ThinkbackCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 {
		inv.Args = []string{"generate"}
	}
	if len(inv.Args) != 1 {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "status":
		return resultWithIntents(fmt.Sprintf("THINKBACK_STATUS\ncount=%d\nlast_action=%s", cmdCtx.State.ThinkbackCount, normalizeToken(cmdCtx.State.ThinkbackLastAction)), thinkbackStatusIntents(cmdCtx.State.ThinkbackCount, cmdCtx.State.ThinkbackLastAction)...), nil
	case "generate", "play":
		cmdCtx.State.ThinkbackCount++
		cmdCtx.State.ThinkbackLastAction = sub
		return resultWithIntents(fmt.Sprintf("THINKBACK_ACTION\naction=%s\ncount=%d", sub, cmdCtx.State.ThinkbackCount), thinkbackActionIntents(sub, cmdCtx.State.ThinkbackCount)...), nil
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
}

// ThinkbackPlayCommand models hidden deterministic animation playback.
type ThinkbackPlayCommand struct{}

func NewThinkbackPlayCommand() *ThinkbackPlayCommand { return &ThinkbackPlayCommand{} }
func (c *ThinkbackPlayCommand) Name() string         { return "thinkback-play" }
func (c *ThinkbackPlayCommand) Aliases() []string    { return nil }
func (c *ThinkbackPlayCommand) Description() string  { return "Play generated thinkback animation" }
func (c *ThinkbackPlayCommand) Usage() string        { return "/thinkback-play [status|play]" }
func (c *ThinkbackPlayCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "play") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.ThinkbackPlayCount++
		return resultWithIntents(fmt.Sprintf("THINKBACK_PLAY\ncount=%d", cmdCtx.State.ThinkbackPlayCount), thinkbackPlayRunIntents(cmdCtx.State.ThinkbackPlayCount)...), nil
	}
	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		return resultWithIntents(fmt.Sprintf("THINKBACK_PLAY_STATUS\ncount=%d", cmdCtx.State.ThinkbackPlayCount), thinkbackPlayStatusIntents(cmdCtx.State.ThinkbackPlayCount)...), nil
	}
	return Result{}, fmt.Errorf("usage: %s", c.Usage())
}

// TeleportCommand stores deterministic handoff target intent.
type TeleportCommand struct{}

func NewTeleportCommand() *TeleportCommand     { return &TeleportCommand{} }
func (c *TeleportCommand) Name() string        { return "teleport" }
func (c *TeleportCommand) Aliases() []string   { return nil }
func (c *TeleportCommand) Description() string { return "Teleport deterministic execution target" }
func (c *TeleportCommand) Usage() string       { return "/teleport [status|<target>]" }
func (c *TeleportCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || (len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status")) {
		if len(inv.Args) == 0 {
			return resultWithIntents(fmt.Sprintf("TELEPORT_STATUS\ncount=%d\nlast_target=%s", cmdCtx.State.TeleportCount, normalizeToken(cmdCtx.State.TeleportLastTarget)), teleportStatusIntents(cmdCtx.State.TeleportCount, cmdCtx.State.TeleportLastTarget)...), nil
		}
		return resultWithIntents(fmt.Sprintf("TELEPORT_STATUS\ncount=%d\nlast_target=%s", cmdCtx.State.TeleportCount, normalizeToken(cmdCtx.State.TeleportLastTarget)), teleportStatusIntents(cmdCtx.State.TeleportCount, cmdCtx.State.TeleportLastTarget)...), nil
	}
	target := strings.TrimSpace(strings.Join(inv.Args, " "))
	if target == "" {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	cmdCtx.State.TeleportCount++
	cmdCtx.State.TeleportLastTarget = target
	return resultWithIntents(fmt.Sprintf("TELEPORT_SET\ntarget=%s\ncount=%d", normalizeToken(target), cmdCtx.State.TeleportCount), teleportSetIntents(target, cmdCtx.State.TeleportCount)...), nil
}

// SummaryCommand records deterministic summary refresh requests.
type SummaryCommand struct{}

func NewSummaryCommand() *SummaryCommand      { return &SummaryCommand{} }
func (c *SummaryCommand) Name() string        { return "summary" }
func (c *SummaryCommand) Aliases() []string   { return nil }
func (c *SummaryCommand) Description() string { return "Generate deterministic local summary" }
func (c *SummaryCommand) Usage() string       { return "/summary [status|refresh]" }
func (c *SummaryCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "refresh") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.SummaryCount++
		return resultWithIntents(fmt.Sprintf("SUMMARY_REFRESH\ncount=%d", cmdCtx.State.SummaryCount), summaryRefreshIntents(cmdCtx.State.SummaryCount)...), nil
	}
	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		return resultWithIntents(fmt.Sprintf("SUMMARY_STATUS\ncount=%d", cmdCtx.State.SummaryCount), summaryStatusIntents(cmdCtx.State.SummaryCount)...), nil
	}
	return Result{}, fmt.Errorf("usage: %s", c.Usage())
}

// ResetLimitsCommand resets deterministic local counters only.
type ResetLimitsCommand struct{}

func NewResetLimitsCommand() *ResetLimitsCommand { return &ResetLimitsCommand{} }
func (c *ResetLimitsCommand) Name() string       { return "reset-limits" }
func (c *ResetLimitsCommand) Aliases() []string  { return nil }
func (c *ResetLimitsCommand) Description() string {
	return "Reset deterministic local usage counters"
}
func (c *ResetLimitsCommand) Usage() string { return "/reset-limits [status|all]" }
func (c *ResetLimitsCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "all") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.CostInputTokens = 0
		cmdCtx.State.CostOutputTokens = 0
		cmdCtx.State.CostCacheRead = 0
		cmdCtx.State.CostCacheWrite = 0
		cmdCtx.State.RateLimitPrompts = 0
		cmdCtx.State.ResetLimitsCount++
		return resultWithIntents(fmt.Sprintf("RESET_LIMITS\ncount=%d", cmdCtx.State.ResetLimitsCount), resetLimitsApplyIntents(cmdCtx.State.ResetLimitsCount)...), nil
	}
	if len(inv.Args) == 1 && strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		return resultWithIntents(fmt.Sprintf("RESET_LIMITS_STATUS\ncount=%d", cmdCtx.State.ResetLimitsCount), resetLimitsStatusIntents(cmdCtx.State.ResetLimitsCount)...), nil
	}
	return Result{}, fmt.Errorf("usage: %s", c.Usage())
}

// EnvCommand models deterministic env key-value storage.
type EnvCommand struct{}

func NewEnvCommand() *EnvCommand          { return &EnvCommand{} }
func (c *EnvCommand) Name() string        { return "env" }
func (c *EnvCommand) Aliases() []string   { return nil }
func (c *EnvCommand) Description() string { return "Manage deterministic local environment view" }
func (c *EnvCommand) Usage() string       { return "/env [status|set <key> <value>|get <key>]" }
func (c *EnvCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if cmdCtx.State.ConfigValues == nil {
		cmdCtx.State.ConfigValues = make(map[string]string)
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		count := 0
		for key := range cmdCtx.State.ConfigValues {
			if strings.HasPrefix(key, "env.") {
				count++
			}
		}
		message := fmt.Sprintf("ENV_STATUS\nkeys=%d\nsets=%d", count, cmdCtx.State.EnvSetCount)
		return resultWithIntents(message, envStatusIntents(count, cmdCtx.State.EnvSetCount)...), nil
	}
	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "set":
		if len(inv.Args) < 3 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		key := strings.TrimSpace(inv.Args[1])
		if key == "" {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		value := strings.TrimSpace(strings.Join(inv.Args[2:], " "))
		if value == "" {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.ConfigValues["env."+key] = value
		cmdCtx.State.EnvSetCount++
		message := fmt.Sprintf("ENV_SET\nkey=%s\nvalue=%s\nsets=%d", normalizeToken(key), normalizeToken(value), cmdCtx.State.EnvSetCount)
		return resultWithIntents(message, envSetIntents(key, value, cmdCtx.State.EnvSetCount)...), nil
	case "get":
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		key := strings.TrimSpace(inv.Args[1])
		if key == "" {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		value, ok := cmdCtx.State.ConfigValues["env."+key]
		if !ok {
			message := fmt.Sprintf("ENV_GET\nkey=%s\nfound=false", normalizeToken(key))
			return resultWithIntents(message, envLookupIntents(key, false, "")...), nil
		}
		message := fmt.Sprintf("ENV_GET\nkey=%s\nfound=true\nvalue=%s", normalizeToken(key), normalizeToken(value))
		return resultWithIntents(message, envLookupIntents(key, true, value)...), nil
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
}

// IssueCommand provides provider-agnostic issue lifecycle operations.
type IssueCommand struct{}

func NewIssueCommand() *IssueCommand        { return &IssueCommand{} }
func (c *IssueCommand) Name() string        { return "issue" }
func (c *IssueCommand) Aliases() []string   { return []string{"issues"} }
func (c *IssueCommand) Description() string { return "Manage provider-agnostic issue state" }
func (c *IssueCommand) Usage() string {
	return "/issue [status|list|provider <name>|create <title>|open <id>|close <id>|assign <id> <assignee>|label <id> <label>|unlabel <id> <label>]"
}

func (c *IssueCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	return executeIssueCommand(cmdCtx, inv)
}

// WorkflowsCommand provides provider-agnostic workflow operations.
type WorkflowsCommand struct{}

func NewWorkflowsCommand() *WorkflowsCommand { return &WorkflowsCommand{} }
func (c *WorkflowsCommand) Name() string     { return "workflows" }
func (c *WorkflowsCommand) Aliases() []string {
	return []string{"workflow"}
}
func (c *WorkflowsCommand) Description() string { return "Manage provider-agnostic workflow runs" }
func (c *WorkflowsCommand) Usage() string {
	return "/workflows [status|list|run <name>|complete <name>|fail <name> [reason]|cancel <name>|rerun <name>]"
}

func (c *WorkflowsCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	return executeWorkflowsCommand(cmdCtx, inv)
}

// ProactiveCommand configures provider-agnostic proactive behavior.
type ProactiveCommand struct{}

func NewProactiveCommand() *ProactiveCommand { return &ProactiveCommand{} }
func (c *ProactiveCommand) Name() string     { return "proactive" }
func (c *ProactiveCommand) Aliases() []string {
	return []string{"pro"}
}
func (c *ProactiveCommand) Description() string { return "Manage proactive automation policy" }
func (c *ProactiveCommand) Usage() string {
	return "/proactive [status|on|off|rule list|rule add <rule>|rule remove <rule>|trigger <event>]"
}

func (c *ProactiveCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		message := fmt.Sprintf("PROACTIVE_STATUS\nenabled=%t\nrules=%d\nlast_action=%s", cmdCtx.State.ProactiveEnabled, len(cmdCtx.State.ProactiveRules), normalizeToken(cmdCtx.State.ProactiveLastAction))
		return resultWithIntents(message, proactiveStatusIntents(cmdCtx.State.ProactiveEnabled, cmdCtx.State.ProactiveRules, cmdCtx.State.ProactiveLastAction)...), nil
	}
	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "on":
		if len(inv.Args) != 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.ProactiveEnabled = true
		cmdCtx.State.ProactiveLastAction = "on"
		return resultWithIntents("PROACTIVE_SET\nenabled=true", proactiveMutationIntents("Proactive enabled", detailRow("Enabled", "yes", "enabled", "Proactive automation enabled."))...), nil
	case "off":
		if len(inv.Args) != 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.ProactiveEnabled = false
		cmdCtx.State.ProactiveLastAction = "off"
		return resultWithIntents("PROACTIVE_SET\nenabled=false", proactiveMutationIntents("Proactive disabled", detailRow("Enabled", "no", "disabled", "Proactive automation disabled."))...), nil
	case "rule":
		if len(inv.Args) < 2 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		action := strings.ToLower(strings.TrimSpace(inv.Args[1]))
		switch action {
		case "list":
			if len(inv.Args) != 2 {
				return Result{}, fmt.Errorf("usage: %s", c.Usage())
			}
			rules := uniqueSortedStrings(append([]string(nil), cmdCtx.State.ProactiveRules...))
			lines := []string{"PROACTIVE_RULES", fmt.Sprintf("count=%d", len(rules))}
			for i, rule := range rules {
				lines = append(lines, fmt.Sprintf("rule.%d=%s", i+1, normalizeToken(rule)))
			}
			return resultWithIntents(strings.Join(lines, "\n"), proactiveStatusIntents(cmdCtx.State.ProactiveEnabled, rules, cmdCtx.State.ProactiveLastAction)...), nil
		case "add":
			if len(inv.Args) < 3 {
				return Result{}, fmt.Errorf("usage: %s", c.Usage())
			}
			rule := strings.TrimSpace(strings.Join(inv.Args[2:], " "))
			if rule == "" {
				return Result{}, fmt.Errorf("usage: %s", c.Usage())
			}
			before := len(cmdCtx.State.ProactiveRules)
			cmdCtx.State.ProactiveRules = uniqueSortedStrings(append(cmdCtx.State.ProactiveRules, rule))
			cmdCtx.State.ProactiveLastAction = "rule-add"
			message := fmt.Sprintf("PROACTIVE_RULE_ADD\nrule=%s\nadded=%t\ncount=%d", normalizeToken(rule), len(cmdCtx.State.ProactiveRules) > before, len(cmdCtx.State.ProactiveRules))
			return resultWithIntents(message, proactiveMutationIntents("Proactive rule added", detailRow("Rule", normalizeToken(rule), "rule", "Rule value."), detailRow("Added", boolState(len(cmdCtx.State.ProactiveRules) > before, "yes", "no"), "state", "Whether rule was newly added."), detailRow("Count", itoa(len(cmdCtx.State.ProactiveRules)), "count", "Total rules."))...), nil
		case "remove":
			if len(inv.Args) < 3 {
				return Result{}, fmt.Errorf("usage: %s", c.Usage())
			}
			rule := strings.TrimSpace(strings.Join(inv.Args[2:], " "))
			if rule == "" {
				return Result{}, fmt.Errorf("usage: %s", c.Usage())
			}
			removed := containsString(cmdCtx.State.ProactiveRules, rule)
			cmdCtx.State.ProactiveRules = removeStringValue(cmdCtx.State.ProactiveRules, rule)
			cmdCtx.State.ProactiveLastAction = "rule-remove"
			message := fmt.Sprintf("PROACTIVE_RULE_REMOVE\nrule=%s\nremoved=%t\ncount=%d", normalizeToken(rule), removed, len(cmdCtx.State.ProactiveRules))
			return resultWithIntents(message, proactiveMutationIntents("Proactive rule removed", detailRow("Rule", normalizeToken(rule), "rule", "Rule value."), detailRow("Removed", boolState(removed, "yes", "no"), "state", "Whether rule was removed."), detailRow("Count", itoa(len(cmdCtx.State.ProactiveRules)), "count", "Remaining rules."))...), nil
		default:
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
	case "trigger":
		if len(inv.Args) < 2 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		event := strings.TrimSpace(strings.Join(inv.Args[1:], " "))
		if event == "" {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.ProactiveLastAction = "trigger"
		status := "ignored"
		if cmdCtx.State.ProactiveEnabled {
			status = "queued"
		}
		message := fmt.Sprintf("PROACTIVE_TRIGGER\nevent=%s\nenabled=%t\nstatus=%s", normalizeToken(event), cmdCtx.State.ProactiveEnabled, status)
		return resultWithIntents(message, proactiveMutationIntents("Proactive trigger", detailRow("Event", normalizeToken(event), "event", "Trigger event."), detailRow("Enabled", boolState(cmdCtx.State.ProactiveEnabled, "yes", "no"), "state", "Whether proactive mode is enabled."), detailRow("Status", normalizeToken(status), normalizeToken(status), "Trigger handling result."))...), nil
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
}

// AssistantCommand provides provider-agnostic assistant runtime controls.
type AssistantCommand struct{}

func NewAssistantCommand() *AssistantCommand { return &AssistantCommand{} }
func (c *AssistantCommand) Name() string     { return "assistant" }
func (c *AssistantCommand) Aliases() []string {
	return []string{"assist"}
}
func (c *AssistantCommand) Description() string { return "Manage assistant mode and session" }
func (c *AssistantCommand) Usage() string {
	return "/assistant [status|mode <chat|plan|review>|session [id]|reset]"
}

func (c *AssistantCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		mode := normalizeToken(cmdCtx.State.AssistantMode)
		if mode == "-" {
			mode = "chat"
		}
		message := fmt.Sprintf("ASSISTANT_STATUS\nmode=%s\nsession_id=%s\nlast_action=%s", mode, normalizeToken(cmdCtx.State.AssistantSessionID), normalizeToken(cmdCtx.State.AssistantLastAction))
		return resultWithIntents(message, assistantStatusIntents(mode, cmdCtx.State.AssistantSessionID, cmdCtx.State.AssistantLastAction)...), nil
	}
	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "mode":
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		mode := strings.ToLower(strings.TrimSpace(inv.Args[1]))
		switch mode {
		case "chat", "plan", "review":
			cmdCtx.State.AssistantMode = mode
			cmdCtx.State.AssistantLastAction = "mode"
			message := fmt.Sprintf("ASSISTANT_MODE\nmode=%s", mode)
			return resultWithIntents(message, assistantMutationIntents("Assistant mode", detailRow("Mode", normalizeToken(mode), "mode", "Assistant mode."))...), nil
		default:
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
	case "session":
		if len(inv.Args) > 2 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		if len(inv.Args) == 1 {
			if strings.TrimSpace(cmdCtx.State.AssistantSessionID) == "" {
				cmdCtx.State.AssistantSessionID = "assistant-1"
			}
			cmdCtx.State.AssistantLastAction = "session"
			message := fmt.Sprintf("ASSISTANT_SESSION\nsession_id=%s\ncreated=false", normalizeToken(cmdCtx.State.AssistantSessionID))
			return resultWithIntents(message, assistantMutationIntents("Assistant session", detailRow("Session", normalizeToken(cmdCtx.State.AssistantSessionID), "session", "Current assistant session id."), detailRow("Created", "no", "state", "Whether a new session was created."))...), nil
		}
		sessionID := strings.TrimSpace(inv.Args[1])
		if sessionID == "" {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		created := strings.TrimSpace(cmdCtx.State.AssistantSessionID) == ""
		cmdCtx.State.AssistantSessionID = sessionID
		cmdCtx.State.AssistantLastAction = "session"
		message := fmt.Sprintf("ASSISTANT_SESSION\nsession_id=%s\ncreated=%t", normalizeToken(sessionID), created)
		return resultWithIntents(message, assistantMutationIntents("Assistant session", detailRow("Session", normalizeToken(sessionID), "session", "Requested session id."), detailRow("Created", boolState(created, "yes", "no"), "state", "Whether session was created."))...), nil
	case "reset":
		if len(inv.Args) != 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.AssistantMode = ""
		cmdCtx.State.AssistantSessionID = ""
		cmdCtx.State.AssistantLastAction = "reset"
		return resultWithIntents("ASSISTANT_RESET\nmode=chat\nsession_id=-", assistantMutationIntents("Assistant reset", detailRow("Mode", "chat", "mode", "Reset to chat mode."), detailRow("Session", "-", "session", "Session cleared."))...), nil
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
}

// ShareCommand provides provider-agnostic share link operations.
type ShareCommand struct{}

func NewShareCommand() *ShareCommand        { return &ShareCommand{} }
func (c *ShareCommand) Name() string        { return "share" }
func (c *ShareCommand) Aliases() []string   { return []string{"publish"} }
func (c *ShareCommand) Description() string { return "Manage deterministic share links" }
func (c *ShareCommand) Usage() string {
	return "/share [status|list|create [scope] [private|public]|revoke <id>]"
}

func (c *ShareCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		active := 0
		for _, link := range cmdCtx.State.ShareLinks {
			if !link.Revoked {
				active++
			}
		}
		message := fmt.Sprintf("SHARE_STATUS\ncount=%d\nactive=%d\nrevoked=%d\nlast_action=%s", len(cmdCtx.State.ShareLinks), active, len(cmdCtx.State.ShareLinks)-active, normalizeToken(cmdCtx.State.ShareLastAction))
		return resultWithIntents(message, shareStatusIntents(cmdCtx.State.ShareLinks, cmdCtx.State.ShareLastAction)...), nil
	}
	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "list":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		links := append([]ShareLink(nil), cmdCtx.State.ShareLinks...)
		sort.Slice(links, func(i, j int) bool { return links[i].ID < links[j].ID })
		lines := []string{"SHARE_LIST", fmt.Sprintf("count=%d", len(links))}
		for i, link := range links {
			idx := i + 1
			lines = append(lines, fmt.Sprintf("share.%d.id=%s", idx, normalizeToken(link.ID)))
			lines = append(lines, fmt.Sprintf("share.%d.scope=%s", idx, normalizeToken(link.Scope)))
			lines = append(lines, fmt.Sprintf("share.%d.visibility=%s", idx, normalizeToken(link.Visibility)))
			lines = append(lines, fmt.Sprintf("share.%d.revoked=%t", idx, link.Revoked))
			lines = append(lines, fmt.Sprintf("share.%d.url=%s", idx, normalizeToken(link.URL)))
		}
		return resultWithIntents(strings.Join(lines, "\n"), shareListIntents(links)...), nil
	case "create":
		if len(inv.Args) > 3 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		scope := "session"
		visibility := "private"
		if len(inv.Args) >= 2 {
			scope = strings.TrimSpace(inv.Args[1])
			if scope == "" {
				scope = "session"
			}
		}
		if len(inv.Args) == 3 {
			visibility = strings.ToLower(strings.TrimSpace(inv.Args[2]))
			switch visibility {
			case "private", "public":
			default:
				return Result{}, fmt.Errorf("usage: %s", c.Usage())
			}
		}
		id := fmt.Sprintf("share-%d", len(cmdCtx.State.ShareLinks)+1)
		url := fmt.Sprintf("https://share.example.invalid/%s", id)
		cmdCtx.State.ShareLinks = append(cmdCtx.State.ShareLinks, ShareLink{ID: id, Scope: scope, Visibility: visibility, URL: url})
		cmdCtx.State.ShareLastAction = "create"
		message := fmt.Sprintf("SHARE_CREATE\nid=%s\nscope=%s\nvisibility=%s\nrevoked=false\nurl=%s", normalizeToken(id), normalizeToken(scope), normalizeToken(visibility), normalizeToken(url))
		return resultWithIntents(message, shareMutationIntents("Share created", detailRow("ID", normalizeToken(id), "share", "Share link id."), detailRow("Scope", normalizeToken(scope), "scope", "Share scope."), detailRow("Visibility", normalizeToken(visibility), "visibility", "Link visibility."), detailRow("URL", normalizeToken(url), "url", "Share URL."))...), nil
	case "revoke":
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		id := strings.TrimSpace(inv.Args[1])
		if id == "" {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		idx := findShareLinkIndex(cmdCtx.State.ShareLinks, id)
		if idx < 0 {
			return Result{}, fmt.Errorf("share link not found: %s", id)
		}
		cmdCtx.State.ShareLinks[idx].Revoked = true
		cmdCtx.State.ShareLastAction = "revoke"
		message := fmt.Sprintf("SHARE_REVOKE\nid=%s\nrevoked=true\nurl=%s", normalizeToken(id), normalizeToken(cmdCtx.State.ShareLinks[idx].URL))
		return resultWithIntents(message, shareMutationIntents("Share revoked", detailRow("ID", normalizeToken(id), "share", "Share link id."), detailRow("Revoked", "yes", "revoked", "Revocation state."), detailRow("URL", normalizeToken(cmdCtx.State.ShareLinks[idx].URL), "url", "Share URL."))...), nil
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
}

// OAuthRefreshCommand tracks provider-agnostic OAuth refresh lifecycle.
type OAuthRefreshCommand struct{}

func NewOAuthRefreshCommand() *OAuthRefreshCommand { return &OAuthRefreshCommand{} }
func (c *OAuthRefreshCommand) Name() string        { return "oauth-refresh" }
func (c *OAuthRefreshCommand) Aliases() []string   { return []string{"oauth"} }
func (c *OAuthRefreshCommand) Description() string { return "Refresh or inspect OAuth provider tokens" }
func (c *OAuthRefreshCommand) Usage() string {
	return "/oauth-refresh [status|refresh <provider>|error <provider> <reason>|clear <provider>]"
}

func (c *OAuthRefreshCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if cmdCtx.State.OAuthRefreshStates == nil {
		cmdCtx.State.OAuthRefreshStates = make(map[string]OAuthRefreshState)
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		providers := make([]string, 0, len(cmdCtx.State.OAuthRefreshStates))
		for provider := range cmdCtx.State.OAuthRefreshStates {
			providers = append(providers, provider)
		}
		sort.Strings(providers)
		lines := []string{"OAUTH_REFRESH_STATUS", fmt.Sprintf("count=%d", len(providers)), fmt.Sprintf("last_action=%s", normalizeToken(cmdCtx.State.OAuthRefreshLastAction))}
		for i, provider := range providers {
			idx := i + 1
			state := cmdCtx.State.OAuthRefreshStates[provider]
			lines = append(lines, fmt.Sprintf("provider.%d.name=%s", idx, normalizeToken(provider)))
			lines = append(lines, fmt.Sprintf("provider.%d.refreshed=%t", idx, state.Refreshed))
			lines = append(lines, fmt.Sprintf("provider.%d.expires_in=%d", idx, state.ExpiresIn))
			lines = append(lines, fmt.Sprintf("provider.%d.error=%s", idx, normalizeToken(state.Error)))
		}
		return resultWithIntents(strings.Join(lines, "\n"), oauthRefreshStatusIntents(cmdCtx.State.OAuthRefreshStates, cmdCtx.State.OAuthRefreshLastAction)...), nil
	}
	if len(inv.Args) < 2 {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	provider := strings.ToLower(strings.TrimSpace(inv.Args[1]))
	if provider == "" {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	switch sub {
	case "refresh":
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		tokenPrefix := provider
		if len(tokenPrefix) > 6 {
			tokenPrefix = tokenPrefix[:6]
		}
		state := OAuthRefreshState{Provider: provider, Refreshed: true, TokenPrefix: tokenPrefix + "...", ExpiresIn: 3600, Error: ""}
		cmdCtx.State.OAuthRefreshStates[provider] = state
		cmdCtx.State.OAuthRefreshLastAction = "refresh"
		return resultWithIntents(fmt.Sprintf("OAUTH_REFRESH_RESULT\nprovider=%s\nrefreshed=true\nexpires_in=3600\ntoken_prefix=%s\nerror=-", normalizeToken(provider), normalizeToken(state.TokenPrefix)), oauthRefreshResultIntents(provider, true, 3600, state.TokenPrefix, "")...), nil
	case "error":
		if len(inv.Args) < 3 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		reason := strings.TrimSpace(strings.Join(inv.Args[2:], " "))
		if reason == "" {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		state := OAuthRefreshState{Provider: provider, Refreshed: false, TokenPrefix: "-", ExpiresIn: 0, Error: reason}
		cmdCtx.State.OAuthRefreshStates[provider] = state
		cmdCtx.State.OAuthRefreshLastAction = "error"
		return resultWithIntents(fmt.Sprintf("OAUTH_REFRESH_RESULT\nprovider=%s\nrefreshed=false\nexpires_in=0\ntoken_prefix=-\nerror=%s", normalizeToken(provider), normalizeToken(reason)), oauthRefreshResultIntents(provider, false, 0, "", reason)...), nil
	case "clear":
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		_, existed := cmdCtx.State.OAuthRefreshStates[provider]
		delete(cmdCtx.State.OAuthRefreshStates, provider)
		cmdCtx.State.OAuthRefreshLastAction = "clear"
		return resultWithIntents(fmt.Sprintf("OAUTH_REFRESH_CLEAR\nprovider=%s\nremoved=%t", normalizeToken(provider), existed), oauthRefreshClearIntents(provider, existed)...), nil
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
}

// BridgeCommand tracks deterministic bridge-mode runtime state.
type BridgeCommand struct{}

func NewBridgeCommand() *BridgeCommand       { return &BridgeCommand{} }
func (c *BridgeCommand) Name() string        { return "bridge" }
func (c *BridgeCommand) Aliases() []string   { return nil }
func (c *BridgeCommand) Description() string { return "Toggle deterministic bridge mode" }
func (c *BridgeCommand) Usage() string       { return "/bridge [status|on|off|toggle]" }
func (c *BridgeCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return resultWithIntents(
			fmt.Sprintf("BRIDGE_STATUS\nenabled=%t\ntransitions=%d", cmdCtx.State.BridgeEnabled, cmdCtx.State.BridgeTransitions),
			summaryCardIntent("Bridge mode", "Bridge runtime routing state.", field("enabled", fmt.Sprintf("%t", cmdCtx.State.BridgeEnabled)), field("transitions", fmt.Sprintf("%d", cmdCtx.State.BridgeTransitions))),
			actionHintsIntent("Actions", hint("Enable", "/bridge on"), hint("Disable", "/bridge off"), hint("Toggle", "/bridge toggle")),
		), nil
	}
	if len(inv.Args) != 1 {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "on":
		if !cmdCtx.State.BridgeEnabled {
			cmdCtx.State.BridgeTransitions++
		}
		cmdCtx.State.BridgeEnabled = true
	case "off":
		if cmdCtx.State.BridgeEnabled {
			cmdCtx.State.BridgeTransitions++
		}
		cmdCtx.State.BridgeEnabled = false
	case "toggle":
		cmdCtx.State.BridgeEnabled = !cmdCtx.State.BridgeEnabled
		cmdCtx.State.BridgeTransitions++
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	return resultWithIntents(
		fmt.Sprintf("BRIDGE_SET\nenabled=%t\ntransitions=%d", cmdCtx.State.BridgeEnabled, cmdCtx.State.BridgeTransitions),
		summaryCardIntent("Bridge updated", "Bridge mode mutation applied.", field("enabled", fmt.Sprintf("%t", cmdCtx.State.BridgeEnabled)), field("transitions", fmt.Sprintf("%d", cmdCtx.State.BridgeTransitions))),
	), nil
}

// AntTraceCommand tracks deterministic ant-trace markers.
type AntTraceCommand struct{}

func NewAntTraceCommand() *AntTraceCommand     { return &AntTraceCommand{} }
func (c *AntTraceCommand) Name() string        { return "ant-trace" }
func (c *AntTraceCommand) Aliases() []string   { return nil }
func (c *AntTraceCommand) Description() string { return "Track deterministic ant trace checkpoints" }
func (c *AntTraceCommand) Usage() string       { return "/ant-trace [status|on|off|mark <label>|list|clear]" }
func (c *AntTraceCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return resultWithIntents(fmt.Sprintf("ANT_TRACE_STATUS\nenabled=%t\ncount=%d\nmarks=%d", cmdCtx.State.AntTraceEnabled, cmdCtx.State.AntTraceCount, len(cmdCtx.State.AntTraceMarks)), antTraceStatusIntents(cmdCtx.State.AntTraceEnabled, cmdCtx.State.AntTraceCount, len(cmdCtx.State.AntTraceMarks))...), nil
	}
	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "on", "off":
		if len(inv.Args) != 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.AntTraceEnabled = sub == "on"
		cmdCtx.State.AntTraceCount++
		return resultWithIntents(fmt.Sprintf("ANT_TRACE_SET\nenabled=%t\ncount=%d", cmdCtx.State.AntTraceEnabled, cmdCtx.State.AntTraceCount), antTraceSetIntents(cmdCtx.State.AntTraceEnabled, cmdCtx.State.AntTraceCount)...), nil
	case "mark":
		if len(inv.Args) < 2 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		label := strings.TrimSpace(strings.Join(inv.Args[1:], " "))
		if label == "" {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.AntTraceMarks = append(cmdCtx.State.AntTraceMarks, label)
		cmdCtx.State.AntTraceCount++
		return resultWithIntents(fmt.Sprintf("ANT_TRACE_MARK\nlabel=%s\nmarks=%d\ncount=%d", normalizeToken(label), len(cmdCtx.State.AntTraceMarks), cmdCtx.State.AntTraceCount), antTraceMarkIntents(label, len(cmdCtx.State.AntTraceMarks), cmdCtx.State.AntTraceCount)...), nil
	case "list":
		if len(inv.Args) != 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		lines := []string{"ANT_TRACE_LIST", fmt.Sprintf("count=%d", len(cmdCtx.State.AntTraceMarks))}
		for i, mark := range cmdCtx.State.AntTraceMarks {
			lines = append(lines, fmt.Sprintf("mark.%d=%s", i+1, normalizeToken(mark)))
		}
		return resultWithIntents(strings.Join(lines, "\n"), antTraceListIntents(cmdCtx.State.AntTraceMarks)...), nil
	case "clear":
		if len(inv.Args) != 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.AntTraceMarks = nil
		cmdCtx.State.AntTraceCount++
		return resultWithIntents(fmt.Sprintf("ANT_TRACE_CLEAR\ncount=%d", cmdCtx.State.AntTraceCount), antTraceClearIntents(cmdCtx.State.AntTraceCount)...), nil
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
}

// AutofixPRCommand tracks deterministic autofix PR workflow state.
type AutofixPRCommand struct{}

func NewAutofixPRCommand() *AutofixPRCommand    { return &AutofixPRCommand{} }
func (c *AutofixPRCommand) Name() string        { return "autofix-pr" }
func (c *AutofixPRCommand) Aliases() []string   { return nil }
func (c *AutofixPRCommand) Description() string { return "Queue deterministic autofix PR actions" }
func (c *AutofixPRCommand) Usage() string       { return "/autofix-pr [status|run [ref]|plan [ref]|cancel]" }
func (c *AutofixPRCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return resultWithIntents(fmt.Sprintf("AUTOFIX_PR_STATUS\ncount=%d\nlast_action=%s\nlast_ref=%s", cmdCtx.State.AutofixPRCount, normalizeToken(cmdCtx.State.AutofixPRLastAction), normalizeToken(cmdCtx.State.AutofixPRLastRef)), autofixPRStatusIntents(cmdCtx.State.AutofixPRCount, cmdCtx.State.AutofixPRLastAction, cmdCtx.State.AutofixPRLastRef)...), nil
	}
	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "run", "plan":
		if len(inv.Args) > 2 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		ref := "current"
		if len(inv.Args) == 2 {
			ref = strings.TrimSpace(inv.Args[1])
			if ref == "" {
				return Result{}, fmt.Errorf("usage: %s", c.Usage())
			}
		}
		cmdCtx.State.AutofixPRCount++
		cmdCtx.State.AutofixPRLastAction = sub
		cmdCtx.State.AutofixPRLastRef = ref
		return resultWithIntents(fmt.Sprintf("AUTOFIX_PR_%s\nref=%s\ncount=%d", strings.ToUpper(sub), normalizeToken(ref), cmdCtx.State.AutofixPRCount), autofixPRActionIntents(sub, ref, cmdCtx.State.AutofixPRCount)...), nil
	case "cancel":
		if len(inv.Args) != 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.AutofixPRCount++
		cmdCtx.State.AutofixPRLastAction = "cancel"
		return resultWithIntents(fmt.Sprintf("AUTOFIX_PR_CANCEL\ncount=%d", cmdCtx.State.AutofixPRCount), autofixPRCancelIntents(cmdCtx.State.AutofixPRCount)...), nil
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
}

// BackfillSessionsCommand tracks deterministic session backfill runs.
type BackfillSessionsCommand struct{}

func NewBackfillSessionsCommand() *BackfillSessionsCommand { return &BackfillSessionsCommand{} }
func (c *BackfillSessionsCommand) Name() string            { return "backfill-sessions" }
func (c *BackfillSessionsCommand) Aliases() []string       { return nil }
func (c *BackfillSessionsCommand) Description() string {
	return "Backfill deterministic session metadata"
}
func (c *BackfillSessionsCommand) Usage() string {
	return "/backfill-sessions [status|run [count]|dry-run [count]]"
}
func (c *BackfillSessionsCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return resultWithIntents(fmt.Sprintf("BACKFILL_STATUS\nruns=%d\nlast_count=%d\nlast_action=%s", cmdCtx.State.BackfillRuns, cmdCtx.State.BackfillLastCount, normalizeToken(cmdCtx.State.BackfillLastAction)), backfillStatusIntents(cmdCtx.State.BackfillRuns, cmdCtx.State.BackfillLastCount, cmdCtx.State.BackfillLastAction)...), nil
	}
	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	if sub != "run" && sub != "dry-run" {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	if len(inv.Args) > 2 {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	count := 10
	if len(inv.Args) == 2 {
		parsed, err := parseNonNegativeInt(inv.Args[1])
		if err != nil {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		count = parsed
	}
	cmdCtx.State.BackfillRuns++
	cmdCtx.State.BackfillLastCount = count
	cmdCtx.State.BackfillLastAction = sub
	return resultWithIntents(fmt.Sprintf("BACKFILL_RUN\naction=%s\ncount=%d\nruns=%d", normalizeToken(sub), count, cmdCtx.State.BackfillRuns), backfillRunIntents(sub, count, cmdCtx.State.BackfillRuns)...), nil
}

// BreakCacheCommand tracks deterministic cache invalidation scopes.
type BreakCacheCommand struct{}

func NewBreakCacheCommand() *BreakCacheCommand { return &BreakCacheCommand{} }
func (c *BreakCacheCommand) Name() string      { return "break-cache" }
func (c *BreakCacheCommand) Aliases() []string { return nil }
func (c *BreakCacheCommand) Description() string {
	return "Invalidate deterministic local cache scopes"
}
func (c *BreakCacheCommand) Usage() string { return "/break-cache [status|all|models|history|tools]" }
func (c *BreakCacheCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if cmdCtx.State.CacheBreakScopes == nil {
		cmdCtx.State.CacheBreakScopes = map[string]int{}
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		lines := []string{"BREAK_CACHE_STATUS", fmt.Sprintf("count=%d", cmdCtx.State.CacheBreakCount), fmt.Sprintf("last_scope=%s", normalizeToken(cmdCtx.State.CacheBreakLastScope))}
		scopes := []string{"all", "models", "history", "tools"}
		for _, scope := range scopes {
			lines = append(lines, fmt.Sprintf("scope.%s=%d", scope, cmdCtx.State.CacheBreakScopes[scope]))
		}
		return resultWithIntents(strings.Join(lines, "\n"), breakCacheStatusIntents(cmdCtx.State.CacheBreakCount, cmdCtx.State.CacheBreakLastScope, cmdCtx.State.CacheBreakScopes)...), nil
	}
	if len(inv.Args) != 1 {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	scope := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch scope {
	case "all", "models", "history", "tools":
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	cmdCtx.State.CacheBreakCount++
	cmdCtx.State.CacheBreakLastScope = scope
	cmdCtx.State.CacheBreakScopes[scope]++
	return resultWithIntents(fmt.Sprintf("BREAK_CACHE_APPLY\nscope=%s\ncount=%d", scope, cmdCtx.State.CacheBreakCount), breakCacheApplyIntents(scope, cmdCtx.State.CacheBreakCount)...), nil
}

// BughunterCommand tracks deterministic bug reproduction runs.
type BughunterCommand struct{}

func NewBughunterCommand() *BughunterCommand    { return &BughunterCommand{} }
func (c *BughunterCommand) Name() string        { return "bughunter" }
func (c *BughunterCommand) Aliases() []string   { return nil }
func (c *BughunterCommand) Description() string { return "Run deterministic bughunter diagnostics" }
func (c *BughunterCommand) Usage() string       { return "/bughunter [status|run [scope]]" }
func (c *BughunterCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return resultWithIntents(fmt.Sprintf("BUGHUNTER_STATUS\ncount=%d\nlast_scope=%s", cmdCtx.State.BughunterCount, normalizeToken(cmdCtx.State.BughunterLastScope)), bughunterStatusIntents(cmdCtx.State.BughunterCount, cmdCtx.State.BughunterLastScope)...), nil
	}
	if strings.ToLower(strings.TrimSpace(inv.Args[0])) != "run" || len(inv.Args) > 2 {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	scope := "workspace"
	if len(inv.Args) == 2 {
		scope = strings.TrimSpace(inv.Args[1])
		if scope == "" {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
	}
	cmdCtx.State.BughunterCount++
	cmdCtx.State.BughunterLastScope = scope
	return resultWithIntents(fmt.Sprintf("BUGHUNTER_RUN\nscope=%s\ncount=%d", normalizeToken(scope), cmdCtx.State.BughunterCount), bughunterRunIntents(scope, cmdCtx.State.BughunterCount)...), nil
}

// CtxVizCommand tracks deterministic context visualization refreshes.
type CtxVizCommand struct{}

func NewCtxVizCommand() *CtxVizCommand     { return &CtxVizCommand{} }
func (c *CtxVizCommand) Name() string      { return "ctx-viz" }
func (c *CtxVizCommand) Aliases() []string { return nil }
func (c *CtxVizCommand) Description() string {
	return "Render deterministic context visualization state"
}
func (c *CtxVizCommand) Usage() string { return "/ctx-viz [status|render [focus]|clear]" }
func (c *CtxVizCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return resultWithIntents(fmt.Sprintf("CTX_VIZ_STATUS\ncount=%d\nlast_action=%s", cmdCtx.State.ContextVizCount, normalizeToken(cmdCtx.State.ContextVizLastAction)), ctxVizStatusIntents(cmdCtx.State.ContextVizCount, cmdCtx.State.ContextVizLastAction)...), nil
	}
	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "render":
		if len(inv.Args) > 2 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		focus := "workspace"
		if len(inv.Args) == 2 {
			focus = strings.TrimSpace(inv.Args[1])
			if focus == "" {
				return Result{}, fmt.Errorf("usage: %s", c.Usage())
			}
		}
		cmdCtx.State.ContextVizCount++
		cmdCtx.State.ContextVizLastAction = "render"
		return resultWithIntents(fmt.Sprintf("CTX_VIZ_RENDER\nfocus=%s\ncount=%d", normalizeToken(focus), cmdCtx.State.ContextVizCount), ctxVizRenderIntents(focus, cmdCtx.State.ContextVizCount)...), nil
	case "clear":
		if len(inv.Args) != 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.ContextVizCount++
		cmdCtx.State.ContextVizLastAction = "clear"
		return resultWithIntents(fmt.Sprintf("CTX_VIZ_CLEAR\ncount=%d", cmdCtx.State.ContextVizCount), ctxVizClearIntents(cmdCtx.State.ContextVizCount)...), nil
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
}

// DebugToolCallCommand tracks deterministic tool-call diagnostics.
type DebugToolCallCommand struct{}

func NewDebugToolCallCommand() *DebugToolCallCommand { return &DebugToolCallCommand{} }
func (c *DebugToolCallCommand) Name() string         { return "debug-tool-call" }
func (c *DebugToolCallCommand) Aliases() []string    { return nil }
func (c *DebugToolCallCommand) Description() string {
	return "Inspect deterministic tool-call debug state"
}
func (c *DebugToolCallCommand) Usage() string { return "/debug-tool-call [status|log <tool>|clear]" }
func (c *DebugToolCallCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return resultWithIntents(fmt.Sprintf("DEBUG_TOOL_CALL_STATUS\ncount=%d\nlast_tool=%s", cmdCtx.State.DebugToolCallCount, normalizeToken(cmdCtx.State.DebugToolCallLastTool)), debugToolCallStatusIntents(cmdCtx.State.DebugToolCallCount, cmdCtx.State.DebugToolCallLastTool)...), nil
	}
	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "log":
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		tool := strings.TrimSpace(inv.Args[1])
		if tool == "" {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.DebugToolCallCount++
		cmdCtx.State.DebugToolCallLastTool = tool
		return resultWithIntents(fmt.Sprintf("DEBUG_TOOL_CALL_LOG\ntool=%s\ncount=%d", normalizeToken(tool), cmdCtx.State.DebugToolCallCount), debugToolCallLogIntents(tool, cmdCtx.State.DebugToolCallCount)...), nil
	case "clear":
		if len(inv.Args) != 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.DebugToolCallCount++
		cmdCtx.State.DebugToolCallLastTool = ""
		return resultWithIntents(fmt.Sprintf("DEBUG_TOOL_CALL_CLEAR\ncount=%d", cmdCtx.State.DebugToolCallCount), debugToolCallClearIntents(cmdCtx.State.DebugToolCallCount)...), nil
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
}

// GoodClaudeCommand toggles deterministic "good claude" mode.
type GoodClaudeCommand struct{}

func NewGoodClaudeCommand() *GoodClaudeCommand { return &GoodClaudeCommand{} }
func (c *GoodClaudeCommand) Name() string      { return "good-claude" }
func (c *GoodClaudeCommand) Aliases() []string { return nil }
func (c *GoodClaudeCommand) Description() string {
	return "Toggle deterministic good-claude helper mode"
}
func (c *GoodClaudeCommand) Usage() string { return "/good-claude [status|on|off|ask <question>]" }
func (c *GoodClaudeCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return resultWithIntents(fmt.Sprintf("GOOD_CLAUDE_STATUS\nenabled=%t\nask_count=%d", cmdCtx.State.GoodClaudeEnabled, cmdCtx.State.GoodClaudeAskCount), legacyOutputIntents(fmt.Sprintf("GOOD_CLAUDE_STATUS\nenabled=%t\nask_count=%d", cmdCtx.State.GoodClaudeEnabled, cmdCtx.State.GoodClaudeAskCount))...), nil
	}
	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "on", "off":
		if len(inv.Args) != 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.GoodClaudeEnabled = sub == "on"
		return resultWithIntents(fmt.Sprintf("GOOD_CLAUDE_SET\nenabled=%t", cmdCtx.State.GoodClaudeEnabled), legacyOutputIntents(fmt.Sprintf("GOOD_CLAUDE_SET\nenabled=%t", cmdCtx.State.GoodClaudeEnabled))...), nil
	case "ask":
		if len(inv.Args) < 2 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		question := strings.TrimSpace(strings.Join(inv.Args[1:], " "))
		if question == "" {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.GoodClaudeAskCount++
		return resultWithIntents(fmt.Sprintf("GOOD_CLAUDE_ASK\nquestion=%s\ncount=%d", normalizeToken(question), cmdCtx.State.GoodClaudeAskCount), legacyOutputIntents(fmt.Sprintf("GOOD_CLAUDE_ASK\nquestion=%s\ncount=%d", normalizeToken(question), cmdCtx.State.GoodClaudeAskCount))...), nil
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
}

// HeapdumpCommand records deterministic heapdump captures.
type HeapdumpCommand struct{}

func NewHeapdumpCommand() *HeapdumpCommand     { return &HeapdumpCommand{} }
func (c *HeapdumpCommand) Name() string        { return "heapdump" }
func (c *HeapdumpCommand) Aliases() []string   { return nil }
func (c *HeapdumpCommand) Description() string { return "Capture deterministic heapdump markers" }
func (c *HeapdumpCommand) Usage() string       { return "/heapdump [status|capture [reason]|clear]" }
func (c *HeapdumpCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return resultWithIntents(fmt.Sprintf("HEAPDUMP_STATUS\ncount=%d\nlast_reason=%s", cmdCtx.State.HeapdumpCount, normalizeToken(cmdCtx.State.HeapdumpLastReason)), legacyOutputIntents(fmt.Sprintf("HEAPDUMP_STATUS\ncount=%d\nlast_reason=%s", cmdCtx.State.HeapdumpCount, normalizeToken(cmdCtx.State.HeapdumpLastReason)))...), nil
	}
	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "capture":
		if len(inv.Args) > 2 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		reason := "manual"
		if len(inv.Args) == 2 {
			reason = strings.TrimSpace(inv.Args[1])
			if reason == "" {
				return Result{}, fmt.Errorf("usage: %s", c.Usage())
			}
		}
		cmdCtx.State.HeapdumpCount++
		cmdCtx.State.HeapdumpLastReason = reason
		return resultWithIntents(fmt.Sprintf("HEAPDUMP_CAPTURE\nreason=%s\ncount=%d", normalizeToken(reason), cmdCtx.State.HeapdumpCount), legacyOutputIntents(fmt.Sprintf("HEAPDUMP_CAPTURE\nreason=%s\ncount=%d", normalizeToken(reason), cmdCtx.State.HeapdumpCount))...), nil
	case "clear":
		if len(inv.Args) != 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.HeapdumpLastReason = ""
		return resultWithIntents("HEAPDUMP_CLEAR\nlast_reason=-", legacyOutputIntents("HEAPDUMP_CLEAR\nlast_reason=-")...), nil
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
}

// InstallCommand tracks deterministic install flow state.
type InstallCommand struct{}

func NewInstallCommand() *InstallCommand      { return &InstallCommand{} }
func (c *InstallCommand) Name() string        { return "install" }
func (c *InstallCommand) Aliases() []string   { return nil }
func (c *InstallCommand) Description() string { return "Run deterministic install/setup steps" }
func (c *InstallCommand) Usage() string       { return "/install [status|plugins|setup|doctor]" }
func (c *InstallCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return resultWithIntents(fmt.Sprintf("INSTALL_STATUS\ncount=%d\nlast_action=%s", cmdCtx.State.InstallCount, normalizeToken(cmdCtx.State.InstallLastAction)), legacyOutputIntents(fmt.Sprintf("INSTALL_STATUS\ncount=%d\nlast_action=%s", cmdCtx.State.InstallCount, normalizeToken(cmdCtx.State.InstallLastAction)))...), nil
	}
	if len(inv.Args) != 1 {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "plugins", "setup", "doctor":
		cmdCtx.State.InstallCount++
		cmdCtx.State.InstallLastAction = sub
		return resultWithIntents(fmt.Sprintf("INSTALL_RUN\naction=%s\ncount=%d", normalizeToken(sub), cmdCtx.State.InstallCount), legacyOutputIntents(fmt.Sprintf("INSTALL_RUN\naction=%s\ncount=%d", normalizeToken(sub), cmdCtx.State.InstallCount))...), nil
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
}

// MockLimitsCommand tracks deterministic rate-limit mocking state.
type MockLimitsCommand struct{}

func NewMockLimitsCommand() *MockLimitsCommand   { return &MockLimitsCommand{} }
func (c *MockLimitsCommand) Name() string        { return "mock-limits" }
func (c *MockLimitsCommand) Aliases() []string   { return nil }
func (c *MockLimitsCommand) Description() string { return "Toggle deterministic mock limit behavior" }
func (c *MockLimitsCommand) Usage() string       { return "/mock-limits [status|on [count]|off]" }
func (c *MockLimitsCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return resultWithIntents(fmt.Sprintf("MOCK_LIMITS_STATUS\nenabled=%t\nvalue=%d\nchanges=%d", cmdCtx.State.MockLimitsEnabled, cmdCtx.State.MockLimitsValue, cmdCtx.State.MockLimitsChanges), legacyOutputIntents(fmt.Sprintf("MOCK_LIMITS_STATUS\nenabled=%t\nvalue=%d\nchanges=%d", cmdCtx.State.MockLimitsEnabled, cmdCtx.State.MockLimitsValue, cmdCtx.State.MockLimitsChanges))...), nil
	}
	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "on":
		if len(inv.Args) > 2 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		value := cmdCtx.State.MockLimitsValue
		if value == 0 {
			value = 1
		}
		if len(inv.Args) == 2 {
			parsed, err := parseNonNegativeInt(inv.Args[1])
			if err != nil {
				return Result{}, fmt.Errorf("usage: %s", c.Usage())
			}
			value = parsed
		}
		cmdCtx.State.MockLimitsEnabled = true
		cmdCtx.State.MockLimitsValue = value
		cmdCtx.State.MockLimitsChanges++
		return resultWithIntents(fmt.Sprintf("MOCK_LIMITS_SET\nenabled=true\nvalue=%d\nchanges=%d", value, cmdCtx.State.MockLimitsChanges), legacyOutputIntents(fmt.Sprintf("MOCK_LIMITS_SET\nenabled=true\nvalue=%d\nchanges=%d", value, cmdCtx.State.MockLimitsChanges))...), nil
	case "off":
		if len(inv.Args) != 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.MockLimitsEnabled = false
		cmdCtx.State.MockLimitsChanges++
		return resultWithIntents(fmt.Sprintf("MOCK_LIMITS_SET\nenabled=false\nvalue=%d\nchanges=%d", cmdCtx.State.MockLimitsValue, cmdCtx.State.MockLimitsChanges), legacyOutputIntents(fmt.Sprintf("MOCK_LIMITS_SET\nenabled=false\nvalue=%d\nchanges=%d", cmdCtx.State.MockLimitsValue, cmdCtx.State.MockLimitsChanges))...), nil
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
}

// OnboardingCommand tracks deterministic onboarding status for parity.
type OnboardingCommand struct{}

func NewOnboardingCommand() *OnboardingCommand   { return &OnboardingCommand{} }
func (c *OnboardingCommand) Name() string        { return "onboarding" }
func (c *OnboardingCommand) Aliases() []string   { return nil }
func (c *OnboardingCommand) Description() string { return "Inspect deterministic onboarding state" }
func (c *OnboardingCommand) Usage() string       { return "/onboarding [status|run|reset]" }
func (c *OnboardingCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return resultWithIntents(fmt.Sprintf("ONBOARDING_STATUS\ncompleted=%t\nruns=%d", cmdCtx.State.OnboardingCompleted, cmdCtx.State.OnboardingRuns), legacyOutputIntents(fmt.Sprintf("ONBOARDING_STATUS\ncompleted=%t\nruns=%d", cmdCtx.State.OnboardingCompleted, cmdCtx.State.OnboardingRuns))...), nil
	}
	if len(inv.Args) != 1 {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "run":
		cmdCtx.State.OnboardingRuns++
		cmdCtx.State.OnboardingCompleted = true
		return resultWithIntents(fmt.Sprintf("ONBOARDING_RUN\ncompleted=true\nruns=%d", cmdCtx.State.OnboardingRuns), legacyOutputIntents(fmt.Sprintf("ONBOARDING_RUN\ncompleted=true\nruns=%d", cmdCtx.State.OnboardingRuns))...), nil
	case "reset":
		cmdCtx.State.OnboardingCompleted = false
		cmdCtx.State.OnboardingRuns++
		return resultWithIntents(fmt.Sprintf("ONBOARDING_RESET\ncompleted=false\nruns=%d", cmdCtx.State.OnboardingRuns), legacyOutputIntents(fmt.Sprintf("ONBOARDING_RESET\ncompleted=false\nruns=%d", cmdCtx.State.OnboardingRuns))...), nil
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
}

// PerfIssueCommand tracks deterministic performance issue filing state.
type PerfIssueCommand struct{}

func NewPerfIssueCommand() *PerfIssueCommand    { return &PerfIssueCommand{} }
func (c *PerfIssueCommand) Name() string        { return "perf-issue" }
func (c *PerfIssueCommand) Aliases() []string   { return nil }
func (c *PerfIssueCommand) Description() string { return "File deterministic performance issue notes" }
func (c *PerfIssueCommand) Usage() string       { return "/perf-issue [status|open <title>]" }
func (c *PerfIssueCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return resultWithIntents(fmt.Sprintf("PERF_ISSUE_STATUS\ncount=%d\nlast_title=%s", cmdCtx.State.PerfIssueCount, normalizeToken(cmdCtx.State.PerfIssueLastTitle)), legacyOutputIntents(fmt.Sprintf("PERF_ISSUE_STATUS\ncount=%d\nlast_title=%s", cmdCtx.State.PerfIssueCount, normalizeToken(cmdCtx.State.PerfIssueLastTitle)))...), nil
	}
	if len(inv.Args) < 2 || !strings.EqualFold(strings.TrimSpace(inv.Args[0]), "open") {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	title := strings.TrimSpace(strings.Join(inv.Args[1:], " "))
	if title == "" {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	cmdCtx.State.PerfIssueCount++
	cmdCtx.State.PerfIssueLastTitle = title
	return resultWithIntents(fmt.Sprintf("PERF_ISSUE_OPEN\ntitle=%s\ncount=%d", normalizeToken(title), cmdCtx.State.PerfIssueCount), legacyOutputIntents(fmt.Sprintf("PERF_ISSUE_OPEN\ntitle=%s\ncount=%d", normalizeToken(title), cmdCtx.State.PerfIssueCount))...), nil
}

// SandboxToggleCommand toggles deterministic sandbox mode quickly.
type SandboxToggleCommand struct{}

func NewSandboxToggleCommand() *SandboxToggleCommand { return &SandboxToggleCommand{} }
func (c *SandboxToggleCommand) Name() string         { return "sandbox-toggle" }
func (c *SandboxToggleCommand) Aliases() []string    { return nil }
func (c *SandboxToggleCommand) Description() string  { return "Toggle deterministic sandbox mode" }
func (c *SandboxToggleCommand) Usage() string        { return "/sandbox-toggle [status|on|off|toggle]" }
func (c *SandboxToggleCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if strings.TrimSpace(cmdCtx.State.SandboxMode) == "" {
		cmdCtx.State.SandboxMode = "off"
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return resultWithIntents(fmt.Sprintf("SANDBOX_TOGGLE_STATUS\nmode=%s", normalizeToken(cmdCtx.State.SandboxMode)), legacyOutputIntents(fmt.Sprintf("SANDBOX_TOGGLE_STATUS\nmode=%s", normalizeToken(cmdCtx.State.SandboxMode)))...), nil
	}
	if len(inv.Args) != 1 {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "on":
		cmdCtx.State.SandboxMode = "workspace-write"
	case "off":
		cmdCtx.State.SandboxMode = "off"
	case "toggle":
		if cmdCtx.State.SandboxMode == "off" {
			cmdCtx.State.SandboxMode = "workspace-write"
		} else {
			cmdCtx.State.SandboxMode = "off"
		}
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	return resultWithIntents(fmt.Sprintf("SANDBOX_TOGGLE_SET\nmode=%s", normalizeToken(cmdCtx.State.SandboxMode)), legacyOutputIntents(fmt.Sprintf("SANDBOX_TOGGLE_SET\nmode=%s", normalizeToken(cmdCtx.State.SandboxMode)))...), nil
}

// RemoteSetupCommand mirrors remote setup command family for parity.
type RemoteSetupCommand struct{}

func NewRemoteSetupCommand() *RemoteSetupCommand { return &RemoteSetupCommand{} }
func (c *RemoteSetupCommand) Name() string       { return "remote-setup" }
func (c *RemoteSetupCommand) Aliases() []string  { return nil }
func (c *RemoteSetupCommand) Description() string {
	return "Configure deterministic remote setup connection"
}
func (c *RemoteSetupCommand) Usage() string { return "/remote-setup [status|connect|disconnect]" }
func (c *RemoteSetupCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return resultWithIntents(fmt.Sprintf("REMOTE_SETUP_STATUS\nconnected=%t\ncount=%d\nurl=%s", cmdCtx.State.WebSetupConnected, cmdCtx.State.WebSetupCount, normalizeToken(cmdCtx.State.RemoteSessionURL)), legacyOutputIntents(fmt.Sprintf("REMOTE_SETUP_STATUS\nconnected=%t\ncount=%d\nurl=%s", cmdCtx.State.WebSetupConnected, cmdCtx.State.WebSetupCount, normalizeToken(cmdCtx.State.RemoteSessionURL)))...), nil
	}
	if len(inv.Args) != 1 {
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "connect":
		cmdCtx.State.WebSetupConnected = true
		cmdCtx.State.WebSetupCount++
		if strings.TrimSpace(cmdCtx.State.RemoteSessionURL) == "" {
			cmdCtx.State.RemoteSessionURL = "https://claude.ai/code"
		}
	case "disconnect":
		cmdCtx.State.WebSetupConnected = false
		cmdCtx.State.WebSetupCount++
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
	return resultWithIntents(fmt.Sprintf("REMOTE_SETUP_APPLY\naction=%s\nconnected=%t\ncount=%d", sub, cmdCtx.State.WebSetupConnected, cmdCtx.State.WebSetupCount), legacyOutputIntents(fmt.Sprintf("REMOTE_SETUP_APPLY\naction=%s\nconnected=%t\ncount=%d", sub, cmdCtx.State.WebSetupConnected, cmdCtx.State.WebSetupCount))...), nil
}

// UltraplanCommand tracks deterministic ultraplan runs.
type UltraplanCommand struct{}

func NewUltraplanCommand() *UltraplanCommand    { return &UltraplanCommand{} }
func (c *UltraplanCommand) Name() string        { return "ultraplan" }
func (c *UltraplanCommand) Aliases() []string   { return nil }
func (c *UltraplanCommand) Description() string { return "Run deterministic ultraplan planner" }
func (c *UltraplanCommand) Usage() string       { return "/ultraplan [status|run [target]|clear]" }
func (c *UltraplanCommand) Execute(_ context.Context, cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || strings.EqualFold(strings.TrimSpace(inv.Args[0]), "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		return resultWithIntents(fmt.Sprintf("ULTRAPLAN_STATUS\ncount=%d\nlast_action=%s\nlast_target=%s", cmdCtx.State.UltraplanCount, normalizeToken(cmdCtx.State.UltraplanLastAction), normalizeToken(cmdCtx.State.UltraplanLastTarget)), legacyOutputIntents(fmt.Sprintf("ULTRAPLAN_STATUS\ncount=%d\nlast_action=%s\nlast_target=%s", cmdCtx.State.UltraplanCount, normalizeToken(cmdCtx.State.UltraplanLastAction), normalizeToken(cmdCtx.State.UltraplanLastTarget)))...), nil
	}
	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "run":
		if len(inv.Args) > 2 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		target := "current"
		if len(inv.Args) == 2 {
			target = strings.TrimSpace(inv.Args[1])
			if target == "" {
				return Result{}, fmt.Errorf("usage: %s", c.Usage())
			}
		}
		cmdCtx.State.UltraplanCount++
		cmdCtx.State.UltraplanLastAction = "run"
		cmdCtx.State.UltraplanLastTarget = target
		return resultWithIntents(fmt.Sprintf("ULTRAPLAN_RUN\ntarget=%s\ncount=%d", normalizeToken(target), cmdCtx.State.UltraplanCount), legacyOutputIntents(fmt.Sprintf("ULTRAPLAN_RUN\ntarget=%s\ncount=%d", normalizeToken(target), cmdCtx.State.UltraplanCount))...), nil
	case "clear":
		if len(inv.Args) != 1 {
			return Result{}, fmt.Errorf("usage: %s", c.Usage())
		}
		cmdCtx.State.UltraplanCount++
		cmdCtx.State.UltraplanLastAction = "clear"
		cmdCtx.State.UltraplanLastTarget = ""
		return resultWithIntents(fmt.Sprintf("ULTRAPLAN_CLEAR\ncount=%d", cmdCtx.State.UltraplanCount), legacyOutputIntents(fmt.Sprintf("ULTRAPLAN_CLEAR\ncount=%d", cmdCtx.State.UltraplanCount))...), nil
	default:
		return Result{}, fmt.Errorf("usage: %s", c.Usage())
	}
}

func normalizePluginState(state *RuntimeState) {
	state.PluginsInstalled = uniqueSortedStrings(state.PluginsInstalled)
	state.PluginsEnabled = uniqueSortedStrings(state.PluginsEnabled)
	state.PluginMarketplaces = uniqueSortedStrings(state.PluginMarketplaces)
}

func renderPluginList(state *RuntimeState, serviceResult pluginspkg.ServiceListResult, conflicts []skillspkg.SkillConflictDiagnostic) Result {
	lines := []string{
		"PLUGIN_LIST",
		fmt.Sprintf("installed=%d", len(state.PluginsInstalled)),
		fmt.Sprintf("enabled=%d", len(state.PluginsEnabled)),
		fmt.Sprintf("pending_reload=%t", state.PluginReloadPending),
	}
	if len(serviceResult.Diagnostics) > 0 {
		lines = append(lines, fmt.Sprintf("diagnostic_groups=%d", len(serviceResult.Diagnostics)))
	}
	if len(serviceResult.History) > 0 {
		lines = append(lines, fmt.Sprintf("history=%d", len(serviceResult.History)))
	}
	if len(conflicts) > 0 {
		lines = append(lines, fmt.Sprintf("conflicts=%d", len(conflicts)))
	}
	for i, name := range state.PluginsInstalled {
		idx := i + 1
		lines = append(lines, fmt.Sprintf("plugin.%d.name=%s", idx, normalizeToken(name)))
		lines = append(lines, fmt.Sprintf("plugin.%d.enabled=%t", idx, containsString(state.PluginsEnabled, name)))
		if diags := serviceResult.Diagnostics[name]; len(diags) > 0 {
			lines = append(lines, fmt.Sprintf("plugin.%d.diagnostics=%d", idx, len(diags)))
		}
	}
	for i, conflict := range conflicts {
		idx := i + 1
		lines = append(lines, fmt.Sprintf("conflict.%d.name=%s", idx, normalizeToken(conflict.Name)))
		lines = append(lines, fmt.Sprintf("conflict.%d.winner=%s", idx, normalizeToken(conflict.WinnerSource)))
	}
	return resultWithIntents(strings.Join(lines, "\n"), legacyOutputIntents(strings.Join(lines, "\n"))...)
}

func renderPluginMarketplaces(state *RuntimeState) Result {
	lines := []string{"PLUGIN_MARKETPLACES", fmt.Sprintf("count=%d", len(state.PluginMarketplaces))}
	for i, value := range state.PluginMarketplaces {
		lines = append(lines, fmt.Sprintf("marketplace.%d=%s", i+1, normalizeToken(value)))
	}
	return resultWithIntents(strings.Join(lines, "\n"), legacyOutputIntents(strings.Join(lines, "\n"))...)
}

func removeStringValue(values []string, target string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == strings.TrimSpace(target) {
			continue
		}
		out = append(out, value)
	}
	return uniqueSortedStrings(out)
}

func pluginServiceFromWorkingDir() (*pluginspkg.Service, string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, "", fmt.Errorf("resolve working directory: %w", err)
	}
	root := filepath.Join(workingDir, ".alliecode", "plugins")
	policyPath := filepath.Join(root, "policy.yaml")
	indexPath := filepath.Join(root, "index.yaml")
	return pluginspkg.NewService(root, policyPath, indexPath), workingDir, nil
}

func syncPluginRuntimeState(state *RuntimeState, serviceResult pluginspkg.ServiceListResult) {
	installed := make([]string, 0, len(serviceResult.Plugins))
	enabled := make([]string, 0, len(serviceResult.Plugins))
	for _, plugin := range serviceResult.Plugins {
		id := strings.TrimSpace(plugin.Manifest.ID)
		if id == "" {
			continue
		}
		installed = append(installed, id)
		if plugin.Enabled {
			enabled = append(enabled, id)
		}
	}
	state.PluginsInstalled = uniqueSortedStrings(installed)
	state.PluginsEnabled = uniqueSortedStrings(enabled)
	normalizePluginState(state)
}

func resolvedPluginCommandsForPluginCommand(service *pluginspkg.Service) ([]pluginspkg.ResolvedCommand, error) {
	state, err := service.Store.Load()
	if err != nil {
		return nil, err
	}

	base, err := pluginspkg.LoadRuntime(service.InstallRoot, pluginspkg.Policy{})
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	ids := make([]string, 0, len(base.Plugins()))
	for _, plugin := range base.Plugins() {
		ids = append(ids, plugin.Manifest.ID)
	}
	runtime, err := pluginspkg.LoadRuntime(service.InstallRoot, state.RuntimePolicy(ids))
	if err != nil {
		return nil, err
	}
	commands := runtime.Commands()
	out := make([]pluginspkg.ResolvedCommand, len(commands))
	copy(out, commands)
	return out, nil
}

func pluginSkillConflictsForWorkingDir(_ string, pluginCommands []pluginspkg.ResolvedCommand) []skillspkg.SkillConflictDiagnostic {
	mgr := skillspkg.NewManager(nil)
	if err := mgr.Load(); err != nil {
		return nil
	}
	base := mgr.List()
	merged := skillspkg.MergeSkillsWithPluginCommandsDetailed(base, pluginCommands, skillspkg.PreferSkillFiles)
	out := make([]skillspkg.SkillConflictDiagnostic, len(merged.Conflicts))
	copy(out, merged.Conflicts)
	return out
}

func splitPluginRefName(value string) string {
	value = strings.TrimSpace(value)
	parts := strings.SplitN(value, "@", 2)
	return strings.TrimSpace(parts[0])
}

func isValidEffort(value string) bool {
	switch value {
	case "low", "medium", "high", "max", "auto", "unset":
		return true
	default:
		return false
	}
}

func isSupportedColor(value string) bool {
	switch value {
	case "red", "orange", "yellow", "green", "blue", "cyan", "magenta", "pink", "white", "black", "teal":
		return true
	default:
		return false
	}
}

func parseMode(raw string) (permissions.Mode, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "plan":
		return permissions.ModePlan, true
	case "default":
		return permissions.ModeDefault, true
	case "auto":
		return permissions.ModeAuto, true
	case "bypass":
		return permissions.ModeBypass, true
	default:
		return permissions.ModeDefault, false
	}
}

func modeString(mode permissions.Mode) string {
	switch mode {
	case permissions.ModePlan:
		return "plan"
	case permissions.ModeDefault:
		return "default"
	case permissions.ModeAuto:
		return "auto"
	case permissions.ModeBypass:
		return "bypass"
	default:
		return "default"
	}
}

func parseModelSelection(raw, defaultProvider string) (string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", fmt.Errorf("model cannot be empty")
	}
	if strings.Contains(raw, "/") {
		parts := strings.SplitN(raw, "/", 2)
		providerName := strings.TrimSpace(parts[0])
		modelName := strings.TrimSpace(parts[1])
		if providerName == "" || modelName == "" {
			return "", "", fmt.Errorf("model must be in provider/model format")
		}
		return providerName, modelName, nil
	}
	providerName := strings.TrimSpace(defaultProvider)
	if providerName == "" {
		return "", "", fmt.Errorf("model must include provider as provider/model when no provider is configured")
	}
	return providerName, raw, nil
}

func resolveModelForStatus(rawModel, fallbackProvider string) (string, string, bool) {
	rawModel = strings.TrimSpace(rawModel)
	if rawModel == "" || strings.EqualFold(rawModel, "unknown") {
		return "", "", false
	}
	providerName, modelName, err := parseModelSelection(rawModel, fallbackProvider)
	if err != nil {
		return "", "", false
	}
	return providerName, modelName, true
}

func isSupportedProviderName(name string) bool {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "anthropic", "gemini", "ollama", "openai":
		return true
	default:
		return false
	}
}
