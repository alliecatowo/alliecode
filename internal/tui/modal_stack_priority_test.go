package tui

import "testing"

func TestModalStackPriorityPermissionOverSearch(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.startQuickOpen()
	app.permissionQueue = []permissionPromptRequest{{toolName: "bash", description: "run command"}}
	app.ensurePermissionPromptVisible()
	if got := app.activeModalSurface(); got != modalSurfacePermission {
		t.Fatalf("expected permission to own top modal, got %q", got)
	}
	app.permissionQueue = nil
	app.setState(stateSearch)
	if got := app.activeModalSurface(); got != modalSurfaceSearch {
		t.Fatalf("expected search to own top modal after permission closes, got %q", got)
	}
}
