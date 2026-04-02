package state

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestResolvePathsDeterministic(t *testing.T) {
	homeBase := t.TempDir()
	projectBase := t.TempDir()

	first, err := ResolvePaths(filepath.Join(homeBase, "."), filepath.Join(projectBase, "a", ".."))
	if err != nil {
		t.Fatalf("ResolvePaths() error = %v", err)
	}
	second, err := ResolvePaths(homeBase, projectBase)
	if err != nil {
		t.Fatalf("ResolvePaths() second error = %v", err)
	}

	if first.HomeDir != second.HomeDir {
		t.Fatalf("HomeDir mismatch: %q != %q", first.HomeDir, second.HomeDir)
	}
	if first.ProjectDir != second.ProjectDir {
		t.Fatalf("ProjectDir mismatch: %q != %q", first.ProjectDir, second.ProjectDir)
	}
	if first.GlobalHistoryFile != filepath.Join(second.HomeDir, globalHistoryFileName) {
		t.Fatalf("GlobalHistoryFile mismatch: %q", first.GlobalHistoryFile)
	}
	if first.SessionsDir != filepath.Join(second.HomeDir, sessionsDirName) {
		t.Fatalf("SessionsDir mismatch: %q", first.SessionsDir)
	}
	if first.ProjectsStateFile != filepath.Join(second.HomeDir, projectsStateFileName) {
		t.Fatalf("ProjectsStateFile mismatch: %q", first.ProjectsStateFile)
	}
	if first.SettingsCacheFile != filepath.Join(second.HomeDir, settingsCacheFileName) {
		t.Fatalf("SettingsCacheFile mismatch: %q", first.SettingsCacheFile)
	}
	if first.AuthStateFile != filepath.Join(second.HomeDir, authStateFileName) {
		t.Fatalf("AuthStateFile mismatch: %q", first.AuthStateFile)
	}
	if first.SessionMetadataFile != filepath.Join(second.HomeDir, sessionMetadataFileName) {
		t.Fatalf("SessionMetadataFile mismatch: %q", first.SessionMetadataFile)
	}
	if first.RuntimeStateFile != filepath.Join(second.HomeDir, runtimeStateFileName) {
		t.Fatalf("RuntimeStateFile mismatch: %q", first.RuntimeStateFile)
	}
	if first.ConfigDiagnosticsFile != filepath.Join(second.HomeDir, configDiagnosticsFileName) {
		t.Fatalf("ConfigDiagnosticsFile mismatch: %q", first.ConfigDiagnosticsFile)
	}
	if first.MigrationsLedgerFile != filepath.Join(second.HomeDir, migrationsLedgerFileName) {
		t.Fatalf("MigrationsLedgerFile mismatch: %q", first.MigrationsLedgerFile)
	}
	if first.ProjectOnboardingStateFile != filepath.Join(second.ProjectDir, projectOnboardingStateFileName) {
		t.Fatalf("ProjectOnboardingStateFile mismatch: %q", first.ProjectOnboardingStateFile)
	}
}

func TestEnsureStateCreatesCanonicalLayout(t *testing.T) {
	paths, err := ResolvePaths(t.TempDir(), t.TempDir())
	if err != nil {
		t.Fatalf("ResolvePaths() error = %v", err)
	}

	if err := EnsureState(paths); err != nil {
		t.Fatalf("EnsureState() error = %v", err)
	}

	assertExists(t, paths.HomeDir)
	assertExists(t, paths.ProjectDir)
	assertExists(t, paths.SessionsDir)
	assertExists(t, paths.GlobalHistoryFile)
	assertExists(t, paths.ProjectsStateFile)
	assertExists(t, paths.SettingsCacheFile)
	assertExists(t, paths.AuthStateFile)
	assertExists(t, paths.SessionMetadataFile)
	assertExists(t, paths.RuntimeStateFile)
	assertExists(t, paths.ConfigDiagnosticsFile)
	assertExists(t, paths.MigrationsLedgerFile)
	assertExists(t, paths.ProjectOnboardingStateFile)

	if runtime.GOOS != "windows" {
		assertMode(t, paths.ProjectsStateFile, stateFilePerm)
		assertMode(t, paths.AuthStateFile, stateFilePerm)
	}
}

func TestEnsureStateDoesNotOverwriteExistingFile(t *testing.T) {
	paths, err := ResolvePaths(t.TempDir(), t.TempDir())
	if err != nil {
		t.Fatalf("ResolvePaths() error = %v", err)
	}
	if err := os.MkdirAll(paths.HomeDir, stateDirPerm); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	want := []byte("keep-me")
	if err := os.WriteFile(paths.AuthStateFile, want, stateFilePerm); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := EnsureHomeState(paths); err != nil {
		t.Fatalf("EnsureHomeState() error = %v", err)
	}
	got, err := os.ReadFile(paths.AuthStateFile)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("auth state overwritten: got %q want %q", got, want)
	}
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %q to exist: %v", path, err)
	}
}

func assertMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat(%q) error = %v", path, err)
	}
	got := info.Mode().Perm()
	if got != want {
		t.Fatalf("file mode for %q = %#o, want %#o", path, got, want)
	}
}
