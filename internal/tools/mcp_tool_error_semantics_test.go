package tools

import (
	"errors"
	"testing"
)

func TestNewMCPToolErrorIncludesHintAndCode(t *testing.T) {
	out := newMCPToolError("MCP_TOOL_INVOKE_FAILED", errors.New("service unavailable"))
	if out.ErrorCode == "" || out.Hint == "" || !out.Retryable {
		t.Fatalf("expected enriched error semantics: %+v", out)
	}
}
