package state

import "testing"

func TestNewRepositories(t *testing.T) {
	paths, err := ResolvePaths(t.TempDir(), t.TempDir())
	if err != nil {
		t.Fatalf("ResolvePaths() error = %v", err)
	}
	repos := NewRepositories(paths)
	if repos.Projects == nil || repos.Sessions == nil || repos.ConfigDiagnostics == nil || repos.SettingsCache == nil || repos.Runtime == nil || repos.Auth == nil || repos.Onboarding == nil {
		t.Fatalf("expected all repositories to be initialized")
	}
}
