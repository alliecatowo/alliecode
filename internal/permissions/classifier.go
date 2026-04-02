package permissions

import (
	"regexp"
	"sort"
	"strings"
)

// RiskLevel categorizes the danger level of a shell command.
type RiskLevel int

const (
	RiskNone     RiskLevel = iota // ls, cat, echo, pwd
	RiskLow                       // git add, git commit, mv, cp
	RiskMedium                    // rm (files), pip install, curl | sh
	RiskHigh                      // rm -rf, git push --force, DROP TABLE
	RiskCritical                  // rm -rf /, dd if=, fork bombs
)

// String returns a human-readable risk level.
func (r RiskLevel) String() string {
	switch r {
	case RiskNone:
		return "none"
	case RiskLow:
		return "low"
	case RiskMedium:
		return "medium"
	case RiskHigh:
		return "high"
	case RiskCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// RiskReason describes why a command was classified at a given risk level.
type RiskReason struct {
	Code     string    `json:"code"`
	Detail   string    `json:"detail"`
	Level    RiskLevel `json:"level"`
	Snippet  string    `json:"snippet,omitempty"`
	Category string    `json:"category,omitempty"`
	Family   string    `json:"family,omitempty"`
	Group    string    `json:"group,omitempty"`
}

// Classification is the structured output for shell command risk analysis.
type Classification struct {
	Level   RiskLevel    `json:"level"`
	Reasons []RiskReason `json:"reasons"`
}

// classificationRule maps a pattern to a risk level.
type classificationRule struct {
	pattern  *regexp.Regexp
	level    RiskLevel
	code     string
	detail   string
	category string
	family   string
	group    string
}

type readOnlyFamily struct {
	name     string
	detail   string
	patterns []*regexp.Regexp
}

// rules are evaluated in order. More specific rules come first.
var classificationRules = []classificationRule{
	// Critical — system-destroying commands
	{regexp.MustCompile(`\brm\s+-[a-zA-Z]*r[a-zA-Z]*f[a-zA-Z]*\s+(--no-preserve-root\b|/\s*$|/\*\s*$)`), RiskCritical, "rm_recursive_force_root", "recursive forced deletion targets root path", "destructive", "filesystem_mutation", "filesystem_wipe"},
	{regexp.MustCompile(`\brm\s+-[a-zA-Z]*f[a-zA-Z]*r[a-zA-Z]*\s+(--no-preserve-root\b|/\s*$|/\*\s*$)`), RiskCritical, "rm_recursive_force_root", "recursive forced deletion targets root path", "destructive", "filesystem_mutation", "filesystem_wipe"},
	{regexp.MustCompile(`\brm\s+(-[a-zA-Z]*f[a-zA-Z]*\s+)?/\s*$`), RiskCritical, "rm_root", "remove command targets root path", "destructive", "filesystem_mutation", "filesystem_wipe"},
	{regexp.MustCompile(`\bdd\s+.*\b(of=/dev/(sd|nvme|vd|xvd))`), RiskCritical, "dd_raw_device_write", "dd writes directly to block device", "destructive", "device_mutation", "raw_device_write"},
	{regexp.MustCompile(`\bdd\s+.*\b(if=|of=)`), RiskCritical, "dd_disk_image", "dd with explicit input/output can overwrite data", "destructive", "device_mutation", "disk_image_write"},
	{regexp.MustCompile(`\bmkfs(\.[a-zA-Z0-9]+)?\b`), RiskCritical, "mkfs_filesystem_create", "mkfs recreates filesystem and destroys existing data", "destructive", "filesystem_mutation", "filesystem_reformat"},
	{regexp.MustCompile(`:\(\)\{.*\|.*&\}`), RiskCritical, "fork_bomb", "fork bomb pattern detected", "destructive", "process_control", "resource_exhaustion"},
	{regexp.MustCompile(`>\s*/dev/(sd|nvme|vd|xvd)`), RiskCritical, "raw_device_redirect", "redirecting output to block device", "destructive", "device_mutation", "raw_device_write"},
	{regexp.MustCompile(`\bchmod\s+-R\s+777\s+(/\s*$|/\*\s*$)`), RiskCritical, "chmod_world_writable_root", "recursive chmod 777 on root path", "destructive", "permission_mutation", "permission_overexposure"},

	// High — destructive but recoverable
	{regexp.MustCompile(`\brm\s+-[a-zA-Z]*r`), RiskHigh, "rm_recursive", "recursive remove command detected", "destructive", "filesystem_mutation", "recursive_delete"},
	{regexp.MustCompile(`\brm\s+-[a-zA-Z]*f`), RiskHigh, "rm_force", "forced remove command detected", "destructive", "filesystem_mutation", "forced_delete"},
	{regexp.MustCompile(`\bgit\s+push\s+--force`), RiskHigh, "git_force_push", "force push rewrites remote history", "destructive", "vcs_mutation", "history_rewrite"},
	{regexp.MustCompile(`\bgit\s+push\s+-f\b`), RiskHigh, "git_force_push", "force push rewrites remote history", "destructive", "vcs_mutation", "history_rewrite"},
	{regexp.MustCompile(`\bgit\s+reset\s+--hard`), RiskHigh, "git_reset_hard", "hard reset discards local changes", "destructive", "vcs_mutation", "workspace_discard"},
	{regexp.MustCompile(`\bgit\s+clean\s+-[a-zA-Z]*f`), RiskHigh, "git_clean_force", "force clean removes untracked files", "destructive", "vcs_mutation", "workspace_discard"},
	{regexp.MustCompile(`\bgit\s+branch\s+-D`), RiskHigh, "git_branch_delete_force", "force branch deletion", "destructive", "vcs_mutation", "branch_delete"},
	{regexp.MustCompile(`(?i)\bDROP\s+(TABLE|DATABASE)`), RiskHigh, "sql_drop", "DROP statement deletes database objects", "destructive", "database_mutation", "schema_drop"},
	{regexp.MustCompile(`(?i)\bTRUNCATE\s+TABLE`), RiskHigh, "sql_truncate", "TRUNCATE removes all table rows", "destructive", "database_mutation", "data_truncate"},
	{regexp.MustCompile(`\bkill\s+-9`), RiskHigh, "kill_sigkill", "SIGKILL can abruptly terminate processes", "destructive", "process_control", "force_terminate"},
	{regexp.MustCompile(`\bkillall\b`), RiskHigh, "killall", "killall can terminate multiple processes", "destructive", "process_control", "mass_terminate"},
	{regexp.MustCompile(`\bchmod\s+777\b`), RiskHigh, "chmod_world_writable", "chmod 777 makes files world-writable", "destructive", "permission_mutation", "permission_overexposure"},
	{regexp.MustCompile(`\bchmod\s+-R\b`), RiskHigh, "chmod_recursive", "recursive chmod can broadly alter permissions", "destructive", "permission_mutation", "recursive_permission_change"},
	{regexp.MustCompile(`\bchown\s+-R\b`), RiskHigh, "chown_recursive", "recursive chown can broadly alter ownership", "destructive", "permission_mutation", "recursive_owner_change"},
	{regexp.MustCompile(`>\s*/etc/`), RiskHigh, "etc_redirect", "redirecting output into /etc may alter system config", "destructive", "filesystem_mutation", "system_config_overwrite"},
	{regexp.MustCompile(`\bsystemctl\s+(stop|disable|mask)`), RiskHigh, "systemctl_service_disrupt", "service stop/disable can disrupt the system", "destructive", "service_control", "service_disruption"},

	// Medium — potentially dangerous
	{regexp.MustCompile(`\brm\s+`), RiskMedium, "rm", "remove command detected", "destructive", "filesystem_mutation", "delete"},
	{regexp.MustCompile(`\bgit\s+checkout\s+--`), RiskMedium, "git_checkout_discard", "checkout -- discards worktree changes", "destructive", "vcs_mutation", "workspace_discard"},
	{regexp.MustCompile(`\bgit\s+restore\s+`), RiskMedium, "git_restore", "restore can overwrite local edits", "destructive", "vcs_mutation", "workspace_overwrite"},
	{regexp.MustCompile(`\bpip3?\s+install\b`), RiskMedium, "package_install_python", "python package install executes downloaded code", "mutation", "package_management", "dependency_install"},
	{regexp.MustCompile(`\bnpm\s+install\s+-g\b`), RiskMedium, "package_install_global", "global package installation modifies system tools", "mutation", "package_management", "global_install"},
	{regexp.MustCompile(`\bcurl\s+.*\|\s*(sh|bash)\b`), RiskHigh, "curl_pipe_shell", "network content is piped directly to shell", "destructive", "remote_execution", "pipe_to_shell"},
	{regexp.MustCompile(`\bwget\s+.*\|\s*(sh|bash)\b`), RiskHigh, "wget_pipe_shell", "network content is piped directly to shell", "destructive", "remote_execution", "pipe_to_shell"},
	{regexp.MustCompile(`\bsudo\s+`), RiskMedium, "sudo", "sudo elevates privileges", "mutation", "privilege_escalation", "elevated_execution"},
	{regexp.MustCompile(`\bdocker\s+rm\b`), RiskMedium, "docker_rm", "docker rm removes containers", "destructive", "container_mutation", "container_delete"},
	{regexp.MustCompile(`\bdocker\s+system\s+prune\b`), RiskMedium, "docker_system_prune", "docker system prune removes unused resources", "destructive", "container_mutation", "resource_prune"},

	// Low — normal operations with side effects
	{regexp.MustCompile(`\bgit\s+add\b`), RiskLow, "git_add", "stages files for commit", "mutation", "vcs_mutation", "staging"},
	{regexp.MustCompile(`\bgit\s+commit\b`), RiskLow, "git_commit", "creates commit objects", "mutation", "vcs_mutation", "commit"},
	{regexp.MustCompile(`\bgit\s+push\b`), RiskLow, "git_push", "push updates remote branch", "mutation", "vcs_mutation", "push"},
	{regexp.MustCompile(`\bgit\s+merge\b`), RiskLow, "git_merge", "merge updates history", "mutation", "vcs_mutation", "merge"},
	{regexp.MustCompile(`\bgit\s+rebase\b`), RiskLow, "git_rebase", "rebase rewrites local history", "mutation", "vcs_mutation", "rebase"},
	{regexp.MustCompile(`\bgit\s+stash\b`), RiskLow, "git_stash", "stash stores worktree changes", "mutation", "vcs_mutation", "stash"},
	{regexp.MustCompile(`\bmv\s+`), RiskLow, "move", "move command changes filesystem state", "mutation", "filesystem_mutation", "move"},
	{regexp.MustCompile(`\bcp\s+`), RiskLow, "copy", "copy command writes files", "mutation", "filesystem_mutation", "copy"},
	{regexp.MustCompile(`\bmkdir\s+`), RiskLow, "mkdir", "directory creation changes filesystem state", "mutation", "filesystem_mutation", "mkdir"},
	{regexp.MustCompile(`\btouch\s+`), RiskLow, "touch", "touch creates or updates files", "mutation", "filesystem_mutation", "touch"},
	{regexp.MustCompile(`\bnpm\s+install\b`), RiskLow, "package_install_node", "npm install modifies dependencies", "mutation", "package_management", "dependency_install"},
	{regexp.MustCompile(`\byarn\s+add\b`), RiskLow, "package_add_yarn", "yarn add modifies dependencies", "mutation", "package_management", "dependency_install"},
	{regexp.MustCompile(`\bgo\s+get\b`), RiskLow, "package_add_go", "go get updates module dependencies", "mutation", "package_management", "dependency_install"},
	{regexp.MustCompile(`\bcargo\s+add\b`), RiskLow, "package_add_cargo", "cargo add updates Cargo.toml", "mutation", "package_management", "dependency_install"},
}

// readOnlyFamilies define allowlisted read-only command groups.
var readOnlyFamilies = []readOnlyFamily{
	{
		name:   "filesystem_listing",
		detail: "filesystem and metadata listing commands",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`^ls(\s|$)`),
			regexp.MustCompile(`^tree\s*`),
			regexp.MustCompile(`^find\s+`),
			regexp.MustCompile(`^file\s+`),
			regexp.MustCompile(`^pwd\s*$`),
		},
	},
	{
		name:   "text_inspection",
		detail: "text inspection and search commands",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`^cat\s+`),
			regexp.MustCompile(`^head\s+`),
			regexp.MustCompile(`^tail\s+`),
			regexp.MustCompile(`^wc\s+`),
			regexp.MustCompile(`^sort\s+`),
			regexp.MustCompile(`^uniq\s+`),
			regexp.MustCompile(`^diff\s+`),
			regexp.MustCompile(`^grep\s+`),
			regexp.MustCompile(`^rg\s+`),
			regexp.MustCompile(`^fd\s+`),
			regexp.MustCompile(`^ag\s+`),
		},
	},
	{
		name:   "shell_introspection",
		detail: "shell identity and environment inspection",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`^echo\s+`),
			regexp.MustCompile(`^printf\s+`),
			regexp.MustCompile(`^whoami\s*$`),
			regexp.MustCompile(`^hostname\s*$`),
			regexp.MustCompile(`^date\s*$`),
			regexp.MustCompile(`^uname`),
			regexp.MustCompile(`^which\s+`),
			regexp.MustCompile(`^type\s+`),
			regexp.MustCompile(`^env\s*$`),
			regexp.MustCompile(`^printenv`),
		},
	},
	{
		name:   "git_read_only",
		detail: "git inspection commands that do not mutate refs or worktree",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`^git\s+status`),
			regexp.MustCompile(`^git\s+log`),
			regexp.MustCompile(`^git\s+diff`),
			regexp.MustCompile(`^git\s+show`),
			regexp.MustCompile(`^git\s+branch(\s+-[^dD]|\s*$)`),
			regexp.MustCompile(`^git\s+tag(\s+-l|\s*$)`),
			regexp.MustCompile(`^git\s+remote\s+-v`),
			regexp.MustCompile(`^git\s+rev-parse`),
			regexp.MustCompile(`^git\s+ls-files`),
		},
	},
	{
		name:   "runtime_version_query",
		detail: "runtime and toolchain version/environment queries",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`^go\s+(version|env|list)`),
			regexp.MustCompile(`^node\s+--version`),
			regexp.MustCompile(`^python[3]?\s+--version`),
			regexp.MustCompile(`^rustc\s+--version`),
		},
	},
}

// ClassifyCommand determines the risk level of a shell command.
func ClassifyCommand(command string) RiskLevel {
	return ClassifyCommandDetailed(command).Level
}

// ClassifyCommandDetailed classifies a shell command and returns structured reasons.
func ClassifyCommandDetailed(command string) Classification {
	command = strings.TrimSpace(command)
	if command == "" {
		return Classification{Level: RiskNone, Reasons: nil}
	}

	segments, syntaxReasons := splitShellSegments(command)
	reasons := make([]RiskReason, 0, len(syntaxReasons)+len(segments))
	reasons = append(reasons, syntaxReasons...)

	for _, segment := range segments {
		reasons = append(reasons, classifySingleDetailed(segment)...)
	}
	reasons = append(reasons, detectCompositePatterns(command, segments)...)
	reasons = append(reasons, detectEscapedOperatorTricks(command)...)
	reasons = append(reasons, detectQuotedCommentDesync(command)...)
	reasons = append(reasons, detectSuspiciousSubstitutionCombos(command)...)

	if exfilReason, ok := detectNetworkExfiltration(command); ok {
		reasons = append(reasons, exfilReason)
	}

	reasons = dedupeReasons(reasons)
	level := RiskNone
	for _, reason := range reasons {
		if reason.Level > level {
			level = reason.Level
		}
	}

	return Classification{Level: level, Reasons: reasons}
}

// ClassifyAmbiguousParseDecision returns a policy hint for ambiguous shell parse cases.
// decision is one of "deny" or "ask" when ok=true.
func ClassifyAmbiguousParseDecision(classification Classification) (decision string, reason string, ok bool) {
	for _, r := range classification.Reasons {
		switch r.Code {
		case "comment_quote_desync":
			return "deny", "quoted comment pattern can desynchronize parser state", true
		case "escaped_newline_operator":
			return "deny", "escaped newline before operator obscures command structure", true
		case "escaped_operator":
			return "ask", "backslash-escaped operator can hide command structure", true
		case "suspicious_substitution_combo":
			return "ask", "mixed substitution constructs require manual review", true
		}
	}
	return "", "", false
}

// classifySingleDetailed classifies a single command segment.
func classifySingleDetailed(command string) []RiskReason {
	command = strings.TrimSpace(command)
	if command == "" {
		return nil
	}

	if reason, ok := classifyReadOnly(command); ok {
		return []RiskReason{reason}
	}

	var reasons []RiskReason

	// Check against classification rules.
	for _, rule := range classificationRules {
		if rule.pattern.MatchString(command) {
			reasons = append(reasons, RiskReason{
				Code:     rule.code,
				Detail:   rule.detail,
				Level:    rule.level,
				Snippet:  command,
				Category: rule.category,
				Family:   rule.family,
				Group:    rule.group,
			})
		}
	}

	if len(reasons) > 0 {
		return reasons
	}

	// Unknown commands default to low risk
	return []RiskReason{{
		Code:     "unclassified_command",
		Detail:   "command is not in known safe/read-only patterns",
		Level:    RiskLow,
		Snippet:  command,
		Category: "unknown",
		Family:   "unclassified",
		Group:    "fallback",
	}}
}

func classifyReadOnly(command string) (RiskReason, bool) {
	for _, family := range readOnlyFamilies {
		for _, pat := range family.patterns {
			if pat.MatchString(command) {
				return RiskReason{
					Code:     "read_only_allowlist",
					Detail:   family.detail,
					Level:    RiskNone,
					Snippet:  command,
					Category: "read_only",
					Family:   family.name,
					Group:    "allowlist_family",
				}, true
			}
		}
	}
	return RiskReason{}, false
}

// IsReadOnlyCommand returns true if the command is known to be safe/read-only.
func IsReadOnlyCommand(command string) bool {
	return ClassifyCommand(command) == RiskNone
}

func splitShellSegments(command string) ([]string, []RiskReason) {
	var (
		segments []string
		reasons  []RiskReason
		start    int
		depth    int
		sq       bool
		dq       bool
		bq       bool
		escape   bool
	)

	appendSegment := func(end int) {
		segment := strings.TrimSpace(command[start:end])
		if segment != "" {
			segments = append(segments, segment)
		}
	}

	for i := 0; i < len(command); i++ {
		ch := command[i]

		if escape {
			escape = false
			continue
		}

		if ch == '\\' {
			escape = true
			continue
		}

		if bq {
			if ch == '`' {
				bq = false
			}
			continue
		}

		if sq {
			if ch == '\'' {
				sq = false
			}
			continue
		}

		if dq {
			if ch == '"' {
				dq = false
				continue
			}
			if ch == '$' && i+1 < len(command) && command[i+1] == '(' {
				reasons = append(reasons, RiskReason{Code: "command_substitution", Detail: "command substitution executes nested shell command", Level: RiskMedium, Snippet: "$()", Category: "syntax", Family: "expansion", Group: "nested_execution"})
				depth++
				i++
			}
			continue
		}

		switch ch {
		case '\'':
			sq = true
			continue
		case '"':
			dq = true
			continue
		case '`':
			bq = true
			reasons = append(reasons, RiskReason{Code: "command_substitution", Detail: "backtick command substitution executes nested shell command", Level: RiskMedium, Snippet: "`...`", Category: "syntax", Family: "expansion", Group: "nested_execution"})
			continue
		case '$':
			if i+1 < len(command) && command[i+1] == '(' {
				reasons = append(reasons, RiskReason{Code: "command_substitution", Detail: "command substitution executes nested shell command", Level: RiskMedium, Snippet: "$()", Category: "syntax", Family: "expansion", Group: "nested_execution"})
				depth++
				i++
				continue
			}
		case '(':
			depth++
			reasons = append(reasons, RiskReason{Code: "subshell", Detail: "subshell grouping detected", Level: RiskLow, Snippet: "(...)", Category: "syntax", Family: "grouping", Group: "subshell"})
			continue
		case ')':
			if depth > 0 {
				depth--
			}
			continue
		}

		if depth > 0 {
			switch ch {
			case '|':
				reasons = append(reasons, RiskReason{Code: "pipe", Detail: "pipeline operator detected", Level: RiskLow, Snippet: "|", Category: "syntax", Family: "compound", Group: "pipeline"})
			case '<', '>':
				reasons = append(reasons, RiskReason{Code: "redirect", Detail: "shell redirection operator detected", Level: RiskLow, Snippet: string(ch), Category: "syntax", Family: "redirection", Group: "io_redirect"})
			}
			continue
		}

		if i+1 < len(command) {
			two := command[i : i+2]
			switch two {
			case "&&", "||", ";;":
				appendSegment(i)
				reasons = append(reasons, RiskReason{Code: "chained_command", Detail: "command chaining operator detected", Level: RiskLow, Snippet: two, Category: "syntax", Family: "compound", Group: "chaining"})
				start = i + 2
				i++
				continue
			case "<<", ">>", "2>":
				reasons = append(reasons, RiskReason{Code: "redirect", Detail: "shell redirection operator detected", Level: RiskLow, Snippet: two, Category: "syntax", Family: "redirection", Group: "io_redirect"})
				i++
				continue
			}
		}

		switch ch {
		case '|':
			appendSegment(i)
			reasons = append(reasons, RiskReason{Code: "pipe", Detail: "pipeline operator detected", Level: RiskLow, Snippet: "|", Category: "syntax", Family: "compound", Group: "pipeline"})
			start = i + 1
		case ';':
			appendSegment(i)
			reasons = append(reasons, RiskReason{Code: "chained_command", Detail: "command chaining operator detected", Level: RiskLow, Snippet: ";", Category: "syntax", Family: "compound", Group: "chaining"})
			start = i + 1
		case '<', '>':
			reasons = append(reasons, RiskReason{Code: "redirect", Detail: "shell redirection operator detected", Level: RiskLow, Snippet: string(ch), Category: "syntax", Family: "redirection", Group: "io_redirect"})
		}
	}

	appendSegment(len(command))
	return segments, reasons
}

func dedupeReasons(reasons []RiskReason) []RiskReason {
	if len(reasons) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(reasons))
	uniq := make([]RiskReason, 0, len(reasons))
	for _, reason := range reasons {
		key := reason.Code + "|" + reason.Snippet
		if seen[key] {
			continue
		}
		seen[key] = true
		uniq = append(uniq, reason)
	}
	sort.SliceStable(uniq, func(i, j int) bool {
		if uniq[i].Level != uniq[j].Level {
			return uniq[i].Level > uniq[j].Level
		}
		return uniq[i].Code < uniq[j].Code
	})
	return uniq
}

func detectNetworkExfiltration(command string) (RiskReason, bool) {
	lower := strings.ToLower(command)
	networkPatterns := []string{"curl", "wget", "nc ", "netcat", "scp ", "ftp ", "sftp ", "ssh "}
	sensitivePatterns := []string{"/etc/passwd", "/etc/shadow", "/etc/", ".ssh/id_rsa", ".aws/credentials", ".env", "id_ed25519"}

	hasNetwork := false
	for _, pat := range networkPatterns {
		if strings.Contains(lower, pat) {
			hasNetwork = true
			break
		}
	}
	if !hasNetwork {
		return RiskReason{}, false
	}

	hasSensitive := false
	for _, pat := range sensitivePatterns {
		if strings.Contains(lower, pat) {
			hasSensitive = true
			break
		}
	}
	if !hasSensitive {
		return RiskReason{}, false
	}

	hasTransferPrimitive := strings.Contains(lower, "|") || strings.Contains(lower, "-f @") || strings.Contains(lower, "--data") || strings.Contains(lower, "--upload-file")
	if !hasTransferPrimitive {
		return RiskReason{}, false
	}

	return RiskReason{
		Code:     "network_exfiltration",
		Detail:   "sensitive file paths combined with network transfer primitives",
		Level:    RiskCritical,
		Snippet:  command,
		Category: "destructive",
		Family:   "network_exfiltration",
		Group:    "sensitive_data_transfer",
	}, true
}

func detectCompositePatterns(command string, segments []string) []RiskReason {
	reasons := make([]RiskReason, 0, 2)
	lower := strings.ToLower(command)

	if strings.Contains(lower, "curl") && strings.Contains(lower, "|") && (strings.Contains(lower, "| sh") || strings.Contains(lower, "|sh") || strings.Contains(lower, "| bash") || strings.Contains(lower, "|bash")) {
		reasons = append(reasons, RiskReason{
			Code:     "curl_pipe_shell",
			Detail:   "network content is piped directly to shell",
			Level:    RiskHigh,
			Snippet:  command,
			Category: "destructive",
			Family:   "remote_execution",
			Group:    "pipe_to_shell",
		})
	}

	if strings.Contains(lower, "wget") && strings.Contains(lower, "|") && (strings.Contains(lower, "| sh") || strings.Contains(lower, "|sh") || strings.Contains(lower, "| bash") || strings.Contains(lower, "|bash")) {
		reasons = append(reasons, RiskReason{
			Code:     "wget_pipe_shell",
			Detail:   "network content is piped directly to shell",
			Level:    RiskHigh,
			Snippet:  command,
			Category: "destructive",
			Family:   "remote_execution",
			Group:    "pipe_to_shell",
		})
	}

	if len(segments) > 1 {
		allReadOnly := true
		for _, segment := range segments {
			if _, ok := classifyReadOnly(segment); !ok {
				allReadOnly = false
				break
			}
		}
		if !allReadOnly {
			reasons = append(reasons, RiskReason{
				Code:     "compound_command_requires_review",
				Detail:   "compound command includes non-read-only segment and should be reviewed",
				Level:    RiskMedium,
				Snippet:  command,
				Category: "syntax",
				Family:   "compound",
				Group:    "multi_segment_mixed_risk",
			})
		}
	}

	return reasons
}

func detectEscapedOperatorTricks(command string) []RiskReason {
	patterns := []struct {
		pattern *regexp.Regexp
		code    string
		detail  string
		snippet string
	}{
		{regexp.MustCompile(`\\[|;&<>]`), "escaped_operator", "escaped shell operator detected; can hide intent in compound input", `\| \; \& \< \>`},
		{regexp.MustCompile(`\\\n\s*(\|\||&&|[|;])`), "escaped_newline_operator", "escaped newline before operator detected", "\\\n&& / \\\n|"},
		{regexp.MustCompile(`\\\s+(\|\||&&|[|;])`), "escaped_whitespace_operator", "escaped whitespace before operator token detected", "\\ <ws> &&"},
	}

	var reasons []RiskReason
	for _, p := range patterns {
		if p.pattern.MatchString(command) {
			reasons = append(reasons, RiskReason{
				Code:     p.code,
				Detail:   p.detail,
				Level:    RiskMedium,
				Snippet:  p.snippet,
				Category: "syntax",
				Family:   "operator_obfuscation",
				Group:    "escaped_operator_trick",
			})
		}
	}

	return reasons
}

func detectQuotedCommentDesync(command string) []RiskReason {
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
				return []RiskReason{{
					Code:     "comment_quote_desync",
					Detail:   "quote characters inside # comment can desynchronize quote tracking",
					Level:    RiskHigh,
					Snippet:  "#' or #\"",
					Category: "syntax",
					Family:   "parser_ambiguity",
					Group:    "comment_quote_desync",
				}}
			}
			i = lineEnd
		}
	}

	return nil
}

func detectSuspiciousSubstitutionCombos(command string) []RiskReason {
	inSingle := false
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
		if ch == '\'' {
			inSingle = !inSingle
			continue
		}
		if inSingle {
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
		if (ch == '<' || ch == '>') && i+1 < len(command) && command[i+1] == '(' {
			hasProcessSub = true
		}
	}

	suspicious := (hasBacktick && hasDollarParen) ||
		(hasProcessSub && (hasDollarParen || hasBacktick)) ||
		(hasLegacyArithmetic && (hasDollarParen || hasBacktick)) ||
		(hasDollarParen && hasDollarBrace)
	if !suspicious {
		return nil
	}

	return []RiskReason{{
		Code:     "suspicious_substitution_combo",
		Detail:   "multiple substitution forms are mixed in one command",
		Level:    RiskMedium,
		Snippet:  "$(), ${}, ``, <(), >(), $[]",
		Category: "syntax",
		Family:   "parser_ambiguity",
		Group:    "substitution_combo",
	}}
}
