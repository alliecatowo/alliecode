package tools

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestValidateBashCommandDenyRules(t *testing.T) {
	tests := []struct {
		name    string
		command string
	}{
		{name: "semicolon", command: "git status; git diff"},
		{name: "or chaining", command: "go test ./... || true"},
		{name: "single ampersand", command: "sleep 1 &"},
		{name: "newline", command: "git status\ngit diff"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := validateBashCommand(tc.command)
			if msg == "" {
				t.Fatalf("expected deny message for %q", tc.command)
			}
		})
	}
}

func TestAnalyzeBashCommandClassifiesRiskyConstructs(t *testing.T) {
	tests := []struct {
		name    string
		command string
		risk    bashRiskKind
	}{
		{name: "subshell", command: "(echo hi)", risk: bashRiskSubshell},
		{name: "command substitution", command: "echo $(whoami)", risk: bashRiskCommandSubstitution},
		{name: "heredoc", command: "cat <<EOF\nhello\nEOF", risk: bashRiskHeredoc},
		{name: "redirect", command: "echo hi > out.txt", risk: bashRiskRedirects},
		{name: "chained", command: "echo hi && echo there", risk: bashRiskChainedOperators},
		{name: "unbalanced quote", command: "echo 'oops", risk: bashRiskUnbalancedQuotes},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			analysis := analyzeBashCommand(tc.command)
			if !hasRisk(analysis.Risks, tc.risk) {
				t.Fatalf("expected risk %q for command %q; got %+v", tc.risk, tc.command, analysis.Risks)
			}
		})
	}
}

func TestEvaluateBashPreflightPolicyDecisions(t *testing.T) {
	workspace := filepath.Clean("/workspace/repo")

	tests := []struct {
		name     string
		command  string
		policy   *bashSandboxPolicyInput
		want     string
		wantRule string
	}{
		{
			name:     "deny chained operators",
			command:  "git status; git diff",
			policy:   nil,
			want:     "deny",
			wantRule: "semicolon",
		},
		{
			name:     "deny quoted comment desync pattern",
			command:  "echo ok # ' \"\nwhoami",
			policy:   nil,
			want:     "deny",
			wantRule: "comment_quote_desync",
		},
		{
			name:     "deny escaped newline before operator",
			command:  "echo ok \\\n&& whoami",
			policy:   nil,
			want:     "deny",
			wantRule: "escaped_newline_operator_obfuscation",
		},
		{
			name:     "ask escaped operator obfuscation",
			command:  `echo ok \| wc -l`,
			policy:   nil,
			want:     "ask",
			wantRule: "escaped_operator_obfuscation",
		},
		{
			name:     "ask suspicious substitution combo",
			command:  "echo $(whoami)`id`",
			policy:   nil,
			want:     "ask",
			wantRule: "suspicious_substitution_combo",
		},
		{
			name:    "ask risky constructs by default policy",
			command: "echo $(date)",
			policy:  nil,
			want:    "ask",
		},
		{
			name:    "deny risky constructs when policy denies",
			command: "echo $(date)",
			policy: &bashSandboxPolicyInput{
				DenyRiskyConstructs:     true,
				RequireApprovalForRisky: false,
			},
			want:     "deny",
			wantRule: "sandbox_deny_risky_constructs",
		},
		{
			name:    "ask outside workspace redirection",
			command: "echo hi > /tmp/out.txt",
			policy:  nil,
			want:    "ask",
		},
		{
			name:    "allow outside workspace by policy",
			command: "echo hi > /tmp/out.txt",
			policy: &bashSandboxPolicyInput{
				AllowOutsideWorkspaceWrites: true,
				RequireApprovalForRisky:     false,
			},
			want: "allow",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			decision := evaluateBashPreflight(tc.command, workspace, tc.policy)
			if decision.Behavior != tc.want {
				t.Fatalf("expected behavior %q, got %q (rule=%q)", tc.want, decision.Behavior, decision.RuleID)
			}
			if tc.wantRule != "" && decision.RuleID != tc.wantRule {
				t.Fatalf("expected rule %q, got %q", tc.wantRule, decision.RuleID)
			}
		})
	}
}

func TestFindOutsideWorkspaceWritePaths(t *testing.T) {
	workspace := filepath.Clean("/workspace/repo")
	analysis := analyzeBashCommand("echo hi > ./ok.txt && echo no > /etc/passwd && echo dyn > $TMP/out")
	paths := findOutsideWorkspaceWritePaths(analysis, workspace)
	if len(paths) == 0 {
		t.Fatalf("expected outside workspace paths")
	}
	joined := strings.Join(paths, ",")
	if !strings.Contains(joined, "/etc/passwd") {
		t.Fatalf("expected /etc/passwd in outside paths; got %v", paths)
	}
	if !strings.Contains(joined, "<dynamic:$TMP/out>") {
		t.Fatalf("expected dynamic marker for $TMP/out; got %v", paths)
	}
}

func TestRenderPreflightResultMessageMachineReadable(t *testing.T) {
	decision := bashPolicyDecision{
		Behavior: "ask",
		Reason:   "sandbox policy requires approval for risky shell constructs",
		RuleID:   "sandbox_ask_risky_constructs",
		Analysis: bashPreflightAnalysis{
			Risks: []bashRisk{{Kind: bashRiskCommandSubstitution, Message: "contains $()"}},
		},
	}
	msg := renderPreflightResultMessage(decision)
	if !strings.Contains(msg, "[bash_preflight]") {
		t.Fatalf("expected preflight structured block, got %q", msg)
	}
	if !strings.Contains(msg, "decision: ask") {
		t.Fatalf("expected decision field, got %q", msg)
	}
	if !strings.Contains(msg, "rule_id: sandbox_ask_risky_constructs") {
		t.Fatalf("expected rule_id field, got %q", msg)
	}
}

func TestValidateBashCommandAllowsSafeDependentChaining(t *testing.T) {
	if msg := validateBashCommand("git add . && git commit -m \"x\""); msg != "" {
		t.Fatalf("expected no deny message, got %q", msg)
	}
}

func TestBashToolCheckPermissionsDeniedForGuardrails(t *testing.T) {
	in, _ := json.Marshal(bashInput{Command: "git status; git diff"})
	perm := (&BashTool{}).CheckPermissions(in, types.ToolContext{IsNonInteractive: true})
	if perm != types.PermissionDenied {
		t.Fatalf("expected PermissionDenied, got %v", perm)
	}
}

func TestBashToolCheckPermissionsAsksForRiskyPreflight(t *testing.T) {
	in, _ := json.Marshal(bashInput{Command: "echo $(date)"})
	perm := (&BashTool{}).CheckPermissions(in, types.ToolContext{WorkingDir: "/workspace/repo", IsNonInteractive: true})
	if perm != types.PermissionAsk {
		t.Fatalf("expected PermissionAsk, got %v", perm)
	}
}

func TestBashToolExecuteDeniedMessageForDestructiveCommand(t *testing.T) {
	in, _ := json.Marshal(bashInput{Command: "rm -rf /tmp/nope"})
	res, err := (&BashTool{}).Execute(context.Background(), in, types.ToolContext{})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected error result")
	}
	if res.Content == "" || res.Content[:14] != "Command denied" {
		t.Fatalf("expected clear deny message, got %q", res.Content)
	}
	if !strings.Contains(res.Content, "[bash_deny]") {
		t.Fatalf("expected bash_deny block, got %q", res.Content)
	}
	if !strings.Contains(res.Content, "deny_type: destructive") {
		t.Fatalf("expected destructive deny type, got %q", res.Content)
	}
	if !strings.Contains(res.Content, "taxonomy_version: 1") {
		t.Fatalf("expected taxonomy version, got %q", res.Content)
	}
	if !strings.Contains(res.Content, "primary_rule_id: rm_force") {
		t.Fatalf("expected primary_rule_id rm_force, got %q", res.Content)
	}
}

func TestBashToolExecuteDeniedMessageForChainingIncludesDeterministicRiskBlock(t *testing.T) {
	in, _ := json.Marshal(bashInput{Command: "git status; git diff"})
	res, err := (&BashTool{}).Execute(context.Background(), in, types.ToolContext{})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected error result")
	}
	if !strings.Contains(res.Content, "Command denied") {
		t.Fatalf("expected deny prefix, got %q", res.Content)
	}
	if !strings.Contains(res.Content, "[bash_preflight]") {
		t.Fatalf("expected bash_preflight block, got %q", res.Content)
	}
	if !strings.Contains(res.Content, "decision: deny") {
		t.Fatalf("expected deny decision, got %q", res.Content)
	}
	if !strings.Contains(res.Content, "rule_id: semicolon") {
		t.Fatalf("expected semicolon rule id, got %q", res.Content)
	}
	if !strings.Contains(res.Content, "risk_kinds: chained_operators") {
		t.Fatalf("expected chained_operators risk kind, got %q", res.Content)
	}
}
