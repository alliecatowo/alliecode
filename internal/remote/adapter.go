package remote

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

// ProviderEventPayload carries normalized provider stream event data.
type ProviderEventPayload struct {
	Type      types.StreamEventType `json:"type"`
	Delta     string                `json:"delta,omitempty"`
	ToolUseID string                `json:"tool_use_id,omitempty"`
	ToolName  string                `json:"tool_name,omitempty"`
	Input     json.RawMessage       `json:"input,omitempty"`
	Message   *types.Message        `json:"message,omitempty"`
	Usage     *types.Usage          `json:"usage,omitempty"`
	Error     string                `json:"error,omitempty"`
}

// ToolEventPayload carries normalized tool lifecycle event data.
type ToolEventPayload struct {
	Type               types.AgentEventType          `json:"type"`
	Turn               int                           `json:"turn,omitempty"`
	AssistantChunk     string                        `json:"assistant_chunk,omitempty"`
	ToolUseID          string                        `json:"tool_use_id,omitempty"`
	ToolName           string                        `json:"tool_name,omitempty"`
	ToolInput          json.RawMessage               `json:"tool_input,omitempty"`
	ToolError          string                        `json:"tool_error,omitempty"`
	PermissionDecision types.AgentPermissionDecision `json:"permission_decision,omitempty"`
	StopReason         types.AgentStopReason         `json:"stop_reason,omitempty"`
	Details            string                        `json:"details,omitempty"`
}

// ActionRequestPayload carries one remote action request payload.
type ActionRequestPayload struct {
	RequestID string          `json:"request_id"`
	Method    string          `json:"method"`
	Params    json.RawMessage `json:"params,omitempty"`
}

// ActionResponsePayload carries one remote action response payload.
type ActionResponsePayload struct {
	RequestID string          `json:"request_id"`
	Result    json.RawMessage `json:"result,omitempty"`
	Error     string          `json:"error,omitempty"`
}

// ActionCancelPayload carries one remote action cancel payload.
type ActionCancelPayload struct {
	RequestID string `json:"request_id"`
	Reason    string `json:"reason,omitempty"`
}

// MessageAdapter maps SDK/provider events to remote protocol messages.
type MessageAdapter struct {
	sessionID string
	seq       int64
	now       func() time.Time
}

// NewMessageAdapter builds an adapter for one session.
func NewMessageAdapter(sessionID string) *MessageAdapter {
	return &MessageAdapter{sessionID: sessionID, now: time.Now}
}

func (a *MessageAdapter) nextSeq() int64 {
	a.seq++
	return a.seq
}

// FromProviderEvent converts one provider stream event into remote message format.
func (a *MessageAdapter) FromProviderEvent(ev types.StreamEvent) (RemoteMessage, error) {
	payload := ProviderEventPayload{
		Type:      ev.Type,
		Delta:     ev.Delta,
		ToolUseID: ev.ToolUseID,
		ToolName:  ev.ToolName,
		Input:     ev.Input,
		Message:   ev.Message,
		Usage:     ev.Usage,
	}
	if ev.Error != nil {
		payload.Error = ev.Error.Error()
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return RemoteMessage{}, fmt.Errorf("marshal provider event: %w", err)
	}

	msg := RemoteMessage{
		SessionID: a.sessionID,
		Kind:      MessageKindProviderEvent,
		Name:      string(ev.Type),
		Seq:       a.nextSeq(),
		Payload:   b,
		SentAt:    a.now().UTC(),
	}
	return msg, msg.Validate()
}

// FromToolEvent converts one tool/agent event into remote message format.
func (a *MessageAdapter) FromToolEvent(ev types.AgentEvent) (RemoteMessage, error) {
	payload := ToolEventPayload{
		Type:               ev.Type,
		Turn:               ev.Turn,
		AssistantChunk:     ev.AssistantChunk,
		ToolUseID:          ev.ToolUseID,
		ToolName:           ev.ToolName,
		ToolInput:          ev.ToolInput,
		ToolError:          ev.ToolError,
		PermissionDecision: ev.PermissionDecision,
		StopReason:         ev.StopReason,
		Details:            ev.Details,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return RemoteMessage{}, fmt.Errorf("marshal tool event: %w", err)
	}

	msg := RemoteMessage{
		SessionID: a.sessionID,
		Kind:      MessageKindToolEvent,
		Name:      string(ev.Type),
		Seq:       a.nextSeq(),
		Payload:   b,
		SentAt:    a.now().UTC(),
	}
	return msg, msg.Validate()
}

// Heartbeat builds a heartbeat protocol message.
func (a *MessageAdapter) Heartbeat() RemoteMessage {
	msg := RemoteMessage{
		SessionID: a.sessionID,
		Kind:      MessageKindHeartbeat,
		Name:      "tick",
		Seq:       a.nextSeq(),
		SentAt:    a.now().UTC(),
	}
	return msg
}

// FromActionRequest builds a request lifecycle message for one remote action.
func (a *MessageAdapter) FromActionRequest(requestID, method string, params json.RawMessage) (RemoteMessage, error) {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return RemoteMessage{}, errors.New("request_id cannot be empty")
	}
	method = strings.TrimSpace(method)
	if method == "" {
		return RemoteMessage{}, errors.New("method cannot be empty")
	}

	payload := ActionRequestPayload{
		RequestID: requestID,
		Method:    method,
		Params:    params,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return RemoteMessage{}, fmt.Errorf("marshal action request: %w", err)
	}

	msg := RemoteMessage{
		SessionID: a.sessionID,
		Kind:      MessageKindRequest,
		Name:      method,
		Seq:       a.nextSeq(),
		Payload:   b,
		SentAt:    a.now().UTC(),
	}
	return msg, msg.Validate()
}

// FromActionResponse builds a response lifecycle message for one remote action.
func (a *MessageAdapter) FromActionResponse(requestID string, result json.RawMessage, errText string) (RemoteMessage, error) {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return RemoteMessage{}, errors.New("request_id cannot be empty")
	}
	errText = strings.TrimSpace(errText)
	if errText != "" && len(result) > 0 {
		return RemoteMessage{}, errors.New("response cannot include both result and error")
	}

	payload := ActionResponsePayload{
		RequestID: requestID,
		Result:    result,
		Error:     errText,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return RemoteMessage{}, fmt.Errorf("marshal action response: %w", err)
	}

	msg := RemoteMessage{
		SessionID: a.sessionID,
		Kind:      MessageKindResponse,
		Name:      "response",
		Seq:       a.nextSeq(),
		Payload:   b,
		SentAt:    a.now().UTC(),
	}
	return msg, msg.Validate()
}

// FromActionCancel builds a cancel lifecycle message for one remote action.
func (a *MessageAdapter) FromActionCancel(requestID, reason string) (RemoteMessage, error) {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return RemoteMessage{}, errors.New("request_id cannot be empty")
	}
	reason = strings.TrimSpace(reason)

	payload := ActionCancelPayload{
		RequestID: requestID,
		Reason:    reason,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return RemoteMessage{}, fmt.Errorf("marshal action cancel: %w", err)
	}

	msg := RemoteMessage{
		SessionID: a.sessionID,
		Kind:      MessageKindCancel,
		Name:      "cancel",
		Seq:       a.nextSeq(),
		Payload:   b,
		SentAt:    a.now().UTC(),
	}
	return msg, msg.Validate()
}
