package tui

import "testing"

func TestModalStackPermissionClearsToNone(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.permissionQueue = []permissionPromptRequest{{toolName: "bash", description: "run"}}
	app.ensurePermissionPromptVisible()
	if got := app.activeModalSurface(); got != modalSurfacePermission {
		t.Fatalf("expected permission modal, got %q", got)
	}
	app.permissionQueue = nil
	app.setState(stateIdle)
	if got := app.activeModalSurface(); got != modalSurfaceNone {
		t.Fatalf("expected no modal after permission clear, got %q", got)
	}
}
