package tui

import "strings"

func permissionDialogEventFromDecision(decision PermissionDecision) permissionDialogEvent {
	switch decision {
	case PermissionYes:
		return permissionDialogAllow
	case PermissionNo:
		return permissionDialogDeny
	case PermissionAlways:
		return permissionDialogAlways
	default:
		return permissionDialogReset
	}
}

func (a *App) ensurePermissionPromptVisible() {
	if len(a.permissionQueue) == 0 {
		if a.activePermissionToolUseID != "" || a.activePermissionQueueKey != "" || a.activePermissionTurn != 0 {
			a.activePermissionToolUseID = ""
			a.activePermissionQueueKey = ""
			a.activePermissionTurn = 0
		}
		if a.stateValue() == statePermissionPrompt {
			a.setState(stateThinking)
			a.spinner = NewSpinner("executing")
		}
		return
	}
	if a.stateValue() == statePermissionPrompt && a.permDialog.stage == permissionDialogPrompt {
		a.syncPermissionQueueMetadata()
		return
	}
	request := a.permissionQueue[0]
	a.permission = NewPermission(request.toolName, request.description)
	a.permission.SetToolDetails(request.toolKind, request.toolDetails)
	a.permission.SetStatus(request.status)
	a.permission.SetQueueIndex(1, len(a.permissionQueue))
	a.permDialog = a.permDialog.transition(permissionDialogShow)
	a.setState(statePermissionPrompt)
	a.activePermissionToolUseID = request.toolUseID
	a.activePermissionQueueKey = request.queueKey
	a.activePermissionTurn = request.turn
	a.syncPermissionQueueMetadata()
}

func (a *App) dequeueActivePermissionPrompt() {
	if len(a.permissionQueue) == 0 {
		a.activePermissionToolUseID = ""
		a.activePermissionQueueKey = ""
		a.activePermissionTurn = 0
		return
	}
	if a.activePermissionToolUseID == "" {
		if strings.TrimSpace(a.activePermissionQueueKey) != "" {
			for i, request := range a.permissionQueue {
				if request.queueKey != a.activePermissionQueueKey {
					continue
				}
				a.permissionQueue = append(a.permissionQueue[:i], a.permissionQueue[i+1:]...)
				a.activePermissionQueueKey = ""
				a.activePermissionTurn = 0
				return
			}
		}
		a.permissionQueue = a.permissionQueue[1:]
		a.activePermissionQueueKey = ""
		a.activePermissionTurn = 0
		return
	}
	for i, request := range a.permissionQueue {
		if request.toolUseID != a.activePermissionToolUseID && request.queueKey != a.activePermissionQueueKey {
			continue
		}
		a.permissionQueue = append(a.permissionQueue[:i], a.permissionQueue[i+1:]...)
		a.activePermissionToolUseID = ""
		a.activePermissionQueueKey = ""
		a.activePermissionTurn = 0
		return
	}
	a.permissionQueue = a.permissionQueue[1:]
	a.activePermissionToolUseID = ""
	a.activePermissionQueueKey = ""
	a.activePermissionTurn = 0
}

func (a *App) syncPermissionQueueMetadata() {
	if len(a.permissionQueue) == 0 {
		a.permission.SetQueueIndex(0, 0)
		a.permission.SetQueuePreview(nil)
		a.permission.SetQueueStack(nil)
		a.permission.SetRecentDecisions(nil)
		return
	}
	position := 1
	for i, request := range a.permissionQueue {
		if strings.TrimSpace(a.activePermissionQueueKey) != "" {
			if request.queueKey != a.activePermissionQueueKey {
				continue
			}
			position = i + 1
			a.permission.SetToolDetails(request.toolKind, request.toolDetails)
			a.permission.SetStatus(request.status)
			break
		}
		if request.toolUseID == a.activePermissionToolUseID {
			position = i + 1
			a.permission.SetToolDetails(request.toolKind, request.toolDetails)
			a.permission.SetStatus(request.status)
			break
		}
	}
	a.permission.SetQueueIndex(position, len(a.permissionQueue))
	if position < 1 {
		position = 1
	}
	if position > len(a.permissionQueue) {
		position = len(a.permissionQueue)
	}
	next := make([]string, 0, 3)
	stack := make([]string, 0, 4)
	for i := position - 1; i < len(a.permissionQueue) && len(stack) < 4; i++ {
		if i < 0 {
			continue
		}
		request := a.permissionQueue[i]
		marker := "[next]"
		if i == position-1 {
			marker = "[active]"
		}
		row := marker + " " + strings.TrimSpace(request.toolName)
		if request.status != "" {
			row += " [" + strings.ToUpper(string(request.status)) + "]"
		}
		if request.turn > 0 {
			row += " turn " + itoa(request.turn)
		}
		if detail := strings.TrimSpace(request.description); detail != "" {
			row += " :: " + detail
		}
		stack = append(stack, truncateDisplayWidth(row, 64, "..."))
	}
	for i := position; i < len(a.permissionQueue) && len(next) < 3; i++ {
		request := a.permissionQueue[i]
		label := strings.TrimSpace(request.toolName)
		if label == "" {
			label = "tool"
		}
		if request.turn > 0 {
			label += " (turn " + itoa(request.turn) + ")"
		}
		if request.status != "" {
			label += " [" + strings.ToUpper(string(request.status)) + "]"
		}
		preview := strings.TrimSpace(request.description)
		if preview == "" {
			next = append(next, label)
			continue
		}
		next = append(next, truncateDisplayWidth(label+": "+preview, 64, "..."))
	}
	a.permission.SetQueuePreview(next)
	a.permission.SetQueueStack(stack)
	a.permission.SetRecentDecisions(a.permissionRecentPreview(3))
}

func (a *App) permissionRecentPreview(limit int) []string {
	if limit <= 0 {
		limit = 3
	}
	if len(a.permissionHistory) == 0 {
		return nil
	}
	out := make([]string, 0, limit)
	for i := len(a.permissionHistory) - 1; i >= 0 && len(out) < limit; i-- {
		record := a.permissionHistory[i]
		decision := "allow"
		switch record.Decision {
		case PermissionNo:
			decision = "deny"
		case PermissionAlways:
			decision = "always"
		}
		row := strings.TrimSpace(record.ToolName) + " -> " + decision
		if record.Turn > 0 {
			row += " (turn " + itoa(record.Turn) + ")"
		}
		out = append(out, truncateDisplayWidth(row, 64, "..."))
	}
	return out
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	digits := [20]byte{}
	i := len(digits)
	for v > 0 {
		i--
		digits[i] = byte('0' + (v % 10))
		v /= 10
	}
	if neg {
		i--
		digits[i] = '-'
	}
	return string(digits[i:])
}
