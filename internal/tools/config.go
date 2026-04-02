package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alliecatowo/alliecode/internal/types"
)

type ConfigTool struct{}

type configInput struct {
	Setting   string          `json:"setting"`
	Value     json.RawMessage `json:"value,omitempty"`
	Operation string          `json:"operation,omitempty"`
}

type configPair struct {
	Setting string `json:"setting"`
	Value   any    `json:"value"`
}

type configOutput struct {
	Success      bool         `json:"success"`
	Operation    string       `json:"operation"`
	Setting      string       `json:"setting,omitempty"`
	Value        any          `json:"value,omitempty"`
	Previous     any          `json:"previous_value,omitempty"`
	Entries      []configPair `json:"entries,omitempty"`
	StoragePath  string       `json:"storage_path"`
	ErrorMessage string       `json:"error,omitempty"`
}

func (t *ConfigTool) Name() string { return "config" }

func (t *ConfigTool) Description() string {
	return "Gets and sets deterministic local configuration values for this workspace."
}

func (t *ConfigTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"setting":   {Type: "string", Description: "Setting key."},
			"value":     {Type: "string", Description: "JSON value for set operation."},
			"operation": {Type: "string", Description: "get, set, unset, or list", Enum: []string{"get", "set", "unset", "list"}},
		},
	}
}

func (t *ConfigTool) Execute(_ context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	var in configInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("invalid input: %v", err), IsError: true}, nil
	}

	storagePath := filepath.Join(toolCtx.WorkingDir, ".alliecode-config.json")
	store, err := readConfigStore(storagePath)
	if err != nil {
		return configResult(configOutput{Success: false, Operation: "error", StoragePath: storagePath, ErrorMessage: err.Error()}, true)
	}

	op := strings.ToLower(strings.TrimSpace(in.Operation))
	if op == "" {
		if len(in.Value) > 0 {
			op = "set"
		} else if strings.TrimSpace(in.Setting) == "" {
			op = "list"
		} else {
			op = "get"
		}
	}

	setting := strings.TrimSpace(in.Setting)
	switch op {
	case "list":
		keys := make([]string, 0, len(store))
		for k := range store {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		entries := make([]configPair, 0, len(keys))
		for _, k := range keys {
			entries = append(entries, configPair{Setting: k, Value: decodeJSONRaw(store[k])})
		}
		return configResult(configOutput{Success: true, Operation: "list", Entries: entries, StoragePath: storagePath}, false)
	case "get":
		if setting == "" {
			return configResult(configOutput{Success: false, Operation: "get", StoragePath: storagePath, ErrorMessage: "setting is required for get"}, true)
		}
		raw, ok := store[setting]
		if !ok {
			return configResult(configOutput{Success: false, Operation: "get", Setting: setting, StoragePath: storagePath, ErrorMessage: "setting not found"}, true)
		}
		return configResult(configOutput{Success: true, Operation: "get", Setting: setting, Value: decodeJSONRaw(raw), StoragePath: storagePath}, false)
	case "set":
		if setting == "" {
			return configResult(configOutput{Success: false, Operation: "set", StoragePath: storagePath, ErrorMessage: "setting is required for set"}, true)
		}
		if len(in.Value) == 0 {
			return configResult(configOutput{Success: false, Operation: "set", Setting: setting, StoragePath: storagePath, ErrorMessage: "value is required for set"}, true)
		}
		prev := decodeJSONRaw(store[setting])
		store[setting] = normalizeJSONRaw(in.Value)
		if err := writeConfigStore(storagePath, store); err != nil {
			return configResult(configOutput{Success: false, Operation: "set", Setting: setting, StoragePath: storagePath, ErrorMessage: err.Error()}, true)
		}
		return configResult(configOutput{Success: true, Operation: "set", Setting: setting, Value: decodeJSONRaw(store[setting]), Previous: prev, StoragePath: storagePath}, false)
	case "unset":
		if setting == "" {
			return configResult(configOutput{Success: false, Operation: "unset", StoragePath: storagePath, ErrorMessage: "setting is required for unset"}, true)
		}
		prevRaw, ok := store[setting]
		if !ok {
			return configResult(configOutput{Success: false, Operation: "unset", Setting: setting, StoragePath: storagePath, ErrorMessage: "setting not found"}, true)
		}
		delete(store, setting)
		if err := writeConfigStore(storagePath, store); err != nil {
			return configResult(configOutput{Success: false, Operation: "unset", Setting: setting, StoragePath: storagePath, ErrorMessage: err.Error()}, true)
		}
		return configResult(configOutput{Success: true, Operation: "unset", Setting: setting, Previous: decodeJSONRaw(prevRaw), StoragePath: storagePath}, false)
	default:
		return configResult(configOutput{Success: false, Operation: op, StoragePath: storagePath, ErrorMessage: "operation must be one of: get, set, unset, list"}, true)
	}
}

func readConfigStore(path string) (map[string]json.RawMessage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]json.RawMessage{}, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return map[string]json.RawMessage{}, nil
	}
	out := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func writeConfigStore(path string, store map[string]json.RawMessage) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o644)
}

func decodeJSONRaw(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw)
	}
	return v
}

func normalizeJSONRaw(raw json.RawMessage) json.RawMessage {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return json.RawMessage(strconvQuote(string(raw)))
	}
	b, err := json.Marshal(v)
	if err != nil {
		return raw
	}
	return b
}

func strconvQuote(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		return "\"\""
	}
	return string(b)
}

func configResult(out configOutput, isError bool) (types.ToolResult, error) {
	b, err := json.Marshal(out)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}
	return types.ToolResult{Content: string(b), IsError: isError}, nil
}

func (t *ConfigTool) IsReadOnly(input types.ToolInput) bool {
	var in configInput
	if err := json.Unmarshal(input, &in); err != nil {
		return false
	}
	op := strings.ToLower(strings.TrimSpace(in.Operation))
	if op == "" {
		return len(in.Value) == 0
	}
	return op == "get" || op == "list"
}

func (t *ConfigTool) IsDestructive(input types.ToolInput) bool {
	var in configInput
	if err := json.Unmarshal(input, &in); err != nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(in.Operation), "unset")
}

func (t *ConfigTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *ConfigTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return types.PermissionAllowed
}
