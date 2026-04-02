package tools

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTightenPermissionForHighRiskShellAmbiguousPatterns(t *testing.T) {
	input := []byte(`{"command":"echo hi \\| cat"}`)
	got := tightenPermissionForHighRisk("bash", input, types.ToolContext{}, types.PermissionAllowed)
	if got != types.PermissionAsk {
		t.Fatalf("expected ask for escaped operator ambiguity, got %v", got)
	}
}

func TestTightenPermissionForHighRiskDotfileTargets(t *testing.T) {
	input := []byte(`{"file_path":"/home/user/.gitconfig","content":"x"}`)
	got := tightenPermissionForHighRisk("write", input, types.ToolContext{IsNonInteractive: true}, types.PermissionAllowed)
	if got != types.PermissionDenied {
		t.Fatalf("expected deny for non-interactive dotfile write, got %v", got)
	}
}
