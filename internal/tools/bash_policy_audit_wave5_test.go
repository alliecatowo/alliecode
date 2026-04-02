package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestBashExecutionIncludesRiskAuditFields(t *testing.T) {
	in, _ := json.Marshal(bashInput{Command: "pwd"})
	res, err := (&BashTool{}).Execute(context.Background(), in, types.ToolContext{WorkingDir: t.TempDir()})
	if err != nil || res.IsError {
		t.Fatalf("bash failed: err=%v content=%q", err, res.Content)
	}
	for _, needle := range []string{"risk_level:", "risk_codes:", "policy_decision:", "policy_rule_id:"} {
		if !strings.Contains(res.Content, needle) {
			t.Fatalf("expected %q in shell execution metadata: %s", needle, res.Content)
		}
	}
}
