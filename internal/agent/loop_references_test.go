package agent

import (
	"context"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestRunEmitsInputProcessedEventWithResolvedReferences(t *testing.T) {
	provider := &scriptedProvider{scripts: []func(chan types.StreamEvent){
		func(ch chan types.StreamEvent) {
			ch <- types.StreamEvent{Type: types.StreamContentDelta, Delta: "done"}
			ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopEndTurn}
		},
	}}

	a := New(Config{Provider: provider, MaxTurns: 1, WorkingDir: t.TempDir()})
	var inputEvent *types.AgentEvent
	a.SetEventCallback(func(ev types.AgentEvent) {
		if ev.Type == types.AgentEventInputProcessed {
			cp := ev
			inputEvent = &cp
		}
	})

	if err := a.Run(context.Background(), "please check @missing.txt and @src/main.go:12"); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if inputEvent == nil {
		t.Fatalf("expected input_processed event")
	}
	if inputEvent.Input == "" {
		t.Fatalf("expected input payload on input_processed event")
	}
	if len(inputEvent.ResolvedReferences) != 2 {
		t.Fatalf("len(resolved_refs) = %d, want 2", len(inputEvent.ResolvedReferences))
	}
	if inputEvent.ResolvedReferences[1].Line != 12 {
		t.Fatalf("line = %d, want 12", inputEvent.ResolvedReferences[1].Line)
	}
}
