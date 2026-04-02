package tools

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTightenPermissionForHighRiskWritePath(t *testing.T) {
	input := []byte(`{"file_path":"/etc/passwd","content":"x"}`)
	got := tightenPermissionForHighRisk("write", input, types.ToolContext{IsNonInteractive: true}, types.PermissionAllowed)
	if got != types.PermissionDenied {
		t.Fatalf("expected deny for non-interactive sensitive path, got %v", got)
	}
}

func TestTightenPermissionForHighRiskWriteContent(t *testing.T) {
	input := []byte(`{"file_path":"./a.txt","content":"-----BEGIN PRIVATE KEY-----"}`)
	got := tightenPermissionForHighRisk("write", input, types.ToolContext{}, types.PermissionAllowed)
	if got != types.PermissionAsk {
		t.Fatalf("expected ask for sensitive content, got %v", got)
	}
}

func TestTightenPermissionForHighRiskShell(t *testing.T) {
	input := []byte(`{"command":"rm -rf /"}`)
	got := tightenPermissionForHighRisk("bash", input, types.ToolContext{}, types.PermissionAllowed)
	if got != types.PermissionDenied {
		t.Fatalf("expected deny for critical shell command, got %v", got)
	}
}
