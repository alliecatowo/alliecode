package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/alliecatowo/alliecode/internal/permissions"
	"github.com/alliecatowo/alliecode/internal/types"
)

// BashTool executes shell commands.
type BashTool struct{}

type bashInput struct {
	Command                   string                  `json:"command"`
	Description               string                  `json:"description"`
	Timeout                   int                     `json:"timeout"`
	DangerouslyDisableSandbox bool                    `json:"dangerouslyDisableSandbox"`
	SandboxPolicy             *bashSandboxPolicyInput `json:"sandboxPolicy,omitempty"`
}

type bashSandboxPolicyInput struct {
	AllowOutsideWorkspaceWrites bool `json:"allowOutsideWorkspaceWrites"`
	RequireApprovalForRisky     bool `json:"requireApprovalForRisky"`
	DenyRiskyConstructs         bool `json:"denyRiskyConstructs"`
}

type bashSandboxPolicy struct {
	Enabled                     bool
	AllowOutsideWorkspaceWrites bool
	RequireApprovalForRisky     bool
	DenyRiskyConstructs         bool
}

type bashRiskKind string

const (
	bashRiskSubshell              bashRiskKind = "subshell"
	bashRiskCommandSubstitution   bashRiskKind = "command_substitution"
	bashRiskHeredoc               bashRiskKind = "heredoc"
	bashRiskUnbalancedQuotes      bashRiskKind = "unbalanced_quotes"
	bashRiskChainedOperators      bashRiskKind = "chained_operators"
	bashRiskRedirects             bashRiskKind = "redirects"
	bashRiskEscapedOperator       bashRiskKind = "escaped_operator_obfuscation"
	bashRiskEscapedNewlineOp      bashRiskKind = "escaped_newline_operator_obfuscation"
	bashRiskCommentQuoteDesync    bashRiskKind = "comment_quote_desync"
	bashRiskSuspiciousSubstCombo  bashRiskKind = "suspicious_substitution_combo"
	bashRiskOutsideWorkspaceWrite bashRiskKind = "outside_workspace_write"
)

type bashRisk struct {
	Kind    bashRiskKind
	Message string
}

type bashToken struct {
	Type  string
	Value string
	Pos   int
}

type bashRedirect struct {
	Operator string
	Target   string
	Pos      int
}

type bashPreflightAnalysis struct {
	Tokens            []bashToken
	Risks             []bashRisk
	Redirects         []bashRedirect
	OutsideWritePaths []string
}

type bashPolicyDecision struct {
	Behavior string
	Reason   string
	RuleID   string
	Analysis bashPreflightAnalysis
}

func (t *BashTool) Name() string { return "Bash" }

func (t *BashTool) Description() string {
	return "Executes a shell command and returns its output. Use this for running programs, scripts, git commands, and system operations."
}

func (t *BashTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"command": {
				Type:        "string",
				Description: "The shell command to execute.",
			},
			"description": {
				Type:        "string",
				Description: "A brief description of what the command does.",
			},
			"timeout": {
				Type:        "integer",
				Description: "Timeout in milliseconds. Defaults to 120000 (2 minutes).",
			},
		},
		Required: []string{"command"},
	}
}

func (t *BashTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	started := time.Now()
	var in bashInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Invalid input: %v", err), IsError: true}, nil
	}

	command := strings.TrimSpace(in.Command)
	if command == "" {
		return types.ToolResult{Content: "command is required", IsError: true}, nil
	}

	preflight := evaluateBashPreflight(command, toolCtx.WorkingDir, in.SandboxPolicy)
	if preflight.Behavior == "deny" || preflight.Behavior == "ask" {
		return types.ToolResult{Content: renderPreflightResultMessage(preflight), IsError: true}, nil
	}
	if denyMsg := denyDestructiveCommand(command); denyMsg != "" {
		return types.ToolResult{Content: denyMsg, IsError: true}, nil
	}

	timeout := 120 * time.Second
	if in.Timeout > 0 {
		timeout = time.Duration(in.Timeout) * time.Millisecond
	}

	provider := selectShellProvider(t.Name())
	res, err := executeWithShellProvider(ctx, toolCtx, provider, command, timeout)
	if err != nil {
		return res, err
	}
	classify := permissions.ClassifyCommandDetailed(command)
	riskCodes := make([]string, 0, len(classify.Reasons))
	for _, reason := range classify.Reasons {
		riskCodes = append(riskCodes, reason.Code)
	}
	audit := &shellExecutionAudit{
		RiskLevel:      classify.Level.String(),
		RiskCodes:      riskCodes,
		PolicyDecision: preflight.Behavior,
		PolicyRuleID:   preflight.RuleID,
		PolicyReason:   preflight.Reason,
	}
	meta := formatShellExecutionBlock(t.Name(), provider.Type(), toolCtx.WorkingDir, timeout, time.Since(started), command, res.Content, res.IsError, audit)
	if strings.TrimSpace(res.Content) == "" {
		res.Content = meta
	} else {
		res.Content = res.Content + "\n\n" + meta
	}
	return res, nil
}

func (t *BashTool) IsReadOnly(input types.ToolInput) bool {
	return false
}

// destructivePatterns are shell command patterns considered dangerous.
var destructiveRules = []struct {
	id      string
	pattern *regexp.Regexp
	reason  string
}{
	{id: "rm_force", pattern: regexp.MustCompile(`(?i)\brm\s+(-[a-zA-Z]*r[a-zA-Z]*f|--force)\b`), reason: "forceful recursive deletion detected"},
	{id: "raw_disk_write", pattern: regexp.MustCompile(`(?i)\b(dd|mkfs)\b`), reason: "raw disk mutation command detected"},
	{id: "git_force_push", pattern: regexp.MustCompile(`(?i)\bgit\s+push\s+(?:--force|-f)\b`), reason: "git force push detected"},
	{id: "git_hard_reset", pattern: regexp.MustCompile(`(?i)\bgit\s+reset\s+--hard\b`), reason: "git hard reset detected"},
	{id: "git_clean_force", pattern: regexp.MustCompile(`(?i)\bgit\s+clean\s+-[a-zA-Z]*f\b`), reason: "git clean force deletion detected"},
	{id: "git_checkout_discard", pattern: regexp.MustCompile(`(?i)\bgit\s+checkout\s+--\s*\.`), reason: "git checkout discard-all detected"},
	{id: "git_branch_delete_force", pattern: regexp.MustCompile(`(?i)\bgit\s+branch\s+-D\b`), reason: "git force branch deletion detected"},
	{id: "chmod_777_recursive", pattern: regexp.MustCompile(`(?i)\bchmod\s+-R\s+777\b`), reason: "recursive chmod 777 detected"},
	{id: "device_redirect", pattern: regexp.MustCompile(`(?i)\b>\s*/dev/sd[a-z]`), reason: "direct write to block device detected"},
	{id: "pipe_to_shell", pattern: regexp.MustCompile(`(?i)\bcurl\b.*\|\s*(sh|bash)\b`), reason: "remote script piped to shell detected"},
}

type bashRuleMatch struct {
	id     string
	reason string
}

func renderBashDenyBlock(denyType, primaryRuleID string, matches []bashRuleMatch) string {
	matchedRuleIDs := make([]string, 0, len(matches))
	matchedReasons := make([]string, 0, len(matches))
	for _, match := range matches {
		matchedRuleIDs = append(matchedRuleIDs, match.id)
		matchedReasons = append(matchedReasons, match.reason)
	}

	return renderStructuredBlock("bash_deny", []structuredField{
		{Key: "blocked", Value: "true"},
		{Key: "taxonomy_version", Value: "1"},
		{Key: "deny_type", Value: denyType},
		{Key: "primary_rule_id", Value: primaryRuleID},
		{Key: "matched_rule_ids", Value: strings.Join(matchedRuleIDs, ",")},
		{Key: "matched_reasons", Value: strings.Join(matchedReasons, " | ")},
	})
}

func renderBashPreflightBlock(decision bashPolicyDecision) string {
	riskKinds := make([]string, 0, len(decision.Analysis.Risks))
	for _, risk := range decision.Analysis.Risks {
		riskKinds = append(riskKinds, string(risk.Kind))
	}
	sort.Strings(riskKinds)

	outsidePaths := append([]string(nil), decision.Analysis.OutsideWritePaths...)
	sort.Strings(outsidePaths)

	return renderStructuredBlock("bash_preflight", []structuredField{
		{Key: "taxonomy_version", Value: "2"},
		{Key: "decision", Value: decision.Behavior},
		{Key: "rule_id", Value: decision.RuleID},
		{Key: "reason", Value: decision.Reason},
		{Key: "risk_kinds", Value: strings.Join(riskKinds, ",")},
		{Key: "redirect_count", Value: fmt.Sprintf("%d", len(decision.Analysis.Redirects))},
		{Key: "outside_write_count", Value: fmt.Sprintf("%d", len(outsidePaths))},
		{Key: "outside_write_paths", Value: strings.Join(outsidePaths, ",")},
	})
}

func renderPreflightResultMessage(decision bashPolicyDecision) string {
	prefix := "Command denied by bash preflight policy"
	if decision.Behavior == "ask" {
		prefix = "Command requires approval by bash preflight policy"
	}
	return prefix + ": " + decision.Reason + "\n\n" + renderBashPreflightBlock(decision)
}

// classifyCommand checks destructive rule matches in deterministic order.
func classifyCommand(command string) []bashRuleMatch {
	matches := make([]bashRuleMatch, 0, len(destructiveRules))
	for _, rule := range destructiveRules {
		if rule.pattern.MatchString(command) {
			matches = append(matches, bashRuleMatch{id: rule.id, reason: rule.reason})
		}
	}
	return matches
}

func validateBashCommand(command string) string {
	decision := evaluateBashPreflight(command, "", nil)
	if decision.Behavior == "deny" {
		return renderPreflightResultMessage(decision)
	}
	return ""
}

func denyDestructiveCommand(command string) string {
	matches := classifyCommand(command)
	if len(matches) > 0 {
		return "Command denied: potential destructive operation detected. Refine the command to a safe read-only operation or request explicit permission for the risky action." + "\n\n" + renderBashDenyBlock("destructive", matches[0].id, matches)
	}
	return ""
}

func (t *BashTool) IsDestructive(input types.ToolInput) bool {
	var in bashInput
	if err := json.Unmarshal(input, &in); err != nil {
		return true // err on the side of caution
	}
	command := strings.TrimSpace(in.Command)
	return len(classifyCommand(command)) > 0
}

func (t *BashTool) IsConcurrencySafe(input types.ToolInput) bool {
	return false
}

func (t *BashTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	var in bashInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.PermissionDenied
	}
	command := strings.TrimSpace(in.Command)
	if command == "" {
		return types.PermissionDenied
	}
	baseDecision := evaluateToolPermissionByFamily(toolFamilyShell, "bash", input, toolCtx, types.PermissionAsk)
	preflightDecision := evaluateBashPreflight(command, toolCtx.WorkingDir, in.SandboxPolicy)
	if preflightDecision.Behavior == "deny" {
		return types.PermissionDenied
	}
	if preflightDecision.Behavior == "ask" {
		return maxPermissionSeverity(baseDecision, types.PermissionAsk)
	}
	// Destructive commands require explicit confirmation
	if len(classifyCommand(command)) > 0 {
		return maxPermissionSeverity(baseDecision, types.PermissionAsk)
	}
	return baseDecision
}

// parseCommand is a helper that extracts the command string from raw input.
func parseCommand(input types.ToolInput) string {
	var in bashInput
	if err := json.Unmarshal(input, &in); err != nil {
		return ""
	}
	return strings.TrimSpace(in.Command)
}

func evaluateBashPreflight(command, workspaceRoot string, policyInput *bashSandboxPolicyInput) bashPolicyDecision {
	policy := resolveSandboxPolicy(policyInput)
	analysis := analyzeBashCommand(command)
	analysis.OutsideWritePaths = findOutsideWorkspaceWritePaths(analysis, workspaceRoot)
	if len(analysis.OutsideWritePaths) > 0 {
		analysis.Risks = append(analysis.Risks, bashRisk{
			Kind:    bashRiskOutsideWorkspaceWrite,
			Message: "write target resolves outside workspace",
		})
	}

	if deniedRule, deniedReason := deniedChainingRule(analysis); deniedRule != "" {
		return bashPolicyDecision{
			Behavior: "deny",
			Reason:   deniedReason,
			RuleID:   deniedRule,
			Analysis: analysis,
		}
	}

	if ambiguousRule, ambiguousReason, deny := classifyAmbiguousParseRisk(analysis); ambiguousRule != "" {
		behavior := "ask"
		if deny {
			behavior = "deny"
		}
		return bashPolicyDecision{
			Behavior: behavior,
			Reason:   ambiguousReason,
			RuleID:   ambiguousRule,
			Analysis: analysis,
		}
	}

	if hasRisk(analysis.Risks, bashRiskUnbalancedQuotes) {
		return bashPolicyDecision{
			Behavior: "deny",
			Reason:   "unbalanced shell quotes detected",
			RuleID:   "unbalanced_quotes",
			Analysis: analysis,
		}
	}

	if hasRisk(analysis.Risks, bashRiskOutsideWorkspaceWrite) && !policy.AllowOutsideWorkspaceWrites {
		return bashPolicyDecision{
			Behavior: "ask",
			Reason:   "command writes outside the workspace and policy does not allow it",
			RuleID:   "outside_workspace_write",
			Analysis: analysis,
		}
	}

	if policy.Enabled {
		hasRiskyConstruct := hasAnyRisk(analysis.Risks, []bashRiskKind{
			bashRiskSubshell,
			bashRiskCommandSubstitution,
			bashRiskHeredoc,
			bashRiskRedirects,
		})

		if hasRiskyConstruct && policy.DenyRiskyConstructs {
			return bashPolicyDecision{
				Behavior: "deny",
				Reason:   "sandbox policy denies risky shell constructs",
				RuleID:   "sandbox_deny_risky_constructs",
				Analysis: analysis,
			}
		}
		if hasRiskyConstruct && policy.RequireApprovalForRisky {
			return bashPolicyDecision{
				Behavior: "ask",
				Reason:   "sandbox policy requires approval for risky shell constructs",
				RuleID:   "sandbox_ask_risky_constructs",
				Analysis: analysis,
			}
		}
	}

	return bashPolicyDecision{Behavior: "allow", Reason: "preflight checks passed", RuleID: "allow", Analysis: analysis}
}

func resolveSandboxPolicy(in *bashSandboxPolicyInput) bashSandboxPolicy {
	policy := bashSandboxPolicy{
		Enabled:                     true,
		AllowOutsideWorkspaceWrites: false,
		RequireApprovalForRisky:     true,
		DenyRiskyConstructs:         false,
	}
	if in == nil {
		return policy
	}
	policy.AllowOutsideWorkspaceWrites = in.AllowOutsideWorkspaceWrites
	policy.RequireApprovalForRisky = in.RequireApprovalForRisky
	policy.DenyRiskyConstructs = in.DenyRiskyConstructs
	return policy
}

func hasRisk(risks []bashRisk, target bashRiskKind) bool {
	for _, risk := range risks {
		if risk.Kind == target {
			return true
		}
	}
	return false
}

func hasAnyRisk(risks []bashRisk, targets []bashRiskKind) bool {
	for _, target := range targets {
		if hasRisk(risks, target) {
			return true
		}
	}
	return false
}

func addRiskUnique(risks []bashRisk, risk bashRisk) []bashRisk {
	for _, existing := range risks {
		if existing.Kind == risk.Kind {
			return risks
		}
	}
	return append(risks, risk)
}

func analyzeBashCommand(command string) bashPreflightAnalysis {
	analysis := bashPreflightAnalysis{}
	for _, risk := range detectAdvancedObfuscationRisks(command) {
		analysis.Risks = addRiskUnique(analysis.Risks, risk)
	}

	var current strings.Builder
	currentStart := -1

	flushWord := func(endPos int) {
		if current.Len() == 0 {
			return
		}
		analysis.Tokens = append(analysis.Tokens, bashToken{Type: "word", Value: current.String(), Pos: currentStart})
		current.Reset()
		currentStart = endPos
	}

	inSingle := false
	inDouble := false
	inBacktick := false
	escaped := false

	for i := 0; i < len(command); i++ {
		ch := command[i]

		if currentStart == -1 {
			currentStart = i
		}

		if escaped {
			current.WriteByte(ch)
			escaped = false
			continue
		}

		if ch == '\\' {
			current.WriteByte(ch)
			escaped = true
			continue
		}

		if inSingle {
			current.WriteByte(ch)
			if ch == '\'' {
				inSingle = false
			}
			continue
		}

		if inDouble {
			current.WriteByte(ch)
			if ch == '"' {
				inDouble = false
			}
			if ch == '$' && i+1 < len(command) && command[i+1] == '(' {
				analysis.Risks = addRiskUnique(analysis.Risks, bashRisk{Kind: bashRiskCommandSubstitution, Message: "contains $() command substitution"})
			}
			continue
		}

		if inBacktick {
			current.WriteByte(ch)
			if ch == '`' {
				inBacktick = false
			}
			continue
		}

		switch ch {
		case '\'':
			inSingle = true
			current.WriteByte(ch)
		case '"':
			inDouble = true
			current.WriteByte(ch)
		case '`':
			analysis.Risks = addRiskUnique(analysis.Risks, bashRisk{Kind: bashRiskCommandSubstitution, Message: "contains backtick command substitution"})
			inBacktick = true
			current.WriteByte(ch)
		case ' ', '\t', '\r':
			flushWord(i)
			currentStart = -1
		case '\n':
			flushWord(i)
			analysis.Tokens = append(analysis.Tokens, bashToken{Type: "operator", Value: "\\n", Pos: i})
			analysis.Risks = addRiskUnique(analysis.Risks, bashRisk{Kind: bashRiskChainedOperators, Message: "contains newline command separator"})
			currentStart = -1
		case '&':
			flushWord(i)
			op := "&"
			if i+1 < len(command) && command[i+1] == '&' {
				op = "&&"
				i++
			}
			analysis.Tokens = append(analysis.Tokens, bashToken{Type: "operator", Value: op, Pos: i})
			analysis.Risks = addRiskUnique(analysis.Risks, bashRisk{Kind: bashRiskChainedOperators, Message: "contains chained operator"})
			currentStart = -1
		case '|':
			flushWord(i)
			op := "|"
			if i+1 < len(command) && command[i+1] == '|' {
				op = "||"
				i++
				analysis.Risks = addRiskUnique(analysis.Risks, bashRisk{Kind: bashRiskChainedOperators, Message: "contains chained operator"})
			}
			analysis.Tokens = append(analysis.Tokens, bashToken{Type: "operator", Value: op, Pos: i})
			currentStart = -1
		case ';':
			flushWord(i)
			analysis.Tokens = append(analysis.Tokens, bashToken{Type: "operator", Value: ";", Pos: i})
			analysis.Risks = addRiskUnique(analysis.Risks, bashRisk{Kind: bashRiskChainedOperators, Message: "contains chained operator"})
			currentStart = -1
		case '(':
			if i > 0 && command[i-1] == '$' {
				analysis.Risks = addRiskUnique(analysis.Risks, bashRisk{Kind: bashRiskCommandSubstitution, Message: "contains $() command substitution"})
			}
			if i == 0 || isBoundaryByte(command[i-1]) {
				analysis.Risks = addRiskUnique(analysis.Risks, bashRisk{Kind: bashRiskSubshell, Message: "contains subshell expression"})
			}
			current.WriteByte(ch)
		case '>':
			flushWord(i)
			op, next := parseRedirectOperator(command, i)
			target, end := parseRedirectTarget(command, next)
			analysis.Redirects = append(analysis.Redirects, bashRedirect{Operator: op, Target: target, Pos: i})
			analysis.Tokens = append(analysis.Tokens, bashToken{Type: "redirect", Value: op, Pos: i})
			analysis.Risks = addRiskUnique(analysis.Risks, bashRisk{Kind: bashRiskRedirects, Message: "contains output redirection"})
			if strings.HasPrefix(op, "<<") {
				analysis.Risks = addRiskUnique(analysis.Risks, bashRisk{Kind: bashRiskHeredoc, Message: "contains heredoc"})
			}
			i = end - 1
			currentStart = -1
		case '<':
			flushWord(i)
			op, next := parseRedirectOperator(command, i)
			target, end := parseRedirectTarget(command, next)
			analysis.Redirects = append(analysis.Redirects, bashRedirect{Operator: op, Target: target, Pos: i})
			analysis.Tokens = append(analysis.Tokens, bashToken{Type: "redirect", Value: op, Pos: i})
			analysis.Risks = addRiskUnique(analysis.Risks, bashRisk{Kind: bashRiskRedirects, Message: "contains redirection"})
			if strings.HasPrefix(op, "<<") {
				analysis.Risks = addRiskUnique(analysis.Risks, bashRisk{Kind: bashRiskHeredoc, Message: "contains heredoc"})
			}
			i = end - 1
			currentStart = -1
		default:
			current.WriteByte(ch)
		}
	}

	flushWord(len(command))
	if inSingle || inDouble || inBacktick {
		analysis.Risks = addRiskUnique(analysis.Risks, bashRisk{Kind: bashRiskUnbalancedQuotes, Message: "contains unbalanced shell quotes"})
	}

	return analysis
}

func isBoundaryByte(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == ';' || ch == '&' || ch == '|'
}

func parseRedirectOperator(command string, i int) (string, int) {
	if i >= len(command) {
		return "", i
	}
	if command[i] == '>' {
		if i+1 < len(command) && command[i+1] == '>' {
			return ">>", i + 2
		}
		if i+1 < len(command) && command[i+1] == '|' {
			return ">|", i + 2
		}
		return ">", i + 1
	}
	if command[i] == '<' {
		if i+1 < len(command) && command[i+1] == '<' {
			if i+2 < len(command) && command[i+2] == '-' {
				return "<<-", i + 3
			}
			return "<<", i + 2
		}
		return "<", i + 1
	}
	return string(command[i]), i + 1
}

func parseRedirectTarget(command string, start int) (string, int) {
	i := start
	for i < len(command) && (command[i] == ' ' || command[i] == '\t') {
		i++
	}
	if i >= len(command) {
		return "", i
	}
	if command[i] == '\'' || command[i] == '"' {
		quote := command[i]
		j := i + 1
		for j < len(command) && command[j] != quote {
			if command[j] == '\\' && j+1 < len(command) {
				j += 2
				continue
			}
			j++
		}
		if j < len(command) {
			return command[i : j+1], j + 1
		}
		return command[i:], len(command)
	}
	j := i
	for j < len(command) && !isBoundaryByte(command[j]) && command[j] != '>' && command[j] != '<' {
		j++
	}
	return command[i:j], j
}

func findOutsideWorkspaceWritePaths(analysis bashPreflightAnalysis, workspaceRoot string) []string {
	if workspaceRoot == "" {
		return nil
	}
	workspaceRoot = filepath.Clean(workspaceRoot)
	outside := map[string]struct{}{}

	for _, redir := range analysis.Redirects {
		if !isWriteRedirectOperator(redir.Operator) {
			continue
		}
		if redir.Target == "" || redir.Target == "/dev/null" {
			continue
		}
		resolved, ok := resolvePathForPolicy(redir.Target, workspaceRoot)
		if !ok {
			outside["<dynamic:"+redir.Target+">"] = struct{}{}
			continue
		}
		if !pathWithinWorkspace(workspaceRoot, resolved) {
			outside[resolved] = struct{}{}
		}
	}

	list := make([]string, 0, len(outside))
	for p := range outside {
		list = append(list, p)
	}
	sort.Strings(list)
	return list
}

func isWriteRedirectOperator(op string) bool {
	return op == ">" || op == ">>" || op == ">|"
}

func resolvePathForPolicy(target, workspaceRoot string) (string, bool) {
	target = strings.TrimSpace(target)
	target = strings.Trim(target, "\"'")
	if target == "" {
		return "", false
	}
	if strings.ContainsAny(target, "$`*?[]{}()") {
		return "", false
	}
	if filepath.IsAbs(target) {
		return filepath.Clean(target), true
	}
	return filepath.Clean(filepath.Join(workspaceRoot, target)), true
}

func pathWithinWorkspace(workspaceRoot, candidate string) bool {
	rel, err := filepath.Rel(workspaceRoot, candidate)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func deniedChainingRule(analysis bashPreflightAnalysis) (string, string) {
	for _, tok := range analysis.Tokens {
		if tok.Type != "operator" {
			continue
		}
		switch tok.Value {
		case "\\n":
			return "newline", "newline-separated commands are not allowed; use a single command"
		case ";":
			return "semicolon", "semicolon command chaining is not allowed"
		case "||":
			return "or_chaining", "conditional chaining with || is not allowed"
		case "&":
			return "background", "background execution with & is not allowed"
		}
	}
	return "", ""
}

func classifyAmbiguousParseRisk(analysis bashPreflightAnalysis) (string, string, bool) {
	if hasRisk(analysis.Risks, bashRiskCommentQuoteDesync) {
		return "comment_quote_desync", "quoted comment pattern can desynchronize parser state", true
	}
	if hasRisk(analysis.Risks, bashRiskEscapedNewlineOp) {
		return "escaped_newline_operator_obfuscation", "escaped newline before operator obscures command structure", true
	}
	if hasRisk(analysis.Risks, bashRiskEscapedOperator) {
		return "escaped_operator_obfuscation", "backslash-escaped operator can hide command structure", false
	}
	if hasRisk(analysis.Risks, bashRiskSuspiciousSubstCombo) {
		return "suspicious_substitution_combo", "mixed substitution constructs require manual review", false
	}
	return "", "", false
}

func detectAdvancedObfuscationRisks(command string) []bashRisk {
	var risks []bashRisk

	if hasEscapedOperatorOutsideQuotes(command) {
		risks = append(risks, bashRisk{Kind: bashRiskEscapedOperator, Message: "contains backslash-escaped operator outside quotes"})
	}
	if hasEscapedNewlineBeforeOperator(command) {
		risks = append(risks, bashRisk{Kind: bashRiskEscapedNewlineOp, Message: "contains escaped newline before operator"})
	}
	if hasQuotedCommentDesyncPattern(command) {
		risks = append(risks, bashRisk{Kind: bashRiskCommentQuoteDesync, Message: "contains quote characters in # comment"})
	}
	if hasSuspiciousSubstitutionCombo(command) {
		risks = append(risks, bashRisk{Kind: bashRiskSuspiciousSubstCombo, Message: "contains suspicious mixed substitution constructs"})
	}

	return risks
}

func hasEscapedOperatorOutsideQuotes(command string) bool {
	inSingle := false
	inDouble := false
	escaped := false

	for i := 0; i < len(command); i++ {
		ch := command[i]

		if escaped {
			escaped = false
			continue
		}

		if ch == '\\' && !inSingle {
			if !inDouble && i+1 < len(command) && strings.ContainsRune(";|&<>", rune(command[i+1])) {
				return true
			}
			escaped = true
			continue
		}

		if ch == '\'' && !inDouble {
			inSingle = !inSingle
			continue
		}
		if ch == '"' && !inSingle {
			inDouble = !inDouble
			continue
		}

		if ch == '#' && !inSingle && !inDouble {
			for i < len(command) && command[i] != '\n' {
				i++
			}
		}
	}

	return false
}

func hasEscapedNewlineBeforeOperator(command string) bool {
	inSingle := false
	inDouble := false
	escaped := false

	for i := 0; i < len(command); i++ {
		ch := command[i]

		if escaped {
			escaped = false
			continue
		}

		if ch == '\\' && !inSingle {
			if !inDouble && i+1 < len(command) && command[i+1] == '\n' {
				j := i + 2
				for j < len(command) && (command[j] == ' ' || command[j] == '\t') {
					j++
				}
				if j < len(command) {
					if strings.HasPrefix(command[j:], "&&") || strings.HasPrefix(command[j:], "||") || strings.ContainsRune(";|&<>", rune(command[j])) {
						return true
					}
				}
			}
			escaped = true
			continue
		}

		if ch == '\'' && !inDouble {
			inSingle = !inSingle
			continue
		}
		if ch == '"' && !inSingle {
			inDouble = !inDouble
			continue
		}

		if ch == '#' && !inSingle && !inDouble {
			for i < len(command) && command[i] != '\n' {
				i++
			}
		}
	}

	return false
}

func hasQuotedCommentDesyncPattern(command string) bool {
	inSingle := false
	inDouble := false
	escaped := false

	for i := 0; i < len(command); i++ {
		ch := command[i]

		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' && !inSingle {
			escaped = true
			continue
		}
		if ch == '\'' && !inDouble {
			inSingle = !inSingle
			continue
		}
		if ch == '"' && !inSingle {
			inDouble = !inDouble
			continue
		}

		if ch == '#' && !inSingle && !inDouble {
			lineEnd := strings.IndexByte(command[i:], '\n')
			if lineEnd == -1 {
				lineEnd = len(command)
			} else {
				lineEnd = i + lineEnd
			}
			commentText := command[i+1 : lineEnd]
			if strings.ContainsAny(commentText, "\"'") {
				return true
			}
			i = lineEnd
		}
	}

	return false
}

func hasSuspiciousSubstitutionCombo(command string) bool {
	inSingle := false
	inDouble := false
	escaped := false
	hasDollarParen := false
	hasBacktick := false
	hasDollarBrace := false
	hasProcessSub := false
	hasLegacyArithmetic := false

	for i := 0; i < len(command); i++ {
		ch := command[i]

		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' && !inSingle {
			escaped = true
			continue
		}
		if ch == '\'' && !inDouble {
			inSingle = !inSingle
			continue
		}
		if inSingle {
			continue
		}
		if ch == '"' && !inSingle {
			inDouble = !inDouble
			continue
		}

		if ch == '`' {
			hasBacktick = true
			continue
		}
		if ch == '$' && i+1 < len(command) {
			switch command[i+1] {
			case '(':
				hasDollarParen = true
			case '{':
				hasDollarBrace = true
			case '[':
				hasLegacyArithmetic = true
			}
			continue
		}
		if !inDouble && (ch == '<' || ch == '>') && i+1 < len(command) && command[i+1] == '(' {
			hasProcessSub = true
		}
	}

	if hasBacktick && hasDollarParen {
		return true
	}
	if hasProcessSub && (hasDollarParen || hasBacktick) {
		return true
	}
	if hasLegacyArithmetic && (hasDollarParen || hasBacktick) {
		return true
	}
	if hasDollarParen && hasDollarBrace {
		return true
	}
	return false
}
