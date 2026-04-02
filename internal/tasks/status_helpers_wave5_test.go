package tasks

import (
	"testing"
	"time"
)

func TestStatusSummaryTracksRecentTerminalTask(t *testing.T) {
	now := time.Now().UTC()
	summary := newStatusSummary([]Task{
		{ID: "run", Status: StatusRunning, UpdatedAt: now.Add(-time.Minute)},
		{ID: "done", Status: StatusCompleted, UpdatedAt: now},
	})
	if summary.RecentTerminalTask != "done" {
		t.Fatalf("expected recent terminal task to be done, got %q", summary.RecentTerminalTask)
	}
	if summary.RecentTerminalAt == 0 {
		t.Fatalf("expected recent terminal timestamp to be set")
	}
}
