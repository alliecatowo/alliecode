package agent

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTurnPlannerBeginTurnCanceled(t *testing.T) {
	a := New(Config{Provider: &scriptedProvider{}, MaxTurns: 1})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := newTurnPlanner(a).BeginTurn(ctx, 1)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("BeginTurn() err = %v, want context canceled", err)
	}
}

func TestTurnPlannerEmitsTurnPlan(t *testing.T) {
	provider := &scriptedProvider{scripts: []func(chan types.StreamEvent){
		func(ch chan types.StreamEvent) {
			ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopEndTurn}
		},
	}}
	a := New(Config{Provider: provider, MaxTurns: 1})
	var got types.AgentEvent
	a.SetEventCallback(func(ev types.AgentEvent) {
		if ev.Type == types.AgentEventTurnStart {
			got = ev
		}
	})
	if err := a.Run(context.Background(), "hello"); err != nil {
		t.Fatalf("Run() err = %v", err)
	}
	if !got.TurnPlan.BudgetCheck || !got.TurnPlan.CompactionCheck || got.TurnPlan.Turn != 1 {
		t.Fatalf("unexpected turn plan: %+v", got.TurnPlan)
	}
}

func TestEvaluateContinuationRecovery(t *testing.T) {
	msg := types.NewTextMessage(types.RoleAssistant, strings.Repeat("partial ", 12))
	action := evaluateContinuationRecovery(types.StopMaxTokens, msg, 0, 2)
	if !action.ShouldRetry || action.NextRecoveryCount != 1 {
		t.Fatalf("unexpected action: %+v", action)
	}
}
