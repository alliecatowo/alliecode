package tui

import (
	"strings"
	"testing"
)

func TestWave5PermissionStackIncludesDescriptionSnippet(t *testing.T) {
	app := New(Config{})
	app.permissionQueue = []permissionPromptRequest{{queueKey: "a", toolName: "bash", description: "run test suite for release", turn: 3}}
	app.activePermissionQueueKey = "a"
	app.syncPermissionQueueMetadata()
	if len(app.permission.queueStack) == 0 {
		t.Fatalf("expected stack metadata rows")
	}
	if !strings.Contains(app.permission.queueStack[0], "run test suite") {
		t.Fatalf("expected stack row to include description snippet, got %#v", app.permission.queueStack)
	}
}
