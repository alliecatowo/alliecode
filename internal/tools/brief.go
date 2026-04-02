package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

type BriefTool struct{}

type briefInput struct {
	Message     string   `json:"message"`
	Attachments []string `json:"attachments,omitempty"`
	Status      string   `json:"status"`
}

type briefAttachment struct {
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	IsImage bool   `json:"is_image"`
}

type briefOutput struct {
	Success         bool              `json:"success"`
	Message         string            `json:"message"`
	Status          string            `json:"status"`
	SentAt          string            `json:"sent_at"`
	AttachmentCount int               `json:"attachment_count"`
	Attachments     []briefAttachment `json:"attachments,omitempty"`
	ErrorMessage    string            `json:"error,omitempty"`
}

func (t *BriefTool) Name() string { return "brief" }

func (t *BriefTool) Description() string {
	return "Delivers user-facing messages with optional attachment metadata."
}

func (t *BriefTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"message":     {Type: "string", Description: "Message to deliver to the user."},
			"attachments": {Type: "array", Description: "Optional attachment paths.", Items: &types.PropertySchema{Type: "string"}},
			"status":      {Type: "string", Description: "normal or proactive", Enum: []string{"normal", "proactive"}},
		},
		Required: []string{"message", "status"},
	}
}

func (t *BriefTool) Execute(_ context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	var in briefInput
	if err := json.Unmarshal(input, &in); err != nil {
		return briefResult(briefOutput{Success: false, ErrorMessage: fmt.Sprintf("invalid input: %v", err)}, true)
	}
	message := strings.TrimSpace(in.Message)
	if message == "" {
		return briefResult(briefOutput{Success: false, ErrorMessage: "message is required"}, true)
	}
	status := strings.ToLower(strings.TrimSpace(in.Status))
	if status != "normal" && status != "proactive" {
		return briefResult(briefOutput{Success: false, ErrorMessage: "status must be normal or proactive"}, true)
	}

	attachments := make([]briefAttachment, 0, len(in.Attachments))
	for _, raw := range in.Attachments {
		resolved := raw
		if !filepath.IsAbs(resolved) {
			resolved = filepath.Join(toolCtx.WorkingDir, raw)
		}
		info, err := os.Stat(resolved)
		if err != nil {
			return briefResult(briefOutput{Success: false, ErrorMessage: fmt.Sprintf("attachment not found: %s", raw)}, true)
		}
		if info.IsDir() {
			return briefResult(briefOutput{Success: false, ErrorMessage: fmt.Sprintf("attachment is a directory: %s", raw)}, true)
		}
		attachments = append(attachments, briefAttachment{Path: resolved, Size: info.Size(), IsImage: isImagePath(resolved)})
	}
	sort.Slice(attachments, func(i, j int) bool { return attachments[i].Path < attachments[j].Path })

	sentAt := time.Now().UTC().Format(time.RFC3339)
	orchestrationState.mu.Lock()
	orchestrationState.briefHistory = append(orchestrationState.briefHistory, briefMessageRecord{Message: message, Status: status, AttachmentCount: len(attachments), SentAt: sentAt})
	orchestrationState.mu.Unlock()

	out := briefOutput{Success: true, Message: message, Status: status, SentAt: sentAt, AttachmentCount: len(attachments), Attachments: attachments}
	return briefResult(out, false)
}

func isImagePath(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".webp" || ext == ".svg"
}

func briefResult(out briefOutput, isError bool) (types.ToolResult, error) {
	b, err := json.Marshal(out)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}
	return types.ToolResult{Content: string(b), IsError: isError}, nil
}

func (t *BriefTool) IsReadOnly(input types.ToolInput) bool { return true }

func (t *BriefTool) IsDestructive(input types.ToolInput) bool { return false }

func (t *BriefTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *BriefTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return types.PermissionAllowed
}
