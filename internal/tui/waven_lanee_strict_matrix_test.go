package tui

import (
	"encoding/json"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestWaveNLaneEShiftedKeyReverseMatrixStaysDocked(t *testing.T) {
	t.Run("quick_open", func(t *testing.T) {
		app := readySizedApp(t, 180, 30)
		app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
		app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyDown}, "down")
		app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyShiftTab}, "shift+tab")
		if app.searchMode != searchModeQuickOpen || app.quickOpen.selected != 0 {
			t.Fatalf("expected quick-open reverse to return to first row, mode=%d selected=%d", app.searchMode, app.quickOpen.selected)
		}
	})

	t.Run("slash", func(t *testing.T) {
		app := readySizedApp(t, 180, 30)
		app = typeTestText(t, app, "/")
		app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyTab}, "tab")
		app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyShiftTab}, "shift+tab")
		if app.inputMode != inputModeSlash || !app.slashAutocomplete.isVisible() || app.slashAutocomplete.selected != 0 {
			t.Fatalf("expected slash reverse cycle to remain open at first item, mode=%s selected=%d", app.inputMode, app.slashAutocomplete.selected)
		}
	})

	t.Run("reference", func(t *testing.T) {
		app := readySizedApp(t, 180, 30)
		app = typeTestText(t, app, "@READ")
		app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyTab}, "tab")
		app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyShiftTab}, "shift+tab")
		if app.inputMode != inputModeReference || !app.refAuto.active || app.refAuto.selected != 0 {
			t.Fatalf("expected reference reverse cycle to remain open at first item, mode=%s selected=%d", app.inputMode, app.refAuto.selected)
		}
	})

	t.Run("command_panel", func(t *testing.T) {
		app := readySizedApp(t, 180, 30)
		app = typeTestText(t, app, "/perm")
		app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
		app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyTab}, "tab")
		app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyShiftTab}, "shift+tab")
		if app.inputMode != inputModeCommandPanel || !app.commandPanel.active || app.commandPanel.selected != 0 {
			t.Fatalf("expected panel reverse cycle to remain open at first item, mode=%s selected=%d", app.inputMode, app.commandPanel.selected)
		}
	})

	t.Run("model_picker", func(t *testing.T) {
		app := readySizedApp(t, 180, 30)
		app = typeTestText(t, app, "/model ")
		app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyTab}, "tab")
		app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyShiftTab}, "shift+tab")
		if app.inputMode != inputModeModelPicker || !app.modelPickerActive() || app.slashAutocomplete.modelPicker.selected != 0 {
			t.Fatalf("expected model-picker reverse cycle to remain open at first item, mode=%s selected=%d", app.inputMode, app.slashAutocomplete.modelPicker.selected)
		}
	})
}

func TestWaveNLaneEDrawerDockingOrderStrict(t *testing.T) {
	type dockCase struct {
		name    string
		actions []tea.KeyMsg
		typed   string
		marker  string
	}
	cases := []dockCase{
		{name: "slash", typed: "/", marker: "commands: /"},
		{name: "reference", typed: "@", marker: "references:"},
		{name: "quick_open", actions: []tea.KeyMsg{{Type: tea.KeyCtrlO}}, marker: "search:"},
		{name: "command_panel", typed: "/tasks", actions: []tea.KeyMsg{{Type: tea.KeyEnter}}, marker: "drawer: tasks panel: /tasks"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			app := readySizedApp(t, 140, 30)
			if tc.typed != "" {
				app = typeTestText(t, app, tc.typed)
			}
			for _, msg := range tc.actions {
				app = sendTestKey(t, app, msg, "action")
			}
			plain := stripANSIForTest(app.View())
			drawerIdx := strings.Index(plain, tc.marker)
			inputIdx := strings.Index(plain, "┃")
			buddyIdx := strictBuddyIndex(plain)
			statusIdx := strings.Index(plain, "hints:")
			if drawerIdx < 0 || inputIdx < 0 || buddyIdx < 0 || statusIdx < 0 {
				t.Fatalf("expected drawer/input/buddy/status rows in %s view:\n%s", tc.name, plain)
			}
			if !(drawerIdx < inputIdx && inputIdx < buddyIdx && buddyIdx < statusIdx) {
				t.Fatalf("expected strict docking order drawer->input->buddy->status for %s:\n%s", tc.name, plain)
			}
		})
	}
}

func TestWaveNLaneECompactBuddyStatuslineRemainsStrict(t *testing.T) {
	app := readySizedApp(t, 120, 26)
	plain := stripANSIForTest(app.View())
	if !strings.Contains(plain, "-/unknown  |  default  |  input chat") {
		t.Fatalf("expected compact idle statusline core fields, got:\n%s", plain)
	}
	if !strings.Contains(plain, "`-vvvv-`") && !strings.Contains(plain, "<°~°> buddy") {
		t.Fatalf("expected buddy block visible in compact idle view, got:\n%s", plain)
	}
	if strings.Contains(plain, "runtime: state=") {
		t.Fatalf("expected runtime pane hidden in idle compact view, got:\n%s", plain)
	}
}

func TestWaveNLaneEMultiToolLoopRuntimeContinuityCounters(t *testing.T) {
	app := readySizedApp(t, 180, 30)

	app.handleStreamEvent(types.StreamEvent{Type: types.StreamToolUseStart, ToolUseID: "tool-1", ToolName: "bash", Input: json.RawMessage(`{"command":"go test ./..."}`)})
	app.handleAgentEvent(types.AgentEvent{Type: types.AgentEventPermissionAsk, ToolUseID: "tool-1", ToolName: "bash", Turn: 1, ToolInput: json.RawMessage(`{"command":"go test ./..."}`)})
	app.handleAgentEvent(types.AgentEvent{Type: types.AgentEventPermissionResult, ToolUseID: "tool-1", ToolName: "bash", Turn: 1, PermissionDecision: types.AgentPermissionAllow})
	app.handleStreamEvent(types.StreamEvent{Type: types.StreamToolUseDone, ToolUseID: "tool-1", ToolName: "bash"})

	app.handleStreamEvent(types.StreamEvent{Type: types.StreamToolUseStart, ToolUseID: "tool-2", ToolName: "read", Input: json.RawMessage(`{"file_path":"internal/tui/app.go","offset":1,"limit":20}`)})
	app.handleAgentEvent(types.AgentEvent{Type: types.AgentEventPermissionAsk, ToolUseID: "tool-2", ToolName: "read", Turn: 1, ToolInput: json.RawMessage(`{"file_path":"internal/tui/app.go","offset":1,"limit":20}`)})
	app.handleAgentEvent(types.AgentEvent{Type: types.AgentEventPermissionResult, ToolUseID: "tool-2", ToolName: "read", Turn: 1, PermissionDecision: types.AgentPermissionAllow})
	app.handleStreamEvent(types.StreamEvent{Type: types.StreamToolUseDone, ToolUseID: "tool-2", ToolName: "read"})

	pane := stripANSIForTest(app.renderStatusRuntimePanes())
	if !strings.Contains(pane, "runtime: ") || !strings.Contains(pane, "tools=2") || !strings.Contains(pane, "permissions=4") {
		t.Fatalf("expected runtime continuity pane in idle after loop, got %q", pane)
	}
	if app.cmdState.StatuslineTransitions == 0 {
		t.Fatalf("expected statusline transitions to advance during loop")
	}
	if len(app.timeline) < 2 {
		t.Fatalf("expected timeline tool rows retained, got %d", len(app.timeline))
	}
}

func strictBuddyIndex(view string) int {
	if idx := strings.Index(view, "`-vvvv-`"); idx >= 0 {
		return idx
	}
	return strings.Index(view, "<°~°> buddy")
}
