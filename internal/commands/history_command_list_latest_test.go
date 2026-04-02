package commands

import (
	"context"
	"strings"
	"testing"
)

func TestHistoryListLatestFilter(t *testing.T) {
	cmd := NewHistoryCommand()
	state := &RuntimeState{HistoryEntries: []HistoryEntry{{ID: "a", CreatedAt: "2026-01-01T00:00:00Z"}, {ID: "b", CreatedAt: "2026-01-02T00:00:00Z"}}}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "history", Args: []string{"list", "latest"}})
	if err != nil {
		t.Fatalf("history list latest failed: %v", err)
	}
	if !strings.Contains(res.Message, "count=1") || !strings.Contains(res.Message, "entry.1.id=b") {
		t.Fatalf("unexpected list latest output: %q", res.Message)
	}
}
