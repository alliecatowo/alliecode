package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

// FileEditTool performs string replacement edits in files.
type FileEditTool struct{}

type fileEditInput struct {
	FilePath   string `json:"file_path"`
	OldString  string `json:"old_string"`
	NewString  string `json:"new_string"`
	ReplaceAll bool   `json:"replace_all"`
}

func (t *FileEditTool) Name() string { return "Edit" }

func (t *FileEditTool) Description() string {
	return "Performs exact string replacements in files. The old_string must be unique in the file unless replace_all is true."
}

func (t *FileEditTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"file_path": {
				Type:        "string",
				Description: "The absolute path to the file to edit.",
			},
			"old_string": {
				Type:        "string",
				Description: "The exact string to find and replace.",
			},
			"new_string": {
				Type:        "string",
				Description: "The replacement string.",
			},
			"replace_all": {
				Type:        "boolean",
				Description: "If true, replace all occurrences. Default is false.",
			},
		},
		Required: []string{"file_path", "old_string", "new_string"},
	}
}

func (t *FileEditTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	started := time.Now()
	var in fileEditInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Invalid input: %v", err), IsError: true}, nil
	}

	if in.FilePath == "" {
		return types.ToolResult{Content: "file_path is required", IsError: true}, nil
	}
	if in.OldString == in.NewString {
		return types.ToolResult{Content: "old_string and new_string must be different", IsError: true}, nil
	}

	path := in.FilePath
	if !filepath.IsAbs(path) {
		path = filepath.Join(toolCtx.WorkingDir, path)
	}

	if strings.HasSuffix(strings.ToLower(path), ".ipynb") {
		return types.ToolResult{Content: "File is a Jupyter Notebook. Use the notebook edit flow instead.", IsError: true}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return types.ToolResult{Content: fmt.Sprintf("Error reading file: %v", err), IsError: true}, nil
	}

	if os.IsNotExist(err) {
		if in.OldString != "" {
			return types.ToolResult{Content: fmt.Sprintf("File does not exist: %s", path), IsError: true}, nil
		}
		dir := filepath.Dir(path)
		if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
			return types.ToolResult{Content: fmt.Sprintf("Error creating directories: %v", mkErr), IsError: true}, nil
		}
		if writeErr := os.WriteFile(path, []byte(in.NewString), 0o644); writeErr != nil {
			return types.ToolResult{Content: fmt.Sprintf("Error writing file: %v", writeErr), IsError: true}, nil
		}
		return types.ToolResult{
			Content: fmt.Sprintf(
				"Successfully replaced %d occurrence(s) in %s\n\n%s",
				1,
				path,
				renderStructuredBlock("file_edit_result", []structuredField{
					{Key: "path", Value: path},
					{Key: "replacements", Value: "1"},
					{Key: "replace_all", Value: strconv.FormatBool(in.ReplaceAll)},
					{Key: "old_length", Value: strconv.Itoa(len(in.OldString))},
					{Key: "new_length", Value: strconv.Itoa(len(in.NewString))},
					{Key: "created", Value: "true"},
					{Key: "duration_ms", Value: strconv.FormatInt(time.Since(started).Milliseconds(), 10)},
					{Key: "workspace_dir", Value: toolCtx.WorkingDir},
				}),
			),
		}, nil
	}

	stat, err := os.Stat(path)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Error reading file metadata: %v", err), IsError: true}, nil
	}
	if guardErr := validateReadBeforeModify(toolCtx.Messages, path, stat.ModTime()); guardErr != nil {
		return types.ToolResult{Content: guardErr.Error(), IsError: true}, nil
	}

	content := string(data)
	if in.OldString == "" {
		if strings.TrimSpace(content) != "" {
			return types.ToolResult{Content: "Cannot create new file - file already exists.", IsError: true}, nil
		}
		if err := os.WriteFile(path, []byte(in.NewString), 0o644); err != nil {
			return types.ToolResult{Content: fmt.Sprintf("Error writing file: %v", err), IsError: true}, nil
		}
		return types.ToolResult{
			Content: fmt.Sprintf(
				"Successfully replaced %d occurrence(s) in %s\n\n%s",
				1,
				path,
				renderStructuredBlock("file_edit_result", []structuredField{
					{Key: "path", Value: path},
					{Key: "replacements", Value: "1"},
					{Key: "replace_all", Value: strconv.FormatBool(in.ReplaceAll)},
					{Key: "old_length", Value: strconv.Itoa(len(in.OldString))},
					{Key: "new_length", Value: strconv.Itoa(len(in.NewString))},
					{Key: "created", Value: "false"},
					{Key: "duration_ms", Value: strconv.FormatInt(time.Since(started).Milliseconds(), 10)},
					{Key: "workspace_dir", Value: toolCtx.WorkingDir},
				}),
			),
		}, nil
	}
	count := strings.Count(content, in.OldString)

	if count == 0 {
		return types.ToolResult{
			Content: fmt.Sprintf("old_string not found in %s", path),
			IsError: true,
		}, nil
	}

	if count > 1 && !in.ReplaceAll {
		return types.ToolResult{
			Content: fmt.Sprintf("old_string found %d times in %s. Use replace_all=true to replace all occurrences, or provide a more specific string.", count, path),
			IsError: true,
		}, nil
	}

	var newContent string
	if in.ReplaceAll {
		newContent = strings.ReplaceAll(content, in.OldString, in.NewString)
	} else {
		newContent = strings.Replace(content, in.OldString, in.NewString, 1)
	}

	if err := os.WriteFile(path, []byte(newContent), 0o644); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Error writing file: %v", err), IsError: true}, nil
	}

	replacements := 1
	if in.ReplaceAll {
		replacements = count
	}

	return types.ToolResult{
		Content: fmt.Sprintf(
			"Successfully replaced %d occurrence(s) in %s\n\n%s",
			replacements,
			path,
			renderStructuredBlock("file_edit_result", []structuredField{
				{Key: "path", Value: path},
				{Key: "replacements", Value: strconv.Itoa(replacements)},
				{Key: "replace_all", Value: strconv.FormatBool(in.ReplaceAll)},
				{Key: "old_length", Value: strconv.Itoa(len(in.OldString))},
				{Key: "new_length", Value: strconv.Itoa(len(in.NewString))},
				{Key: "duration_ms", Value: strconv.FormatInt(time.Since(started).Milliseconds(), 10)},
				{Key: "workspace_dir", Value: toolCtx.WorkingDir},
			}),
		),
	}, nil
}

func (t *FileEditTool) IsReadOnly(input types.ToolInput) bool {
	return false
}

func (t *FileEditTool) IsDestructive(input types.ToolInput) bool {
	return false
}

func (t *FileEditTool) IsConcurrencySafe(input types.ToolInput) bool {
	return false
}

func (t *FileEditTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyFile, "edit", input, toolCtx, types.PermissionAsk)
}
