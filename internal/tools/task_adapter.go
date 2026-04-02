package tools

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/alliecatowo/alliecode/internal/tasks"
)

type taskAdapterRecord struct {
	Subject     string
	Description string
	ActiveForm  string
	Owner       string
	Metadata    map[string]any
	Status      string
	Result      string
	Error       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

var (
	taskAdapterMu      sync.RWMutex
	taskAdapterRecords = map[string]taskAdapterRecord{}
)

func taskAdapterGet(ctx context.Context, taskID string) (tasks.Task, bool) {
	if agentTaskManager != nil {
		if t, ok := agentTaskManager.Get(taskID); ok {
			return applyTaskOverride(t), true
		}
	}

	taskAdapterMu.RLock()
	rec, ok := taskAdapterRecords[taskID]
	taskAdapterMu.RUnlock()
	if !ok {
		return tasks.Task{}, false
	}

	now := rec.UpdatedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}
	createdAt := rec.CreatedAt
	if createdAt.IsZero() {
		createdAt = now
	}
	return tasks.Task{
		ID:          taskID,
		Subject:     rec.Subject,
		Description: rec.Description,
		ActiveForm:  rec.ActiveForm,
		Owner:       rec.Owner,
		Metadata:    cloneMetadata(rec.Metadata),
		Status:      tasks.Status(rec.Status),
		Result:      rec.Result,
		Error:       rec.Error,
		CreatedAt:   createdAt,
		UpdatedAt:   now,
	}, true
}

func taskAdapterWait(ctx context.Context, taskID string, pollInterval time.Duration) (tasks.Task, error) {
	if agentTaskManager != nil {
		if _, ok := agentTaskManager.Get(taskID); ok {
			t, err := agentTaskManager.Wait(ctx, taskID, pollInterval)
			if err != nil {
				return tasks.Task{}, err
			}
			return applyTaskOverride(t), nil
		}
	}

	for {
		t, ok := taskAdapterGet(ctx, taskID)
		if !ok {
			return tasks.Task{}, fmt.Errorf("task %q not found", taskID)
		}
		if t.Status != tasks.StatusRunning {
			return t, nil
		}

		select {
		case <-ctx.Done():
			return tasks.Task{}, ctx.Err()
		case <-time.After(pollInterval):
		}
	}
}

func taskAdapterUpdate(taskID, status, result, errText string) taskAdapterRecord {
	if agentTaskManager != nil {
		params := tasks.UpdateParams{}
		if status != "" {
			s := tasks.Status(status)
			params.Status = &s
		}
		if result != "" {
			resultCopy := result
			params.Result = &resultCopy
		}
		if errText != "" {
			errCopy := errText
			params.Error = &errCopy
		}
		if t, ok := agentTaskManager.Update(taskID, params); ok {
			return taskAdapterRecord{
				Status:    string(t.Status),
				Result:    t.Result,
				Error:     t.Error,
				UpdatedAt: t.UpdatedAt,
			}
		}
	}

	now := time.Now().UTC()
	rec := taskAdapterRecord{
		Status:    status,
		Result:    result,
		Error:     errText,
		CreatedAt: now,
		UpdatedAt: now,
	}
	taskAdapterMu.Lock()
	if existing, ok := taskAdapterRecords[taskID]; ok && !existing.CreatedAt.IsZero() {
		rec.CreatedAt = existing.CreatedAt
	}
	taskAdapterRecords[taskID] = rec
	taskAdapterMu.Unlock()

	return rec
}

func taskAdapterSetFields(taskID string, mutate func(rec *taskAdapterRecord)) taskAdapterRecord {
	taskAdapterMu.Lock()
	rec := taskAdapterRecords[taskID]
	mutate(&rec)
	rec.UpdatedAt = time.Now().UTC()
	taskAdapterRecords[taskID] = rec
	taskAdapterMu.Unlock()
	return rec
}

func applyTaskOverride(t tasks.Task) tasks.Task {
	taskAdapterMu.RLock()
	rec, ok := taskAdapterRecords[t.ID]
	taskAdapterMu.RUnlock()
	if !ok {
		return t
	}
	if rec.Status != "" {
		t.Status = tasks.Status(rec.Status)
	}
	if rec.Subject != "" {
		t.Subject = rec.Subject
	}
	if rec.Description != "" {
		t.Description = rec.Description
	}
	if rec.ActiveForm != "" {
		t.ActiveForm = rec.ActiveForm
	}
	if rec.Owner != "" {
		t.Owner = rec.Owner
	}
	if len(rec.Metadata) > 0 {
		t.Metadata = cloneMetadata(rec.Metadata)
	}
	if rec.Result != "" {
		t.Result = rec.Result
	}
	if rec.Error != "" {
		t.Error = rec.Error
	}
	if !rec.UpdatedAt.IsZero() {
		t.UpdatedAt = rec.UpdatedAt
	}
	if !rec.CreatedAt.IsZero() {
		t.CreatedAt = rec.CreatedAt
	}
	return t
}

func cloneMetadata(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func taskAdapterList(ctx context.Context) []tasks.Task {
	if agentTaskManager != nil {
		out := agentTaskManager.List()
		for i := range out {
			out[i] = applyTaskOverride(out[i])
		}
		return out
	}

	taskAdapterMu.RLock()
	ids := make([]string, 0, len(taskAdapterRecords))
	for id := range taskAdapterRecords {
		ids = append(ids, id)
	}
	taskAdapterMu.RUnlock()

	sort.Strings(ids)
	out := make([]tasks.Task, 0, len(ids))
	for _, id := range ids {
		t, ok := taskAdapterGet(ctx, id)
		if !ok {
			continue
		}
		out = append(out, t)
	}
	return out
}

func taskAdapterSummary(ctx context.Context) tasks.StatusSummary {
	if agentTaskManager != nil {
		return agentTaskManager.StatusSummary()
	}
	return tasksStatusSummaryFromList(taskAdapterList(ctx))
}

func tasksStatusSummaryFromList(items []tasks.Task) tasks.StatusSummary {
	out := tasks.StatusSummary{
		ByStatus: map[tasks.Status]int{
			tasks.StatusRunning:   0,
			tasks.StatusCompleted: 0,
			tasks.StatusFailed:    0,
			tasks.StatusCanceled:  0,
		},
		ByOwner: map[string]int{},
	}
	for _, item := range items {
		out.Total++
		out.ByStatus[item.Status]++
		if item.Owner != "" {
			out.ByOwner[item.Owner]++
		}
		if item.Status == tasks.StatusCompleted || item.Status == tasks.StatusFailed || item.Status == tasks.StatusCanceled {
			out.Terminal++
		} else {
			out.NonTerminal++
		}
		if item.UpdatedAt.UnixNano() > out.RecentAt {
			out.RecentAt = item.UpdatedAt.UnixNano()
			out.RecentTask = item.ID
		}
	}
	if out.Total > 0 {
		out.TerminalPct = float64(out.Terminal) / float64(out.Total)
	}
	if len(out.ByOwner) == 0 {
		out.ByOwner = nil
	}
	return out
}
