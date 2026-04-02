package tools

import (
	"errors"
	"testing"

	"github.com/alliecatowo/alliecode/internal/mcp"
)

func TestNewMCPToolErrorCarriesCategory(t *testing.T) {
	err := errors.New("needs auth token")
	out := newMCPToolError("MCP_AUTH_FAILED", err)
	if out.Category != mcp.ErrorCategoryNeedsAuthentication {
		t.Fatalf("unexpected category: %+v", out)
	}
}
