package types

import (
	"context"
	"encoding/json"
)

// ToolInput is the raw JSON input to a tool.
type ToolInput = json.RawMessage

// ToolResult is returned by a tool after execution.
type ToolResult struct {
	Content     string    `json:"content"`
	IsError     bool      `json:"is_error,omitempty"`
	NewMessages []Message `json:"new_messages,omitempty"`
}

// ToolPermission represents the result of a permission check.
type ToolPermission int

const (
	PermissionAllowed ToolPermission = iota
	PermissionDenied
	PermissionAsk
)

// ToolContext carries runtime state available to tools during execution.
type ToolContext struct {
	WorkingDir       string
	AbortChan        <-chan struct{}
	Messages         []Message
	IsNonInteractive bool
	Debug            bool
}

// ToolSchema describes the JSON Schema for a tool's input.
type ToolSchema struct {
	Type       string                    `json:"type"`
	Properties map[string]PropertySchema `json:"properties,omitempty"`
	Required   []string                  `json:"required,omitempty"`
}

// PropertySchema describes a single property in a tool's input schema.
type PropertySchema struct {
	Type        string          `json:"type"`
	Description string          `json:"description,omitempty"`
	Enum        []string        `json:"enum,omitempty"`
	Default     any             `json:"default,omitempty"`
	Items       *PropertySchema `json:"items,omitempty"`
	Minimum     *float64        `json:"minimum,omitempty"`
	Maximum     *float64        `json:"maximum,omitempty"`
}

// Tool defines the interface that all AllieCode tools must implement.
type Tool interface {
	// Name returns the tool's unique identifier.
	Name() string

	// Description returns a human-readable description of what the tool does.
	// May vary based on input context.
	Description() string

	// InputSchema returns the JSON Schema for the tool's expected input.
	InputSchema() ToolSchema

	// Execute runs the tool with the given input and context.
	Execute(ctx context.Context, input ToolInput, toolCtx ToolContext) (ToolResult, error)

	// IsReadOnly returns true if the tool does not modify state.
	IsReadOnly(input ToolInput) bool

	// IsDestructive returns true if the tool may cause irreversible changes.
	IsDestructive(input ToolInput) bool

	// IsConcurrencySafe returns true if multiple instances can run in parallel.
	IsConcurrencySafe(input ToolInput) bool

	// CheckPermissions evaluates whether the tool call should be allowed.
	CheckPermissions(input ToolInput, toolCtx ToolContext) ToolPermission
}

// ToolDef is a serializable tool definition sent to providers.
type ToolDef struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	InputSchema ToolSchema `json:"input_schema"`
}

// ToToolDef converts a Tool to a serializable ToolDef.
func ToToolDef(t Tool) ToolDef {
	return ToolDef{
		Name:        t.Name(),
		Description: t.Description(),
		InputSchema: t.InputSchema(),
	}
}

// ToToolDefs converts a slice of Tools to serializable ToolDefs.
func ToToolDefs(tools []Tool) []ToolDef {
	defs := make([]ToolDef, len(tools))
	for i, t := range tools {
		defs[i] = ToToolDef(t)
	}
	return defs
}
