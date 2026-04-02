package tasks

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

type Status string

const (
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCanceled  Status = "canceled"
)

type Runner func(ctx context.Context, prompt, model string, toolCtx types.ToolContext) (string, error)

type Task struct {
	ID          string         `json:"task_id"`
	Subject     string         `json:"subject,omitempty"`
	Description string         `json:"description,omitempty"`
	ActiveForm  string         `json:"active_form,omitempty"`
	Owner       string         `json:"owner,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	Status      Status         `json:"status"`
	Prompt      string         `json:"prompt,omitempty"`
	Model       string         `json:"model,omitempty"`
	Result      string         `json:"result,omitempty"`
	Error       string         `json:"error,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	History     []Event        `json:"history,omitempty"`
}

type Event struct {
	Index     uint64    `json:"index,omitempty"`
	Type      string    `json:"type"`
	Status    Status    `json:"status,omitempty"`
	Message   string    `json:"message,omitempty"`
	From      string    `json:"from,omitempty"`
	To        string    `json:"to,omitempty"`
	Summary   string    `json:"summary,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type UpdateParams struct {
	Status      *Status
	Result      *string
	Error       *string
	Subject     *string
	Description *string
	ActiveForm  *string
	Owner       *string
	Metadata    map[string]any
}

// Manager tracks subagent tasks and their lifecycle.
type Manager struct {
	runner  Runner
	onEvent func(Task, Event)

	mu    sync.RWMutex
	tasks map[string]*Task
	cxl   map[string]context.CancelFunc

	seq atomic.Uint64
}

type Stats struct {
	Total     int `json:"total"`
	Running   int `json:"running"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
	Canceled  int `json:"canceled"`
}

func NewManager(runner Runner) *Manager {
	return &Manager{
		runner: runner,
		tasks:  make(map[string]*Task),
		cxl:    make(map[string]context.CancelFunc),
	}
}

func (m *Manager) SetEventCallback(cb func(Task, Event)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onEvent = cb
}

func (m *Manager) Start(ctx context.Context, prompt, model string, toolCtx types.ToolContext) (Task, error) {
	if m.runner == nil {
		return Task{}, fmt.Errorf("task manager runner is not configured")
	}

	now := time.Now().UTC()
	id := nextTaskID(now, m.seq.Add(1))
	createdEvent := newCreatedEvent(m.seq.Add(1), now)
	statusEvent := newStatusEvent(m.seq.Add(1), StatusRunning, "", now)
	t := &Task{
		ID:        id,
		Status:    StatusRunning,
		Prompt:    prompt,
		Model:     model,
		CreatedAt: now,
		UpdatedAt: now,
		History:   []Event{createdEvent},
	}
	t.History = append(t.History, statusEvent)

	runCtx, cancel := context.WithCancel(ctx)

	m.mu.Lock()
	m.tasks[id] = t
	m.cxl[id] = cancel
	cb := m.onEvent
	initial := cloneTask(t)
	m.mu.Unlock()

	if cb != nil {
		cb(initial, createdEvent)
		cb(initial, statusEvent)
	}

	go m.runTask(runCtx, id, prompt, model, toolCtx)

	return cloneTask(t), nil
}

func (m *Manager) Get(id string) (Task, bool) {
	m.mu.RLock()
	t, ok := m.tasks[id]
	m.mu.RUnlock()
	if !ok {
		return Task{}, false
	}
	return cloneTask(t), true
}

func (m *Manager) Wait(ctx context.Context, id string, pollInterval time.Duration) (Task, error) {
	if pollInterval <= 0 {
		pollInterval = 200 * time.Millisecond
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		t, ok := m.Get(id)
		if !ok {
			return Task{}, fmt.Errorf("task %q not found", id)
		}
		if isTerminalStatus(t.Status) {
			return t, nil
		}

		select {
		case <-ctx.Done():
			return Task{}, ctx.Err()
		case <-ticker.C:
		}
	}
}

func (m *Manager) List() []Task {
	m.mu.RLock()
	out := make([]Task, 0, len(m.tasks))
	for _, t := range m.tasks {
		out = append(out, cloneTask(t))
	}
	m.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

func (m *Manager) Stats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var out Stats
	out.Total = len(m.tasks)
	for _, t := range m.tasks {
		switch t.Status {
		case StatusRunning:
			out.Running++
		case StatusCompleted:
			out.Completed++
		case StatusFailed:
			out.Failed++
		case StatusCanceled:
			out.Canceled++
		}
	}
	return out
}

func (m *Manager) StatusSummary() StatusSummary {
	list := m.List()
	return newStatusSummary(list)
}

func (m *Manager) StatusSummaryByOwner(owner string) StatusSummary {
	owner = strings.TrimSpace(owner)
	list := m.List()
	if owner == "" {
		return newStatusSummary(list)
	}
	filtered := make([]Task, 0, len(list))
	for _, task := range list {
		if strings.EqualFold(strings.TrimSpace(task.Owner), owner) {
			filtered = append(filtered, task)
		}
	}
	return newStatusSummary(filtered)
}

func (m *Manager) Query(query TaskQuery) []Task {
	items, _ := ApplyTaskQuery(m.List(), query)
	return items
}

func (m *Manager) QuerySummary(query TaskQuery) TaskQuerySummary {
	_, summary := ApplyTaskQuery(m.List(), query)
	return summary
}

func (m *Manager) Update(id string, params UpdateParams) (Task, bool) {
	m.mu.Lock()
	t, ok := m.tasks[id]
	if !ok {
		m.mu.Unlock()
		return Task{}, false
	}

	var emitted *Event

	if params.Status != nil {
		if !isValidStatusTransition(t.Status, *params.Status) {
			out := cloneTask(t)
			m.mu.Unlock()
			return out, true
		}
		t.Status = *params.Status
		ev := Event{Index: m.seq.Add(1), Type: "status", Status: *params.Status, CreatedAt: time.Now().UTC()}
		t.History = append(t.History, ev)
		emitted = &ev
	}
	if params.Result != nil {
		t.Result = *params.Result
	}
	if params.Error != nil {
		t.Error = *params.Error
	}
	if params.Subject != nil {
		t.Subject = *params.Subject
	}
	if params.Description != nil {
		t.Description = *params.Description
	}
	if params.ActiveForm != nil {
		t.ActiveForm = *params.ActiveForm
	}
	if params.Owner != nil {
		t.Owner = *params.Owner
	}
	if params.Metadata != nil {
		t.Metadata = copyMetadata(params.Metadata)
	}
	t.UpdatedAt = time.Now().UTC()
	out := cloneTask(t)
	cb := m.onEvent
	m.mu.Unlock()

	if cb != nil && emitted != nil {
		cb(out, *emitted)
	}
	return out, true
}

func (m *Manager) Cancel(id, reason string) (Task, bool) {
	if reason == "" {
		reason = "canceled"
	}

	var cancel context.CancelFunc
	var out Task

	m.mu.Lock()
	t, ok := m.tasks[id]
	if !ok {
		m.mu.Unlock()
		return Task{}, false
	}
	if isTerminalStatus(t.Status) {
		out = cloneTask(t)
		m.mu.Unlock()
		return out, true
	}
	t.Status = StatusCanceled
	t.Error = reason
	t.UpdatedAt = time.Now().UTC()
	ev := Event{Index: m.seq.Add(1), Type: "canceled", Status: StatusCanceled, Message: reason, CreatedAt: t.UpdatedAt}
	t.History = append(t.History, ev)
	cancel = m.cxl[id]
	delete(m.cxl, id)
	out = cloneTask(t)
	cb := m.onEvent
	m.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if cb != nil {
		cb(out, ev)
	}

	return out, true
}

func (m *Manager) RecordMessage(id, from, to, summary, message string) (Task, bool) {
	m.mu.Lock()
	t, ok := m.tasks[id]
	if !ok {
		m.mu.Unlock()
		return Task{}, false
	}
	now := time.Now().UTC()
	ev := Event{
		Index:     m.seq.Add(1),
		Type:      "message",
		From:      from,
		To:        to,
		Summary:   summary,
		Message:   message,
		CreatedAt: now,
	}
	t.History = append(t.History, ev)
	t.UpdatedAt = now
	out := cloneTask(t)
	cb := m.onEvent
	m.mu.Unlock()
	if cb != nil {
		cb(out, ev)
	}
	return out, true
}

func (m *Manager) runTask(ctx context.Context, id, prompt, model string, toolCtx types.ToolContext) {
	result, err := m.runner(ctx, prompt, model, toolCtx)
	now := time.Now().UTC()

	m.mu.Lock()
	t, ok := m.tasks[id]
	if !ok {
		m.mu.Unlock()
		return
	}
	delete(m.cxl, id)

	if t.Status == StatusCanceled {
		m.mu.Unlock()
		return
	}

	if err != nil {
		if errors.Is(err, context.Canceled) {
			t.Status = StatusCanceled
			t.Error = "canceled"
			t.UpdatedAt = now
			ev := newStatusEvent(m.seq.Add(1), StatusCanceled, "", now)
			t.History = append(t.History, ev)
			out := cloneTask(t)
			cb := m.onEvent
			m.mu.Unlock()
			if cb != nil {
				cb(out, ev)
			}
			return
		}
		t.Status = StatusFailed
		t.Error = err.Error()
		t.UpdatedAt = now
		ev := newStatusEvent(m.seq.Add(1), StatusFailed, err.Error(), now)
		t.History = append(t.History, ev)
		out := cloneTask(t)
		cb := m.onEvent
		m.mu.Unlock()
		if cb != nil {
			cb(out, ev)
		}
		return
	}

	t.Status = StatusCompleted
	t.Result = result
	t.UpdatedAt = now
	ev := newStatusEvent(m.seq.Add(1), StatusCompleted, "", now)
	t.History = append(t.History, ev)
	out := cloneTask(t)
	cb := m.onEvent
	m.mu.Unlock()
	if cb != nil {
		cb(out, ev)
	}
}

func cloneTask(t *Task) Task {
	if t == nil {
		return Task{}
	}
	out := *t
	if len(t.History) > 0 {
		out.History = append([]Event(nil), t.History...)
	}
	out.Metadata = copyMetadata(t.Metadata)
	return out
}

func isValidStatusTransition(from, to Status) bool {
	if from == to {
		return true
	}
	switch from {
	case StatusRunning:
		return to == StatusCompleted || to == StatusFailed || to == StatusCanceled
	case StatusCompleted, StatusFailed, StatusCanceled:
		return false
	default:
		return true
	}
}
