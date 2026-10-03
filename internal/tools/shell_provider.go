package tools

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

type shellProvider interface {
	Type() string
	NormalizeCommand(command string) string
	BuildCommand(command string) (string, []string)
	NormalizeEnv(env []string, toolCtx types.ToolContext) []string
}

type bashProvider struct{}

func (p bashProvider) Type() string { return "bash" }

func (p bashProvider) NormalizeCommand(command string) string {
	return strings.TrimSpace(command)
}

func (p bashProvider) BuildCommand(command string) (string, []string) {
	return "sh", []string{"-c", command}
}

func (p bashProvider) NormalizeEnv(env []string, _ types.ToolContext) []string {
	return env
}

type powerShellProvider struct{}

func (p powerShellProvider) Type() string { return "powershell" }

func (p powerShellProvider) NormalizeCommand(command string) string {
	trimmed := strings.TrimSpace(command)
	return strings.ReplaceAll(trimmed, "\r\n", "\n")
}

func (p powerShellProvider) BuildCommand(command string) (string, []string) {
	return "sh", []string{"-c", command}
}

func (p powerShellProvider) NormalizeEnv(env []string, _ types.ToolContext) []string {
	if hasEnvKey(env, "POWERSHELL_TELEMETRY_OPTOUT") {
		return env
	}
	out := append([]string{}, env...)
	out = append(out, "POWERSHELL_TELEMETRY_OPTOUT=1")
	return out
}

func selectShellProvider(toolName string) shellProvider {
	switch strings.ToLower(strings.TrimSpace(toolName)) {
	case "powershell":
		return powerShellProvider{}
	default:
		return bashProvider{}
	}
}

func executeWithShellProvider(ctx context.Context, toolCtx types.ToolContext, provider shellProvider, command string, timeout time.Duration) (types.ToolResult, error) {
	normalizedCommand := provider.NormalizeCommand(command)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	binary, args := provider.BuildCommand(normalizedCommand)
	cmd := exec.CommandContext(ctx, binary, args...)
	workingDir := normalizeToolWorkingDir(toolCtx.WorkingDir)
	cmd.Dir = workingDir
	toolCtx.WorkingDir = workingDir
	cmd.Env = provider.NormalizeEnv(cmd.Environ(), toolCtx)

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	output := buf.String()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return types.ToolResult{Content: formatCommandTimeout(timeout, output), IsError: true}, nil
		}
		return types.ToolResult{Content: formatCommandExit(output, err), IsError: true}, nil
	}

	return types.ToolResult{Content: output}, nil
}

func runShellPreflight(command, workspaceRoot string, policyInput *bashSandboxPolicyInput, requireApproval bool) string {
	decision := evaluateBashPreflight(command, workspaceRoot, policyInput)
	if decision.Behavior == "deny" {
		return renderPreflightResultMessage(decision)
	}
	if decision.Behavior == "ask" && requireApproval {
		return renderPreflightResultMessage(decision)
	}
	if denyMsg := denyDestructiveCommand(command); denyMsg != "" {
		return denyMsg
	}
	return ""
}

func hasEnvKey(env []string, key string) bool {
	prefix := key + "="
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			return true
		}
	}
	return false
}

func normalizeToolWorkingDir(raw string) string {
	workingDir := strings.TrimSpace(raw)
	if workingDir == "" {
		return "."
	}
	return workingDir
}
