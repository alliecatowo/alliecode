package tui

import "testing"

func TestPermissionDialogStateDeterministicTransitions(t *testing.T) {
	state := newPermissionDialogState()
	if state.stage != permissionDialogHidden || state.decision != PermissionUndecided || state.status != permissionPending {
		t.Fatalf("expected hidden undecided pending initial state, got stage=%d decision=%d status=%s", state.stage, state.decision, state.status)
	}

	state = state.transition(permissionDialogAllow)
	if state.stage != permissionDialogHidden || state.decision != PermissionUndecided || state.status != permissionPending {
		t.Fatalf("expected decision ignored while hidden, got stage=%d decision=%d status=%s", state.stage, state.decision, state.status)
	}

	state = state.transition(permissionDialogShow)
	if state.stage != permissionDialogPrompt || state.decision != PermissionUndecided || state.status != permissionPending {
		t.Fatalf("expected prompt pending after show, got stage=%d decision=%d status=%s", state.stage, state.decision, state.status)
	}

	state = state.transition(permissionDialogAlways)
	if state.stage != permissionDialogResolved || state.decision != PermissionAlways || state.status != permissionAlwaysStatus {
		t.Fatalf("expected resolved always state, got stage=%d decision=%d status=%s", state.stage, state.decision, state.status)
	}

	state = state.transition(permissionDialogDeny)
	if state.stage != permissionDialogResolved || state.decision != PermissionAlways || state.status != permissionAlwaysStatus {
		t.Fatalf("expected resolved state to remain stable, got stage=%d decision=%d status=%s", state.stage, state.decision, state.status)
	}

	state = state.transition(permissionDialogReset)
	if state.stage != permissionDialogHidden || state.decision != PermissionUndecided || state.status != permissionPending {
		t.Fatalf("expected reset to initial pending state, got stage=%d decision=%d status=%s", state.stage, state.decision, state.status)
	}
}

func TestPermissionDialogStateEachDecisionFromPrompt(t *testing.T) {
	start := newPermissionDialogState().transition(permissionDialogShow)

	allow := start.transition(permissionDialogAllow)
	if allow.decision != PermissionYes || allow.stage != permissionDialogResolved || allow.status != permissionApproved {
		t.Fatalf("expected allow transition to resolve approved, got stage=%d decision=%d status=%s", allow.stage, allow.decision, allow.status)
	}

	deny := start.transition(permissionDialogDeny)
	if deny.decision != PermissionNo || deny.stage != permissionDialogResolved || deny.status != permissionDenied {
		t.Fatalf("expected deny transition to resolve denied, got stage=%d decision=%d status=%s", deny.stage, deny.decision, deny.status)
	}

	always := start.transition(permissionDialogAlways)
	if always.decision != PermissionAlways || always.stage != permissionDialogResolved || always.status != permissionAlwaysStatus {
		t.Fatalf("expected always transition to resolve always status, got stage=%d decision=%d status=%s", always.stage, always.decision, always.status)
	}
}
