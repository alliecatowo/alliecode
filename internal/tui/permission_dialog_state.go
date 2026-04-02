package tui

type permissionDialogStage int

const (
	permissionDialogHidden permissionDialogStage = iota
	permissionDialogPrompt
	permissionDialogResolved
)

type permissionDialogEvent int

const (
	permissionDialogShow permissionDialogEvent = iota
	permissionDialogAllow
	permissionDialogDeny
	permissionDialogAlways
	permissionDialogReset
)

type permissionDialogState struct {
	stage    permissionDialogStage
	decision PermissionDecision
	status   permissionStatus
	updated  int
}

func newPermissionDialogState() permissionDialogState {
	return permissionDialogState{stage: permissionDialogHidden, decision: PermissionUndecided, status: permissionPending, updated: 0}
}

func (s permissionDialogState) transition(event permissionDialogEvent) permissionDialogState {
	next := s
	switch event {
	case permissionDialogShow:
		next.stage = permissionDialogPrompt
		next.decision = PermissionUndecided
		next.status = permissionPending
		next.updated++
	case permissionDialogAllow:
		if s.stage == permissionDialogPrompt {
			next.stage = permissionDialogResolved
			next.decision = PermissionYes
			next.status = permissionApproved
			next.updated++
		}
	case permissionDialogDeny:
		if s.stage == permissionDialogPrompt {
			next.stage = permissionDialogResolved
			next.decision = PermissionNo
			next.status = permissionDenied
			next.updated++
		}
	case permissionDialogAlways:
		if s.stage == permissionDialogPrompt {
			next.stage = permissionDialogResolved
			next.decision = PermissionAlways
			next.status = permissionAlwaysStatus
			next.updated++
		}
	case permissionDialogReset:
		next = newPermissionDialogState()
		next.updated = s.updated + 1
	}
	return next
}
