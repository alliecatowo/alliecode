package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeStateReadWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime-state.json")
	want := RuntimeState{
		SessionID:               "s-1",
		SessionPath:             "/tmp/s-1.jsonl",
		ProjectPath:             "/tmp/repo",
		WorkingDirectory:        "/tmp/repo",
		RemoteControlAtStartup:  true,
		HydratedFromDisk:        true,
		HydratedFromEnvironment: false,
		HydratedAtUnix:          123,
		RuntimeSurface:          "cli",
		CommandSurface:          "repl",
		StartupCount:            4,
		LastError:               "none",
		LastErrorAtUnix:         124,
	}
	if err := WriteRuntimeState(path, want); err != nil {
		t.Fatalf("WriteRuntimeState() error = %v", err)
	}
	got, err := ReadRuntimeState(path)
	if err != nil {
		t.Fatalf("ReadRuntimeState() error = %v", err)
	}
	if got != want {
		t.Fatalf("runtime state mismatch: got %+v want %+v", got, want)
	}
}

func TestRuntimeStateRejectsInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime-state.json")
	if err := os.WriteFile(path, []byte("{"), stateFilePerm); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	_, err := ReadRuntimeState(path)
	if err == nil || !strings.Contains(err.Error(), "decode runtime state") {
		t.Fatalf("expected decode error, got %v", err)
	}
}

func TestQueryRuntimeStatesNoLimitReturnsAllMatches(t *testing.T) {
	states := []RuntimeState{
		{SessionID: "s1", ProjectPath: "/repo/a", RuntimeSurface: "cli", HydratedAtUnix: 1},
		{SessionID: "s2", ProjectPath: "/repo/b", RuntimeSurface: "cli", HydratedAtUnix: 2},
	}
	out := QueryRuntimeStates(states, RuntimeQuery{RuntimeSurface: "cli"})
	if len(out) != 2 {
		t.Fatalf("len(out) = %d, want 2", len(out))
	}
}
