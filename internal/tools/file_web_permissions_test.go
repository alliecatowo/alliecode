package tools

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestFileAndWebPermissionFamilies(t *testing.T) {
	ro := []struct {
		tool string
		fam  toolFamily
	}{
		{tool: "read", fam: toolFamilyFile},
		{tool: "glob", fam: toolFamilyFile},
		{tool: "grep", fam: toolFamilyFile},
		{tool: "webfetch", fam: toolFamilyWeb},
		{tool: "websearch", fam: toolFamilyWeb},
	}
	for _, tc := range ro {
		got := evaluateToolPermissionByFamily(tc.fam, tc.tool, nil, types.ToolContext{IsNonInteractive: true}, types.PermissionAsk)
		if got != types.PermissionAllowed {
			t.Fatalf("expected allowed for %s, got %v", tc.tool, got)
		}
	}
}
