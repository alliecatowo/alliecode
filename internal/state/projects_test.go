package state

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestProjectsRegistryReadWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "projects-state.json")
	want := ProjectsRegistry{
		Projects: map[string]ProjectMetadata{
			"/tmp/repo-a": {
				LastSessionID:        "sess-a",
				LastSessionPath:      "/tmp/.alliecode/sessions/sess-a.jsonl",
				LastSessionTitle:     "Repo A",
				LastSessionSummary:   "latest prompt",
				LastOpenedAt:         123,
				FirstOpenedAt:        100,
				OpenCount:            2,
				DistinctSessionCount: 1,
			},
		},
		LastSessionID:     "sess-a",
		LastOpenedProject: "/tmp/repo-a",
		TotalProjects:     1,
		TotalOpens:        2,
		LastUpdatedAt:     123,
	}
	if err := WriteProjectsRegistry(path, want); err != nil {
		t.Fatalf("WriteProjectsRegistry() error = %v", err)
	}

	got, err := ReadProjectsRegistry(path)
	if err != nil {
		t.Fatalf("ReadProjectsRegistry() error = %v", err)
	}
	if got.LastSessionID != want.LastSessionID || got.LastOpenedProject != want.LastOpenedProject {
		t.Fatalf("top-level mismatch: got %+v want %+v", got, want)
	}
	if got.Projects["/tmp/repo-a"] != want.Projects["/tmp/repo-a"] {
		t.Fatalf("project metadata mismatch: got %+v want %+v", got.Projects["/tmp/repo-a"], want.Projects["/tmp/repo-a"])
	}
}

func TestProjectsRegistryMissingDefaults(t *testing.T) {
	got, err := ReadProjectsRegistry(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("ReadProjectsRegistry() error = %v", err)
	}
	if got.Projects == nil {
		t.Fatalf("expected initialized project map")
	}
	if len(got.Projects) != 0 {
		t.Fatalf("expected empty project map, got %d entries", len(got.Projects))
	}
}

func TestProjectsRegistryRejectsInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "projects-state.json")
	if err := os.WriteFile(path, []byte("{"), stateFilePerm); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	_, err := ReadProjectsRegistry(path)
	if err == nil || !strings.Contains(err.Error(), "decode projects registry") {
		t.Fatalf("expected decode error, got %v", err)
	}
}

func TestProjectsRegistryNormalizesValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "projects-state.json")
	if err := WriteProjectsRegistry(path, ProjectsRegistry{
		Projects: map[string]ProjectMetadata{
			"   ": {LastOpenedAt: -1, OpenCount: -5},
			" /repo ": {
				LastSessionID:      "  session ",
				LastSessionPath:    " /tmp/sessions/session.jsonl ",
				LastSessionTitle:   "  title ",
				LastSessionSummary: "  summary ",
				LastOpenedAt:       -3,
				OpenCount:          -2,
			},
		},
		LastSessionID:     "  top-session ",
		LastOpenedProject: " /repo ",
	}); err != nil {
		t.Fatalf("WriteProjectsRegistry() error = %v", err)
	}

	got, err := ReadProjectsRegistry(path)
	if err != nil {
		t.Fatalf("ReadProjectsRegistry() error = %v", err)
	}
	if _, ok := got.Projects[""]; ok {
		t.Fatalf("expected empty key to be dropped")
	}
	meta, ok := got.Projects["/repo"]
	if !ok {
		t.Fatalf("expected trimmed project key to exist")
	}
	if meta.LastSessionID != "session" || meta.LastSessionPath != "/tmp/sessions/session.jsonl" || meta.LastSessionTitle != "title" || meta.LastSessionSummary != "summary" || meta.LastOpenedAt != 0 || meta.OpenCount != 0 {
		t.Fatalf("unexpected normalized metadata: %+v", meta)
	}
	if got.LastSessionID != "top-session" || got.LastOpenedProject != "/repo" {
		t.Fatalf("unexpected normalized top-level values: %+v", got)
	}
}

func TestRecordProjectOpenUpdatesRegistry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "projects-state.json")
	first, err := RecordProjectOpen(path, ProjectOpenRecord{
		ProjectPath:    "/tmp/repo-a",
		SessionID:      "session-1",
		SessionPath:    "/tmp/.alliecode/sessions/session-1.jsonl",
		SessionTitle:   "Initial Title",
		SessionSummary: "first summary",
	})
	if err != nil {
		t.Fatalf("RecordProjectOpen() first error = %v", err)
	}
	meta := first.Projects["/tmp/repo-a"]
	if meta.OpenCount != 1 || meta.LastSessionID != "session-1" {
		t.Fatalf("unexpected first metadata: %+v", meta)
	}

	second, err := RecordProjectOpen(path, ProjectOpenRecord{
		ProjectPath:    "/tmp/repo-a",
		SessionID:      "session-2",
		SessionPath:    "/tmp/.alliecode/sessions/session-2.jsonl",
		SessionTitle:   "Next Title",
		SessionSummary: "next summary",
	})
	if err != nil {
		t.Fatalf("RecordProjectOpen() second error = %v", err)
	}
	meta = second.Projects["/tmp/repo-a"]
	if meta.OpenCount != 2 || meta.LastSessionID != "session-2" {
		t.Fatalf("unexpected second metadata: %+v", meta)
	}
	if second.LastOpenedProject != "/tmp/repo-a" || second.LastSessionID != "session-2" {
		t.Fatalf("unexpected top-level registry values: %+v", second)
	}
}

func TestQueryProjectsByPathAndSort(t *testing.T) {
	path := filepath.Join(t.TempDir(), "projects-state.json")
	now := time.Now().UTC()
	_, _ = RecordProjectOpen(path, ProjectOpenRecord{ProjectPath: "/tmp/repo-a", SessionID: "s1", SessionTitle: "A", SessionSummary: "alpha", OpenedAt: now.Add(-time.Hour)})
	_, _ = RecordProjectOpen(path, ProjectOpenRecord{ProjectPath: "/tmp/repo-b", SessionID: "s2", SessionTitle: "B", SessionSummary: "beta", OpenedAt: now})
	_, _ = RecordProjectOpen(path, ProjectOpenRecord{ProjectPath: "/tmp/repo-b", SessionID: "s3", SessionTitle: "B2", SessionSummary: "beta2", OpenedAt: now.Add(time.Minute)})

	items, err := QueryProjects(path, ProjectQuery{ContainsPath: "repo", SortByOpenCount: true, Limit: 2})
	if err != nil {
		t.Fatalf("QueryProjects() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	if got := []string{items[0].Path, items[1].Path}; !reflect.DeepEqual(got, []string{"/tmp/repo-b", "/tmp/repo-a"}) {
		t.Fatalf("path order mismatch: got %v", got)
	}
}
