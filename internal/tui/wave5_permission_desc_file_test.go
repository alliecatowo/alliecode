package tui

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestWave5PermissionDescriptionIncludesFileRangeContext(t *testing.T) {
	desc := permissionPromptDescription("read", json.RawMessage(`{"file_path":"internal/tui/app.go","offset":"20","limit":"40"}`))
	if !strings.Contains(desc, "internal/tui/app.go") {
		t.Fatalf("expected file path in description, got %q", desc)
	}
	if !strings.Contains(desc, "line 20") {
		t.Fatalf("expected line offset in description, got %q", desc)
	}
	if !strings.Contains(desc, "40 lines") {
		t.Fatalf("expected line count in description, got %q", desc)
	}
}
