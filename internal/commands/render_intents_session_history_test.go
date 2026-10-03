package commands

import (
	"context"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestSessionStatusEmitsSummaryIntent(t *testing.T) {
	state := &RuntimeState{SessionID: "session-1"}
	res, err := NewSessionCommand().Execute(context.Background(), Context{State: state}, Invocation{Name: "session", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("session status failed: %v", err)
	}
	if len(res.RenderIntents) == 0 || res.RenderIntents[0].Kind != types.RenderIntentSummaryCard {
		t.Fatalf("expected summary card intent, got %#v", res.RenderIntents)
	}
}

func TestHistoryListEmitsTableIntent(t *testing.T) {
	state := &RuntimeState{HistoryEntries: []HistoryEntry{{ID: "a1", Model: "gpt-4o", Turns: 3, Title: "demo"}}}
	res, err := NewHistoryCommand().Execute(context.Background(), Context{State: state}, Invocation{Name: "history", Args: []string{"list"}})
	if err != nil {
		t.Fatalf("history list failed: %v", err)
	}
	if len(res.RenderIntents) == 0 || res.RenderIntents[0].Kind != types.RenderIntentTable {
		t.Fatalf("expected table intent, got %#v", res.RenderIntents)
	}
}
