// Package permissions implements the permission engine for AllieCode.
// It checks whether a tool call should be allowed, denied, or needs user confirmation.
package permissions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// Mode determines the overall permission behavior.
type Mode int

const (
	// ModePlan uses default permission behavior while planning.
	ModePlan Mode = iota
	// ModeDefault asks for confirmation on non-read-only tools.
	ModeDefault
	// ModeAuto allows safe tools, asks for destructive ones.
	ModeAuto
	// ModeBypass allows everything without asking.
	ModeBypass
	// ModeAcceptEdits is a compatibility alias with auto-like behavior.
	ModeAcceptEdits
)

// String renders the canonical mode label.
func (m Mode) String() string {
	switch m {
	case ModePlan:
		return "plan"
	case ModeDefault:
		return "default"
	case ModeAuto:
		return "auto"
	case ModeAcceptEdits:
		return "accept_edits"
	case ModeBypass:
		return "bypass"
	default:
		return "default"
	}
}

// ParseMode parses a mode string, including compatibility aliases.
func ParseMode(raw string) (Mode, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "plan":
		return ModePlan, true
	case "default":
		return ModeDefault, true
	case "auto":
		return ModeAuto, true
	case "accept_edits", "acceptedits", "accept-edits":
		return ModeAcceptEdits, true
	case "bypass", "bypasspermissions", "bypass_permissions":
		return ModeBypass, true
	default:
		return ModeDefault, false
	}
}

// Decision is the outcome of a permission check.
type Decision int

const (
	DecisionAllow Decision = iota
	DecisionDeny
	DecisionAsk
)

// RuleSource identifies where a rule originated.
type RuleSource string

const (
	RuleSourcePolicy  RuleSource = "policy"
	RuleSourceUser    RuleSource = "user"
	RuleSourceProject RuleSource = "project"
	RuleSourceSession RuleSource = "session"
)

// Rule defines a permission rule that matches tool calls by name, file path, or bash command.
type Rule struct {
	Source    RuleSource `json:"source,omitempty" yaml:"source,omitempty"`
	Tool      string     `json:"tool" yaml:"tool"`             // matcher DSL (defaults to glob)
	FileGlob  string     `json:"file_glob" yaml:"file_glob"`   // matcher DSL for file paths (defaults to glob)
	BashRegex string     `json:"bash_regex" yaml:"bash_regex"` // matcher DSL for bash commands (defaults to regex)
	Decision  Decision   `json:"decision" yaml:"decision"`
}

// persistentEntry is the serializable form of a persistent decision.
type persistentEntry struct {
	Tool     string   `json:"tool"`
	InputKey string   `json:"input_key"`
	Decision Decision `json:"decision"`
}

// Engine evaluates permission checks for tool calls.
type Engine struct {
	mode       Mode
	rules      []Rule
	persistent map[string]Decision // keyed by "toolName:inputHash"
	mu         sync.RWMutex
}

// Checker is the permission-check interface used by the agent loop.
type Checker interface {
	Check(toolName string, input json.RawMessage) Decision
}

// DecisionResult is a structured permission outcome with a human reason.
type DecisionResult struct {
	Decision Decision
	Reason   string
	Code     string
}

// NewEngine creates a new permission engine with the given mode and rules.
func NewEngine(mode Mode, rules []Rule) *Engine {
	normalized := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		normalized = append(normalized, normalizeRule(rule))
	}
	sortRulesBySourcePrecedence(normalized)

	return &Engine{
		mode:       mode,
		rules:      normalized,
		persistent: make(map[string]Decision),
	}
}

// Check evaluates whether a tool call should be allowed, denied, or needs confirmation.
func (e *Engine) Check(toolName string, input json.RawMessage) Decision {
	return e.CheckDetailed(toolName, input).Decision
}

// CheckDetailed evaluates permissions and includes a decision reason.
func (e *Engine) CheckDetailed(toolName string, input json.RawMessage) DecisionResult {
	// Check persistent (remembered) decisions first.
	e.mu.RLock()
	key := persistentKey(toolName, input)
	if decision, ok := e.persistent[key]; ok {
		e.mu.RUnlock()
		return DecisionResult{Decision: decision, Reason: "persistent decision", Code: "persistent_decision"}
	}
	e.mu.RUnlock()

	// Check explicit rules in order; first match wins.
	for _, rule := range e.rules {
		if matchesRule(rule, toolName, input) {
			return DecisionResult{Decision: rule.Decision, Reason: "matched permission rule", Code: "matched_rule"}
		}
	}

	isBashTool := strings.EqualFold(toolName, "bash") || strings.EqualFold(toolName, "shell")
	if isBashTool {
		if decision, reason, ok := checkBashPathAndRedirectionConstraints(input); ok {
			return DecisionResult{Decision: decision, Reason: reason, Code: "bash_path_constraint"}
		}
		classification := ClassifyCommandDetailed(extractBashCommand(input))
		if classifyDecision, classifyReason, ok := ClassifyAmbiguousParseDecision(classification); ok {
			if classifyDecision == "deny" {
				return DecisionResult{Decision: DecisionDeny, Reason: classifyReason, Code: "ambiguous_shell_deny"}
			}
			return DecisionResult{Decision: DecisionAsk, Reason: classifyReason, Code: "ambiguous_shell_ask"}
		}
	}

	// Fall back to mode-based defaults.
	switch e.mode {
	case ModePlan, ModeDefault:
		if isReadOnlyTool(toolName) {
			return DecisionResult{Decision: DecisionAllow, Reason: "read-only tool in default/plan mode", Code: "readonly_default_allow"}
		}
		if isBashTool {
			if classifyBashDecision(input) == DecisionDeny {
				return DecisionResult{Decision: DecisionDeny, Reason: "critical bash command denied", Code: "critical_bash_deny"}
			}
		}
		return DecisionResult{Decision: DecisionAsk, Reason: "default confirmation required", Code: "default_ask"}

	case ModeAuto, ModeAcceptEdits:
		if isReadOnlyTool(toolName) {
			return DecisionResult{Decision: DecisionAllow, Reason: "read-only tool in auto mode", Code: "readonly_auto_allow"}
		}
		// For bash commands, classify the risk.
		if isBashTool {
			riskDecision := classifyBashDecision(input)
			switch riskDecision {
			case DecisionAllow:
				return DecisionResult{Decision: DecisionAllow, Reason: "low-risk bash command in auto mode", Code: "bash_auto_allow"}
			case DecisionDeny:
				return DecisionResult{Decision: DecisionDeny, Reason: "critical-risk bash command in auto mode", Code: "bash_auto_deny"}
			default:
				return DecisionResult{Decision: DecisionAsk, Reason: "non-low-risk bash command in auto mode", Code: "bash_auto_ask"}
			}
		}
		return DecisionResult{Decision: DecisionAsk, Reason: "non-read-only tool requires confirmation in auto mode", Code: "auto_nonreadonly_ask"}

	case ModeBypass:
		// Bypass mode still applies explicit/persistent safety decisions above.
		// It only bypasses the mode default once no objectioning rule matched.
		return DecisionResult{Decision: DecisionAllow, Reason: "bypass mode", Code: "bypass_allow"}

	default:
		return DecisionResult{Decision: DecisionAsk, Reason: "unknown mode", Code: "unknown_mode"}
	}
}

// Remember stores a persistent decision for a specific tool+input combination
// (e.g., "always allow").
func (e *Engine) Remember(toolName string, input json.RawMessage, decision Decision) {
	e.mu.Lock()
	defer e.mu.Unlock()
	key := persistentKey(toolName, input)
	e.persistent[key] = decision
}

// LoadPersistent loads remembered decisions from a JSON file.
func (e *Engine) LoadPersistent(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // no file is fine
		}
		return err
	}

	var entries []persistentEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	for _, entry := range entries {
		key := entry.Tool + ":" + entry.InputKey
		e.persistent[key] = entry.Decision
	}
	return nil
}

// SavePersistent writes remembered decisions to a JSON file.
func (e *Engine) SavePersistent(path string) error {
	e.mu.RLock()
	entries := make([]persistentEntry, 0, len(e.persistent))
	for key, decision := range e.persistent {
		parts := strings.SplitN(key, ":", 2)
		tool := key
		inputKey := ""
		if len(parts) == 2 {
			tool = parts[0]
			inputKey = parts[1]
		}
		entries = append(entries, persistentEntry{
			Tool:     tool,
			InputKey: inputKey,
			Decision: decision,
		})
	}
	e.mu.RUnlock()

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// persistentKey creates a lookup key from tool name and input.
func persistentKey(toolName string, input json.RawMessage) string {
	// Use the compact JSON of input as the key suffix.
	compact := string(input)
	if len(compact) > 200 {
		compact = compact[:200]
	}
	return toolName + ":" + compact
}

// isReadOnlyTool returns true for tools known to be read-only.
func isReadOnlyTool(name string) bool {
	return isReadOnlyToolName(name)
}

func classifyBashDecision(input json.RawMessage) Decision {
	var params struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return DecisionAsk
	}
	risk := ClassifyCommand(params.Command)
	switch {
	case risk <= RiskLow:
		return DecisionAllow
	case risk >= RiskCritical:
		return DecisionDeny
	default:
		return DecisionAsk
	}
}

func extractBashCommand(input json.RawMessage) string {
	var params struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return ""
	}
	return strings.TrimSpace(params.Command)
}

func normalizeRuleSource(source RuleSource) RuleSource {
	normalized := RuleSource(strings.ToLower(strings.TrimSpace(string(source))))
	if normalized == "" {
		return RuleSourceUser
	}
	return normalized
}

func ruleSourcePrecedence(source RuleSource) int {
	switch normalizeRuleSource(source) {
	case RuleSourcePolicy:
		return 1
	case RuleSourceUser:
		return 2
	case RuleSourceProject:
		return 3
	case RuleSourceSession:
		return 4
	default:
		return 0
	}
}

func sortRulesBySourcePrecedence(rules []Rule) {
	if len(rules) <= 1 {
		return
	}
	for i, j := 0, len(rules)-1; i < j; i, j = i+1, j-1 {
		rules[i], rules[j] = rules[j], rules[i]
	}
	sort.SliceStable(rules, func(i, j int) bool {
		return ruleSourcePrecedence(rules[i].Source) > ruleSourcePrecedence(rules[j].Source)
	})
}

// matchesRule checks if a rule applies to the given tool call.
func matchesRule(rule Rule, toolName string, input json.RawMessage) bool {
	// Check tool name pattern.
	if rule.Tool != "" {
		matched, err := matchDSL(rule.Tool, toolName, matchModeGlob)
		if err != nil || !matched {
			return false
		}
	}

	// Check file glob against file_path in input.
	if rule.FileGlob != "" {
		var params struct {
			FilePath string `json:"file_path"`
			Path     string `json:"path"`
		}
		if err := json.Unmarshal(input, &params); err == nil {
			p := params.FilePath
			if p == "" {
				p = params.Path
			}
			if p != "" {
				matched, err := matchDSL(rule.FileGlob, p, matchModeGlob)
				if err != nil || !matched {
					return false
				}
			} else {
				return false
			}
		} else {
			return false
		}
	}

	// Check bash regex against command in input.
	if rule.BashRegex != "" {
		var params struct {
			Command string `json:"command"`
		}
		if err := json.Unmarshal(input, &params); err == nil {
			matched, err := matchDSL(rule.BashRegex, params.Command, matchModeShell)
			if err != nil || !matched {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

type bashConstraintInput struct {
	Command string `json:"command"`
	Workdir string `json:"workdir"`
	Cwd     string `json:"cwd"`
	Path    string `json:"path"`
}

func checkBashPathAndRedirectionConstraints(input json.RawMessage) (Decision, string, bool) {
	var params bashConstraintInput
	if err := json.Unmarshal(input, &params); err != nil {
		return DecisionAsk, fmt.Sprintf("invalid bash input: %v", err), false
	}

	cmd := strings.TrimSpace(params.Command)
	if cmd == "" {
		return DecisionAsk, "missing command", false
	}

	baseDir := firstNonEmpty(strings.TrimSpace(params.Workdir), strings.TrimSpace(params.Cwd), strings.TrimSpace(params.Path))
	if baseDir != "" {
		baseDir = filepath.Clean(baseDir)
	}

	if target, ok := findDynamicRedirectionTarget(cmd); ok {
		return DecisionAsk, "redirection target uses dynamic shell expansion: " + target, true
	}

	if target, ok := findSensitiveRedirectionTarget(cmd, baseDir); ok {
		return DecisionAsk, "redirection target requires approval: " + target, true
	}

	if baseDir == "" {
		return DecisionAsk, "", false
	}

	argv := shellSplit(cmd)
	if len(argv) == 0 {
		return DecisionAsk, "", false
	}

	base := strings.ToLower(argv[0])
	args := argv[1:]
	paths, checks := extractCandidatePaths(base, args)
	if !checks {
		return DecisionAsk, "", false
	}

	for _, p := range paths {
		resolved, ok := resolvePathForConstraintCheck(p, baseDir)
		if !ok {
			return DecisionAsk, "path uses dynamic shell expansion and requires approval: " + p, true
		}
		if !pathWithinBase(baseDir, resolved) {
			return DecisionAsk, "path is outside allowed working directory: " + resolved, true
		}
		if isSensitivePath(resolved) {
			return DecisionAsk, "sensitive path requires approval: " + resolved, true
		}
	}

	return DecisionAsk, "", false
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

var redirectionRe = regexp.MustCompile(`(?:^|\s|[;&|])(?:\d*>|\d*>>|>\||>|>>|&>|&>>)\s*([^\s;&|]+)`)

func findDynamicRedirectionTarget(command string) (string, bool) {
	for _, match := range redirectionRe.FindAllStringSubmatch(command, -1) {
		if len(match) < 2 {
			continue
		}
		target := strings.Trim(match[1], "\"'")
		if target == "" {
			continue
		}
		if strings.ContainsAny(target, "$`*?[]{}()") {
			return target, true
		}
	}
	return "", false
}

func findSensitiveRedirectionTarget(command, baseDir string) (string, bool) {
	for _, match := range redirectionRe.FindAllStringSubmatch(command, -1) {
		if len(match) < 2 {
			continue
		}
		target := strings.Trim(match[1], "\"'")
		if target == "" || target == "/dev/null" {
			continue
		}
		resolved, ok := resolvePathForConstraintCheck(target, baseDir)
		if !ok {
			continue
		}
		if baseDir != "" && !pathWithinBase(baseDir, resolved) {
			return resolved, true
		}
		if isSensitivePath(resolved) {
			return resolved, true
		}
	}
	return "", false
}

func shellSplit(command string) []string {
	var out []string
	var b strings.Builder
	inSingle := false
	inDouble := false
	escaped := false

	flush := func() {
		if b.Len() == 0 {
			return
		}
		out = append(out, b.String())
		b.Reset()
	}

	for i := 0; i < len(command); i++ {
		ch := command[i]
		if escaped {
			b.WriteByte(ch)
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if inSingle {
			if ch == '\'' {
				inSingle = false
				continue
			}
			b.WriteByte(ch)
			continue
		}
		if inDouble {
			if ch == '"' {
				inDouble = false
				continue
			}
			b.WriteByte(ch)
			continue
		}
		switch ch {
		case '\'':
			inSingle = true
		case '"':
			inDouble = true
		case ' ', '\t', '\n', ';', '|', '&':
			flush()
		default:
			b.WriteByte(ch)
		}
	}
	flush()
	return out
}

func extractCandidatePaths(base string, args []string) ([]string, bool) {
	switch base {
	case "cd", "ls", "find", "mkdir", "touch", "rm", "rmdir", "mv", "cp", "cat", "head", "tail", "sort", "uniq", "wc", "cut", "paste", "column", "tr", "file", "stat", "diff", "awk", "strings", "hexdump", "od", "base64", "nl", "grep", "rg", "sed", "jq":
		return filterPositionalArgs(args), true
	default:
		return nil, false
	}
}

func filterPositionalArgs(args []string) []string {
	paths := make([]string, 0, len(args))
	afterDoubleDash := false
	for i, arg := range args {
		if arg == "" {
			continue
		}
		if !afterDoubleDash && arg == "--" {
			afterDoubleDash = true
			continue
		}
		if !afterDoubleDash && strings.HasPrefix(arg, "-") {
			continue
		}
		if i == 0 && (arg == "." || arg == "..") {
			continue
		}
		paths = append(paths, arg)
	}
	return paths
}

func resolvePathForConstraintCheck(candidate, baseDir string) (string, bool) {
	candidate = strings.TrimSpace(strings.Trim(candidate, "\"'"))
	if candidate == "" {
		return "", false
	}
	if strings.HasPrefix(candidate, "~") {
		return "", false
	}
	if strings.ContainsAny(candidate, "$`*?[]{}()") {
		return "", false
	}
	if filepath.IsAbs(candidate) {
		return filepath.Clean(candidate), true
	}
	if baseDir == "" {
		return filepath.Clean(candidate), true
	}
	return filepath.Clean(filepath.Join(baseDir, candidate)), true
}

func pathWithinBase(baseDir, path string) bool {
	baseDir = filepath.Clean(baseDir)
	path = filepath.Clean(path)
	if baseDir == "" {
		return true
	}
	rel, err := filepath.Rel(baseDir, path)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func isSensitivePath(path string) bool {
	clean := filepath.ToSlash(filepath.Clean(path))
	sensitive := []string{"/.git", "/.claude", "/.vscode", "/.ssh", "/etc"}
	for _, prefix := range sensitive {
		if clean == prefix || strings.HasPrefix(clean, prefix+"/") {
			return true
		}
	}
	return false
}
