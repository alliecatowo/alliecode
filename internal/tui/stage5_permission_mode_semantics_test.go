package tui

import (
	"encoding/json"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStage5PermissionModeSemanticsShiftTabCyclesReverseToAlways(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	app.handleAgentEvent(types.AgentEvent{Type: types.AgentEventPermissionAsk, ToolUseID: "tool-1", ToolName: "bash", Turn: 8, ToolInput: json.RawMessage(`{"command":"go test ./..."}`)})
	if app.permission.selected != PermissionYes {
		t.Fatalf("expected default decision yes, got %d", app.permission.selected)
	}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyShiftTab}, "shift+tab")
	if app.permission.selected != PermissionAlways {
		t.Fatalf("expected reverse traversal to wrap to always, got %d", app.permission.selected)
	}
}

func TestStage5PermissionModeSemanticsEscDeniesAndReturnsThinking(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	app.handleAgentEvent(types.AgentEvent{Type: types.AgentEventPermissionAsk, ToolUseID: "tool-2", ToolName: "bash", Turn: 9, ToolInput: json.RawMessage(`{"command":"go vet ./..."}`)})
	app = sendTestKeyAndRun(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")
	if app.state != stateThinking {
		t.Fatalf("expected prompt resolved and state thinking, got %d", app.state)
	}
	if app.permDialog.decision != PermissionNo {
		t.Fatalf("expected esc deny decision, got %d", app.permDialog.decision)
	}
}
