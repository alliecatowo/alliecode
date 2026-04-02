package tui

import (
	"strings"
	"testing"
)

func TestPermissionQueuePreviewIncludesTurnMetadata(t *testing.T) {
	app := New(Config{})
	app.permissionHistory = []permissionDecisionRecord{{ToolName: "bash", Decision: PermissionNo, Turn: 1}}
	app.permissionQueue = []permissionPromptRequest{
		{queueKey: "a", toolName: "bash", description: "run tests", turn: 1},
		{queueKey: "b", toolName: "read", description: "read file", turn: 2},
	}
	app.activePermissionQueueKey = "a"
	app.syncPermissionQueueMetadata()
	if len(app.permission.queueNext) == 0 || !strings.Contains(app.permission.queueNext[0], "turn 2") {
		t.Fatalf("expected queue preview turn metadata, got %#v", app.permission.queueNext)
	}
	if len(app.permission.queueStack) == 0 || !strings.Contains(app.permission.queueStack[0], "active") {
		t.Fatalf("expected stack preview rows, got %#v", app.permission.queueStack)
	}
	if len(app.permission.recent) == 0 || !strings.Contains(app.permission.recent[0], "deny") {
		t.Fatalf("expected recent decision preview rows, got %#v", app.permission.recent)
	}
}
