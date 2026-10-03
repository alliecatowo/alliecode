package tasks

import (
	"testing"
	"time"
)

func TestStatusSummaryRecentTerminalWave6(t *testing.T) {
	now := time.Now().UTC()
	out := newStatusSummary([]Task{{ID: "a", Status: StatusRunning, UpdatedAt: now}, {ID: "b", Status: StatusCompleted, UpdatedAt: now.Add(time.Second)}})
	if out.RecentTask != "b" || out.RecentTerminalTask != "b" || out.Terminal != 1 {
		t.Fatalf("unexpected summary: %+v", out)
	}
}
