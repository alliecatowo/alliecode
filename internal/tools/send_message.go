package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/alliecatowo/alliecode/internal/types"
)

type SendMessageTool struct{}

type sendMessageInput struct {
	To      string `json:"to"`
	Summary string `json:"summary,omitempty"`
	Message string `json:"message"`
	From    string `json:"from,omitempty"`
}

type sendMessageOutput struct {
	Success    bool           `json:"success"`
	Team       string         `json:"team,omitempty"`
	From       string         `json:"from"`
	Recipients []string       `json:"recipients"`
	Stored     int            `json:"stored"`
	Messages   []inboxMessage `json:"messages,omitempty"`
	Error      string         `json:"error,omitempty"`
}

func (t *SendMessageTool) Name() string { return "send_message" }

func (t *SendMessageTool) Description() string {
	return "Sends team messages to one teammate or broadcasts to all teammates."
}

func (t *SendMessageTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"to":      {Type: "string", Description: "Recipient teammate name or * for broadcast."},
			"summary": {Type: "string", Description: "Optional short preview."},
			"message": {Type: "string", Description: "Message body."},
			"from":    {Type: "string", Description: "Optional sender name. Defaults to team_lead."},
		},
		Required: []string{"to", "message"},
	}
}

func (t *SendMessageTool) Execute(_ context.Context, input types.ToolInput, _ types.ToolContext) (types.ToolResult, error) {
	var in sendMessageInput
	if err := json.Unmarshal(input, &in); err != nil {
		return sendMessageResult(sendMessageOutput{Success: false, Error: fmt.Sprintf("invalid input: %v", err)}, true)
	}
	to := strings.TrimSpace(in.To)
	if to == "" {
		return sendMessageResult(sendMessageOutput{Success: false, Error: "to is required"}, true)
	}
	msg := strings.TrimSpace(in.Message)
	if msg == "" {
		return sendMessageResult(sendMessageOutput{Success: false, Error: "message is required"}, true)
	}
	from := strings.TrimSpace(in.From)
	if from == "" {
		from = "team_lead"
	}

	orchestrationState.mu.Lock()
	defer orchestrationState.mu.Unlock()
	teamName := orchestrationState.teamStore.Active()
	if teamName == "" {
		return sendMessageResult(sendMessageOutput{Success: false, From: from, Error: "no active team; create one with team_create"}, true)
	}
	team, ok := orchestrationState.teamStore.Get(teamName)
	if !ok {
		return sendMessageResult(sendMessageOutput{Success: false, Team: teamName, From: from, Error: "active team record missing"}, true)
	}
	if team.Status != "active" {
		return sendMessageResult(sendMessageOutput{Success: false, Team: teamName, From: from, Error: "active team is not available for messaging"}, true)
	}

	recipients := []string{}
	if to == "*" {
		for _, member := range team.Members {
			if member.Name == from {
				continue
			}
			recipients = append(recipients, member.Name)
		}
		sort.Strings(recipients)
	} else {
		if !teamHasMember(team, to) {
			return sendMessageResult(sendMessageOutput{Success: false, Team: teamName, From: from, Error: "recipient is not in the active team"}, true)
		}
		recipients = []string{to}
	}

	stored := make([]inboxMessage, 0, len(recipients))
	for _, recipient := range recipients {
		rec, err := orchestrationState.teamStore.AddMessage(teamName, from, recipient, strings.TrimSpace(in.Summary), msg)
		if err != nil {
			return sendMessageResult(sendMessageOutput{Success: false, Team: teamName, From: from, Error: fmt.Sprintf("failed to store message: %v", err)}, true)
		}
		stored = append(stored, rec)
		if agentTaskManager != nil {
			_, _ = agentTaskManager.RecordMessage(recipient, from, recipient, rec.Summary, rec.Message)
		}
	}

	out := sendMessageOutput{Success: true, Team: teamName, From: from, Recipients: recipients, Stored: len(stored), Messages: stored}
	return sendMessageResult(out, false)
}

func teamHasMember(team teamRecord, name string) bool {
	for _, m := range team.Members {
		if m.Name == name {
			return true
		}
	}
	return false
}

func sendMessageResult(out sendMessageOutput, isError bool) (types.ToolResult, error) {
	b, err := json.Marshal(out)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}
	return types.ToolResult{Content: string(b), IsError: isError}, nil
}

func (t *SendMessageTool) IsReadOnly(input types.ToolInput) bool { return false }

func (t *SendMessageTool) IsDestructive(input types.ToolInput) bool { return false }

func (t *SendMessageTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *SendMessageTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyTeam, "send_message", input, toolCtx, types.PermissionAllowed)
}
