package agent

import (
	"context"
	"sync"
	"testing"

	"github.com/alliecatowo/alliecode/internal/hooks"
	"github.com/alliecatowo/alliecode/internal/types"
)

type hookPayloadSpy struct {
	mu      sync.Mutex
	events  []hooks.Event
	payload []map[string]string
}

func (h *hookPayloadSpy) Fire(_ context.Context, event hooks.Event, data map[string]string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	cp := map[string]string{}
	for k, v := range data {
		cp[k] = v
	}
	h.events = append(h.events, event)
	h.payload = append(h.payload, cp)
	return nil
}

func (h *hookPayloadSpy) HasHooks(_ hooks.Event) bool { return true }

func TestEmitStopAddsStopReasonMetadataToHooks(t *testing.T) {
	provider := &scriptedProvider{scripts: []func(chan types.StreamEvent){func(ch chan types.StreamEvent) {
		ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopError}
	}}}
	spy := &hookPayloadSpy{}
	a := New(Config{Provider: provider, MaxTurns: 1, Hooks: spy})
	if err := a.Run(context.Background(), "hi"); err != nil {
		t.Fatalf("Run() err = %v", err)
	}
	if len(spy.payload) == 0 {
		t.Fatalf("expected hook payloads")
	}
	var stopPayload map[string]string
	for i, ev := range spy.events {
		if ev == hooks.EventStop {
			stopPayload = spy.payload[i]
			break
		}
	}
	if stopPayload == nil {
		t.Fatalf("expected stop hook payload")
	}
	if stopPayload["STOP_REASON_CLASS"] == "" || stopPayload["STOP_REASON_TERMINAL"] == "" {
		t.Fatalf("expected enriched stop payload fields, got %+v", stopPayload)
	}
}
