package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

type CronCreateTool struct{}
type CronDeleteTool struct{}
type CronListTool struct{}

type cronCreateInput struct {
	Cron      string `json:"cron"`
	Prompt    string `json:"prompt"`
	Recurring *bool  `json:"recurring,omitempty"`
}

type cronCreateOutput struct {
	Success      bool   `json:"success"`
	ID           string `json:"id,omitempty"`
	Cron         string `json:"cron,omitempty"`
	Prompt       string `json:"prompt,omitempty"`
	Recurring    bool   `json:"recurring"`
	NextRunAt    string `json:"next_run_at,omitempty"`
	Human        string `json:"human_schedule,omitempty"`
	ErrorMessage string `json:"error,omitempty"`
}

type cronDeleteInput struct {
	ID string `json:"id"`
}

type cronDeleteOutput struct {
	Success      bool   `json:"success"`
	ID           string `json:"id,omitempty"`
	Deleted      bool   `json:"deleted"`
	ErrorMessage string `json:"error,omitempty"`
}

type cronListOutput struct {
	Success bool            `json:"success"`
	Jobs    []cronJobRecord `json:"jobs"`
}

func (t *CronCreateTool) Name() string { return "cron_create" }

func (t *CronCreateTool) Description() string {
	return "Schedules a local cron-like reminder with deterministic validation."
}

func (t *CronCreateTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"cron":      {Type: "string", Description: "Five-field cron expression (M H DoM Mon DoW)."},
			"prompt":    {Type: "string", Description: "Prompt payload to run when matched."},
			"recurring": {Type: "boolean", Description: "Whether the job should recur. Defaults to true."},
		},
		Required: []string{"cron", "prompt"},
	}
}

func (t *CronCreateTool) Execute(_ context.Context, input types.ToolInput, _ types.ToolContext) (types.ToolResult, error) {
	var in cronCreateInput
	if err := json.Unmarshal(input, &in); err != nil {
		return cronCreateResult(cronCreateOutput{Success: false, ErrorMessage: fmt.Sprintf("invalid input: %v", err)}, true)
	}
	cronExpr := strings.TrimSpace(in.Cron)
	if cronExpr == "" {
		return cronCreateResult(cronCreateOutput{Success: false, ErrorMessage: "cron is required"}, true)
	}
	prompt := strings.TrimSpace(in.Prompt)
	if prompt == "" {
		return cronCreateResult(cronCreateOutput{Success: false, ErrorMessage: "prompt is required"}, true)
	}
	if !isValidCronExpression(cronExpr) {
		return cronCreateResult(cronCreateOutput{Success: false, ErrorMessage: "invalid cron expression"}, true)
	}
	next := nextCronRun(cronExpr, time.Now().UTC())
	if next.IsZero() {
		return cronCreateResult(cronCreateOutput{Success: false, ErrorMessage: "cron does not match any time in the next year"}, true)
	}
	recurring := true
	if in.Recurring != nil {
		recurring = *in.Recurring
	}

	orchestrationState.mu.Lock()
	orchestrationState.cronSeq++
	id := fmt.Sprintf("cron-%04d", orchestrationState.cronSeq)
	rec := cronJobRecord{
		ID:        id,
		Cron:      cronExpr,
		Prompt:    prompt,
		Recurring: recurring,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		NextRunAt: next.Format(time.RFC3339),
	}
	orchestrationState.cronJobs[id] = rec
	orchestrationState.mu.Unlock()

	out := cronCreateOutput{Success: true, ID: id, Cron: rec.Cron, Prompt: rec.Prompt, Recurring: rec.Recurring, NextRunAt: rec.NextRunAt, Human: cronHuman(rec.Cron)}
	return cronCreateResult(out, false)
}

func cronCreateResult(out cronCreateOutput, isError bool) (types.ToolResult, error) {
	b, err := json.Marshal(out)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}
	return types.ToolResult{Content: string(b), IsError: isError}, nil
}

func (t *CronCreateTool) IsReadOnly(input types.ToolInput) bool        { return false }
func (t *CronCreateTool) IsDestructive(input types.ToolInput) bool     { return false }
func (t *CronCreateTool) IsConcurrencySafe(input types.ToolInput) bool { return true }
func (t *CronCreateTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return types.PermissionAllowed
}

func (t *CronDeleteTool) Name() string { return "cron_delete" }

func (t *CronDeleteTool) Description() string { return "Deletes a scheduled local cron job by id." }

func (t *CronDeleteTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{Type: "object", Properties: map[string]types.PropertySchema{"id": {Type: "string", Description: "Cron job id."}}, Required: []string{"id"}}
}

func (t *CronDeleteTool) Execute(_ context.Context, input types.ToolInput, _ types.ToolContext) (types.ToolResult, error) {
	var in cronDeleteInput
	if err := json.Unmarshal(input, &in); err != nil {
		return cronDeleteResult(cronDeleteOutput{Success: false, Deleted: false, ErrorMessage: fmt.Sprintf("invalid input: %v", err)}, true)
	}
	id := strings.TrimSpace(in.ID)
	if id == "" {
		return cronDeleteResult(cronDeleteOutput{Success: false, Deleted: false, ErrorMessage: "id is required"}, true)
	}
	orchestrationState.mu.Lock()
	_, ok := orchestrationState.cronJobs[id]
	if ok {
		delete(orchestrationState.cronJobs, id)
	}
	orchestrationState.mu.Unlock()
	if !ok {
		return cronDeleteResult(cronDeleteOutput{Success: false, ID: id, Deleted: false, ErrorMessage: "cron job not found"}, true)
	}
	return cronDeleteResult(cronDeleteOutput{Success: true, ID: id, Deleted: true}, false)
}

func cronDeleteResult(out cronDeleteOutput, isError bool) (types.ToolResult, error) {
	b, err := json.Marshal(out)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}
	return types.ToolResult{Content: string(b), IsError: isError}, nil
}

func (t *CronDeleteTool) IsReadOnly(input types.ToolInput) bool        { return false }
func (t *CronDeleteTool) IsDestructive(input types.ToolInput) bool     { return true }
func (t *CronDeleteTool) IsConcurrencySafe(input types.ToolInput) bool { return true }
func (t *CronDeleteTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return types.PermissionAllowed
}

func (t *CronListTool) Name() string { return "cron_list" }

func (t *CronListTool) Description() string { return "Lists scheduled local cron jobs." }

func (t *CronListTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{Type: "object", Properties: map[string]types.PropertySchema{}}
}

func (t *CronListTool) Execute(_ context.Context, _ types.ToolInput, _ types.ToolContext) (types.ToolResult, error) {
	orchestrationState.mu.Lock()
	jobs := make([]cronJobRecord, 0, len(orchestrationState.cronJobs))
	for _, job := range orchestrationState.cronJobs {
		jobs = append(jobs, job)
	}
	orchestrationState.mu.Unlock()
	sort.Slice(jobs, func(i, j int) bool { return jobs[i].ID < jobs[j].ID })
	b, err := json.Marshal(cronListOutput{Success: true, Jobs: jobs})
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}
	return types.ToolResult{Content: string(b)}, nil
}

func (t *CronListTool) IsReadOnly(input types.ToolInput) bool        { return true }
func (t *CronListTool) IsDestructive(input types.ToolInput) bool     { return false }
func (t *CronListTool) IsConcurrencySafe(input types.ToolInput) bool { return true }
func (t *CronListTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return types.PermissionAllowed
}

func cronHuman(cronExpr string) string {
	fields := strings.Fields(cronExpr)
	if len(fields) != 5 {
		return cronExpr
	}
	return fmt.Sprintf("minute=%s hour=%s dom=%s month=%s dow=%s", fields[0], fields[1], fields[2], fields[3], fields[4])
}

func isValidCronExpression(expr string) bool {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return false
	}
	for idx, field := range fields {
		min := 0
		max := 59
		switch idx {
		case 1:
			max = 23
		case 2:
			min, max = 1, 31
		case 3:
			min, max = 1, 12
		case 4:
			max = 6
		}
		if !isValidCronField(field, min, max) {
			return false
		}
	}
	return true
}

func isValidCronField(field string, min, max int) bool {
	parts := strings.Split(field, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "*" {
			continue
		}
		if strings.HasPrefix(part, "*/") {
			n, err := strconv.Atoi(strings.TrimPrefix(part, "*/"))
			if err != nil || n <= 0 {
				return false
			}
			continue
		}
		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			if len(bounds) != 2 {
				return false
			}
			a, errA := strconv.Atoi(bounds[0])
			b, errB := strconv.Atoi(bounds[1])
			if errA != nil || errB != nil || a < min || b > max || a > b {
				return false
			}
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil || n < min || n > max {
			return false
		}
	}
	return true
}

func cronFieldMatches(field string, value int) bool {
	for _, part := range strings.Split(field, ",") {
		part = strings.TrimSpace(part)
		if part == "*" {
			return true
		}
		if strings.HasPrefix(part, "*/") {
			n, err := strconv.Atoi(strings.TrimPrefix(part, "*/"))
			if err == nil && n > 0 && value%n == 0 {
				return true
			}
			continue
		}
		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			a, errA := strconv.Atoi(bounds[0])
			b, errB := strconv.Atoi(bounds[1])
			if errA == nil && errB == nil && value >= a && value <= b {
				return true
			}
			continue
		}
		n, err := strconv.Atoi(part)
		if err == nil && n == value {
			return true
		}
	}
	return false
}

func nextCronRun(expr string, from time.Time) time.Time {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return time.Time{}
	}
	t := from.UTC().Truncate(time.Minute).Add(time.Minute)
	deadline := t.Add(366 * 24 * time.Hour)
	for !t.After(deadline) {
		if cronFieldMatches(fields[0], t.Minute()) &&
			cronFieldMatches(fields[1], t.Hour()) &&
			cronFieldMatches(fields[2], t.Day()) &&
			cronFieldMatches(fields[3], int(t.Month())) &&
			cronFieldMatches(fields[4], int(t.Weekday())) {
			return t
		}
		t = t.Add(time.Minute)
	}
	return time.Time{}
}
