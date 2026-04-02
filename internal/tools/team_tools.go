package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alliecatowo/alliecode/internal/tasks"
	"github.com/alliecatowo/alliecode/internal/types"
)

type TeamCreateTool struct{}
type TeamDeleteTool struct{}
type TeamListTool struct{}
type TeamStatusTool struct{}
type TeamUpdateTool struct{}

type teamCreateInput struct {
	TeamName    string   `json:"team_name"`
	Description string   `json:"description,omitempty"`
	Members     []string `json:"members,omitempty"`
}

type teamCreateOutput struct {
	Success   bool       `json:"success"`
	Team      teamRecord `json:"team"`
	Activated bool       `json:"activated"`
	Error     string     `json:"error,omitempty"`
}

type teamDeleteInput struct {
	TeamName string `json:"team_name,omitempty"`
	Archive  bool   `json:"archive,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

type teamDeleteOutput struct {
	Success   bool   `json:"success"`
	TeamName  string `json:"team_name,omitempty"`
	Deleted   bool   `json:"deleted"`
	Remaining int    `json:"remaining_teams"`
	Status    string `json:"status,omitempty"`
	Error     string `json:"error,omitempty"`
}

type teamListOutput struct {
	Success  bool                     `json:"success"`
	Total    int                      `json:"total"`
	Active   string                   `json:"active_team,omitempty"`
	Teams    []teamRecord             `json:"teams"`
	Summary  *tasks.TeamStatusSummary `json:"summary,omitempty"`
	Query    *tasks.TeamQuery         `json:"query,omitempty"`
	QueryRun *tasks.TeamQuerySummary  `json:"query_summary,omitempty"`
	Error    string                   `json:"error,omitempty"`
}

type teamStatusInput struct {
	TeamName     string `json:"team_name,omitempty"`
	MessageLimit int    `json:"message_limit,omitempty"`
	HistoryLimit int    `json:"history_limit,omitempty"`
}

type teamStatusOutput struct {
	Success bool                     `json:"success"`
	Team    teamRecord               `json:"team"`
	Summary *tasks.TeamStatusSummary `json:"summary,omitempty"`
	Error   string                   `json:"error,omitempty"`
}

type teamUpdateInput struct {
	TeamName    string   `json:"team_name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Members     []string `json:"members,omitempty"`
}

type teamUpdateOutput struct {
	Success bool       `json:"success"`
	Team    teamRecord `json:"team"`
	Summary any        `json:"summary,omitempty"`
	Error   string     `json:"error,omitempty"`
}

type teamListInput struct {
	Status string `json:"status,omitempty"`
	Member string `json:"member,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

func (t *TeamCreateTool) Name() string { return "team_create" }

func (t *TeamCreateTool) Description() string {
	return "Creates a local team context used by send_message and orchestration tools."
}

func (t *TeamCreateTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"team_name":   {Type: "string", Description: "Team name."},
			"description": {Type: "string", Description: "Optional team purpose."},
			"members":     {Type: "array", Description: "Optional teammate names.", Items: &types.PropertySchema{Type: "string"}},
		},
		Required: []string{"team_name"},
	}
}

func (t *TeamCreateTool) Execute(_ context.Context, input types.ToolInput, _ types.ToolContext) (types.ToolResult, error) {
	var in teamCreateInput
	if err := json.Unmarshal(input, &in); err != nil {
		return teamCreateResult(teamCreateOutput{Success: false, Error: fmt.Sprintf("invalid input: %v", err)}, true)
	}
	name := strings.TrimSpace(in.TeamName)
	if name == "" {
		return teamCreateResult(teamCreateOutput{Success: false, Error: "team_name is required"}, true)
	}

	orchestrationState.mu.Lock()
	defer orchestrationState.mu.Unlock()
	if _, exists := orchestrationState.teamStore.Get(name); exists {
		return teamCreateResult(teamCreateOutput{Success: false, Error: "team already exists"}, true)
	}

	now := time.Now().UTC()
	members := buildTeamMembers(in.Members, now)

	rec, err := orchestrationState.teamStore.Create(name, strings.TrimSpace(in.Description), members)
	if err != nil {
		return teamCreateResult(teamCreateOutput{Success: false, Error: err.Error()}, true)
	}

	return teamCreateResult(teamCreateOutput{Success: true, Team: rec, Activated: true}, false)
}

func teamCreateResult(out teamCreateOutput, isError bool) (types.ToolResult, error) {
	b, err := json.Marshal(out)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}
	return types.ToolResult{Content: string(b), IsError: isError}, nil
}

func (t *TeamCreateTool) IsReadOnly(input types.ToolInput) bool        { return false }
func (t *TeamCreateTool) IsDestructive(input types.ToolInput) bool     { return false }
func (t *TeamCreateTool) IsConcurrencySafe(input types.ToolInput) bool { return true }
func (t *TeamCreateTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyTeam, "team_create", input, toolCtx, types.PermissionAllowed)
}

func (t *TeamDeleteTool) Name() string { return "team_delete" }

func (t *TeamDeleteTool) Description() string {
	return "Cancels a local team lifecycle and deactivates it if active."
}

func (t *TeamDeleteTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"team_name": {Type: "string", Description: "Optional explicit team name. Defaults to active team."},
			"archive":   {Type: "boolean", Description: "Archive instead of cancel/deleting lifecycle metadata."},
			"reason":    {Type: "string", Description: "Optional cancellation/archive reason."},
		},
	}
}

func (t *TeamDeleteTool) Execute(_ context.Context, input types.ToolInput, _ types.ToolContext) (types.ToolResult, error) {
	var in teamDeleteInput
	if err := json.Unmarshal(input, &in); err != nil {
		return teamDeleteResult(teamDeleteOutput{Success: false, Deleted: false, Error: fmt.Sprintf("invalid input: %v", err)}, true)
	}

	orchestrationState.mu.Lock()
	defer orchestrationState.mu.Unlock()
	name := strings.TrimSpace(in.TeamName)
	if name == "" {
		name = orchestrationState.teamStore.Active()
	}
	if name == "" {
		return teamDeleteResult(teamDeleteOutput{Success: true, Deleted: false, Remaining: len(orchestrationState.teamStore.List())}, false)
	}
	reason := strings.TrimSpace(in.Reason)
	if reason == "" {
		reason = "team deleted"
	}
	var team teamRecord
	var ok bool
	if in.Archive {
		team, ok = orchestrationState.teamStore.Archive(name, reason)
	} else {
		team, ok = orchestrationState.teamStore.Cancel(name, reason)
	}
	if !ok {
		return teamDeleteResult(teamDeleteOutput{Success: false, TeamName: name, Deleted: false, Remaining: len(orchestrationState.teamStore.List()), Error: "team not found"}, true)
	}
	return teamDeleteResult(teamDeleteOutput{Success: true, TeamName: name, Deleted: true, Remaining: len(orchestrationState.teamStore.List()), Status: string(team.Status)}, false)
}

func teamDeleteResult(out teamDeleteOutput, isError bool) (types.ToolResult, error) {
	b, err := json.Marshal(out)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}
	return types.ToolResult{Content: string(b), IsError: isError}, nil
}

func (t *TeamDeleteTool) IsReadOnly(input types.ToolInput) bool        { return false }
func (t *TeamDeleteTool) IsDestructive(input types.ToolInput) bool     { return true }
func (t *TeamDeleteTool) IsConcurrencySafe(input types.ToolInput) bool { return true }
func (t *TeamDeleteTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyTeam, "team_delete", input, toolCtx, types.PermissionAllowed)
}

func (t *TeamListTool) Name() string { return "team_list" }

func (t *TeamListTool) Description() string {
	return "Lists teams and lifecycle status for the current process."
}

func (t *TeamListTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{Type: "object", Properties: map[string]types.PropertySchema{
		"status": {Type: "string", Description: "Optional team status filter (active, archived, canceled)."},
		"member": {Type: "string", Description: "Optional team member name filter."},
		"limit":  {Type: "integer", Description: "Optional max number of teams to return."},
	}}
}

func (t *TeamListTool) Execute(_ context.Context, input types.ToolInput, _ types.ToolContext) (types.ToolResult, error) {
	var in teamListInput
	if len(input) > 0 {
		if err := json.Unmarshal(input, &in); err != nil {
			return types.ToolResult{Content: fmt.Sprintf("invalid input: %v", err), IsError: true}, nil
		}
	}
	if in.Limit < 0 {
		return types.ToolResult{Content: "limit must be >= 0", IsError: true}, nil
	}
	query := tasks.TeamQuery{Member: strings.TrimSpace(in.Member), Limit: in.Limit}
	status := strings.ToLower(strings.TrimSpace(in.Status))
	if status != "" {
		switch status {
		case string(tasks.TeamStatusActive), string(tasks.TeamStatusArchived), string(tasks.TeamStatusCanceled):
			query.Statuses = []tasks.TeamStatus{tasks.TeamStatus(status)}
		default:
			return types.ToolResult{Content: "status must be one of: active, archived, canceled", IsError: true}, nil
		}
	}

	orchestrationState.mu.Lock()
	defer orchestrationState.mu.Unlock()
	teams := orchestrationState.teamStore.Query(query)
	summary := tasks.NewTeamStatusSummary(teams)
	querySummary := orchestrationState.teamStore.QuerySummary(query)
	out := teamListOutput{Success: true, Total: len(teams), Active: orchestrationState.teamStore.Active(), Teams: teams, Summary: &summary, Query: &query, QueryRun: &querySummary}
	b, err := json.Marshal(out)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}
	return types.ToolResult{Content: string(b)}, nil
}

func (t *TeamListTool) IsReadOnly(input types.ToolInput) bool        { return true }
func (t *TeamListTool) IsDestructive(input types.ToolInput) bool     { return false }
func (t *TeamListTool) IsConcurrencySafe(input types.ToolInput) bool { return true }
func (t *TeamListTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyTeam, "team_list", input, toolCtx, types.PermissionAllowed)
}

func (t *TeamStatusTool) Name() string { return "team_status" }

func (t *TeamStatusTool) Description() string {
	return "Gets a team record including members and message history."
}

func (t *TeamStatusTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"team_name": {Type: "string", Description: "Optional explicit team name. Defaults to active team."},
		},
	}
}

func (t *TeamStatusTool) Execute(_ context.Context, input types.ToolInput, _ types.ToolContext) (types.ToolResult, error) {
	var in teamStatusInput
	if len(input) > 0 {
		if err := json.Unmarshal(input, &in); err != nil {
			return types.ToolResult{Content: fmt.Sprintf("invalid input: %v", err), IsError: true}, nil
		}
	}

	orchestrationState.mu.Lock()
	defer orchestrationState.mu.Unlock()
	name := strings.TrimSpace(in.TeamName)
	if name == "" {
		name = orchestrationState.teamStore.Active()
	}
	if name == "" {
		return types.ToolResult{Content: "team_name is required when there is no active team", IsError: true}, nil
	}
	team, ok := orchestrationState.teamStore.Get(name)
	if !ok {
		return types.ToolResult{Content: "team not found", IsError: true}, nil
	}
	if in.MessageLimit > 0 && len(team.Messages) > in.MessageLimit {
		team.Messages = team.Messages[len(team.Messages)-in.MessageLimit:]
	}
	if in.HistoryLimit > 0 && len(team.History) > in.HistoryLimit {
		team.History = team.History[len(team.History)-in.HistoryLimit:]
	}

	summary, _ := orchestrationState.teamStore.StatusSummaryFor(name)
	b, err := json.Marshal(teamStatusOutput{Success: true, Team: team, Summary: &summary})
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}
	return types.ToolResult{Content: string(b)}, nil
}

func (t *TeamStatusTool) IsReadOnly(input types.ToolInput) bool        { return true }
func (t *TeamStatusTool) IsDestructive(input types.ToolInput) bool     { return false }
func (t *TeamStatusTool) IsConcurrencySafe(input types.ToolInput) bool { return true }
func (t *TeamStatusTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyTeam, "team_status", input, toolCtx, types.PermissionAllowed)
}

func (t *TeamUpdateTool) Name() string { return "team_update" }

func (t *TeamUpdateTool) Description() string {
	return "Updates team metadata such as description and members."
}

func (t *TeamUpdateTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"team_name":   {Type: "string", Description: "Optional explicit team name. Defaults to active team."},
			"description": {Type: "string", Description: "Optional new team description."},
			"members":     {Type: "array", Description: "Optional full replacement teammate member list.", Items: &types.PropertySchema{Type: "string"}},
		},
	}
}

func (t *TeamUpdateTool) Execute(_ context.Context, input types.ToolInput, _ types.ToolContext) (types.ToolResult, error) {
	var in teamUpdateInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("invalid input: %v", err), IsError: true}, nil
	}

	orchestrationState.mu.Lock()
	defer orchestrationState.mu.Unlock()

	name := strings.TrimSpace(in.TeamName)
	if name == "" {
		name = orchestrationState.teamStore.Active()
	}
	if name == "" {
		return types.ToolResult{Content: "team_name is required when there is no active team", IsError: true}, nil
	}

	params := tasks.TeamUpdateParams{}
	if in.Description != nil {
		d := strings.TrimSpace(*in.Description)
		params.Description = &d
	}
	if in.Members != nil {
		now := time.Now().UTC()
		members := buildTeamMembers(in.Members, now)
		params.Members = members
	}

	if params.Description == nil && params.Members == nil {
		return types.ToolResult{Content: "at least one of description or members is required", IsError: true}, nil
	}

	team, ok := orchestrationState.teamStore.Update(name, params)
	if !ok {
		return types.ToolResult{Content: "team not found", IsError: true}, nil
	}

	summary := orchestrationState.teamStore.QuerySummary(tasks.TeamQuery{})
	b, err := json.Marshal(teamUpdateOutput{Success: true, Team: team, Summary: summary})
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}
	return types.ToolResult{Content: string(b)}, nil
}

func (t *TeamUpdateTool) IsReadOnly(input types.ToolInput) bool        { return false }
func (t *TeamUpdateTool) IsDestructive(input types.ToolInput) bool     { return false }
func (t *TeamUpdateTool) IsConcurrencySafe(input types.ToolInput) bool { return true }
func (t *TeamUpdateTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyTeam, "team_update", input, toolCtx, types.PermissionAllowed)
}
