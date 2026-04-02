package commands

import (
	"context"
	"strings"
	"testing"
)

func TestHistoryListCompoundWithLatestLimit(t *testing.T) {
	cmd := NewHistoryCommand()
	state := &RuntimeState{HistoryEntries: []HistoryEntry{
		{ID: "a1", Model: "gpt-4o", Title: "alpha", CreatedAt: "2026-01-01T00:00:00Z"},
		{ID: "a2", Model: "gpt-4o", Title: "beta", CreatedAt: "2026-01-02T00:00:00Z"},
	}}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "history", Args: []string{"list", "model", "gpt-4o", "latest"}})
	if err != nil {
		t.Fatalf("history compound latest failed: %v", err)
	}
	if !strings.Contains(res.Message, "limit=1") || !strings.Contains(res.Message, "entry.1.id=a2") {
		t.Fatalf("unexpected output: %q", res.Message)
	}
}
