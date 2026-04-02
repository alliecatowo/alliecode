package tui

import "testing"

func TestPermissionDialogResetIncrementsUpdateCounter(t *testing.T) {
	state := newPermissionDialogState().transition(permissionDialogShow)
	updated := state.updated
	state = state.transition(permissionDialogReset)
	if state.updated <= updated {
		t.Fatalf("expected reset to increment update counter")
	}
}
