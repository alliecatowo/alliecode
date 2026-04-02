package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/alliecatowo/alliecode/internal/types"
)

const memoryFileName = ".alliecode-memory.json"

var memoryMu sync.Mutex

type memoryInput struct {
	Action string `json:"action"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

// MemoryTool stores simple key-value notes persisted per workspace.
type MemoryTool struct{}

func (t *MemoryTool) Name() string { return "memory" }

func (t *MemoryTool) Description() string {
	return "Persists simple key-value notes scoped to the current workspace."
}

func (t *MemoryTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"action": {
				Type:        "string",
				Description: "Operation to perform: set, get, delete, or list.",
				Enum:        []string{"set", "get", "delete", "list"},
			},
			"key": {
				Type:        "string",
				Description: "Key for set/get/delete actions.",
			},
			"value": {
				Type:        "string",
				Description: "Value to store for set action.",
			},
		},
		Required: []string{"action"},
	}
}

func (t *MemoryTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	var in memoryInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Invalid input: %v", err), IsError: true}, nil
	}

	workspace := toolCtx.WorkingDir
	if workspace == "" {
		workspace = "."
	}
	path := filepath.Join(workspace, memoryFileName)

	memoryMu.Lock()
	defer memoryMu.Unlock()

	store, err := loadMemoryStore(path)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Error loading memory: %v", err), IsError: true}, nil
	}

	switch in.Action {
	case "set":
		if strings.TrimSpace(in.Key) == "" {
			return types.ToolResult{Content: "key is required for action=set", IsError: true}, nil
		}
		store[in.Key] = in.Value
		if err := saveMemoryStore(path, store); err != nil {
			return types.ToolResult{Content: fmt.Sprintf("Error saving memory: %v", err), IsError: true}, nil
		}
		return types.ToolResult{Content: fmt.Sprintf("Stored key %q", in.Key)}, nil
	case "get":
		if strings.TrimSpace(in.Key) == "" {
			return types.ToolResult{Content: "key is required for action=get", IsError: true}, nil
		}
		v, ok := store[in.Key]
		if !ok {
			return types.ToolResult{Content: fmt.Sprintf("Key %q not found", in.Key), IsError: true}, nil
		}
		return types.ToolResult{Content: v}, nil
	case "delete":
		if strings.TrimSpace(in.Key) == "" {
			return types.ToolResult{Content: "key is required for action=delete", IsError: true}, nil
		}
		delete(store, in.Key)
		if err := saveMemoryStore(path, store); err != nil {
			return types.ToolResult{Content: fmt.Sprintf("Error saving memory: %v", err), IsError: true}, nil
		}
		return types.ToolResult{Content: fmt.Sprintf("Deleted key %q", in.Key)}, nil
	case "list":
		if len(store) == 0 {
			return types.ToolResult{Content: "(no memory entries)"}, nil
		}
		keys := make([]string, 0, len(store))
		for k := range store {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		lines := make([]string, 0, len(keys))
		for _, k := range keys {
			lines = append(lines, fmt.Sprintf("%s=%s", k, store[k]))
		}
		return types.ToolResult{Content: strings.Join(lines, "\n")}, nil
	default:
		return types.ToolResult{Content: "action must be one of: set, get, delete, list", IsError: true}, nil
	}
}

func loadMemoryStore(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return map[string]string{}, nil
	}
	store := map[string]string{}
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	return store, nil
}

func saveMemoryStore(path string, store map[string]string) error {
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func (t *MemoryTool) IsReadOnly(input types.ToolInput) bool {
	var in memoryInput
	if err := json.Unmarshal(input, &in); err != nil {
		return false
	}
	return in.Action == "get" || in.Action == "list"
}

func (t *MemoryTool) IsDestructive(input types.ToolInput) bool {
	var in memoryInput
	if err := json.Unmarshal(input, &in); err != nil {
		return false
	}
	return in.Action == "delete"
}

func (t *MemoryTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *MemoryTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return types.PermissionAllowed
}
