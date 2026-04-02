package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

type RemoteTriggerTool struct{}

type remoteTriggerInput struct {
	Action    string          `json:"action"`
	TriggerID string          `json:"trigger_id,omitempty"`
	Body      json.RawMessage `json:"body,omitempty"`
}

type remoteTriggerOutput struct {
	Success   bool                  `json:"success"`
	Action    string                `json:"action"`
	Status    int                   `json:"status"`
	Trigger   *remoteTriggerRecord  `json:"trigger,omitempty"`
	Triggers  []remoteTriggerRecord `json:"triggers,omitempty"`
	Message   string                `json:"message,omitempty"`
	ErrorText string                `json:"error,omitempty"`
}

func (t *RemoteTriggerTool) Name() string { return "remote_trigger" }

func (t *RemoteTriggerTool) Description() string {
	return "Manages provider-agnostic remote triggers using deterministic local state."
}

func (t *RemoteTriggerTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"action":     {Type: "string", Description: "list, get, create, update, or run", Enum: []string{"list", "get", "create", "update", "run"}},
			"trigger_id": {Type: "string", Description: "Trigger identifier for get/update/run."},
			"body":       {Type: "string", Description: "JSON payload for create/update."},
		},
		Required: []string{"action"},
	}
}

func (t *RemoteTriggerTool) Execute(_ context.Context, input types.ToolInput, _ types.ToolContext) (types.ToolResult, error) {
	var in remoteTriggerInput
	if err := json.Unmarshal(input, &in); err != nil {
		return remoteTriggerResult(remoteTriggerOutput{Success: false, Action: "error", Status: 400, ErrorText: fmt.Sprintf("invalid input: %v", err)}, true)
	}
	action := strings.ToLower(strings.TrimSpace(in.Action))
	id := strings.TrimSpace(in.TriggerID)

	orchestrationState.mu.Lock()
	defer orchestrationState.mu.Unlock()

	switch action {
	case "list":
		triggers := make([]remoteTriggerRecord, 0, len(orchestrationState.triggers))
		for _, rec := range orchestrationState.triggers {
			triggers = append(triggers, rec)
		}
		sort.Slice(triggers, func(i, j int) bool { return triggers[i].ID < triggers[j].ID })
		return remoteTriggerResult(remoteTriggerOutput{Success: true, Action: "list", Status: 200, Triggers: triggers}, false)
	case "get":
		if id == "" {
			return remoteTriggerResult(remoteTriggerOutput{Success: false, Action: "get", Status: 400, ErrorText: "trigger_id is required"}, true)
		}
		rec, ok := orchestrationState.triggers[id]
		if !ok {
			return remoteTriggerResult(remoteTriggerOutput{Success: false, Action: "get", Status: 404, ErrorText: "trigger not found"}, true)
		}
		copyRec := rec
		return remoteTriggerResult(remoteTriggerOutput{Success: true, Action: "get", Status: 200, Trigger: &copyRec}, false)
	case "create":
		if len(in.Body) == 0 {
			return remoteTriggerResult(remoteTriggerOutput{Success: false, Action: "create", Status: 400, ErrorText: "body is required"}, true)
		}
		orchestrationState.triggerSeq++
		newID := fmt.Sprintf("trigger-%04d", orchestrationState.triggerSeq)
		now := time.Now().UTC().Format(time.RFC3339)
		rec := remoteTriggerRecord{ID: newID, Body: normalizeRemoteBody(in.Body), CreatedAt: now, UpdatedAt: now}
		orchestrationState.triggers[newID] = rec
		copyRec := rec
		return remoteTriggerResult(remoteTriggerOutput{Success: true, Action: "create", Status: 201, Trigger: &copyRec}, false)
	case "update":
		if id == "" {
			return remoteTriggerResult(remoteTriggerOutput{Success: false, Action: "update", Status: 400, ErrorText: "trigger_id is required"}, true)
		}
		if len(in.Body) == 0 {
			return remoteTriggerResult(remoteTriggerOutput{Success: false, Action: "update", Status: 400, ErrorText: "body is required"}, true)
		}
		rec, ok := orchestrationState.triggers[id]
		if !ok {
			return remoteTriggerResult(remoteTriggerOutput{Success: false, Action: "update", Status: 404, ErrorText: "trigger not found"}, true)
		}
		rec.Body = normalizeRemoteBody(in.Body)
		rec.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
		orchestrationState.triggers[id] = rec
		copyRec := rec
		return remoteTriggerResult(remoteTriggerOutput{Success: true, Action: "update", Status: 200, Trigger: &copyRec}, false)
	case "run":
		if id == "" {
			return remoteTriggerResult(remoteTriggerOutput{Success: false, Action: "run", Status: 400, ErrorText: "trigger_id is required"}, true)
		}
		rec, ok := orchestrationState.triggers[id]
		if !ok {
			return remoteTriggerResult(remoteTriggerOutput{Success: false, Action: "run", Status: 404, ErrorText: "trigger not found"}, true)
		}
		rec.RunCount++
		rec.LastRunAt = time.Now().UTC().Format(time.RFC3339)
		rec.UpdatedAt = rec.LastRunAt
		orchestrationState.triggers[id] = rec
		copyRec := rec
		return remoteTriggerResult(remoteTriggerOutput{Success: true, Action: "run", Status: 200, Trigger: &copyRec, Message: "trigger run recorded"}, false)
	default:
		return remoteTriggerResult(remoteTriggerOutput{Success: false, Action: action, Status: 400, ErrorText: "action must be one of: list, get, create, update, run"}, true)
	}
}

func normalizeRemoteBody(raw json.RawMessage) string {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return string(raw)
	}
	return string(b)
}

func remoteTriggerResult(out remoteTriggerOutput, isError bool) (types.ToolResult, error) {
	b, err := json.Marshal(out)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}
	return types.ToolResult{Content: string(b), IsError: isError}, nil
}

func (t *RemoteTriggerTool) IsReadOnly(input types.ToolInput) bool {
	var in remoteTriggerInput
	if err := json.Unmarshal(input, &in); err != nil {
		return false
	}
	action := strings.ToLower(strings.TrimSpace(in.Action))
	return action == "list" || action == "get"
}

func (t *RemoteTriggerTool) IsDestructive(input types.ToolInput) bool { return false }

func (t *RemoteTriggerTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *RemoteTriggerTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return types.PermissionAllowed
}
