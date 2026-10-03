package tui

import (
	"strings"
	"testing"
)

func TestWave5PermissionStackIncludesDescriptionSnippet(t *testing.T) {
	app := New(Config{})
	app.permissionQueue = []permissionPromptRequest{{queueKey: "a", toolName: "bash", toolKind: "bash", toolDetails: []string{"command: go test ./..."}, description: "run test suite for release", turn: 3, status: permissionPending}}
	app.activePermissionQueueKey = "a"
	app.syncPermissionQueueMetadata()
	if len(app.permission.queueStack) == 0 {
		t.Fatalf("expected stack metadata rows")
	}
	if !strings.Contains(app.permission.queueStack[0], "run test suite") {
		t.Fatalf("expected stack row to include description snippet, got %#v", app.permission.queueStack)
	}
	if app.permission.toolKind != "bash" || len(app.permission.toolDetails) == 0 {
		t.Fatalf("expected active tool detail metadata synced into model, got kind=%q details=%#v", app.permission.toolKind, app.permission.toolDetails)
	}
}
