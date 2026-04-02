package commands

import (
	"context"
	"strings"
	"testing"
)

func TestHistoryLatestAndLimit(t *testing.T) {
	cmd := NewHistoryCommand()
	state := &RuntimeState{HistoryEntries: []HistoryEntry{{ID: "s1", CreatedAt: "2026-01-01T00:00:00Z", Title: "one"}, {ID: "s2", CreatedAt: "2026-01-02T00:00:00Z", Title: "two"}}}
	latest, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "history", Args: []string{"latest"}})
	if err != nil || !strings.Contains(latest.Message, "id=s2") {
		t.Fatalf("latest failed: %v %q", err, latest.Message)
	}
	list, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "history", Args: []string{"list", "limit", "1"}})
	if err != nil || !strings.Contains(list.Message, "count=1") {
		t.Fatalf("limit failed: %v %q", err, list.Message)
	}
}
