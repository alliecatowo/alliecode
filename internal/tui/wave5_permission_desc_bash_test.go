package tui

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestWave5PermissionDescriptionIncludesBashWorkdirAndTimeout(t *testing.T) {
	desc := permissionPromptDescription("bash", json.RawMessage(`{"command":"go test ./...","workdir":"/repo","timeoutMs":"120000"}`))
	if !strings.Contains(desc, "go test ./...") {
		t.Fatalf("expected command in description, got %q", desc)
	}
	if !strings.Contains(desc, "/repo") {
		t.Fatalf("expected workdir in description, got %q", desc)
	}
	if !strings.Contains(desc, "120000") {
		t.Fatalf("expected timeout in description, got %q", desc)
	}
}
