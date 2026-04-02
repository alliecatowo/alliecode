package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSessionMetadataRegistryReadWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session-metadata.json")
	want := SessionMetadataRegistry{
		Sessions: map[string]SessionMetadata{
			"session-a": {
				SessionID:      "session-a",
				SessionPath:    "/tmp/.alliecode/sessions/session-a.jsonl",
				ProjectPath:    "/tmp/repo-a",
				Title:          "Sprint Planning",
				Summary:        "events=5 messages=2",
				UpdatedAt:      123,
				CreatedAt:      120,
				MessageCount:   2,
				EventCount:     5,
				LastRole:       "assistant",
				RuntimeSurface: "cli",
				CommandSurface: "repl",
			},
		},
		LastSessionID:     "session-a",
		LastUpdatedAt:     123,
		TotalSessions:     1,
		TotalMessageCount: 2,
	}
	if err := WriteSessionMetadataRegistry(path, want); err != nil {
		t.Fatalf("WriteSessionMetadataRegistry() error = %v", err)
	}

	got, err := ReadSessionMetadataRegistry(path)
	if err != nil {
		t.Fatalf("ReadSessionMetadataRegistry() error = %v", err)
	}
	meta := got.Sessions["session-a"]
	if meta.SessionID != want.Sessions["session-a"].SessionID || meta.Title != want.Sessions["session-a"].Title {
		t.Fatalf("metadata mismatch: got %+v want %+v", meta, want.Sessions["session-a"])
	}
}

func TestUpsertSessionMetadata(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session-metadata.json")
	now := time.Now().UTC().Unix()
	registry, err := UpsertSessionMetadata(path, SessionMetadataUpdate{
		SessionID:      "s1",
		SessionPath:    "/tmp/s1.jsonl",
		ProjectPath:    "/tmp/repo",
		Title:          "First",
		Summary:        "summary",
		UpdatedAt:      now,
		CreatedAt:      now - 10,
		MessageCount:   3,
		EventCount:     5,
		LastRole:       "assistant",
		RuntimeSurface: "cli",
		CommandSurface: "repl",
	})
	if err != nil {
		t.Fatalf("UpsertSessionMetadata() error = %v", err)
	}
	meta := registry.Sessions["s1"]
	if meta.SessionID != "s1" || meta.MessageCount != 3 || meta.EventCount != 5 {
		t.Fatalf("unexpected metadata after upsert: %+v", meta)
	}
	if registry.LastSessionID != "s1" || registry.TotalSessions != 1 || registry.TotalMessageCount != 3 {
		t.Fatalf("unexpected registry after upsert: %+v", registry)
	}
}

func TestSessionMetadataRegistryMissingDefaults(t *testing.T) {
	got, err := ReadSessionMetadataRegistry(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("ReadSessionMetadataRegistry() error = %v", err)
	}
	if got.Sessions == nil {
		t.Fatalf("expected initialized sessions map")
	}
	if len(got.Sessions) != 0 {
		t.Fatalf("expected empty sessions map, got %d entries", len(got.Sessions))
	}
}

func TestSessionMetadataRegistryRejectsInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session-metadata.json")
	if err := os.WriteFile(path, []byte("{"), stateFilePerm); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	_, err := ReadSessionMetadataRegistry(path)
	if err == nil || !strings.Contains(err.Error(), "decode session metadata") {
		t.Fatalf("expected decode error, got %v", err)
	}
}
