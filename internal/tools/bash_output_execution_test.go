package tools

import (
	"strings"
	"testing"
	"time"
)

func TestFormatShellExecutionBlock(t *testing.T) {
	block := formatShellExecutionBlock("Bash", "bash", "/tmp", 2*time.Second, 10*time.Millisecond, "echo hi", "ok", false, &shellExecutionAudit{RiskLevel: "low", RiskCodes: []string{"git_commit"}, PolicyDecision: "allow", PolicyRuleID: "allow", PolicyReason: "preflight checks passed"})
	for _, needle := range []string{"[shell_execution]", "tool: bash", "timeout_ms: 2000", "duration_ms:", "output_bytes:", "workspace_root: /tmp", "path_scope: workspace"} {
		if !strings.Contains(block, needle) {
			t.Fatalf("missing %q in block: %s", needle, block)
		}
	}
	for _, needle := range []string{"risk_level: low", "risk_codes: git_commit", "policy_decision: allow", "policy_rule_id: allow"} {
		if !strings.Contains(block, needle) {
			t.Fatalf("missing %q in block: %s", needle, block)
		}
	}
}
