package remote

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestAdapterProviderEventFidelity(t *testing.T) {
	adapter := NewMessageAdapter("sid-adapter")
	fixed := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	adapter.now = func() time.Time { return fixed }

	ev := types.StreamEvent{
		Type:      types.StreamToolUseDone,
		ToolUseID: "tool-1",
		ToolName:  "bash",
		Input:     json.RawMessage(`{"command":"ls"}`),
		Usage:     &types.Usage{InputTokens: 10, OutputTokens: 5},
		Error:     errors.New("provider boom"),
	}

	msg, err := adapter.FromProviderEvent(ev)
	if err != nil {
		t.Fatalf("FromProviderEvent() error = %v", err)
	}
	if msg.Kind != MessageKindProviderEvent || msg.Name != string(types.StreamToolUseDone) {
		t.Fatalf("message meta mismatch: %+v", msg)
	}
	if msg.Seq != 1 || !msg.SentAt.Equal(fixed) {
		t.Fatalf("message sequencing mismatch: %+v", msg)
	}

	var payload ProviderEventPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload error = %v", err)
	}
	if payload.ToolUseID != "tool-1" || payload.ToolName != "bash" {
		t.Fatalf("tool payload mismatch: %+v", payload)
	}
	if payload.Error != "provider boom" {
		t.Fatalf("payload error = %q, want %q", payload.Error, "provider boom")
	}
}

func TestAdapterToolEventFidelityAndHeartbeat(t *testing.T) {
	adapter := NewMessageAdapter("sid-adapter")
	fixed := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
	adapter.now = func() time.Time { return fixed }

	ev := types.AgentEvent{
		Type:               types.AgentEventToolStart,
		Turn:               7,
		ToolUseID:          "tool-2",
		ToolName:           "read",
		ToolInput:          json.RawMessage(`{"file":"a.txt"}`),
		PermissionDecision: types.AgentPermissionAllow,
	}

	msg, err := adapter.FromToolEvent(ev)
	if err != nil {
		t.Fatalf("FromToolEvent() error = %v", err)
	}
	if msg.Kind != MessageKindToolEvent || msg.Name != string(types.AgentEventToolStart) {
		t.Fatalf("message meta mismatch: %+v", msg)
	}
	if msg.Seq != 1 || !msg.SentAt.Equal(fixed) {
		t.Fatalf("message sequencing mismatch: %+v", msg)
	}

	var payload ToolEventPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload error = %v", err)
	}
	if payload.ToolUseID != "tool-2" || payload.ToolName != "read" || payload.Turn != 7 {
		t.Fatalf("tool payload mismatch: %+v", payload)
	}

	hb := adapter.Heartbeat()
	if hb.Kind != MessageKindHeartbeat || hb.Name != "tick" {
		t.Fatalf("heartbeat mismatch: %+v", hb)
	}
	if hb.Seq != 2 {
		t.Fatalf("heartbeat seq = %d, want 2", hb.Seq)
	}
}

func TestAdapterActionLifecycleMessages(t *testing.T) {
	adapter := NewMessageAdapter("sid-adapter")
	fixed := time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)
	adapter.now = func() time.Time { return fixed }

	req, err := adapter.FromActionRequest("req-1", "tools/call", json.RawMessage(`{"name":"read"}`))
	if err != nil {
		t.Fatalf("FromActionRequest() error = %v", err)
	}
	if req.Kind != MessageKindRequest || req.Name != "tools/call" {
		t.Fatalf("request message mismatch: %+v", req)
	}
	if req.Seq != 1 || !req.SentAt.Equal(fixed) {
		t.Fatalf("request sequencing mismatch: %+v", req)
	}
	var reqPayload ActionRequestPayload
	if err := json.Unmarshal(req.Payload, &reqPayload); err != nil {
		t.Fatalf("unmarshal request payload error = %v", err)
	}
	if reqPayload.RequestID != "req-1" || reqPayload.Method != "tools/call" {
		t.Fatalf("request payload mismatch: %+v", reqPayload)
	}

	resp, err := adapter.FromActionResponse("req-1", json.RawMessage(`{"ok":true}`), "")
	if err != nil {
		t.Fatalf("FromActionResponse(success) error = %v", err)
	}
	if resp.Kind != MessageKindResponse || resp.Seq != 2 {
		t.Fatalf("response message mismatch: %+v", resp)
	}
	var respPayload ActionResponsePayload
	if err := json.Unmarshal(resp.Payload, &respPayload); err != nil {
		t.Fatalf("unmarshal response payload error = %v", err)
	}
	if respPayload.RequestID != "req-1" || string(respPayload.Result) != `{"ok":true}` || respPayload.Error != "" {
		t.Fatalf("response payload mismatch: %+v", respPayload)
	}

	cancel, err := adapter.FromActionCancel("req-1", "user_cancelled")
	if err != nil {
		t.Fatalf("FromActionCancel() error = %v", err)
	}
	if cancel.Kind != MessageKindCancel || cancel.Seq != 3 {
		t.Fatalf("cancel message mismatch: %+v", cancel)
	}
	var cancelPayload ActionCancelPayload
	if err := json.Unmarshal(cancel.Payload, &cancelPayload); err != nil {
		t.Fatalf("unmarshal cancel payload error = %v", err)
	}
	if cancelPayload.RequestID != "req-1" || cancelPayload.Reason != "user_cancelled" {
		t.Fatalf("cancel payload mismatch: %+v", cancelPayload)
	}

	if _, err := adapter.FromActionResponse("req-2", json.RawMessage(`{"ok":true}`), "boom"); err == nil || err.Error() != "response cannot include both result and error" {
		t.Fatalf("expected deterministic mixed response error, got %v", err)
	}
}
