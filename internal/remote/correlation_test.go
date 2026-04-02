package remote

import (
	"encoding/json"
	"testing"
	"time"
)

func TestRequestCorrelatorRetrySafeTransitions(t *testing.T) {
	c := NewRequestCorrelator()
	now := time.Now().UTC()

	reqPayload, _ := json.Marshal(ActionRequestPayload{RequestID: "req-1", Method: "tools/call"})
	if err := c.Observe(RemoteMessage{Kind: MessageKindRequest, Seq: 1, SentAt: now, Payload: reqPayload}); err != nil {
		t.Fatalf("Observe(request#1) error = %v", err)
	}
	if err := c.Observe(RemoteMessage{Kind: MessageKindRequest, Seq: 2, SentAt: now.Add(time.Millisecond), Payload: reqPayload}); err != nil {
		t.Fatalf("Observe(request#2 retry) error = %v", err)
	}

	respPayload, _ := json.Marshal(ActionResponsePayload{RequestID: "req-1", Result: json.RawMessage(`{"ok":true}`)})
	if err := c.Observe(RemoteMessage{Kind: MessageKindResponse, Seq: 3, SentAt: now.Add(2 * time.Millisecond), Payload: respPayload}); err != nil {
		t.Fatalf("Observe(response) error = %v", err)
	}

	stats := c.Stats()
	if stats.Pending != 0 || stats.Responded != 1 || stats.Canceled != 0 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
	if stats.RetryAttempts != 1 {
		t.Fatalf("RetryAttempts = %d, want 1", stats.RetryAttempts)
	}

	snap, ok := c.Snapshot("req-1")
	if !ok {
		t.Fatalf("Snapshot(req-1) missing")
	}
	if snap.State != RequestLifecycleStateResponded || snap.Attempts != 2 || snap.LastSeq != 3 {
		t.Fatalf("unexpected snapshot: %+v", snap)
	}

	if err := c.Observe(RemoteMessage{Kind: MessageKindCancel, Seq: 4, SentAt: now.Add(3 * time.Millisecond), Payload: mustJSON(t, ActionCancelPayload{RequestID: "req-1", Reason: "late"})}); err == nil {
		t.Fatalf("expected terminal conflict error for late cancel")
	}
}

func TestRequestCorrelatorReplayAndOrphanTerminals(t *testing.T) {
	c := NewRequestCorrelator()
	now := time.Now().UTC()

	respPayload := mustJSON(t, ActionResponsePayload{RequestID: "orphan", Error: "denied"})
	if err := c.Observe(RemoteMessage{Kind: MessageKindResponse, Seq: 10, SentAt: now, Payload: respPayload}); err != nil {
		t.Fatalf("Observe(orphan response) error = %v", err)
	}
	if err := c.Observe(RemoteMessage{Kind: MessageKindResponse, Seq: 9, SentAt: now.Add(time.Millisecond), Payload: respPayload}); err != nil {
		t.Fatalf("Observe(replay response) error = %v", err)
	}

	stats := c.Stats()
	if stats.Orphaned != 1 || stats.ReplayDrops != 1 || stats.Responded != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	snap, ok := c.Snapshot("orphan")
	if !ok || !snap.OrphanTerminal || snap.State != RequestLifecycleStateResponded {
		t.Fatalf("unexpected orphan snapshot: %+v ok=%t", snap, ok)
	}
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	return b
}
