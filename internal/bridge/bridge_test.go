package bridge

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNormalizeSessionIDCompatibility(t *testing.T) {
	t.Run("plain", func(t *testing.T) {
		got, err := NormalizeSessionID("abc-123")
		if err != nil {
			t.Fatalf("NormalizeSessionID() error = %v", err)
		}
		if got != "abc-123" {
			t.Fatalf("NormalizeSessionID() = %q, want %q", got, "abc-123")
		}
	})

	t.Run("legacy prefix", func(t *testing.T) {
		got, err := NormalizeSessionID("session:legacy_1")
		if err != nil {
			t.Fatalf("NormalizeSessionID() error = %v", err)
		}
		if got != "legacy_1" {
			t.Fatalf("NormalizeSessionID() = %q, want %q", got, "legacy_1")
		}
	})

	t.Run("pointer", func(t *testing.T) {
		pointer, err := EncodeSessionPointer("pointer-id")
		if err != nil {
			t.Fatalf("EncodeSessionPointer() error = %v", err)
		}
		got, err := NormalizeSessionID(pointer)
		if err != nil {
			t.Fatalf("NormalizeSessionID() error = %v", err)
		}
		if got != "pointer-id" {
			t.Fatalf("NormalizeSessionID() = %q, want %q", got, "pointer-id")
		}
	})

	t.Run("invalid", func(t *testing.T) {
		if _, err := NormalizeSessionID("bad space"); err == nil || err.Error() != `invalid session ID "bad space"` {
			t.Fatalf("expected deterministic invalid session ID error, got %v", err)
		}
	})

	t.Run("empty after legacy prefix", func(t *testing.T) {
		if _, err := NormalizeSessionID("session:"); !errors.Is(err, errSessionIDEmpty) {
			t.Fatalf("expected empty session ID error, got %v", err)
		}
	})

	t.Run("compat and infra transforms", func(t *testing.T) {
		if got := ToCompatSessionID("cse_123"); got != "session_123" {
			t.Fatalf("ToCompatSessionID() = %q, want %q", got, "session_123")
		}
		if got := ToInfraSessionID("session_123"); got != "cse_123" {
			t.Fatalf("ToInfraSessionID() = %q, want %q", got, "cse_123")
		}
		if got := ToCompatSessionID("plain-id"); got != "plain-id" {
			t.Fatalf("ToCompatSessionID(no-op) = %q", got)
		}
		if got := ToInfraSessionID("plain-id"); got != "plain-id" {
			t.Fatalf("ToInfraSessionID(no-op) = %q", got)
		}
		if got := ToCompatSessionID(" cse_123 "); got != " cse_123 " {
			t.Fatalf("ToCompatSessionID(whitespace no-op) = %q", got)
		}
		if got := ToInfraSessionID(" session_123 "); got != " session_123 " {
			t.Fatalf("ToInfraSessionID(whitespace no-op) = %q", got)
		}
	})
}

func TestValidateEnvelopeStrictCompatibility(t *testing.T) {
	now := time.Now().UTC()

	valid := Envelope{
		ID:            "req-1",
		SessionID:     "session:legacy_1",
		From:          "remote",
		To:            "local",
		Kind:          MessageKindData,
		CorrelationID: "req-1",
		CreatedAt:     now,
	}
	if err := ValidateEnvelope(valid); err != nil {
		t.Fatalf("ValidateEnvelope(valid) error = %v", err)
	}

	t.Run("session-id wrapped errors", func(t *testing.T) {
		badSession := valid
		badSession.SessionID = "bad space"
		err := ValidateEnvelope(badSession)
		if err == nil || !strings.Contains(err.Error(), `envelope session_id invalid: invalid session ID "bad space"`) {
			t.Fatalf("expected wrapped deterministic session error, got %v", err)
		}

		badSession = valid
		badSession.SessionID = " sid-1 "
		err = ValidateEnvelope(badSession)
		if err == nil || err.Error() != `envelope session_id " sid-1 " has leading or trailing whitespace` {
			t.Fatalf("expected deterministic session whitespace error, got %v", err)
		}
	})

	pointer, err := EncodeSessionPointer("pointer-session")
	if err != nil {
		t.Fatalf("EncodeSessionPointer() error = %v", err)
	}
	valid.SessionID = pointer
	if err := ValidateEnvelope(valid); err != nil {
		t.Fatalf("ValidateEnvelope(pointer session) error = %v", err)
	}

	valid.Kind = MessageKind(ProtocolTypePermission)
	if err := ValidateEnvelope(valid); err != nil {
		t.Fatalf("ValidateEnvelope(permission kind) error = %v", err)
	}

	invalidKind := valid
	invalidKind.Kind = MessageKind("unknown")
	if err := ValidateEnvelope(invalidKind); err == nil {
		t.Fatalf("expected error for unsupported envelope kind")
	}

	zeroCreated := valid
	zeroCreated.CreatedAt = time.Time{}
	if err := ValidateEnvelope(zeroCreated); err == nil {
		t.Fatalf("expected error for zero created_at")
	}

	badID := valid
	badID.ID = " req-1 "
	if err := ValidateEnvelope(badID); err == nil {
		t.Fatalf("expected error for ID whitespace")
	}

	badCorrelation := valid
	badCorrelation.CorrelationID = " req-1 "
	if err := ValidateEnvelope(badCorrelation); err == nil {
		t.Fatalf("expected error for correlation ID whitespace")
	}

	badFrom := valid
	badFrom.From = " remote "
	if err := ValidateEnvelope(badFrom); err == nil || err.Error() != "envelope from cannot include leading or trailing whitespace" {
		t.Fatalf("expected deterministic from whitespace error, got %v", err)
	}

	badTo := valid
	badTo.To = " local "
	if err := ValidateEnvelope(badTo); err == nil || err.Error() != "envelope to cannot include leading or trailing whitespace" {
		t.Fatalf("expected deterministic to whitespace error, got %v", err)
	}
}

func TestSessionPointerRoundTrip(t *testing.T) {
	pointer, err := EncodeSessionPointer("roundtrip-id")
	if err != nil {
		t.Fatalf("EncodeSessionPointer() error = %v", err)
	}

	decoded, err := DecodeSessionPointer(pointer)
	if err != nil {
		t.Fatalf("DecodeSessionPointer() error = %v", err)
	}

	if decoded.SessionID != "roundtrip-id" {
		t.Fatalf("decoded SessionID = %q, want %q", decoded.SessionID, "roundtrip-id")
	}
}

func TestInMemoryTransportRoutesByReceiver(t *testing.T) {
	transport := NewInMemoryTransport()
	ctx := context.Background()

	if err := transport.Publish(ctx, Envelope{
		ID:        "1",
		SessionID: "sid-1",
		From:      "local",
		To:        "remote-a",
		Kind:      MessageKindData,
		Payload:   []byte("a-1"),
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("Publish(remote-a #1) error = %v", err)
	}
	if err := transport.Publish(ctx, Envelope{
		ID:        "2",
		SessionID: "sid-1",
		From:      "local",
		To:        "remote-b",
		Kind:      MessageKindData,
		Payload:   []byte("b-1"),
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("Publish(remote-b) error = %v", err)
	}
	if err := transport.Publish(ctx, Envelope{
		ID:        "3",
		SessionID: "sid-1",
		From:      "local",
		To:        "remote-a",
		Kind:      MessageKindData,
		Payload:   []byte("a-2"),
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("Publish(remote-a #2) error = %v", err)
	}

	batchA, err := transport.Poll(ctx, "remote-a", PollConfig{BatchSize: 8, WaitTimeout: time.Millisecond})
	if err != nil {
		t.Fatalf("Poll(remote-a) error = %v", err)
	}
	if len(batchA) != 2 {
		t.Fatalf("len(remote-a) = %d, want 2", len(batchA))
	}
	if string(batchA[0].Payload) != "a-1" || string(batchA[1].Payload) != "a-2" {
		t.Fatalf("remote-a payload order = %q, %q; want a-1, a-2", string(batchA[0].Payload), string(batchA[1].Payload))
	}

	batchB, err := transport.Poll(ctx, "remote-b", PollConfig{BatchSize: 8, WaitTimeout: time.Millisecond})
	if err != nil {
		t.Fatalf("Poll(remote-b) error = %v", err)
	}
	if len(batchB) != 1 {
		t.Fatalf("len(remote-b) = %d, want 1", len(batchB))
	}
	if string(batchB[0].Payload) != "b-1" {
		t.Fatalf("remote-b payload = %q, want %q", string(batchB[0].Payload), "b-1")
	}

	empty, err := transport.Poll(ctx, "remote-a", PollConfig{WaitTimeout: time.Millisecond})
	if err != nil {
		t.Fatalf("Poll(remote-a empty) error = %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("len(remote-a empty) = %d, want 0", len(empty))
	}
}

func TestNormalizePollConfig(t *testing.T) {
	cfg, err := NormalizePollConfig(PollConfig{})
	if err != nil {
		t.Fatalf("NormalizePollConfig() error = %v", err)
	}
	if cfg.Interval <= 0 || cfg.BatchSize <= 0 || cfg.WaitTimeout <= 0 {
		t.Fatalf("defaults not applied: %+v", cfg)
	}

	if _, err := NormalizePollConfig(PollConfig{WaitTimeout: -time.Millisecond}); err == nil {
		t.Fatalf("expected error for negative wait timeout")
	}
}

func TestRouterRoutesByType(t *testing.T) {
	router := NewRouter()
	var called bool
	if err := router.Register(ProtocolTypePermission, func(ctx context.Context, header EnvelopeHeader, payload json.RawMessage) error {
		called = true
		if header.Type != ProtocolTypePermission {
			t.Fatalf("header type = %q, want %q", header.Type, ProtocolTypePermission)
		}
		if string(payload) != `{"permission":"exec"}` {
			t.Fatalf("payload = %s", string(payload))
		}
		return nil
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	err := router.Route(context.Background(), Envelope{
		SessionID: "sid-router",
		From:      "remote",
		To:        "local",
		Kind:      MessageKind(ProtocolTypePermission),
		Payload:   []byte(`{"permission":"exec"}`),
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("Route() error = %v", err)
	}
	if !called {
		t.Fatalf("expected registered route to be called")
	}

	if err := router.Route(context.Background(), Envelope{
		SessionID: "sid-router",
		From:      "remote",
		To:        "local",
		Kind:      MessageKind(ProtocolTypeData),
		Payload:   []byte("{}"),
		CreatedAt: time.Now().UTC(),
	}); err == nil {
		t.Fatalf("expected error for unknown route")
	}

	router.SetUnknownHandler(func(ctx context.Context, header EnvelopeHeader, payload json.RawMessage) error {
		if header.Type != ProtocolTypeData {
			t.Fatalf("unknown header type = %q, want %q", header.Type, ProtocolTypeData)
		}
		return nil
	})
	if err := router.Route(context.Background(), Envelope{
		SessionID: "sid-router",
		From:      "remote",
		To:        "local",
		Kind:      MessageKind(ProtocolTypeData),
		Payload:   []byte("{}"),
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("Route() with unknown handler error = %v", err)
	}
}

func TestDecodeInboundEnvelopeAckUsesIDForCorrelation(t *testing.T) {
	decoded, err := DecodeInboundEnvelope[json.RawMessage](Envelope{
		ID:        "ack-42",
		SessionID: "sid-router",
		From:      "remote",
		To:        "local",
		Kind:      MessageKindAck,
		Payload:   []byte(`{"ok":true}`),
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("DecodeInboundEnvelope() error = %v", err)
	}
	if decoded.RequestID != "ack-42" {
		t.Fatalf("RequestID = %q, want %q", decoded.RequestID, "ack-42")
	}
}

func TestRouterAckUsesIDForHeaderRequestID(t *testing.T) {
	router := NewRouter()
	if err := router.Register(ProtocolTypeAck, func(ctx context.Context, header EnvelopeHeader, payload json.RawMessage) error {
		if header.RequestID != "ack-1" {
			t.Fatalf("header.RequestID = %q, want ack-1", header.RequestID)
		}
		return nil
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if err := router.Route(context.Background(), Envelope{
		ID:        "ack-1",
		SessionID: "sid-router",
		From:      "remote",
		To:        "local",
		Kind:      MessageKindAck,
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("Route(ack) error = %v", err)
	}
}

func TestProtocolRequestResponseCancelWrappers(t *testing.T) {
	createdAt := time.Now().UTC()

	requestWire, err := EncodeRequestEnvelope(OutboundEnvelope[RequestWrapper]{
		SessionID: "sid-wrapper",
		From:      "local",
		To:        "remote",
		RequestID: "req-1",
		CreatedAt: createdAt,
		Body: RequestWrapper{
			ID:     "req-1",
			Method: "tools/call",
			Params: json.RawMessage(`{"name":"read"}`),
		},
	})
	if err != nil {
		t.Fatalf("EncodeRequestEnvelope() error = %v", err)
	}
	if requestWire.Kind != MessageKindRequest {
		t.Fatalf("request wire kind = %q, want %q", requestWire.Kind, MessageKindRequest)
	}

	decodedReq, err := DecodeRequestEnvelope(requestWire)
	if err != nil {
		t.Fatalf("DecodeRequestEnvelope() error = %v", err)
	}
	if decodedReq.Body.Method != "tools/call" {
		t.Fatalf("decoded request method = %q, want tools/call", decodedReq.Body.Method)
	}

	responseWire, err := EncodeResponseEnvelope(OutboundEnvelope[ResponseWrapper]{
		SessionID: "sid-wrapper",
		From:      "remote",
		To:        "local",
		CreatedAt: createdAt,
		Body: ResponseWrapper{
			RequestID: "req-1",
			Result:    json.RawMessage(`{"ok":true}`),
		},
	})
	if err != nil {
		t.Fatalf("EncodeResponseEnvelope() error = %v", err)
	}
	if responseWire.Kind != MessageKindResponse {
		t.Fatalf("response wire kind = %q, want %q", responseWire.Kind, MessageKindResponse)
	}

	decodedResp, err := DecodeResponseEnvelope(responseWire)
	if err != nil {
		t.Fatalf("DecodeResponseEnvelope() error = %v", err)
	}
	if decodedResp.Body.RequestID != "req-1" {
		t.Fatalf("decoded response request_id = %q, want req-1", decodedResp.Body.RequestID)
	}

	cancelWire, err := EncodeCancelEnvelope(OutboundEnvelope[CancelWrapper]{
		SessionID: "sid-wrapper",
		From:      "local",
		To:        "remote",
		CreatedAt: createdAt,
		Body: CancelWrapper{
			RequestID: "req-1",
			Reason:    "user_cancelled",
		},
	})
	if err != nil {
		t.Fatalf("EncodeCancelEnvelope() error = %v", err)
	}
	if cancelWire.Kind != MessageKindCancel {
		t.Fatalf("cancel wire kind = %q, want %q", cancelWire.Kind, MessageKindCancel)
	}

	decodedCancel, err := DecodeCancelEnvelope(cancelWire)
	if err != nil {
		t.Fatalf("DecodeCancelEnvelope() error = %v", err)
	}
	if decodedCancel.Body.Reason != "user_cancelled" {
		t.Fatalf("decoded cancel reason = %q, want user_cancelled", decodedCancel.Body.Reason)
	}
}

func TestProtocolWrapperValidationDeterministicErrors(t *testing.T) {
	if err := ValidateRequestWrapper(RequestWrapper{ID: " req-1", Method: "tools/call"}); err == nil || err.Error() != "request wrapper id cannot include leading or trailing whitespace" {
		t.Fatalf("unexpected request validation error: %v", err)
	}
	if err := ValidateResponseWrapper(ResponseWrapper{RequestID: "", Result: json.RawMessage(`{"ok":true}`)}); err == nil || err.Error() != "response wrapper request_id cannot be empty" {
		t.Fatalf("unexpected response validation error: %v", err)
	}
	if err := ValidateResponseWrapper(ResponseWrapper{RequestID: "req-1", Result: json.RawMessage(`{"ok":true}`), Error: &ResponseError{Code: "E_FAIL", Message: "boom"}}); err == nil || err.Error() != "response wrapper cannot include both result and error" {
		t.Fatalf("unexpected mixed response validation error: %v", err)
	}
	if err := ValidateCancelWrapper(CancelWrapper{RequestID: " req-1"}); err == nil || err.Error() != "cancel wrapper request_id cannot include leading or trailing whitespace" {
		t.Fatalf("unexpected cancel validation error: %v", err)
	}
}

func TestProtocolLifecycleEnvelopeHelpers(t *testing.T) {
	createdAt := time.Now().UTC()

	req, err := NewRequestLifecycleEnvelope("sid-lifecycle", "local", "remote", "req-10", "tools/call", json.RawMessage(`{"name":"read"}`), createdAt)
	if err != nil {
		t.Fatalf("NewRequestLifecycleEnvelope() error = %v", err)
	}
	if req.Kind != MessageKindRequest || req.CorrelationID != "req-10" {
		t.Fatalf("request helper mismatch: %+v", req)
	}

	resp, err := NewResponseLifecycleEnvelope("sid-lifecycle", "remote", "local", "req-10", json.RawMessage(`{"ok":true}`), createdAt)
	if err != nil {
		t.Fatalf("NewResponseLifecycleEnvelope() error = %v", err)
	}
	if resp.Kind != MessageKindResponse || resp.CorrelationID != "req-10" {
		t.Fatalf("response helper mismatch: %+v", resp)
	}

	errResp, err := NewErrorResponseLifecycleEnvelope("sid-lifecycle", "remote", "local", "req-10", "E_DENIED", "denied", createdAt)
	if err != nil {
		t.Fatalf("NewErrorResponseLifecycleEnvelope() error = %v", err)
	}
	decodedErrResp, err := DecodeResponseEnvelope(errResp)
	if err != nil {
		t.Fatalf("DecodeResponseEnvelope(error helper) error = %v", err)
	}
	if decodedErrResp.Body.Error == nil || decodedErrResp.Body.Error.Code != "E_DENIED" {
		t.Fatalf("error response helper payload mismatch: %+v", decodedErrResp.Body)
	}

	cancel, err := NewCancelLifecycleEnvelope("sid-lifecycle", "local", "remote", "req-10", "user_cancelled", createdAt)
	if err != nil {
		t.Fatalf("NewCancelLifecycleEnvelope() error = %v", err)
	}
	if cancel.Kind != MessageKindCancel || cancel.CorrelationID != "req-10" {
		t.Fatalf("cancel helper mismatch: %+v", cancel)
	}
}

func TestDecodeInboundEnvelopeUsesIDFallbackForNonAck(t *testing.T) {
	decoded, err := DecodeInboundEnvelope[json.RawMessage](Envelope{
		ID:        "req-99",
		SessionID: "sid-router",
		From:      "remote",
		To:        "local",
		Kind:      MessageKindData,
		Payload:   []byte(`{"ok":true}`),
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("DecodeInboundEnvelope() error = %v", err)
	}
	if decoded.RequestID != "req-99" {
		t.Fatalf("RequestID = %q, want %q", decoded.RequestID, "req-99")
	}
}

func TestAckCorrelatorResolveAndTimeout(t *testing.T) {
	correlator := NewAckCorrelator()

	errCh := make(chan error, 1)
	go func() {
		_, err := correlator.Wait(context.Background(), "req-timeout", 10*time.Millisecond)
		errCh <- err
	}()

	err := <-errCh
	if !errors.Is(err, errAckTimeout) {
		t.Fatalf("Wait(timeout) error = %v, want errAckTimeout", err)
	}

	gotCh := make(chan Envelope, 1)
	errCh = make(chan error, 1)
	go func() {
		env, err := correlator.Wait(context.Background(), "req-resolve", time.Second)
		if err != nil {
			errCh <- err
			return
		}
		gotCh <- env
	}()

	deadline := time.Now().Add(250 * time.Millisecond)
	resolved := false
	for time.Now().Before(deadline) {
		if correlator.Resolve(Envelope{
			ID:        "req-resolve",
			SessionID: "sid-ack",
			From:      "remote",
			To:        "local",
			Kind:      MessageKindAck,
			CreatedAt: time.Now().UTC(),
		}) {
			resolved = true
			break
		}
		time.Sleep(time.Millisecond)
	}
	if !resolved {
		t.Fatalf("Resolve() never matched waiting request")
	}

	select {
	case got := <-gotCh:
		if got.ID != "req-resolve" {
			t.Fatalf("resolved envelope ID = %q, want req-resolve", got.ID)
		}
	case err := <-errCh:
		t.Fatalf("Wait(resolve) error = %v", err)
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timed out waiting for resolved ack")
	}

	if correlator.Resolve(Envelope{
		ID:        "ignored",
		SessionID: "sid-ack",
		From:      "remote",
		To:        "local",
		Kind:      MessageKindData,
		CreatedAt: time.Now().UTC(),
	}) {
		t.Fatalf("Resolve() should ignore non-ack envelopes")
	}
}

func TestPendingRequestCorrelatorResponseAndCancel(t *testing.T) {
	correlator := NewPendingRequestCorrelator()

	canceled := 0
	if err := correlator.Register("req-1", func() { canceled++ }); err != nil {
		t.Fatalf("Register(req-1) error = %v", err)
	}
	if !correlator.ResolveResponse(ResponseWrapper{RequestID: "req-1"}) {
		t.Fatalf("ResolveResponse(req-1) = false, want true")
	}
	if canceled != 0 {
		t.Fatalf("cancel callback called on response path")
	}

	if err := correlator.Register("req-2", func() { canceled++ }); err != nil {
		t.Fatalf("Register(req-2) error = %v", err)
	}
	if !correlator.ResolveCancel(CancelWrapper{RequestID: "req-2", Reason: "user_cancelled"}) {
		t.Fatalf("ResolveCancel(req-2) = false, want true")
	}
	if canceled != 1 {
		t.Fatalf("cancel callback count = %d, want 1", canceled)
	}
	if correlator.ResolveCancel(CancelWrapper{RequestID: "req-missing"}) {
		t.Fatalf("ResolveCancel(req-missing) should be false")
	}

	if err := correlator.Register("req-retry", nil); err != nil {
		t.Fatalf("Register(req-retry #1) error = %v", err)
	}
	if err := correlator.Register("req-retry", nil); err != nil {
		t.Fatalf("Register(req-retry #2 retry) error = %v", err)
	}
	snap, ok := correlator.Snapshot("req-retry")
	if !ok {
		t.Fatalf("Snapshot(req-retry) missing")
	}
	if snap.Attempts != 2 || snap.State != RequestLifecycleStatePending {
		t.Fatalf("unexpected retry snapshot: %+v", snap)
	}
	stats := correlator.Stats()
	if stats.RetryAttempts != 1 {
		t.Fatalf("RetryAttempts = %d, want 1", stats.RetryAttempts)
	}
}

type runnerTestSession struct {
	heartbeats int
	stopAfter  int
}

func (s *runnerTestSession) Connect(ctx context.Context) error    { return nil }
func (s *runnerTestSession) Disconnect(ctx context.Context) error { return nil }
func (s *runnerTestSession) Heartbeat(ctx context.Context) error {
	s.heartbeats++
	if s.stopAfter > 0 && s.heartbeats >= s.stopAfter {
		return errors.New("heartbeat failed")
	}
	return nil
}

func TestSessionRunnerCancelsPendingOnDisconnectPath(t *testing.T) {
	session := &runnerTestSession{stopAfter: 1}
	runner, err := NewSessionRunner(session)
	if err != nil {
		t.Fatalf("NewSessionRunner() error = %v", err)
	}
	runner.Heartbeat = time.Millisecond

	canceled := 0
	if err := runner.RegisterPendingRequest("req-1", func() { canceled++ }); err != nil {
		t.Fatalf("RegisterPendingRequest() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if err := runner.Run(ctx); err == nil {
		t.Fatalf("Run() expected error from heartbeat failure")
	}

	if canceled != 1 {
		t.Fatalf("pending cancel callback count = %d, want 1", canceled)
	}
}

func TestPermissionBrokerCallbackFlow(t *testing.T) {
	broker := NewPermissionBroker(PermissionCallbackFunc(func(ctx context.Context, ask PermissionAsk) (PermissionDecision, error) {
		if ask.SessionID != "sid-perm" {
			t.Fatalf("ask SessionID = %q, want sid-perm", ask.SessionID)
		}
		if ask.Permission != "workspace.read" {
			t.Fatalf("ask Permission = %q, want workspace.read", ask.Permission)
		}
		if ask.RequestedAt.IsZero() {
			t.Fatalf("expected RequestedAt to be set")
		}
		return PermissionDecision{Allowed: true, Message: "ok", TTL: time.Minute}, nil
	}))

	decision, err := broker.Ask(context.Background(), PermissionAsk{
		SessionID:  "sid-perm",
		DeviceID:   "dev-1",
		Permission: "workspace.read",
		Reason:     "sync",
	})
	if err != nil {
		t.Fatalf("Ask() error = %v", err)
	}
	if !decision.Allowed || decision.Message != "ok" {
		t.Fatalf("unexpected decision: %+v", decision)
	}

	broker.SetCallback(PermissionCallbackFunc(func(ctx context.Context, ask PermissionAsk) (PermissionDecision, error) {
		return PermissionDecision{}, errors.New("denied")
	}))
	if _, err := broker.Ask(context.Background(), PermissionAsk{
		SessionID:  "sid-perm",
		Permission: "workspace.read",
	}); err == nil {
		t.Fatalf("expected callback error")
	}

	broker.SetCallback(nil)
	if _, err := broker.Ask(context.Background(), PermissionAsk{
		SessionID:  "sid-perm",
		Permission: "workspace.read",
	}); err == nil {
		t.Fatalf("expected error when callback is nil")
	}
}

func TestTrustedDeviceRegistryLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "trusted-devices.json")

	registry, err := NewTrustedDeviceRegistry(path)
	if err != nil {
		t.Fatalf("NewTrustedDeviceRegistry() error = %v", err)
	}

	stored, err := registry.Add(TrustedDevice{DeviceID: "dev-b", Label: "tablet"})
	if err != nil {
		t.Fatalf("Add(dev-b) error = %v", err)
	}
	if stored.AddedAt.IsZero() {
		t.Fatalf("expected AddedAt to be populated")
	}
	if _, err := registry.Add(TrustedDevice{DeviceID: "dev-a", Label: "laptop"}); err != nil {
		t.Fatalf("Add(dev-a) error = %v", err)
	}

	listed := registry.List()
	if len(listed) != 2 {
		t.Fatalf("List() len = %d, want 2", len(listed))
	}
	if listed[0].DeviceID != "dev-a" || listed[1].DeviceID != "dev-b" {
		t.Fatalf("List() order = %q, %q; want dev-a, dev-b", listed[0].DeviceID, listed[1].DeviceID)
	}
	if !registry.IsTrusted("dev-a") {
		t.Fatalf("expected dev-a to be trusted")
	}

	if err := registry.Revoke("dev-a"); err != nil {
		t.Fatalf("Revoke(dev-a) error = %v", err)
	}
	if registry.IsTrusted("dev-a") {
		t.Fatalf("expected dev-a to be revoked")
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected persisted file, stat error = %v", err)
	}

	reloaded, err := NewTrustedDeviceRegistry(path)
	if err != nil {
		t.Fatalf("reload registry error = %v", err)
	}
	reloadedList := reloaded.List()
	if len(reloadedList) != 1 || reloadedList[0].DeviceID != "dev-b" {
		t.Fatalf("reloaded List() = %+v, want only dev-b", reloadedList)
	}
}

func TestConfigStoreLoadValidateToggle(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "bridge-config.json")
	store := NewConfigStore(path)

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load(default) error = %v", err)
	}
	if cfg.Enabled {
		t.Fatalf("default Enabled = true, want false")
	}

	saved, err := store.Save(BridgeConfig{
		Enabled: true,
		Poll: PollConfig{
			Interval:    10 * time.Millisecond,
			BatchSize:   4,
			WaitTimeout: 20 * time.Millisecond,
		},
		HeartbeatInterval: 5 * time.Second,
		ReconnectBackoff: BackoffPolicy{
			Initial: 100 * time.Millisecond,
			Max:     time.Second,
			Factor:  1.5,
			Jitter:  0.1,
		},
	})
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if !saved.Enabled {
		t.Fatalf("saved Enabled = false, want true")
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load(saved) error = %v", err)
	}
	if !loaded.Enabled || loaded.Poll.BatchSize != 4 {
		t.Fatalf("loaded config mismatch: %+v", loaded)
	}

	toggled, err := store.SetEnabled(false)
	if err != nil {
		t.Fatalf("SetEnabled(false) error = %v", err)
	}
	if toggled.Enabled {
		t.Fatalf("toggled Enabled = true, want false")
	}
}
