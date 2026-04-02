package tui

import (
	"encoding/json"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestAppPermissionAskTransitionsToPrompt(t *testing.T) {
	app := New(Config{})

	app.handleAgentEvent(types.AgentEvent{
		Type:      types.AgentEventPermissionAsk,
		Turn:      2,
		ToolUseID: "tool-1",
		ToolName:  "bash",
		ToolInput: json.RawMessage(`{"command":"go test ./..."}`),
	})

	if app.state != statePermissionPrompt {
		t.Fatalf("expected permission prompt state after ask, got %d", app.state)
	}
	if app.permDialog.stage != permissionDialogPrompt {
		t.Fatalf("expected permission dialog prompt stage, got %d", app.permDialog.stage)
	}
	if app.permission.toolName != "bash" {
		t.Fatalf("expected permission model for bash tool, got %q", app.permission.toolName)
	}
	if !strings.Contains(app.permission.description, "go test ./...") {
		t.Fatalf("expected permission description to include command, got %q", app.permission.description)
	}
	if len(app.timeline) == 0 || app.timeline[len(app.timeline)-1].permissionState != permissionPending {
		t.Fatalf("expected pending permission timeline row, got %#v", app.timeline)
	}
	if app.activePermissionToolUseID != "tool-1" {
		t.Fatalf("expected active permission tool use id set, got %q", app.activePermissionToolUseID)
	}
}

func TestAppPermissionDecisionResolvesDialogState(t *testing.T) {
	app := New(Config{})
	app.handleAgentEvent(types.AgentEvent{Type: types.AgentEventPermissionAsk, ToolName: "bash"})

	_, _ = app.Update(permissionDecisionMsg{decision: PermissionAlways})

	if app.state != stateThinking {
		t.Fatalf("expected thinking state after decision, got %d", app.state)
	}
	if app.permDialog.stage != permissionDialogResolved {
		t.Fatalf("expected resolved permission dialog stage, got %d", app.permDialog.stage)
	}
	if app.permDialog.decision != PermissionAlways {
		t.Fatalf("expected always decision, got %d", app.permDialog.decision)
	}
	if len(app.timeline) == 0 || app.timeline[len(app.timeline)-1].permissionState != permissionAlwaysStatus {
		t.Fatalf("expected live permission status to show always, got %#v", app.timeline)
	}
}

func TestAppPermissionDecisionIgnoredOutsidePrompt(t *testing.T) {
	app := New(Config{})

	_, _ = app.Update(permissionDecisionMsg{decision: PermissionYes})

	if app.state != stateIdle {
		t.Fatalf("expected idle state when decision arrives outside prompt, got %d", app.state)
	}
	if app.permDialog.stage != permissionDialogHidden {
		t.Fatalf("expected hidden dialog stage, got %d", app.permDialog.stage)
	}
}

func TestAppQuickOpenAndHistorySearchModeWiring(t *testing.T) {
	app := New(Config{})
	app.input.history = []string{"deploy release", "open logs"}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	if app.state != stateSearch {
		t.Fatalf("expected search state for quick-open, got %d", app.state)
	}
	if app.searchMode != searchModeQuickOpen {
		t.Fatalf("expected quick-open mode, got %d", app.searchMode)
	}
	if app.quickOpen.selected != 0 {
		t.Fatalf("expected initial quick-open selection at index 0, got %d", app.quickOpen.selected)
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	if app.quickOpen.selected != 1 {
		t.Fatalf("expected quick-open ctrl+n to advance selection, got %d", app.quickOpen.selected)
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	if app.searchMode != searchModeHistory {
		t.Fatalf("expected history-search mode, got %d", app.searchMode)
	}
	if len(app.history.entries) != len(app.input.history) {
		t.Fatalf("expected history entries copied from input history, got %d entries", len(app.history.entries))
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if app.history.query != "r" {
		t.Fatalf("expected history query sync, got %q", app.history.query)
	}
	if len(app.history.matches) == 0 {
		t.Fatalf("expected history matches after typing query, got none")
	}
}

func TestAppQuickOpenEnterAppliesSelectionAndUpdatesStatus(t *testing.T) {
	app := New(Config{})
	app.width = 120

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	if app.searchMode != searchModeQuickOpen {
		t.Fatalf("expected quick-open mode, got %d", app.searchMode)
	}
	if !strings.Contains(app.renderStatusBar(), "search:quick-open") {
		t.Fatalf("expected status line to include quick-open mode")
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if app.state != stateIdle {
		t.Fatalf("expected idle state after applying quick-open selection, got %d", app.state)
	}
	if app.input.Value() != "/model " {
		t.Fatalf("expected command.model quick-open action to stage /model input, got %q", app.input.Value())
	}
}

func TestAppSearchArrowAndTabNavigation(t *testing.T) {
	app := New(Config{})

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	if app.quickOpen.selected != 0 {
		t.Fatalf("expected initial quick-open selection at index 0, got %d", app.quickOpen.selected)
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyDown})
	if app.quickOpen.selected != 1 {
		t.Fatalf("expected down key to advance quick-open selection, got %d", app.quickOpen.selected)
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if app.quickOpen.selected != 0 {
		t.Fatalf("expected shift+tab to reverse quick-open selection, got %d", app.quickOpen.selected)
	}

	if !strings.Contains(app.renderSearchLine(), "jump:ctrl+n/ctrl+p up/down tab pgup/pgdown") {
		t.Fatalf("expected quick-open search line to advertise extended navigation")
	}
	if !strings.Contains(app.renderSearchLine(), "selected: 1/") {
		t.Fatalf("expected quick-open search line to include selected summary")
	}
}

func TestAppSearchPageNavigationMovesByFiveEntries(t *testing.T) {
	app := New(Config{})
	app.quickOpen = newQuickOpenState([]quickOpenItem{
		{label: "one", value: "one"},
		{label: "two", value: "two"},
		{label: "three", value: "three"},
		{label: "four", value: "four"},
		{label: "five", value: "five"},
		{label: "six", value: "six"},
		{label: "seven", value: "seven"},
	})

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	if app.quickOpen.selected != 0 {
		t.Fatalf("expected initial selection at zero, got %d", app.quickOpen.selected)
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if app.quickOpen.selected != 5 {
		t.Fatalf("expected pgdown to advance by 5, got %d", app.quickOpen.selected)
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	if app.quickOpen.selected != 0 {
		t.Fatalf("expected pgup to reverse by 5, got %d", app.quickOpen.selected)
	}
}

func TestAppHistorySelectionPreservedOnQueryRefinement(t *testing.T) {
	app := New(Config{})
	app.input.history = []string{"deploy release", "ship release", "open logs"}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyDown})
	selectedBefore, ok := app.history.selectedEntry()
	if !ok || selectedBefore.text != "deploy release" {
		t.Fatalf("expected deploy release selected before refinement, got ok=%t text=%q", ok, selectedBefore.text)
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

	selectedAfter, ok := app.history.selectedEntry()
	if !ok || selectedAfter.text != "deploy release" {
		t.Fatalf("expected deploy release to remain selected after refinement, got ok=%t text=%q", ok, selectedAfter.text)
	}
}

func TestAppPermissionQueueCyclesAndShowsQueueIndex(t *testing.T) {
	app := New(Config{})

	app.handleAgentEvent(types.AgentEvent{Type: types.AgentEventPermissionAsk, ToolUseID: "tool-1", ToolName: "bash", Turn: 1})
	app.handleAgentEvent(types.AgentEvent{Type: types.AgentEventPermissionAsk, ToolUseID: "tool-2", ToolName: "read", Turn: 1})

	if app.state != statePermissionPrompt {
		t.Fatalf("expected permission prompt state, got %d", app.state)
	}
	if app.activePermissionToolUseID != "tool-1" {
		t.Fatalf("expected first queued tool to be active, got %q", app.activePermissionToolUseID)
	}
	if app.permission.queueIndex != 1 || app.permission.queueTotal != 2 {
		t.Fatalf("expected queue index 1/2, got %d/%d", app.permission.queueIndex, app.permission.queueTotal)
	}
	if len(app.permission.queueNext) != 1 || !strings.Contains(app.permission.queueNext[0], "read") {
		t.Fatalf("expected queue preview for next action, got %#v", app.permission.queueNext)
	}
	if !strings.Contains(app.permission.View(), "next in queue") {
		t.Fatalf("expected permission view to render queue heading")
	}

	_, _ = app.Update(permissionDecisionMsg{decision: PermissionYes})
	app.handleAgentEvent(types.AgentEvent{Type: types.AgentEventPermissionResult, ToolUseID: "tool-1", ToolName: "bash", Turn: 1, PermissionDecision: types.AgentPermissionAllow})

	if app.activePermissionToolUseID != "tool-2" {
		t.Fatalf("expected second queued tool to become active, got %q", app.activePermissionToolUseID)
	}
	if app.state != statePermissionPrompt {
		t.Fatalf("expected prompt to remain open for queued request, got %d", app.state)
	}
	if app.permission.queueIndex != 1 || app.permission.queueTotal != 1 {
		t.Fatalf("expected queue index 1/1 after dequeue, got %d/%d", app.permission.queueIndex, app.permission.queueTotal)
	}
	if len(app.permission.queueNext) != 0 {
		t.Fatalf("expected no queued preview once one request remains, got %#v", app.permission.queueNext)
	}
}

func TestAppQuickOpenWorkflowActions(t *testing.T) {
	app := New(Config{})
	app.width = 120

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	for i, idx := range app.quickOpen.visible {
		if app.quickOpen.items[idx].value == "search.history" {
			app.quickOpen.selected = i
			break
		}
	}
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if app.searchMode != searchModeHistory {
		t.Fatalf("expected quick-open history action to switch mode, got %d", app.searchMode)
	}

	app.exitSearch()
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.input.Value() != "@" {
		t.Fatalf("expected reference workflow quick-open action to stage @ input, got %q", app.input.Value())
	}
}

func TestAppSlashAndReferenceKeyboardParityKeys(t *testing.T) {
	app := New(Config{})

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if !app.slashAutocomplete.isVisible() {
		t.Fatalf("expected slash autocomplete to open")
	}
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if app.slashAutocomplete.selected != 5 {
		t.Fatalf("expected slash pgdown to move by 5, got %d", app.slashAutocomplete.selected)
	}
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyHome})
	if app.slashAutocomplete.selected != 0 {
		t.Fatalf("expected slash home to jump to start, got %d", app.slashAutocomplete.selected)
	}

	app.slashAutocomplete.clear()
	app.clearReferenceAutocomplete()
	app.input.SetValue("see @int")
	app.syncReferenceAutocomplete()
	if !app.refAuto.active {
		t.Fatalf("expected reference autocomplete to open")
	}
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnd})
	if app.refAuto.selected != len(app.refAuto.suggestions)-1 {
		t.Fatalf("expected reference end to jump to end, got %d", app.refAuto.selected)
	}
}

func TestAppSearchEnterWithNoMatchesStaysInSearchForQuickOpenAndHistory(t *testing.T) {
	app := New(Config{})

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.state != stateSearch || app.searchMode != searchModeQuickOpen {
		t.Fatalf("expected quick-open enter with no matches to stay in search, got state=%d mode=%d", app.state, app.searchMode)
	}

	app.input.history = []string{"deploy release"}
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.state != stateSearch || app.searchMode != searchModeHistory {
		t.Fatalf("expected history enter with no matches to stay in search, got state=%d mode=%d", app.state, app.searchMode)
	}
}

func TestAppPermissionResultUpdatesExistingTimelineRow(t *testing.T) {
	app := New(Config{})
	app.handleAgentEvent(types.AgentEvent{Type: types.AgentEventPermissionAsk, ToolUseID: "tool-2", ToolName: "bash", Turn: 4})
	if len(app.timeline) != 1 {
		t.Fatalf("expected single pending permission row, got %d", len(app.timeline))
	}

	_, _ = app.Update(permissionDecisionMsg{decision: PermissionAlways})
	app.handleAgentEvent(types.AgentEvent{Type: types.AgentEventPermissionResult, ToolUseID: "tool-2", ToolName: "bash", Turn: 4, PermissionDecision: types.AgentPermissionAllow})

	if len(app.timeline) != 1 {
		t.Fatalf("expected permission result to update existing row, got %d rows", len(app.timeline))
	}
	if app.timeline[0].permissionState != permissionAlwaysStatus {
		t.Fatalf("expected permission row to become always, got %s", app.timeline[0].permissionState)
	}
}

func TestAppHistorySearchEnterStagesSelectedEntry(t *testing.T) {
	app := New(Config{})
	app.width = 120
	app.input.history = []string{"deploy release", "open logs"}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	if app.searchMode != searchModeHistory {
		t.Fatalf("expected history-search mode, got %d", app.searchMode)
	}
	if !strings.Contains(app.renderStatusBar(), "search:history-search") {
		t.Fatalf("expected status line to include history-search mode")
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if app.state != stateIdle {
		t.Fatalf("expected idle state after accepting history selection, got %d", app.state)
	}
	if app.input.Value() != "open logs" {
		t.Fatalf("expected selected history entry to populate input, got %q", app.input.Value())
	}
}

func TestAppStatuslineRuntimeMetricsReflectEvents(t *testing.T) {
	app := New(Config{})
	app.width = 120
	app.height = 40
	app.recalcLayout()

	app.handleStreamEvent(types.StreamEvent{Type: types.StreamToolUseStart, ToolUseID: "tool-1", ToolName: "bash"})
	app.handleAgentEvent(types.AgentEvent{Type: types.AgentEventPermissionAsk, ToolUseID: "tool-1", ToolName: "bash", Turn: 1})

	bar := app.renderStatusBar()
	if !strings.Contains(bar, "tools:1") {
		t.Fatalf("expected status bar to include tool metric, got %q", bar)
	}
	if !strings.Contains(bar, "perm:1") {
		t.Fatalf("expected status bar to include permission metric, got %q", bar)
	}
	if app.cmdState.StatuslineTransitions == 0 {
		t.Fatalf("expected statusline transitions to increment")
	}
	if app.cmdState.StatuslineLastRenderMS < 0 {
		t.Fatalf("expected non-negative render duration, got %d", app.cmdState.StatuslineLastRenderMS)
	}
}

func TestAppInputHintSurfaceAppearsInIdleView(t *testing.T) {
	app := New(Config{})
	updated, _ := app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	ready, ok := updated.(*App)
	if !ok {
		t.Fatalf("window update returned %T", updated)
	}
	app = ready

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	view := app.View()
	if !strings.Contains(view, "input: slash command mode") {
		t.Fatalf("expected slash-mode input hint in view, got %q", view)
	}
}

func TestAppQuickOpenResetsQueryAcrossReopen(t *testing.T) {
	app := New(Config{})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	if app.searchQuery != "" {
		t.Fatalf("expected quick-open query reset across reopen, got %q", app.searchQuery)
	}
}

func TestAppSearchHomeEndJumpSelection(t *testing.T) {
	app := New(Config{})
	app.quickOpen = newQuickOpenState([]quickOpenItem{{label: "one"}, {label: "two"}, {label: "three"}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnd})
	if app.quickOpen.selected != 2 {
		t.Fatalf("expected end to jump to last search item, got %d", app.quickOpen.selected)
	}
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyHome})
	if app.quickOpen.selected != 0 {
		t.Fatalf("expected home to jump to first search item, got %d", app.quickOpen.selected)
	}
}

func TestAppPermissionHistoryMetricRecorded(t *testing.T) {
	app := New(Config{})
	app.handleAgentEvent(types.AgentEvent{Type: types.AgentEventPermissionAsk, ToolUseID: "tool-1", ToolName: "bash", Turn: 2})
	_, _ = app.Update(permissionDecisionMsg{decision: PermissionNo})
	app.handleAgentEvent(types.AgentEvent{Type: types.AgentEventPermissionResult, ToolUseID: "tool-1", ToolName: "bash", Turn: 2, PermissionDecision: types.AgentPermissionDeny})
	if len(app.permissionHistory) != 1 {
		t.Fatalf("expected one permission decision in history, got %d", len(app.permissionHistory))
	}
}
