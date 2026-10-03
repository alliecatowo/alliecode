package tui

import (
	"encoding/json"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTeaPermissionQueueEscapeDeniesPromptAndUpdatesTimeline(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	app.handleAgentEvent(types.AgentEvent{Type: types.AgentEventPermissionAsk, ToolUseID: "tool-1", ToolName: "bash", Turn: 3, ToolInput: json.RawMessage(`{"command":"go test ./..."}`)})
	if app.inputMode != inputModePermission {
		t.Fatalf("expected permission input mode, got %s", app.inputMode)
	}

	app = sendTestKeyAndRun(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")
	if app.state != stateThinking {
		t.Fatalf("expected permission escape to resolve prompt and return thinking, got %d", app.state)
	}
	if app.permDialog.decision != PermissionNo || app.permDialog.status != permissionDenied {
		t.Fatalf("expected permission escape to deny, got decision=%d status=%s", app.permDialog.decision, app.permDialog.status)
	}
	if len(app.timeline) == 0 || app.timeline[len(app.timeline)-1].permissionState != permissionDenied {
		t.Fatalf("expected denied permission timeline state, got %#v", app.timeline)
	}
}

func TestTeaPermissionQueueShiftTabReversesSelectionToAlways(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	app.handleAgentEvent(types.AgentEvent{Type: types.AgentEventPermissionAsk, ToolUseID: "tool-1", ToolName: "bash", Turn: 4, ToolInput: json.RawMessage(`{"command":"go test ./..."}`)})
	if app.permission.selected != PermissionYes {
		t.Fatalf("expected default selection yes, got %d", app.permission.selected)
	}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyShiftTab}, "shift+tab")
	if app.permission.selected != PermissionAlways {
		t.Fatalf("expected permission shift+tab to reverse-wrap to always, got %d", app.permission.selected)
	}
}

func TestTeaPermissionQueueShiftTabSequenceReversesSelectionToAlways(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	app.handleAgentEvent(types.AgentEvent{Type: types.AgentEventPermissionAsk, ToolUseID: "tool-1", ToolName: "bash", Turn: 4, ToolInput: json.RawMessage(`{"command":"go test ./..."}`)})
	if app.permission.selected != PermissionYes {
		t.Fatalf("expected default selection yes, got %d", app.permission.selected)
	}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'', '[', 'Z'}}, "shift-tab-csi-z")
	if app.permission.selected != PermissionAlways {
		t.Fatalf("expected permission shift+tab sequence alias to reverse-wrap to always, got %d", app.permission.selected)
	}
}
