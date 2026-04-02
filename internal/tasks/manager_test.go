package tasks

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestManagerStartAndWaitCompleted(t *testing.T) {
	mgr := NewManager(func(ctx context.Context, prompt, model string, toolCtx types.ToolContext) (string, error) {
		return "ok:" + prompt + ":" + model, nil
	})

	task, err := mgr.Start(context.Background(), "do work", "m1", types.ToolContext{})
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if task.ID == "" {
		t.Fatal("expected task id")
	}

	finished, err := mgr.Wait(context.Background(), task.ID, 5*time.Millisecond)
	if err != nil {
		t.Fatalf("Wait returned error: %v", err)
	}
	if finished.Status != StatusCompleted {
		t.Fatalf("expected completed status, got %s", finished.Status)
	}
	if finished.Result != "ok:do work:m1" {
		t.Fatalf("unexpected result: %q", finished.Result)
	}
}

func TestManagerWaitFailed(t *testing.T) {
	mgr := NewManager(func(ctx context.Context, prompt, model string, toolCtx types.ToolContext) (string, error) {
		return "", errors.New("boom")
	})

	task, err := mgr.Start(context.Background(), "do work", "", types.ToolContext{})
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	finished, err := mgr.Wait(context.Background(), task.ID, 5*time.Millisecond)
	if err != nil {
		t.Fatalf("Wait returned error: %v", err)
	}
	if finished.Status != StatusFailed {
		t.Fatalf("expected failed status, got %s", finished.Status)
	}
	if finished.Error == "" {
		t.Fatal("expected error message")
	}
}

func TestManagerWaitContextCanceled(t *testing.T) {
	mgr := NewManager(func(ctx context.Context, prompt, model string, toolCtx types.ToolContext) (string, error) {
		time.Sleep(100 * time.Millisecond)
		return "late", nil
	})

	task, err := mgr.Start(context.Background(), "do work", "", types.ToolContext{})
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err = mgr.Wait(ctx, task.ID, 5*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}

func TestManagerLifecycleUpdateCancelAndHistory(t *testing.T) {
	mgr := NewManager(func(ctx context.Context, prompt, model string, toolCtx types.ToolContext) (string, error) {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(150 * time.Millisecond):
			return "late", nil
		}
	})

	task, err := mgr.Start(context.Background(), "do work", "m1", types.ToolContext{})
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	res := "partial"
	status := StatusRunning
	updated, ok := mgr.Update(task.ID, UpdateParams{Status: &status, Result: &res})
	if !ok {
		t.Fatalf("expected Update to find task")
	}
	if updated.Result != "partial" {
		t.Fatalf("unexpected update result: %q", updated.Result)
	}

	rec, ok := mgr.RecordMessage(task.ID, "team_lead", "agent-1", "sync", "please continue")
	if !ok {
		t.Fatalf("expected RecordMessage to find task")
	}
	if len(rec.History) < 1 || rec.History[len(rec.History)-1].Type != "message" {
		t.Fatalf("expected message event in history")
	}

	canceled, ok := mgr.Cancel(task.ID, "manual cancel")
	if !ok {
		t.Fatalf("expected Cancel to find task")
	}
	if canceled.Status != StatusCanceled {
		t.Fatalf("status = %q, want %q", canceled.Status, StatusCanceled)
	}

	finished, err := mgr.Wait(context.Background(), task.ID, 5*time.Millisecond)
	if err != nil {
		t.Fatalf("Wait returned error: %v", err)
	}
	if finished.Status != StatusCanceled {
		t.Fatalf("expected canceled status, got %s", finished.Status)
	}

	all := mgr.List()
	if len(all) != 1 {
		t.Fatalf("expected one task in list, got %d", len(all))
	}
	if len(all[0].History) < 4 {
		t.Fatalf("expected rich lifecycle history, got %d events", len(all[0].History))
	}

	stats := mgr.Stats()
	if stats.Total != 1 || stats.Canceled != 1 || stats.Running != 0 || stats.Completed != 0 || stats.Failed != 0 {
		t.Fatalf("unexpected stats after cancel: %+v", stats)
	}
}

func TestManagerStatsCountsStates(t *testing.T) {
	mgr := NewManager(nil)
	mgr.tasks["r"] = &Task{ID: "r", Status: StatusRunning}
	mgr.tasks["c"] = &Task{ID: "c", Status: StatusCompleted}
	mgr.tasks["f"] = &Task{ID: "f", Status: StatusFailed}
	mgr.tasks["x"] = &Task{ID: "x", Status: StatusCanceled}

	stats := mgr.Stats()
	if stats.Total != 4 || stats.Running != 1 || stats.Completed != 1 || stats.Failed != 1 || stats.Canceled != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}

func TestManagerRejectsTerminalStatusTransition(t *testing.T) {
	mgr := NewManager(nil)
	mgr.tasks["x"] = &Task{ID: "x", Status: StatusCompleted}

	next := StatusRunning
	updated, ok := mgr.Update("x", UpdateParams{Status: &next})
	if !ok {
		t.Fatalf("expected update to find task")
	}
	if updated.Status != StatusCompleted {
		t.Fatalf("terminal task should remain completed, got %q", updated.Status)
	}
}

func TestManagerUpdateMetadataAndFields(t *testing.T) {
	mgr := NewManager(nil)
	mgr.tasks["x"] = &Task{ID: "x", Status: StatusRunning}

	subject := "Ship"
	desc := "Deploy"
	active := "Deploying"
	owner := "dev1"
	meta := map[string]any{"priority": "high"}

	updated, ok := mgr.Update("x", UpdateParams{
		Subject:     &subject,
		Description: &desc,
		ActiveForm:  &active,
		Owner:       &owner,
		Metadata:    meta,
	})
	if !ok {
		t.Fatalf("expected update to find task")
	}
	if updated.Subject != subject || updated.Description != desc || updated.ActiveForm != active || updated.Owner != owner {
		t.Fatalf("unexpected updated fields: %+v", updated)
	}
	if updated.Metadata["priority"] != "high" {
		t.Fatalf("expected metadata priority high, got %+v", updated.Metadata)
	}
	meta["priority"] = "low"
	if updated.Metadata["priority"] != "high" {
		t.Fatalf("metadata must be copied defensively")
	}
}
