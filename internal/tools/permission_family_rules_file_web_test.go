package tools

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestPermissionPolicyRulesIncludeMutationAskForFileFamily(t *testing.T) {
	rules := permissionPolicyRules(toolFamilyFile, "write", types.ToolContext{IsNonInteractive: true})
	if len(rules) == 0 {
		t.Fatalf("expected policy rule for non-interactive file mutation")
	}
}
