package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

// FileWriteTool writes or creates files.
type FileWriteTool struct{}

type fileWriteInput struct {
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
}

func (t *FileWriteTool) Name() string { return "Write" }

func (t *FileWriteTool) Description() string {
	return "Writes content to a file, creating it and any parent directories if they do not exist. Overwrites existing files."
}

func (t *FileWriteTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"file_path": {
				Type:        "string",
				Description: "The absolute path to the file to write.",
			},
			"content": {
				Type:        "string",
				Description: "The content to write to the file.",
			},
		},
		Required: []string{"file_path", "content"},
	}
}

func (t *FileWriteTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	started := time.Now()
	var in fileWriteInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Invalid input: %v", err), IsError: true}, nil
	}

	if in.FilePath == "" {
		return types.ToolResult{Content: "file_path is required", IsError: true}, nil
	}

	path := in.FilePath
	if !filepath.IsAbs(path) {
		path = filepath.Join(toolCtx.WorkingDir, path)
	}

	// Create parent directories if needed.
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Error creating directories: %v", err), IsError: true}, nil
	}

	existedBefore := false
	previousSize := int64(0)
	if stat, err := os.Stat(path); err == nil {
		existedBefore = true
		previousSize = stat.Size()
		if guardErr := validateReadBeforeModify(toolCtx.Messages, path, stat.ModTime()); guardErr != nil {
			return types.ToolResult{Content: guardErr.Error(), IsError: true}, nil
		}
	} else if !os.IsNotExist(err) {
		return types.ToolResult{Content: fmt.Sprintf("Error reading file metadata: %v", err), IsError: true}, nil
	}

	if err := os.WriteFile(path, []byte(in.Content), 0o644); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Error writing file: %v", err), IsError: true}, nil
	}

	block := renderStructuredBlock("file_write_result", []structuredField{
		{Key: "path", Value: path},
		{Key: "write_type", Value: map[bool]string{true: "update", false: "create"}[existedBefore]},
		{Key: "bytes_written", Value: strconv.Itoa(len(in.Content))},
		{Key: "content_chars", Value: strconv.Itoa(len([]rune(in.Content)))},
		{Key: "existed_before", Value: strconv.FormatBool(existedBefore)},
		{Key: "previous_size", Value: strconv.FormatInt(previousSize, 10)},
		{Key: "read_guard_applied", Value: strconv.FormatBool(existedBefore)},
		{Key: "duration_ms", Value: strconv.FormatInt(time.Since(started).Milliseconds(), 10)},
		{Key: "workspace_dir", Value: toolCtx.WorkingDir},
	})
	return types.ToolResult{Content: fmt.Sprintf("Successfully wrote %d bytes to %s\n\n%s", len(in.Content), path, block)}, nil
}

func (t *FileWriteTool) IsReadOnly(input types.ToolInput) bool {
	return false
}

func (t *FileWriteTool) IsDestructive(input types.ToolInput) bool {
	var in fileWriteInput
	if err := json.Unmarshal(input, &in); err != nil {
		return true
	}
	// Destructive if the file already exists (overwrite).
	path := in.FilePath
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func (t *FileWriteTool) IsConcurrencySafe(input types.ToolInput) bool {
	return false
}

func (t *FileWriteTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyFile, "write", input, toolCtx, types.PermissionAsk)
}
