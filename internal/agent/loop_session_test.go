package agent

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/alliecatowo/alliecode/internal/session"
	"github.com/alliecatowo/alliecode/internal/types"
)

func TestResumeFromSessionID(t *testing.T) {
	root := t.TempDir()
	store, err := session.Create(root, "resume-id")
	if err != nil {
		t.Fatalf("session.Create() error = %v", err)
	}
	if err := store.AppendMessage(types.NewTextMessage(types.RoleUser, "u1")); err != nil {
		t.Fatalf("AppendMessage(u1) error = %v", err)
	}
	if err := store.AppendMessage(types.NewTextMessage(types.RoleAssistant, "a1")); err != nil {
		t.Fatalf("AppendMessage(a1) error = %v", err)
	}

	a := New(Config{})
	var resumed *types.AgentEvent
	a.SetEventCallback(func(ev types.AgentEvent) {
		if ev.Type == types.AgentEventSessionResumed {
			cp := ev
			resumed = &cp
		}
	})
	if err := a.ResumeFromSessionID(root, "resume-id"); err != nil {
		t.Fatalf("ResumeFromSessionID() error = %v", err)
	}

	msgs := a.Messages()
	if len(msgs) != 2 {
		t.Fatalf("len(Messages) = %d, want 2", len(msgs))
	}
	if msgs[0].GetText() != "u1" || msgs[1].GetText() != "a1" {
		t.Fatalf("unexpected resumed messages: %q / %q", msgs[0].GetText(), msgs[1].GetText())
	}
	if resumed == nil {
		t.Fatalf("expected session_resumed event")
	}
	if resumed.SessionID != "resume-id" || resumed.SessionMessageCount != 2 || resumed.SessionResumedFromPath {
		t.Fatalf("unexpected resumed metadata: %+v", *resumed)
	}
}

func TestResumeFromSessionPath(t *testing.T) {
	root := t.TempDir()
	store, err := session.Create(root, "resume-path")
	if err != nil {
		t.Fatalf("session.Create() error = %v", err)
	}
	if err := store.AppendMessage(types.NewTextMessage(types.RoleUser, "hello")); err != nil {
		t.Fatalf("AppendMessage() error = %v", err)
	}

	a := New(Config{})
	var resumed *types.AgentEvent
	a.SetEventCallback(func(ev types.AgentEvent) {
		if ev.Type == types.AgentEventSessionResumed {
			cp := ev
			resumed = &cp
		}
	})
	path := filepath.Join(root, ".alliecode", "sessions", "resume-path.jsonl")
	if err := a.ResumeFromSessionPath(path); err != nil {
		t.Fatalf("ResumeFromSessionPath() error = %v", err)
	}

	msgs := a.Messages()
	if len(msgs) != 1 {
		t.Fatalf("len(Messages) = %d, want 1", len(msgs))
	}
	if msgs[0].GetText() != "hello" {
		t.Fatalf("resumed text = %q, want %q", msgs[0].GetText(), "hello")
	}
	if resumed == nil {
		t.Fatalf("expected session_resumed event")
	}
	if !resumed.SessionResumedFromPath || resumed.SessionPath == "" {
		t.Fatalf("unexpected resumed metadata: %+v", *resumed)
	}
}

func TestReplayContinuityAcrossCompactionBoundary(t *testing.T) {
	root := t.TempDir()
	store, err := session.Create(root, "resume-compact")
	if err != nil {
		t.Fatalf("session.Create() error = %v", err)
	}

	base := []types.Message{
		types.NewTextMessage(types.RoleUser, "first"),
		types.NewTextMessage(types.RoleAssistant, "a1"),
		types.NewTextMessage(types.RoleUser, "u2"),
		types.NewTextMessage(types.RoleAssistant, "a2"),
		types.NewTextMessage(types.RoleUser, "u3"),
		types.NewTextMessage(types.RoleAssistant, "a3"),
		types.NewTextMessage(types.RoleUser, "u4"),
	}
	for _, msg := range base {
		if err := store.AppendMessage(msg); err != nil {
			t.Fatalf("AppendMessage() error = %v", err)
		}
	}

	_, boundary, err := CompactMessages(context.Background(), base, summaryProvider{}, "test")
	if err != nil {
		t.Fatalf("CompactMessages() error = %v", err)
	}
	if boundary == nil {
		t.Fatalf("expected compaction boundary")
	}
	if err := store.AppendCompactionBoundary(*boundary); err != nil {
		t.Fatalf("AppendCompactionBoundary() error = %v", err)
	}

	buffered := []types.Message{
		types.NewToolResultMessage("tool-1", "r1", false),
		types.NewToolResultMessage("tool-2", "r2", false),
	}
	for _, msg := range buffered {
		if err := store.AppendMessage(msg); err != nil {
			t.Fatalf("AppendMessage(buffered) error = %v", err)
		}
	}

	a := New(Config{})
	if err := a.ResumeFromSessionID(root, "resume-compact"); err != nil {
		t.Fatalf("ResumeFromSessionID() error = %v", err)
	}

	msgs := a.Messages()
	if len(msgs) < 2 {
		t.Fatalf("expected replayed messages, got %d", len(msgs))
	}
	n := len(msgs)
	if got := msgs[n-2].Content[0].ForToolUseID; got != "tool-1" {
		t.Fatalf("replayed buffered message id = %q, want %q", got, "tool-1")
	}
	if got := msgs[n-1].Content[0].ForToolUseID; got != "tool-2" {
		t.Fatalf("replayed buffered message id = %q, want %q", got, "tool-2")
	}
}

func TestReplayContinuityWithBufferedConcurrentToolMessages(t *testing.T) {
	root := t.TempDir()
	store, err := session.Create(root, "resume-buffered-tools")
	if err != nil {
		t.Fatalf("session.Create() error = %v", err)
	}

	slow := &delayTool{name: "slow", delay: 25 * time.Millisecond, concurrencySafe: true, result: "slow-result"}
	fast := &delayTool{name: "fast", delay: 5 * time.Millisecond, concurrencySafe: true, result: "fast-result"}
	provider := &scriptedProvider{scripts: []func(chan types.StreamEvent){
		twoToolUseScript("slow", "tool-slow", json.RawMessage(`{"x":1}`), "fast", "tool-fast", json.RawMessage(`{"y":2}`)),
		textScript("done"),
	}}

	a := New(Config{
		Provider:     provider,
		Tools:        []types.Tool{slow, fast},
		MaxTurns:     4,
		SessionStore: store,
	})
	if err := a.Run(context.Background(), "hello"); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	resumed := New(Config{})
	if err := resumed.ResumeFromSessionID(root, "resume-buffered-tools"); err != nil {
		t.Fatalf("ResumeFromSessionID() error = %v", err)
	}

	var toolIDs []string
	for _, msg := range resumed.Messages() {
		for _, block := range msg.Content {
			if block.Type == types.ContentToolResult {
				toolIDs = append(toolIDs, block.ForToolUseID)
			}
		}
	}
	if len(toolIDs) < 2 {
		t.Fatalf("expected at least 2 replayed tool results, got %d", len(toolIDs))
	}
	if toolIDs[0] != "tool-slow" || toolIDs[1] != "tool-fast" {
		t.Fatalf("replayed tool result order = %v, want [tool-slow tool-fast ...]", toolIDs)
	}
}
