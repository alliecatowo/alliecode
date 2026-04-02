package agent

import (
	"context"
	"errors"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestRunCanceledContextEmitsCanceledStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	a := New(Config{Provider: &scriptedProvider{}, MaxTurns: 1})
	err := a.Run(ctx, "hello")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}

	view := a.QueryReplay(types.AgentReplayQuery{Types: []types.AgentEventType{types.AgentEventStop}, Limit: 1})
	if len(view.Events) != 1 {
		t.Fatalf("expected one stop event, got %d", len(view.Events))
	}
	if view.Events[0].StopReason != types.AgentStopCanceled {
		t.Fatalf("stop_reason = %q, want %q", view.Events[0].StopReason, types.AgentStopCanceled)
	}
}
