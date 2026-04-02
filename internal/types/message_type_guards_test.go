package types

import (
	"encoding/json"
	"testing"
)

func TestMessageTypeGuards(t *testing.T) {
	user := NewTextMessage(RoleUser, "hi")
	assistantTool := NewToolUseMessage("t-1", "Read", json.RawMessage(`{"file":"a"}`))
	assistantThink := Message{Role: RoleAssistant, Content: []ContentBlock{{Type: ContentThinking, Text: "plan"}}}
	result := NewToolResultMessage("t-1", "ok", false)

	if !IsUserMessage(user) || IsAssistantMessage(user) {
		t.Fatalf("unexpected user guard behavior")
	}
	if !IsAssistantMessage(assistantTool) || !MessageHasToolUse(assistantTool) {
		t.Fatalf("expected assistant tool message")
	}
	if !MessageHasThinking(assistantThink) {
		t.Fatalf("expected thinking block detection")
	}
	if !MessageHasToolResult(result) {
		t.Fatalf("expected tool result detection")
	}
}
