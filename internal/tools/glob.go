package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/alliecatowo/alliecode/internal/types"
)

// GlobTool finds files matching glob patterns.
type GlobTool struct{}

type globInput struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path"`
}

const (
	globMaxOutputLines = 500
	globMaxOutputChars = 100000
)

func (t *GlobTool) Name() string { return "Glob" }

func (t *GlobTool) Description() string {
	return "Finds files matching a glob pattern. Supports ** for recursive matching. Returns paths sorted by modification time (most recent first)."
}

func (t *GlobTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"pattern": {
				Type:        "string",
				Description: "The glob pattern to match (e.g. \"**/*.go\", \"src/**/*.ts\").",
			},
			"path": {
				Type:        "string",
				Description: "The directory to search in. Defaults to the working directory.",
			},
		},
		Required: []string{"pattern"},
	}
}

// fileWithTime pairs a path with its modification time for sorting.
type fileWithTime struct {
	path    string
	modTime int64
}

func (t *GlobTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	var in globInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Invalid input: %v", err), IsError: true}, nil
	}

	if in.Pattern == "" {
		return types.ToolResult{Content: "pattern is required", IsError: true}, nil
	}

	baseDir := in.Path
	if baseDir == "" {
		baseDir = toolCtx.WorkingDir
	} else if !filepath.IsAbs(baseDir) {
		baseDir = filepath.Join(toolCtx.WorkingDir, baseDir)
	}

	pattern := in.Pattern

	// Handle ** patterns via filepath.Walk.
	if strings.Contains(pattern, "**") {
		return t.doubleStarGlob(ctx, baseDir, pattern)
	}

	// Simple glob without **.
	fullPattern := filepath.Join(baseDir, pattern)
	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Invalid glob pattern: %v", err), IsError: true}, nil
	}

	return t.sortAndFormat(matches)
}

func (t *GlobTool) doubleStarGlob(ctx context.Context, baseDir, pattern string) (types.ToolResult, error) {
	// Split pattern on ** to get prefix and suffix.
	parts := strings.SplitN(pattern, "**", 2)
	prefix := parts[0]
	suffix := ""
	if len(parts) > 1 {
		suffix = parts[1]
		// Remove leading separator from suffix.
		suffix = strings.TrimPrefix(suffix, "/")
		suffix = strings.TrimPrefix(suffix, string(filepath.Separator))
	}

	searchDir := baseDir
	if prefix != "" {
		searchDir = filepath.Join(baseDir, prefix)
	}

	var matches []string

	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible paths
		}

		// Check context cancellation periodically.
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Skip hidden directories.
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		if suffix == "" {
			matches = append(matches, path)
			return nil
		}

		// Match the suffix pattern against the filename or relative path.
		matched, _ := filepath.Match(suffix, info.Name())
		if matched {
			matches = append(matches, path)
		}

		return nil
	})

	if err != nil && err != context.Canceled {
		return types.ToolResult{Content: fmt.Sprintf("Error walking directory: %v", err), IsError: true}, nil
	}

	return t.sortAndFormat(matches)
}

func (t *GlobTool) sortAndFormat(paths []string) (types.ToolResult, error) {
	if len(paths) == 0 {
		return types.ToolResult{Content: "No files found\n\n" + renderStructuredBlock("glob_result", []structuredField{{Key: "count", Value: "0"}})}, nil
	}

	// Get modification times for sorting.
	files := make([]fileWithTime, 0, len(paths))
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		files = append(files, fileWithTime{path: p, modTime: info.ModTime().UnixNano()})
	}

	// Sort by modification time (most recent first), then path (ascending) for deterministic ties.
	sort.Slice(files, func(i, j int) bool {
		if files[i].modTime == files[j].modTime {
			return files[i].path < files[j].path
		}
		return files[i].modTime > files[j].modTime
	})

	var sb strings.Builder
	for i, f := range files {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(f.path)
	}

	formatted := limitGlobOutput(sb.String(), globMaxOutputLines, globMaxOutputChars)
	return types.ToolResult{Content: formatted + "\n\n" + renderStructuredBlock("glob_result", []structuredField{{Key: "count", Value: strconv.Itoa(len(files))}})}, nil
}

func limitGlobOutput(output string, maxLines, maxChars int) string {
	if !isGrepOutputTruncated(output, maxLines, maxChars) {
		return output
	}

	truncated := truncateCommandOutput(output, maxLines, maxChars)
	return truncated + "\n\n" + renderStructuredBlock("truncation_metadata", []structuredField{
		{Key: "truncated", Value: "true"},
		{Key: "reason", Value: "exceeded_output_limits"},
		{Key: "original_bytes", Value: strconv.Itoa(len(output))},
		{Key: "returned_bytes", Value: strconv.Itoa(len(truncated))},
		{Key: "line_limit", Value: strconv.Itoa(maxLines)},
		{Key: "char_limit", Value: strconv.Itoa(maxChars)},
	})
}

func (t *GlobTool) IsReadOnly(input types.ToolInput) bool {
	return true
}

func (t *GlobTool) IsDestructive(input types.ToolInput) bool {
	return false
}

func (t *GlobTool) IsConcurrencySafe(input types.ToolInput) bool {
	return true
}

func (t *GlobTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyFile, "glob", input, toolCtx, types.PermissionAllowed)
}
