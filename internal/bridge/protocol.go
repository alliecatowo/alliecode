package bridge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

const defaultAckTimeout = 5 * time.Second

var errAckTimeout = errors.New("ack wait timed out")

// ProtocolType identifies one typed bridge protocol message.
type ProtocolType string

const (
	ProtocolTypeData       ProtocolType = "data"
	ProtocolTypeAck        ProtocolType = "ack"
	ProtocolTypePermission ProtocolType = "permission"
	ProtocolTypeHeartbeat  ProtocolType = "heartbeat"
	ProtocolTypeRequest    ProtocolType = "request"
	ProtocolTypeResponse   ProtocolType = "response"
	ProtocolTypeCancel     ProtocolType = "cancel"
)

// RequestWrapper is the canonical request payload inside request envelopes.
type RequestWrapper struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

// ResponseWrapper is the canonical response payload for request envelopes.
type ResponseWrapper struct {
	RequestID string          `json:"request_id"`
	Result    json.RawMessage `json:"result,omitempty"`
	Error     *ResponseError  `json:"error,omitempty"`
}

// ResponseError captures deterministic protocol response errors.
type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// CancelWrapper is the canonical cancel payload for a request envelope.
type CancelWrapper struct {
	RequestID string `json:"request_id"`
	Reason    string `json:"reason,omitempty"`
}

// OutboundEnvelope is the typed message sent to transport.
type OutboundEnvelope[T any] struct {
	SessionID string       `json:"session_id"`
	From      string       `json:"from"`
	To        string       `json:"to"`
	Type      ProtocolType `json:"type"`
	Body      T            `json:"body"`
	RequestID string       `json:"request_id,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
}

// InboundEnvelope is the typed message delivered from transport.
type InboundEnvelope[T any] struct {
	SessionID string       `json:"session_id"`
	From      string       `json:"from"`
	To        string       `json:"to"`
	Type      ProtocolType `json:"type"`
	Body      T            `json:"body"`
	RequestID string       `json:"request_id,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
}

// EnvelopeHeader captures common fields for routing.
type EnvelopeHeader struct {
	SessionID string
	From      string
	To        string
	Type      ProtocolType
	RequestID string
	CreatedAt time.Time
}

// EncodeOutboundEnvelope converts typed outbound data into transport envelope.
func EncodeOutboundEnvelope[T any](env OutboundEnvelope[T]) (Envelope, error) {
	body, err := json.Marshal(env.Body)
	if err != nil {
		return Envelope{}, fmt.Errorf("marshal outbound body: %w", err)
	}
	createdAt := env.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	wire := Envelope{
		ID:            env.RequestID,
		SessionID:     env.SessionID,
		From:          env.From,
		To:            env.To,
		Kind:          MessageKind(env.Type),
		Payload:       body,
		CorrelationID: env.RequestID,
		CreatedAt:     createdAt,
	}
	if err := ValidateEnvelope(wire); err != nil {
		return Envelope{}, err
	}
	return wire, nil
}

// DecodeInboundEnvelope converts wire envelope into a typed inbound envelope.
func DecodeInboundEnvelope[T any](wire Envelope) (InboundEnvelope[T], error) {
	if err := ValidateEnvelope(wire); err != nil {
		return InboundEnvelope[T]{}, err
	}
	var body T
	if len(wire.Payload) > 0 {
		if err := json.Unmarshal(wire.Payload, &body); err != nil {
			return InboundEnvelope[T]{}, fmt.Errorf("decode inbound body: %w", err)
		}
	}
	return InboundEnvelope[T]{
		SessionID: wire.SessionID,
		From:      wire.From,
		To:        wire.To,
		Type:      ProtocolType(wire.Kind),
		Body:      body,
		RequestID: correlatedRequestID(wire),
		CreatedAt: wire.CreatedAt,
	}, nil
}

// EncodeRequestEnvelope validates and encodes a protocol request envelope.
func EncodeRequestEnvelope(env OutboundEnvelope[RequestWrapper]) (Envelope, error) {
	env.Type = ProtocolTypeRequest
	if err := ValidateRequestWrapper(env.Body); err != nil {
		return Envelope{}, err
	}
	return EncodeOutboundEnvelope(env)
}

// DecodeRequestEnvelope validates and decodes a protocol request envelope.
func DecodeRequestEnvelope(wire Envelope) (InboundEnvelope[RequestWrapper], error) {
	if ProtocolType(wire.Kind) != ProtocolTypeRequest {
		return InboundEnvelope[RequestWrapper]{}, fmt.Errorf("expected request envelope kind %q, got %q", ProtocolTypeRequest, wire.Kind)
	}
	decoded, err := DecodeInboundEnvelope[RequestWrapper](wire)
	if err != nil {
		return InboundEnvelope[RequestWrapper]{}, err
	}
	if err := ValidateRequestWrapper(decoded.Body); err != nil {
		return InboundEnvelope[RequestWrapper]{}, err
	}
	return decoded, nil
}

// EncodeResponseEnvelope validates and encodes a protocol response envelope.
func EncodeResponseEnvelope(env OutboundEnvelope[ResponseWrapper]) (Envelope, error) {
	env.Type = ProtocolTypeResponse
	if err := ValidateResponseWrapper(env.Body); err != nil {
		return Envelope{}, err
	}
	if strings.TrimSpace(env.RequestID) == "" {
		env.RequestID = env.Body.RequestID
	}
	return EncodeOutboundEnvelope(env)
}

// DecodeResponseEnvelope validates and decodes a protocol response envelope.
func DecodeResponseEnvelope(wire Envelope) (InboundEnvelope[ResponseWrapper], error) {
	if ProtocolType(wire.Kind) != ProtocolTypeResponse {
		return InboundEnvelope[ResponseWrapper]{}, fmt.Errorf("expected response envelope kind %q, got %q", ProtocolTypeResponse, wire.Kind)
	}
	decoded, err := DecodeInboundEnvelope[ResponseWrapper](wire)
	if err != nil {
		return InboundEnvelope[ResponseWrapper]{}, err
	}
	if err := ValidateResponseWrapper(decoded.Body); err != nil {
		return InboundEnvelope[ResponseWrapper]{}, err
	}
	return decoded, nil
}

// EncodeCancelEnvelope validates and encodes a protocol cancel envelope.
func EncodeCancelEnvelope(env OutboundEnvelope[CancelWrapper]) (Envelope, error) {
	env.Type = ProtocolTypeCancel
	if err := ValidateCancelWrapper(env.Body); err != nil {
		return Envelope{}, err
	}
	if strings.TrimSpace(env.RequestID) == "" {
		env.RequestID = env.Body.RequestID
	}
	return EncodeOutboundEnvelope(env)
}

// DecodeCancelEnvelope validates and decodes a protocol cancel envelope.
func DecodeCancelEnvelope(wire Envelope) (InboundEnvelope[CancelWrapper], error) {
	if ProtocolType(wire.Kind) != ProtocolTypeCancel {
		return InboundEnvelope[CancelWrapper]{}, fmt.Errorf("expected cancel envelope kind %q, got %q", ProtocolTypeCancel, wire.Kind)
	}
	decoded, err := DecodeInboundEnvelope[CancelWrapper](wire)
	if err != nil {
		return InboundEnvelope[CancelWrapper]{}, err
	}
	if err := ValidateCancelWrapper(decoded.Body); err != nil {
		return InboundEnvelope[CancelWrapper]{}, err
	}
	return decoded, nil
}

// NewRequestLifecycleEnvelope builds one request lifecycle envelope.
func NewRequestLifecycleEnvelope(sessionID, from, to, requestID, method string, params json.RawMessage, createdAt time.Time) (Envelope, error) {
	return EncodeRequestEnvelope(OutboundEnvelope[RequestWrapper]{
		SessionID: sessionID,
		From:      from,
		To:        to,
		RequestID: requestID,
		CreatedAt: createdAt,
		Body: RequestWrapper{
			ID:     requestID,
			Method: method,
			Params: params,
		},
	})
}

// NewResponseLifecycleEnvelope builds one success response lifecycle envelope.
func NewResponseLifecycleEnvelope(sessionID, from, to, requestID string, result json.RawMessage, createdAt time.Time) (Envelope, error) {
	return EncodeResponseEnvelope(OutboundEnvelope[ResponseWrapper]{
		SessionID: sessionID,
		From:      from,
		To:        to,
		RequestID: requestID,
		CreatedAt: createdAt,
		Body: ResponseWrapper{
			RequestID: requestID,
			Result:    result,
		},
	})
}

// NewErrorResponseLifecycleEnvelope builds one error response lifecycle envelope.
func NewErrorResponseLifecycleEnvelope(sessionID, from, to, requestID, code, message string, createdAt time.Time) (Envelope, error) {
	return EncodeResponseEnvelope(OutboundEnvelope[ResponseWrapper]{
		SessionID: sessionID,
		From:      from,
		To:        to,
		RequestID: requestID,
		CreatedAt: createdAt,
		Body: ResponseWrapper{
			RequestID: requestID,
			Error: &ResponseError{
				Code:    code,
				Message: message,
			},
		},
	})
}

// NewCancelLifecycleEnvelope builds one cancel lifecycle envelope.
func NewCancelLifecycleEnvelope(sessionID, from, to, requestID, reason string, createdAt time.Time) (Envelope, error) {
	return EncodeCancelEnvelope(OutboundEnvelope[CancelWrapper]{
		SessionID: sessionID,
		From:      from,
		To:        to,
		RequestID: requestID,
		CreatedAt: createdAt,
		Body: CancelWrapper{
			RequestID: requestID,
			Reason:    reason,
		},
	})
}

// ValidateRequestWrapper validates deterministic request wrapper requirements.
func ValidateRequestWrapper(req RequestWrapper) error {
	if strings.TrimSpace(req.ID) == "" {
		return errors.New("request wrapper id cannot be empty")
	}
	if strings.TrimSpace(req.ID) != req.ID {
		return errors.New("request wrapper id cannot include leading or trailing whitespace")
	}
	if strings.TrimSpace(req.Method) == "" {
		return errors.New("request wrapper method cannot be empty")
	}
	if strings.TrimSpace(req.Method) != req.Method {
		return errors.New("request wrapper method cannot include leading or trailing whitespace")
	}
	return nil
}

// ValidateResponseWrapper validates deterministic response wrapper requirements.
func ValidateResponseWrapper(resp ResponseWrapper) error {
	if strings.TrimSpace(resp.RequestID) == "" {
		return errors.New("response wrapper request_id cannot be empty")
	}
	if strings.TrimSpace(resp.RequestID) != resp.RequestID {
		return errors.New("response wrapper request_id cannot include leading or trailing whitespace")
	}
	if resp.Error != nil {
		if strings.TrimSpace(resp.Error.Code) == "" {
			return errors.New("response wrapper error code cannot be empty")
		}
		if strings.TrimSpace(resp.Error.Message) == "" {
			return errors.New("response wrapper error message cannot be empty")
		}
		if len(resp.Result) > 0 {
			return errors.New("response wrapper cannot include both result and error")
		}
	}
	return nil
}

// ValidateCancelWrapper validates deterministic cancel wrapper requirements.
func ValidateCancelWrapper(cancel CancelWrapper) error {
	if strings.TrimSpace(cancel.RequestID) == "" {
		return errors.New("cancel wrapper request_id cannot be empty")
	}
	if strings.TrimSpace(cancel.RequestID) != cancel.RequestID {
		return errors.New("cancel wrapper request_id cannot include leading or trailing whitespace")
	}
	if strings.TrimSpace(cancel.Reason) != cancel.Reason {
		return errors.New("cancel wrapper reason cannot include leading or trailing whitespace")
	}
	return nil
}

// AckCorrelator waits for ack envelopes by request correlation ID.
type AckCorrelator struct {
	mu      sync.Mutex
	waiters map[string]chan Envelope
}

// NewAckCorrelator creates an empty ack correlator.
func NewAckCorrelator() *AckCorrelator {
	return &AckCorrelator{waiters: make(map[string]chan Envelope)}
}

// Wait blocks until an ack is resolved for one request ID.
func (a *AckCorrelator) Wait(ctx context.Context, requestID string, timeout time.Duration) (Envelope, error) {
	if a == nil {
		return Envelope{}, errors.New("ack correlator cannot be nil")
	}
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return Envelope{}, errors.New("request ID cannot be empty")
	}
	if timeout <= 0 {
		timeout = defaultAckTimeout
	}

	ch := make(chan Envelope, 1)
	a.mu.Lock()
	if _, exists := a.waiters[requestID]; exists {
		a.mu.Unlock()
		return Envelope{}, fmt.Errorf("ack waiter already registered for request ID %q", requestID)
	}
	a.waiters[requestID] = ch
	a.mu.Unlock()

	defer func() {
		a.mu.Lock()
		registered, exists := a.waiters[requestID]
		if exists && registered == ch {
			delete(a.waiters, requestID)
		}
		a.mu.Unlock()
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return Envelope{}, ctx.Err()
	case env := <-ch:
		return env, nil
	case <-timer.C:
		return Envelope{}, fmt.Errorf("%w: %s", errAckTimeout, requestID)
	}
}

// Resolve delivers an ack envelope to one waiting request ID.
func (a *AckCorrelator) Resolve(env Envelope) bool {
	if a == nil {
		return false
	}
	if ProtocolType(env.Kind) != ProtocolTypeAck {
		return false
	}
	requestID := correlatedRequestID(env)
	if requestID == "" {
		return false
	}

	a.mu.Lock()
	ch, ok := a.waiters[requestID]
	if ok {
		delete(a.waiters, requestID)
	}
	a.mu.Unlock()
	if !ok {
		return false
	}

	select {
	case ch <- env:
	default:
	}
	return true
}

// RouteHandler handles one inbound payload type.
type RouteHandler func(ctx context.Context, header EnvelopeHeader, payload json.RawMessage) error

// Router dispatches inbound messages by protocol type.
type Router struct {
	mu       sync.RWMutex
	handlers map[ProtocolType]RouteHandler
	unknown  RouteHandler
}

// NewRouter creates an empty protocol router.
func NewRouter() *Router {
	return &Router{handlers: make(map[ProtocolType]RouteHandler)}
}

// Register binds a handler to one protocol type.
func (r *Router) Register(kind ProtocolType, handler RouteHandler) error {
	if kind == "" {
		return errors.New("protocol type cannot be empty")
	}
	if handler == nil {
		return errors.New("route handler cannot be nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[kind] = handler
	return nil
}

// SetUnknownHandler sets fallback handler for unknown protocol types.
func (r *Router) SetUnknownHandler(handler RouteHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.unknown = handler
}

// Route dispatches one wire envelope to registered handlers.
func (r *Router) Route(ctx context.Context, wire Envelope) error {
	if err := ValidateEnvelope(wire); err != nil {
		return err
	}
	header := EnvelopeHeader{
		SessionID: wire.SessionID,
		From:      wire.From,
		To:        wire.To,
		Type:      ProtocolType(wire.Kind),
		RequestID: correlatedRequestID(wire),
		CreatedAt: wire.CreatedAt,
	}

	r.mu.RLock()
	handler, ok := r.handlers[header.Type]
	unknown := r.unknown
	r.mu.RUnlock()

	if ok {
		return handler(ctx, header, json.RawMessage(wire.Payload))
	}
	if unknown != nil {
		return unknown(ctx, header, json.RawMessage(wire.Payload))
	}
	return fmt.Errorf("no route for protocol type %q", header.Type)
}

func correlatedRequestID(wire Envelope) string {
	if correlationID := strings.TrimSpace(wire.CorrelationID); correlationID != "" {
		return correlationID
	}
	return strings.TrimSpace(wire.ID)
}
