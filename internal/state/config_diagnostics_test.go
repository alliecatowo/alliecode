package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigDiagnosticsStateReadWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config-diagnostics.json")
	want := ConfigDiagnosticsState{
		LastConfigPath:             "/tmp/config.yaml",
		LastProjectPath:            "/tmp/repo",
		LayeringMode:               "layered",
		LastValidationPassed:       true,
		LastValidationErrorCount:   0,
		LastValidationWarningCount: 1,
		LastDiagnosticsSummary:     "ok",
		LastValidatedAtUnix:        123,
		LastMigrationVersion:       "20260401_001",
		MigrationAppliedCount:      2,
		MigrationSkippedCount:      1,
	}
	if err := WriteConfigDiagnosticsState(path, want); err != nil {
		t.Fatalf("WriteConfigDiagnosticsState() error = %v", err)
	}
	got, err := ReadConfigDiagnosticsState(path)
	if err != nil {
		t.Fatalf("ReadConfigDiagnosticsState() error = %v", err)
	}
	if got != want {
		t.Fatalf("config diagnostics mismatch: got %+v want %+v", got, want)
	}
}

func TestConfigDiagnosticsStateRejectsInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config-diagnostics.json")
	if err := os.WriteFile(path, []byte("{"), stateFilePerm); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	_, err := ReadConfigDiagnosticsState(path)
	if err == nil || !strings.Contains(err.Error(), "decode config diagnostics") {
		t.Fatalf("expected decode error, got %v", err)
	}
}

func TestConfigDiagnosticsRepositoryMatches(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config-diagnostics.json")
	repo := NewConfigDiagnosticsRepository(path)
	if err := repo.Save(ConfigDiagnosticsState{LayeringMode: "layered", LastValidationPassed: true, LastValidatedAtUnix: 10}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	passed := true
	ok, _, err := repo.Matches(ConfigDiagnosticsQuery{LayeringMode: "layer", ValidationPassed: &passed})
	if err != nil {
		t.Fatalf("Matches() error = %v", err)
	}
	if !ok {
		t.Fatalf("expected repository state to match query")
	}
}
