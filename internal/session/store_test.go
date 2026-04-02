package session

import (
	"path/filepath"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStoreRoundTrip(t *testing.T) {
	root := t.TempDir()
	store, err := Create(root, "s1")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	msg1 := types.NewTextMessage(types.RoleUser, "hello")
	msg2 := types.NewTextMessage(types.RoleAssistant, "world")
	if err := store.AppendMessage(msg1); err != nil {
		t.Fatalf("AppendMessage(msg1) error = %v", err)
	}
	if err := store.AppendMessage(msg2); err != nil {
		t.Fatalf("AppendMessage(msg2) error = %v", err)
	}

	got, err := LoadByID(root, "s1")
	if err != nil {
		t.Fatalf("LoadByID() error = %v", err)
	}

	if got.SessionID != "s1" {
		t.Fatalf("SessionID = %q, want %q", got.SessionID, "s1")
	}
	if len(got.Messages) != 2 {
		t.Fatalf("len(Messages) = %d, want 2", len(got.Messages))
	}
	if got.Messages[0].GetText() != "hello" {
		t.Fatalf("Messages[0] text = %q, want %q", got.Messages[0].GetText(), "hello")
	}
	if got.Messages[1].GetText() != "world" {
		t.Fatalf("Messages[1] text = %q, want %q", got.Messages[1].GetText(), "world")
	}
}

func TestLoadByPathCompactionBoundary(t *testing.T) {
	root := t.TempDir()
	store, err := Create(root, "s2")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	messages := []types.Message{
		types.NewTextMessage(types.RoleUser, "first"),
		types.NewTextMessage(types.RoleAssistant, "a1"),
		types.NewTextMessage(types.RoleUser, "u2"),
		types.NewTextMessage(types.RoleAssistant, "a2"),
		types.NewTextMessage(types.RoleUser, "u3"),
		types.NewTextMessage(types.RoleAssistant, "a3"),
	}
	for _, m := range messages {
		if err := store.AppendMessage(m); err != nil {
			t.Fatalf("AppendMessage() error = %v", err)
		}
	}

	boundary := types.CompactionBoundary{
		StartIndex: 1,
		EndIndex:   4,
		Replacement: []types.Message{
			types.NewTextMessage(types.RoleUser, "summary user"),
			types.NewTextMessage(types.RoleAssistant, "summary ack"),
		},
	}
	if err := store.AppendCompactionBoundary(boundary); err != nil {
		t.Fatalf("AppendCompactionBoundary() error = %v", err)
	}

	got, err := LoadByPath(filepath.Join(root, ".alliecode", "sessions", "s2.jsonl"))
	if err != nil {
		t.Fatalf("LoadByPath() error = %v", err)
	}

	if len(got.Boundaries) != 1 {
		t.Fatalf("len(Boundaries) = %d, want 1", len(got.Boundaries))
	}
	if len(got.Messages) != 5 {
		t.Fatalf("len(Messages) = %d, want 5", len(got.Messages))
	}
	if got.Messages[0].GetText() != "first" {
		t.Fatalf("Messages[0] = %q, want %q", got.Messages[0].GetText(), "first")
	}
	if got.Messages[1].GetText() != "summary user" {
		t.Fatalf("Messages[1] = %q, want %q", got.Messages[1].GetText(), "summary user")
	}
	if got.Messages[2].GetText() != "summary ack" {
		t.Fatalf("Messages[2] = %q, want %q", got.Messages[2].GetText(), "summary ack")
	}
	if got.Messages[3].GetText() != "u3" {
		t.Fatalf("Messages[3] = %q, want %q", got.Messages[3].GetText(), "u3")
	}
	if got.Messages[4].GetText() != "a3" {
		t.Fatalf("Messages[4] = %q, want %q", got.Messages[4].GetText(), "a3")
	}
}

func TestApplyCompactionBoundaryInvalidRange(t *testing.T) {
	messages := []types.Message{types.NewTextMessage(types.RoleUser, "x")}
	_, err := ApplyCompactionBoundary(messages, types.CompactionBoundary{StartIndex: 0, EndIndex: 2})
	if err == nil {
		t.Fatalf("expected error for invalid boundary range")
	}
}

func TestCreateWithOptionsAndMetadataHydration(t *testing.T) {
	root := t.TempDir()
	store, err := CreateWithOptions(root, "s-meta", SessionStartOptions{
		ProjectPath:    "/tmp/repo",
		RuntimeSurface: "cli",
		CommandSurface: "repl",
		Provider:       "openai",
		Model:          "gpt-4o-mini",
	})
	if err != nil {
		t.Fatalf("CreateWithOptions() error = %v", err)
	}
	if err := store.AppendMessage(types.NewTextMessage(types.RoleUser, "hi")); err != nil {
		t.Fatalf("AppendMessage() error = %v", err)
	}
	tr, err := LoadByID(root, "s-meta")
	if err != nil {
		t.Fatalf("LoadByID() error = %v", err)
	}
	if tr.ProjectPath != "/tmp/repo" || tr.RuntimeSurface != "cli" || tr.CommandSurface != "repl" || tr.Provider != "openai" || tr.Model != "gpt-4o-mini" {
		t.Fatalf("unexpected transcript metadata: %+v", tr)
	}
	if tr.EventCount < 2 || tr.MessageCount != 1 || tr.LastMessageRole != string(types.RoleUser) {
		t.Fatalf("unexpected transcript counters: %+v", tr)
	}

	meta, err := LoadMetadataByPath(tr.Path)
	if err != nil {
		t.Fatalf("LoadMetadataByPath() error = %v", err)
	}
	if len(meta.Messages) != 0 || len(meta.Boundaries) != 0 {
		t.Fatalf("expected metadata-only load without transcript payload")
	}
}
