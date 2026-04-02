package agent

import (
	"context"
	"errors"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestRecoverPromptTooLongCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	a := New(Config{Provider: &scriptedProvider{}})
	_, err := newRecoveryCoordinator(a).RecoverPromptTooLong(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}

	view := a.QueryReplay(types.AgentReplayQuery{Types: []types.AgentEventType{types.AgentEventRetry}, Limit: 1})
	if len(view.Events) != 1 {
		t.Fatalf("expected one retry event, got %d", len(view.Events))
	}
	if view.Events[0].Recovery.Branch != "prompt_too_long" {
		t.Fatalf("branch = %q, want prompt_too_long", view.Events[0].Recovery.Branch)
	}
}
