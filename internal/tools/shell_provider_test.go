package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestSelectShellProvider(t *testing.T) {
	if got := selectShellProvider("Bash").Type(); got != "bash" {
		t.Fatalf("expected bash provider, got %q", got)
	}
	if got := selectShellProvider("powershell").Type(); got != "powershell" {
		t.Fatalf("expected powershell provider, got %q", got)
	}
	if got := selectShellProvider("unknown").Type(); got != "bash" {
		t.Fatalf("expected unknown provider to default to bash, got %q", got)
	}
}

func TestPowerShellProviderNormalization(t *testing.T) {
	p := powerShellProvider{}
	normalized := p.NormalizeCommand("  echo one\r\necho two  ")
	if strings.Contains(normalized, "\r\n") {
		t.Fatalf("expected CRLF normalization, got %q", normalized)
	}
	if normalized != "echo one\necho two" {
		t.Fatalf("unexpected normalized command: %q", normalized)
	}

	env := p.NormalizeEnv([]string{"PATH=/tmp/bin"}, types.ToolContext{})
	if !hasEnvKey(env, "POWERSHELL_TELEMETRY_OPTOUT") {
		t.Fatalf("expected POWERSHELL_TELEMETRY_OPTOUT to be injected")
	}

	existing := p.NormalizeEnv([]string{"POWERSHELL_TELEMETRY_OPTOUT=0"}, types.ToolContext{})
	count := 0
	for _, entry := range existing {
		if strings.HasPrefix(entry, "POWERSHELL_TELEMETRY_OPTOUT=") {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected no duplicate POWERSHELL_TELEMETRY_OPTOUT, got %d", count)
	}
}

func TestExecuteWithShellProviderUsesProviderExecutionPath(t *testing.T) {
	t.Run("bash provider", func(t *testing.T) {
		res, err := executeWithShellProvider(
			context.Background(),
			types.ToolContext{WorkingDir: t.TempDir()},
			selectShellProvider("Bash"),
			"printf 'ok'",
			2*time.Second,
		)
		if err != nil {
			t.Fatalf("executeWithShellProvider returned error: %v", err)
		}
		if res.IsError {
			t.Fatalf("expected success, got error: %s", res.Content)
		}
		if strings.TrimSpace(res.Content) != "ok" {
			t.Fatalf("unexpected output: %q", res.Content)
		}
	})

	t.Run("powershell provider env normalization", func(t *testing.T) {
		res, err := executeWithShellProvider(
			context.Background(),
			types.ToolContext{WorkingDir: t.TempDir()},
			selectShellProvider("powershell"),
			`printf '%s' "$POWERSHELL_TELEMETRY_OPTOUT"`,
			2*time.Second,
		)
		if err != nil {
			t.Fatalf("executeWithShellProvider returned error: %v", err)
		}
		if res.IsError {
			t.Fatalf("expected success, got error: %s", res.Content)
		}
		if strings.TrimSpace(res.Content) != "1" {
			t.Fatalf("expected env override to be visible, got %q", res.Content)
		}
	})
}

func TestPowerShellToolExecutesThroughSharedProvider(t *testing.T) {
	tool := &PowerShellTool{}
	in, _ := json.Marshal(powerShellInput{Command: `printf '%s' "$POWERSHELL_TELEMETRY_OPTOUT"`})
	res, err := tool.Execute(context.Background(), in, types.ToolContext{WorkingDir: t.TempDir()})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected successful result, got error: %s", res.Content)
	}
	if !strings.Contains(res.Content, "1") {
		t.Fatalf("expected shared provider env override, got %q", res.Content)
	}
	if !strings.Contains(res.Content, "[shell_execution]") {
		t.Fatalf("expected shell execution metadata block, got %q", res.Content)
	}
}

func TestRunShellPreflightRequireApprovalBehavior(t *testing.T) {
	cmd := "echo $(date)"
	if msg := runShellPreflight(cmd, "/workspace/repo", nil, true); msg == "" {
		t.Fatalf("expected approval-required preflight message")
	}
	if msg := runShellPreflight(cmd, "/workspace/repo", nil, false); msg != "" {
		t.Fatalf("expected approval-only risk to pass when approval not required, got %q", msg)
	}
}
