package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alliecatowo/alliecode/internal/types"
)

// NotebookEditTool edits Jupyter notebook cells.
type NotebookEditTool struct{}

type notebookEditInput struct {
	NotebookPath string `json:"notebook_path"`
	Action       string `json:"action"`
	CellIndex    int    `json:"cell_index"`
	CellType     string `json:"cell_type"`
	NewSource    string `json:"new_source"`
}

// notebookFile represents the top-level structure of a .ipynb file.
type notebookFile struct {
	Cells         []notebookCell         `json:"cells"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	NBFormat      int                    `json:"nbformat"`
	NBFormatMinor int                    `json:"nbformat_minor"`
}

// notebookCell represents a single cell in a Jupyter notebook.
type notebookCell struct {
	CellType       string                 `json:"cell_type"`
	Source         []string               `json:"source"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Outputs        []interface{}          `json:"outputs,omitempty"`
	ExecutionCount *int                   `json:"execution_count,omitempty"`
}

func (t *NotebookEditTool) Name() string { return "NotebookEdit" }

func (t *NotebookEditTool) Description() string {
	return "Edits a cell in a Jupyter notebook (.ipynb file). Modifies the source of the specified cell."
}

func (t *NotebookEditTool) InputSchema() types.ToolSchema {
	zero := float64(0)
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"notebook_path": {
				Type:        "string",
				Description: "Path to the .ipynb notebook file.",
			},
			"action": {
				Type:        "string",
				Description: "Cell operation: update (default), add, or delete.",
				Enum:        []string{"update", "add", "delete"},
			},
			"cell_index": {
				Type:        "integer",
				Description: "The 0-based index of the target cell.",
				Minimum:     &zero,
			},
			"cell_type": {
				Type:        "string",
				Description: "Cell type for add action: code (default) or markdown.",
				Enum:        []string{"code", "markdown"},
			},
			"new_source": {
				Type:        "string",
				Description: "The new source content for update/add actions.",
			},
		},
		Required: []string{"notebook_path", "cell_index"},
	}
}

func (t *NotebookEditTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	var in notebookEditInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Invalid input: %v", err), IsError: true}, nil
	}

	if in.NotebookPath == "" {
		return types.ToolResult{Content: "notebook_path is required", IsError: true}, nil
	}

	path := in.NotebookPath
	if !filepath.IsAbs(path) {
		path = filepath.Join(toolCtx.WorkingDir, path)
	}

	// Read the notebook file.
	data, err := os.ReadFile(path)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Error reading notebook: %v", err), IsError: true}, nil
	}

	var nb notebookFile
	if err := json.Unmarshal(data, &nb); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Error parsing notebook JSON: %v", err), IsError: true}, nil
	}

	action := in.Action
	if action == "" {
		action = "update"
	}

	if err := applyNotebookCellAction(&nb, in, action); err != nil {
		return types.ToolResult{Content: err.Error(), IsError: true}, nil
	}

	// Write back.
	output, err := json.MarshalIndent(nb, "", " ")
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Error marshaling notebook: %v", err), IsError: true}, nil
	}
	// Ensure trailing newline.
	output = append(output, '\n')

	if err := os.WriteFile(path, output, 0o644); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Error writing notebook: %v", err), IsError: true}, nil
	}

	message := fmt.Sprintf("Successfully %sed cell %d in %s", action, in.CellIndex, path)
	if action == "add" {
		message = fmt.Sprintf("Successfully added cell at index %d in %s", in.CellIndex, path)
	}
	if action == "delete" {
		message = fmt.Sprintf("Successfully deleted cell %d in %s", in.CellIndex, path)
	}

	return types.ToolResult{
		Content: message,
	}, nil
}

func applyNotebookCellAction(nb *notebookFile, in notebookEditInput, action string) error {
	if in.CellIndex < 0 {
		return fmt.Errorf("cell_index %d is out of range (must be >= 0)", in.CellIndex)
	}

	switch action {
	case "update":
		if in.NewSource == "" {
			return fmt.Errorf("new_source is required for update action")
		}
		if in.CellIndex >= len(nb.Cells) {
			return fmt.Errorf("cell_index %d is out of range (notebook has %d cells)", in.CellIndex, len(nb.Cells))
		}
		nb.Cells[in.CellIndex].Source = sourceToLines(in.NewSource)
		return nil
	case "add":
		if in.NewSource == "" {
			return fmt.Errorf("new_source is required for add action")
		}
		if in.CellIndex > len(nb.Cells) {
			return fmt.Errorf("cell_index %d is out of range for add (notebook has %d cells)", in.CellIndex, len(nb.Cells))
		}
		cellType := in.CellType
		if cellType == "" {
			cellType = "code"
		}
		if cellType != "code" && cellType != "markdown" {
			return fmt.Errorf("cell_type must be one of: code, markdown")
		}
		newCell := notebookCell{CellType: cellType, Source: sourceToLines(in.NewSource), Metadata: map[string]interface{}{}}
		if cellType == "code" {
			newCell.Outputs = []interface{}{}
		}
		nb.Cells = append(nb.Cells, notebookCell{})
		copy(nb.Cells[in.CellIndex+1:], nb.Cells[in.CellIndex:])
		nb.Cells[in.CellIndex] = newCell
		return nil
	case "delete":
		if in.CellIndex >= len(nb.Cells) {
			return fmt.Errorf("cell_index %d is out of range (notebook has %d cells)", in.CellIndex, len(nb.Cells))
		}
		nb.Cells = append(nb.Cells[:in.CellIndex], nb.Cells[in.CellIndex+1:]...)
		return nil
	default:
		return fmt.Errorf("action must be one of: update, add, delete")
	}
}

// sourceToLines splits source text into the line format used by .ipynb files.
// Each line except the last ends with \n.
func sourceToLines(source string) []string {
	if source == "" {
		return []string{}
	}

	var lines []string
	start := 0
	for i := 0; i < len(source); i++ {
		if source[i] == '\n' {
			lines = append(lines, source[start:i+1])
			start = i + 1
		}
	}
	// Add remaining content (last line without trailing newline).
	if start < len(source) {
		lines = append(lines, source[start:])
	}

	return lines
}

func (t *NotebookEditTool) IsReadOnly(input types.ToolInput) bool {
	return false
}

func (t *NotebookEditTool) IsDestructive(input types.ToolInput) bool {
	return false
}

func (t *NotebookEditTool) IsConcurrencySafe(input types.ToolInput) bool {
	return false
}

func (t *NotebookEditTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return types.PermissionAsk
}
