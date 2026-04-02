package commands

import (
	"context"
	"strings"
	"testing"
)

func TestHistoryListCompoundFilterAndLimit(t *testing.T) {
	cmd := NewHistoryCommand()
	state := &RuntimeState{HistoryEntries: []HistoryEntry{
		{ID: "h1", Model: "gpt-4o", Title: "bugfix", CreatedAt: "2026-01-01T00:00:00Z"},
		{ID: "h2", Model: "gpt-4o", Title: "feature", CreatedAt: "2026-01-02T00:00:00Z"},
		{ID: "h3", Model: "gpt-4o-mini", Title: "bugfix", CreatedAt: "2026-01-03T00:00:00Z"},
	}}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "history", Args: []string{"list", "model", "gpt-4o", "text", "feature", "limit", "1"}})
	if err != nil {
		t.Fatalf("history compound filter failed: %v", err)
	}
	if !strings.Contains(res.Message, "filter.type=compound") || !strings.Contains(res.Message, "count=1") || !strings.Contains(res.Message, "entry.1.id=h2") {
		t.Fatalf("unexpected compound filter output: %q", res.Message)
	}
}
