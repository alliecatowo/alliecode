package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alliecatowo/alliecode/internal/types"
)

type lsInput struct {
	Path string `json:"path"`
}

// LSTool lists files and directories.
type LSTool struct{}

func (t *LSTool) Name() string { return "ls" }

func (t *LSTool) Description() string {
	return "Lists files and directories in a path. Returns one entry per line, with trailing / for directories."
}

func (t *LSTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"path": {
				Type:        "string",
				Description: "Directory to list. Defaults to the working directory.",
			},
		},
	}
}

func (t *LSTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	var in lsInput
	if len(input) > 0 {
		if err := json.Unmarshal(input, &in); err != nil {
			return types.ToolResult{Content: fmt.Sprintf("Invalid input: %v", err), IsError: true}, nil
		}
	}

	target := in.Path
	if target == "" {
		target = toolCtx.WorkingDir
	} else if !filepath.IsAbs(target) {
		target = filepath.Join(toolCtx.WorkingDir, target)
	}

	entries, err := os.ReadDir(target)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Error reading directory: %v", err), IsError: true}, nil
	}

	if len(entries) == 0 {
		return types.ToolResult{Content: "(empty directory)"}, nil
	}

	var sb strings.Builder
	for i, entry := range entries {
		if i > 0 {
			sb.WriteString("\n")
		}
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		sb.WriteString(name)
	}

	return types.ToolResult{Content: sb.String()}, nil
}

func (t *LSTool) IsReadOnly(input types.ToolInput) bool { return true }

func (t *LSTool) IsDestructive(input types.ToolInput) bool { return false }

func (t *LSTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *LSTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return types.PermissionAllowed
}
