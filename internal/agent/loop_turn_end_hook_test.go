package agent

import (
	"context"
	"testing"

	"github.com/alliecatowo/alliecode/internal/hooks"
	"github.com/alliecatowo/alliecode/internal/types"
)

func TestEmitStopFiresTurnEndHook(t *testing.T) {
	provider := &scriptedProvider{scripts: []func(chan types.StreamEvent){func(ch chan types.StreamEvent) {
		ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopEndTurn}
	}}}
	spy := &hookSpy{}
	a := New(Config{Provider: provider, MaxTurns: 1, Hooks: spy})
	if err := a.Run(context.Background(), "hi"); err != nil {
		t.Fatalf("Run() err = %v", err)
	}
	if !spy.seen(hooks.EventTurnEnd) {
		t.Fatalf("expected turn_end hook to fire")
	}
}
