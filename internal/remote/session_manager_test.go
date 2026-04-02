package remote

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type flakyTransport struct {
	*InMemoryTransport
	failConnects int
	connectCalls int
}

func newFlakyTransport(failConnects int) *flakyTransport {
	return &flakyTransport{InMemoryTransport: NewInMemoryTransport(), failConnects: failConnects}
}

func (t *flakyTransport) Connect(ctx context.Context, sessionID string) error {
	t.connectCalls++
	if t.connectCalls <= t.failConnects {
		return errors.New("temporary connect failure")
	}
	return t.InMemoryTransport.Connect(ctx, sessionID)
}

func TestSessionManagerStateTransitions(t *testing.T) {
	transport := NewInMemoryTransport()
	mgr, err := NewSessionManager(transport, "sid-1", RetryPolicy{}, time.Now())
	if err != nil {
		t.Fatalf("NewSessionManager() error = %v", err)
	}

	if got := mgr.State(); got != SessionStateClosed {
		t.Fatalf("initial state = %q, want %q", got, SessionStateClosed)
	}

	if err := mgr.Connect(context.Background()); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if got := mgr.State(); got != SessionStateActive {
		t.Fatalf("state after connect = %q, want %q", got, SessionStateActive)
	}

	if err := mgr.HandleDisconnect(); err != nil {
		t.Fatalf("HandleDisconnect() error = %v", err)
	}
	if got := mgr.State(); got != SessionStateReconnecting {
		t.Fatalf("state after disconnect = %q, want %q", got, SessionStateReconnecting)
	}

	if err := mgr.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if got := mgr.State(); got != SessionStateClosed {
		t.Fatalf("state after close = %q, want %q", got, SessionStateClosed)
	}
}

func TestSessionManagerReconnectSucceedsAfterRetries(t *testing.T) {
	transport := newFlakyTransport(2)
	policy := RetryPolicy{
		InitialDelay: time.Millisecond,
		MaxDelay:     2 * time.Millisecond,
		Multiplier:   1,
		Jitter:       0,
		MaxAttempts:  4,
	}
	mgr, err := NewSessionManager(transport, "sid-2", policy, time.Now())
	if err != nil {
		t.Fatalf("NewSessionManager() error = %v", err)
	}

	if err := mgr.Connect(context.Background()); err == nil {
		t.Fatalf("Connect() expected initial failure")
	}
	if got := mgr.State(); got != SessionStateReconnecting {
		t.Fatalf("state after failed connect = %q, want %q", got, SessionStateReconnecting)
	}

	if err := mgr.Reconnect(context.Background()); err != nil {
		t.Fatalf("Reconnect() error = %v", err)
	}
	if got := mgr.State(); got != SessionStateActive {
		t.Fatalf("state after reconnect = %q, want %q", got, SessionStateActive)
	}
	if transport.connectCalls != 3 {
		t.Fatalf("connect calls = %d, want 3", transport.connectCalls)
	}
}

func TestSessionManagerHeartbeatTriggersReconnect(t *testing.T) {
	now := time.Now().UTC()
	transport := NewInMemoryTransport()
	policy := RetryPolicy{
		InitialDelay: time.Millisecond,
		MaxDelay:     time.Millisecond,
		Multiplier:   1,
		Jitter:       0,
		MaxAttempts:  1,
	}
	mgr, err := NewSessionManager(transport, "sid-3", policy, now)
	if err != nil {
		t.Fatalf("NewSessionManager() error = %v", err)
	}

	if err := mgr.Connect(context.Background()); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	beat := RemoteMessage{
		SessionID: "sid-3",
		Kind:      MessageKindHeartbeat,
		Name:      "tick",
		Seq:       1,
		SentAt:    now.Add(time.Second),
	}
	mgr.TickHeartbeat(beat)

	if tripped := mgr.CheckHeartbeat(now.Add(2500 * time.Millisecond)); tripped {
		t.Fatalf("CheckHeartbeat() unexpectedly tripped before timeout")
	}
	if tripped := mgr.CheckHeartbeat(now.Add(5 * time.Second)); !tripped {
		t.Fatalf("CheckHeartbeat() expected reconnect transition")
	}
	if got := mgr.State(); got != SessionStateReconnecting {
		t.Fatalf("state after heartbeat timeout = %q, want %q", got, SessionStateReconnecting)
	}
}

func TestSessionManagerHandleDisconnectIdempotent(t *testing.T) {
	transport := NewInMemoryTransport()
	mgr, err := NewSessionManager(transport, "sid-idempotent", RetryPolicy{}, time.Now())
	if err != nil {
		t.Fatalf("NewSessionManager() error = %v", err)
	}

	if err := mgr.Connect(context.Background()); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	if err := mgr.HandleDisconnect(); err != nil {
		t.Fatalf("HandleDisconnect(first) error = %v", err)
	}
	if err := mgr.HandleDisconnect(); err != nil {
		t.Fatalf("HandleDisconnect(second) should be idempotent, got %v", err)
	}
	if got := mgr.State(); got != SessionStateReconnecting {
		t.Fatalf("state after repeated disconnect = %q, want %q", got, SessionStateReconnecting)
	}
}

func TestSessionManagerConnectContextCanceledCloses(t *testing.T) {
	transport := NewInMemoryTransport()
	mgr, err := NewSessionManager(transport, "sid-canceled", RetryPolicy{}, time.Now())
	if err != nil {
		t.Fatalf("NewSessionManager() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := mgr.Connect(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Connect() error = %v, want context canceled", err)
	}
	if got := mgr.State(); got != SessionStateClosed {
		t.Fatalf("state after canceled connect = %q, want %q", got, SessionStateClosed)
	}
}

func TestSessionManagerReconnectAbortsWhenClosed(t *testing.T) {
	transport := newFlakyTransport(10)
	policy := RetryPolicy{
		InitialDelay: 5 * time.Millisecond,
		MaxDelay:     5 * time.Millisecond,
		Multiplier:   1,
		Jitter:       0,
		MaxAttempts:  0,
	}
	mgr, err := NewSessionManager(transport, "sid-close-abort", policy, time.Now())
	if err != nil {
		t.Fatalf("NewSessionManager() error = %v", err)
	}

	if err := mgr.Connect(context.Background()); err == nil {
		t.Fatalf("Connect() expected failure")
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- mgr.Reconnect(context.Background())
	}()

	time.Sleep(15 * time.Millisecond)
	if err := mgr.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	err = <-errCh
	if !errors.Is(err, errReconnectAbortedClosed) {
		t.Fatalf("Reconnect() error = %v, want %v", err, errReconnectAbortedClosed)
	}
	if got := mgr.State(); got != SessionStateClosed {
		t.Fatalf("state after close+reconnect = %q, want %q", got, SessionStateClosed)
	}
}

func TestSessionManagerReconnectDelayInterruptedByClose(t *testing.T) {
	transport := newFlakyTransport(10)
	policy := RetryPolicy{
		InitialDelay: 200 * time.Millisecond,
		MaxDelay:     200 * time.Millisecond,
		Multiplier:   1,
		Jitter:       0,
		MaxAttempts:  0,
	}
	mgr, err := NewSessionManager(transport, "sid-close-interrupt", policy, time.Now())
	if err != nil {
		t.Fatalf("NewSessionManager() error = %v", err)
	}

	if err := mgr.Connect(context.Background()); err == nil {
		t.Fatalf("Connect() expected failure")
	}

	errCh := make(chan error, 1)
	start := time.Now()
	go func() {
		errCh <- mgr.Reconnect(context.Background())
	}()

	time.Sleep(20 * time.Millisecond)
	if err := mgr.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	err = <-errCh
	if !errors.Is(err, errReconnectAbortedClosed) {
		t.Fatalf("Reconnect() error = %v, want %v", err, errReconnectAbortedClosed)
	}
	if elapsed := time.Since(start); elapsed >= 180*time.Millisecond {
		t.Fatalf("Reconnect() close interrupt took too long: %v", elapsed)
	}
}

func TestSessionManagerCloseAllowsFreshReconnectCycle(t *testing.T) {
	transport := NewInMemoryTransport()
	mgr, err := NewSessionManager(transport, "sid-fresh-cycle", RetryPolicy{}, time.Now())
	if err != nil {
		t.Fatalf("NewSessionManager() error = %v", err)
	}

	if err := mgr.Connect(context.Background()); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if err := mgr.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	if err := mgr.Connect(context.Background()); err != nil {
		t.Fatalf("Connect() after close error = %v", err)
	}
	if got := mgr.State(); got != SessionStateActive {
		t.Fatalf("state after reconnect cycle = %q, want %q", got, SessionStateActive)
	}
}

func TestSessionManagerReconnectRejectsConcurrentCalls(t *testing.T) {
	transport := newFlakyTransport(10)
	policy := RetryPolicy{
		InitialDelay: 200 * time.Millisecond,
		MaxDelay:     200 * time.Millisecond,
		Multiplier:   1,
		Jitter:       0,
		MaxAttempts:  0,
	}
	mgr, err := NewSessionManager(transport, "sid-concurrent-reconnect", policy, time.Now())
	if err != nil {
		t.Fatalf("NewSessionManager() error = %v", err)
	}

	if err := mgr.Connect(context.Background()); err == nil {
		t.Fatalf("Connect() expected failure")
	}

	errFirst := make(chan error, 1)
	go func() {
		errFirst <- mgr.Reconnect(context.Background())
	}()

	time.Sleep(20 * time.Millisecond)
	err = mgr.Reconnect(context.Background())
	if err == nil || err.Error() != "reconnect already in progress" {
		t.Fatalf("concurrent Reconnect() error = %v, want reconnect already in progress", err)
	}

	if closeErr := mgr.Close(); closeErr != nil {
		t.Fatalf("Close() error = %v", closeErr)
	}

	if err = <-errFirst; !errors.Is(err, errReconnectAbortedClosed) {
		t.Fatalf("first Reconnect() error = %v, want %v", err, errReconnectAbortedClosed)
	}
}

func TestSessionManagerTracksTransportStateAndReconnectCause(t *testing.T) {
	transport := newFlakyTransport(1)
	mgr, err := NewSessionManager(transport, "sid-cause", RetryPolicy{}, time.Now())
	if err != nil {
		t.Fatalf("NewSessionManager() error = %v", err)
	}

	err = mgr.Connect(context.Background())
	if err == nil {
		t.Fatalf("Connect() expected failure")
	}
	if got := mgr.TransportState(); got != TransportStateReconnecting {
		t.Fatalf("TransportState() = %q, want %q", got, TransportStateReconnecting)
	}
	cause := mgr.LastReconnectCause()
	if cause.Reason != ReconnectReasonConnectFailure {
		t.Fatalf("LastReconnectCause().Reason = %q, want %q", cause.Reason, ReconnectReasonConnectFailure)
	}
	if cause.ErrorClass != TransportErrorClassTransient {
		t.Fatalf("LastReconnectCause().ErrorClass = %q, want %q", cause.ErrorClass, TransportErrorClassTransient)
	}

	if err := mgr.Reconnect(context.Background()); err != nil {
		t.Fatalf("Reconnect() error = %v", err)
	}
	if got := mgr.TransportState(); got != TransportStateConnected {
		t.Fatalf("TransportState() after reconnect = %q, want %q", got, TransportStateConnected)
	}
	if cause := mgr.LastReconnectCause(); cause.Reason != ReconnectReasonNone {
		t.Fatalf("LastReconnectCause().Reason after reconnect = %q, want %q", cause.Reason, ReconnectReasonNone)
	}
}

func TestSessionManagerHandleDisconnectWithCause(t *testing.T) {
	transport := NewInMemoryTransport()
	mgr, err := NewSessionManager(transport, "sid-disconnect-cause", RetryPolicy{}, time.Now())
	if err != nil {
		t.Fatalf("NewSessionManager() error = %v", err)
	}
	if err := mgr.Connect(context.Background()); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	err = mgr.HandleDisconnectWithCause(ReconnectCause{
		Reason:     ReconnectReasonDisconnect,
		ErrorClass: TransportErrorClassAuth,
		Detail:     "unauthorized",
	})
	if err != nil {
		t.Fatalf("HandleDisconnectWithCause() error = %v", err)
	}

	cause := mgr.LastReconnectCause()
	if cause.ErrorClass != TransportErrorClassAuth {
		t.Fatalf("LastReconnectCause().ErrorClass = %q, want %q", cause.ErrorClass, TransportErrorClassAuth)
	}
	if got := mgr.TransportState(); got != TransportStateReconnecting {
		t.Fatalf("TransportState() = %q, want %q", got, TransportStateReconnecting)
	}
}

func TestSessionManagerHealthSummaryIncludesLifecycleState(t *testing.T) {
	transport := newFlakyTransport(1)
	mgr, err := NewSessionManager(transport, "sid-health", RetryPolicy{}, time.Now())
	if err != nil {
		t.Fatalf("NewSessionManager() error = %v", err)
	}

	if err := mgr.Connect(context.Background()); err == nil {
		t.Fatalf("Connect() expected failure")
	}

	summary := mgr.HealthSummary()
	if summary.State != SessionStateReconnecting {
		t.Fatalf("HealthSummary().State = %q, want %q", summary.State, SessionStateReconnecting)
	}
	if summary.TransportState != TransportStateReconnecting {
		t.Fatalf("HealthSummary().TransportState = %q, want %q", summary.TransportState, TransportStateReconnecting)
	}
	if summary.LastCause.Reason != ReconnectReasonConnectFailure {
		t.Fatalf("HealthSummary().LastCause.Reason = %q, want %q", summary.LastCause.Reason, ReconnectReasonConnectFailure)
	}
}

func TestSessionManagerRequestCorrelationDiagnostics(t *testing.T) {
	transport := NewInMemoryTransport()
	mgr, err := NewSessionManager(transport, "sid-requests", RetryPolicy{}, time.Now())
	if err != nil {
		t.Fatalf("NewSessionManager() error = %v", err)
	}

	now := time.Now().UTC()
	requestPayload, _ := json.Marshal(ActionRequestPayload{RequestID: "req-1", Method: "tools/call"})
	responsePayload, _ := json.Marshal(ActionResponsePayload{RequestID: "req-1", Result: json.RawMessage(`{"ok":true}`)})

	if err := mgr.ObserveMessage(RemoteMessage{Kind: MessageKindRequest, Seq: 1, SentAt: now, Payload: requestPayload}); err != nil {
		t.Fatalf("ObserveMessage(request) error = %v", err)
	}
	if err := mgr.ObserveMessage(RemoteMessage{Kind: MessageKindRequest, Seq: 2, SentAt: now.Add(time.Millisecond), Payload: requestPayload}); err != nil {
		t.Fatalf("ObserveMessage(request retry) error = %v", err)
	}
	if err := mgr.ObserveMessage(RemoteMessage{Kind: MessageKindResponse, Seq: 3, SentAt: now.Add(2 * time.Millisecond), Payload: responsePayload}); err != nil {
		t.Fatalf("ObserveMessage(response) error = %v", err)
	}

	if pending := mgr.PendingRequestIDs(); len(pending) != 0 {
		t.Fatalf("PendingRequestIDs() = %+v, want empty", pending)
	}

	snap, ok := mgr.RequestSnapshot("req-1")
	if !ok {
		t.Fatalf("RequestSnapshot(req-1) missing")
	}
	if snap.Attempts != 2 || snap.State != RequestLifecycleStateResponded {
		t.Fatalf("unexpected request snapshot: %+v", snap)
	}

	summary := mgr.HealthSummary()
	if summary.RequestStats.Responded != 1 || summary.RequestStats.RetryAttempts != 1 || summary.RequestStats.Pending != 0 {
		t.Fatalf("unexpected request stats: %+v", summary.RequestStats)
	}
}
