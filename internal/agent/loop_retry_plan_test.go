package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestLoopRetryEventCarriesRetryPlan(t *testing.T) {
	provider := &scriptedProvider{scripts: []func(chan types.StreamEvent){
		func(ch chan types.StreamEvent) {
			ch <- types.StreamEvent{Type: types.StreamContentDelta, Delta: strings.Repeat("a", 120)}
			ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopMaxTokens}
		},
		func(ch chan types.StreamEvent) {
			ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopEndTurn}
		},
	}}
	a := New(Config{Provider: provider, MaxTurns: 3})
	if err := a.Run(context.Background(), "go"); err != nil {
		t.Fatalf("Run() err = %v", err)
	}
	for _, ev := range a.ReplayEvents(0, 0) {
		if ev.Type == types.AgentEventRetry && ev.RetryKind == "max_output_tokens" {
			if ev.RetryPlan.FromPhase != types.AgentTurnPhaseProviderStream || ev.RetryPlan.ToPhase != types.AgentTurnPhaseRetry {
				t.Fatalf("unexpected retry plan: %+v", ev.RetryPlan)
			}
			return
		}
	}
	t.Fatalf("expected retry event with retry plan")
}
