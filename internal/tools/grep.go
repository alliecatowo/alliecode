package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alliecatowo/alliecode/internal/types"
)

// GrepTool searches file contents using ripgrep or Go's regexp.
type GrepTool struct{}

type grepInput struct {
	Pattern    string `json:"pattern"`
	Path       string `json:"path"`
	Glob       string `json:"glob"`
	Include    string `json:"include"`
	OutputMode string `json:"output_mode"`
	Context    int    `json:"context"`
}

const (
	grepMaxOutputLines = 200
	grepMaxOutputChars = 100000
)

func (t *GrepTool) Name() string { return "Grep" }

func (t *GrepTool) Description() string {
	return "Searches file contents using regular expressions. Uses ripgrep (rg) if available, otherwise falls back to Go's regexp."
}

func (t *GrepTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"pattern": {
				Type:        "string",
				Description: "The regular expression pattern to search for.",
			},
			"path": {
				Type:        "string",
				Description: "File or directory to search in. Defaults to the working directory.",
			},
			"glob": {
				Type:        "string",
				Description: "Glob pattern to filter files (e.g. \"*.go\", \"*.{ts,tsx}\").",
			},
			"include": {
				Type:        "string",
				Description: "File pattern to include in the search (e.g. \"*.js\", \"*.{ts,tsx}\"). Alias of glob.",
			},
			"output_mode": {
				Type:        "string",
				Description: "Output mode: \"content\" shows matching lines, \"files_with_matches\" shows file paths, \"count\" shows match counts.",
				Enum:        []string{"content", "files_with_matches", "count"},
			},
			"context": {
				Type:        "integer",
				Description: "Number of context lines to show before and after each match. Only used with output_mode \"content\".",
			},
		},
		Required: []string{"pattern"},
	}
}

func (t *GrepTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	var in grepInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Invalid input: %v", err), IsError: true}, nil
	}

	if in.Pattern == "" {
		return types.ToolResult{Content: "pattern is required", IsError: true}, nil
	}

	if _, err := regexp.Compile(in.Pattern); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Invalid regex pattern: %v", err), IsError: true}, nil
	}

	outputMode := in.OutputMode
	if outputMode == "" {
		outputMode = "files_with_matches"
	}

	searchPath := in.Path
	if searchPath == "" {
		searchPath = toolCtx.WorkingDir
	} else if !filepath.IsAbs(searchPath) {
		searchPath = filepath.Join(toolCtx.WorkingDir, searchPath)
	}

	// Try ripgrep first.
	rgPath, err := exec.LookPath("rg")
	if err == nil {
		return t.executeWithRipgrep(ctx, rgPath, in, outputMode, searchPath)
	}

	// Fallback to Go regexp.
	return t.executeWithGoRegexp(in, outputMode, searchPath)
}

func (t *GrepTool) executeWithRipgrep(ctx context.Context, rgPath string, in grepInput, outputMode, searchPath string) (types.ToolResult, error) {
	args := buildRipgrepArgs(in, outputMode, searchPath)

	cmd := exec.CommandContext(ctx, rgPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()

	if err != nil {
		// rg returns exit code 1 when no matches found.
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return types.ToolResult{Content: "No matches found.\n\n" + renderStructuredBlock("grep_result", []structuredField{{Key: "matches_found", Value: "false"}})}, nil
		}
		if stderr.Len() > 0 {
			errText := strings.TrimSpace(stderr.String())
			if strings.Contains(strings.ToLower(errText), "regex") {
				return types.ToolResult{Content: fmt.Sprintf("Invalid regex pattern: %s", errText), IsError: true}, nil
			}
			return types.ToolResult{Content: fmt.Sprintf("Error: %s", errText), IsError: true}, nil
		}
		return types.ToolResult{Content: fmt.Sprintf("Error: %v", err), IsError: true}, nil
	}

	if output == "" {
		return types.ToolResult{Content: "No matches found.\n\n" + renderStructuredBlock("grep_result", []structuredField{{Key: "matches_found", Value: "false"}})}, nil
	}

	formatted := strings.TrimRight(output, "\n")
	formatted = limitGrepOutput(formatted, grepMaxOutputLines, grepMaxOutputChars)
	return types.ToolResult{Content: formatted}, nil
}

func (t *GrepTool) executeWithGoRegexp(in grepInput, outputMode, searchPath string) (types.ToolResult, error) {
	re, err := regexp.Compile(in.Pattern)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Invalid regex pattern: %v", err), IsError: true}, nil
	}

	// For now, return a message suggesting ripgrep installation.
	_ = re
	return types.ToolResult{
		Content: "ripgrep (rg) is not installed. Install it for best search performance. Falling back to basic search is not yet fully implemented.",
		IsError: true,
	}, nil
}

func buildRipgrepArgs(in grepInput, outputMode, searchPath string) []string {
	args := []string{"--sort", "path"}

	switch outputMode {
	case "files_with_matches":
		args = append(args, "--files-with-matches")
	case "count":
		args = append(args, "--count")
	case "content":
		args = append(args, "-n")
		if in.Context > 0 {
			args = append(args, fmt.Sprintf("-C%d", in.Context))
		}
	}

	include := strings.TrimSpace(in.Include)
	if include == "" {
		include = strings.TrimSpace(in.Glob)
	}
	if include != "" {
		args = append(args, "--glob", include)
	}

	args = append(args, in.Pattern, searchPath)
	return args
}

func limitGrepOutput(output string, maxLines, maxChars int) string {
	if !isGrepOutputTruncated(output, maxLines, maxChars) {
		return truncateCommandOutput(output, maxLines, maxChars)
	}

	truncated := truncateCommandOutput(output, maxLines, maxChars)
	commonFields := []structuredField{
		{Key: "truncated", Value: "true"},
		{Key: "reason", Value: "exceeded_output_limits"},
		{Key: "original_bytes", Value: fmt.Sprintf("%d", len(output))},
		{Key: "returned_bytes", Value: fmt.Sprintf("%d", len(truncated))},
		{Key: "line_limit", Value: fmt.Sprintf("%d", maxLines)},
		{Key: "char_limit", Value: fmt.Sprintf("%d", maxChars)},
	}

	artifactPath, err := writeGrepOverflowArtifact(output)
	if err != nil {
		return truncated + "\n\n" + renderStructuredBlock("truncation_metadata", append(commonFields,
			structuredField{Key: "artifact_available", Value: "false"},
			structuredField{Key: "artifact_write_error", Value: err.Error()},
		))
	}

	return truncated + "\n\n" + renderStructuredBlock("truncation_metadata", append(commonFields,
		structuredField{Key: "artifact_available", Value: "true"},
		structuredField{Key: "artifact_path", Value: artifactPath},
	))
}

func isGrepOutputTruncated(output string, maxLines, maxChars int) bool {
	if maxLines > 0 {
		if len(strings.Split(output, "\n")) > maxLines {
			return true
		}
	}
	return maxChars > 0 && len(output) > maxChars
}

func writeGrepOverflowArtifact(output string) (string, error) {
	f, err := os.CreateTemp("", "alliecode-grep-overflow-*.txt")
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := f.WriteString(output); err != nil {
		return "", err
	}
	return f.Name(), nil
}

func (t *GrepTool) IsReadOnly(input types.ToolInput) bool {
	return true
}

func (t *GrepTool) IsDestructive(input types.ToolInput) bool {
	return false
}

func (t *GrepTool) IsConcurrencySafe(input types.ToolInput) bool {
	return true
}

func (t *GrepTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyFile, "grep", input, toolCtx, types.PermissionAllowed)
}
