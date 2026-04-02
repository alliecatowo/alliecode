package commands

import (
	"context"
	"strings"
	"testing"
)

func TestHistoryStatusReportsViewAndFilterState(t *testing.T) {
	cmd := NewHistoryCommand()
	state := &RuntimeState{HistoryEntries: []HistoryEntry{{ID: "a1", Model: "gpt-4o", CreatedAt: "2026-01-01T00:00:00Z"}}, HistoryLastFilterType: "model", HistoryLastFilterValue: "gpt-4o", HistoryLastLimit: 5}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "history", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("history status failed: %v", err)
	}
	if !strings.Contains(res.Message, "HISTORY_STATUS") || !strings.Contains(res.Message, "last_filter_type=model") {
		t.Fatalf("unexpected history status: %q", res.Message)
	}
}
