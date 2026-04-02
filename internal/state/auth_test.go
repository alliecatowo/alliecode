package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuthStateReadWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auth-state.json")
	want := AuthState{
		LoggedIn:      true,
		Provider:      "openai",
		Account:       "dev@example.com",
		ProviderReady: true,
		LoginCount:    4,
		LogoutCount:   1,
	}
	if err := WriteAuthState(path, want); err != nil {
		t.Fatalf("WriteAuthState() error = %v", err)
	}

	got, err := ReadAuthState(path)
	if err != nil {
		t.Fatalf("ReadAuthState() error = %v", err)
	}
	if got != want {
		t.Fatalf("auth state mismatch: got %+v want %+v", got, want)
	}
}

func TestAuthStateMissingDefaults(t *testing.T) {
	got, err := ReadAuthState(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("ReadAuthState() error = %v", err)
	}
	if got != (AuthState{}) {
		t.Fatalf("default state mismatch: got %+v", got)
	}
}

func TestAuthStateRejectsInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auth-state.json")
	if err := os.WriteFile(path, []byte("{"), stateFilePerm); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	_, err := ReadAuthState(path)
	if err == nil || !strings.Contains(err.Error(), "decode auth state") {
		t.Fatalf("expected decode error, got %v", err)
	}
}

func TestAuthStateNormalizesFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auth-state.json")
	if err := WriteAuthState(path, AuthState{
		Provider:    "  anthropic  ",
		Account:     " dev@acme ",
		LoginCount:  -1,
		LogoutCount: -3,
	}); err != nil {
		t.Fatalf("WriteAuthState() error = %v", err)
	}

	got, err := ReadAuthState(path)
	if err != nil {
		t.Fatalf("ReadAuthState() error = %v", err)
	}
	if got.Provider != "anthropic" {
		t.Fatalf("Provider = %q, want anthropic", got.Provider)
	}
	if got.Account != "dev@acme" {
		t.Fatalf("Account = %q, want dev@acme", got.Account)
	}
	if got.LoginCount != 0 || got.LogoutCount != 0 {
		t.Fatalf("expected clamped counts, got login=%d logout=%d", got.LoginCount, got.LogoutCount)
	}
}
