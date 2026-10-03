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
	if !strings.Contains(app.renderSearchLine(), "mode:quick-open") {
		t.Fatalf("expected search surface to include quick-open mode")
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
	app.width = 120

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

	line := app.renderSearchLine()
	if !strings.Contains(line, "jump:") || !strings.Contains(line, "ctrl+n/ctrl+p") {
		t.Fatalf("expected quick-open search line to advertise jump navigation")
	}
	if !strings.Contains(app.renderSearchFooterHints(), "shift+tab/up prev") {
		t.Fatalf("expected quick-open footer hints to advertise shift+tab reverse")
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
	if app.quickOpen.offset != 1 {
		t.Fatalf("expected quick-open offset to advance with paged selection, got %d", app.quickOpen.offset)
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	if app.quickOpen.selected != 0 {
		t.Fatalf("expected pgup to reverse by 5, got %d", app.quickOpen.selected)
	}
}

func TestAppSearchPageNavigationClampsAtBounds(t *testing.T) {
	app := New(Config{})
	app.quickOpen = newQuickOpenState([]quickOpenItem{
		{label: "one", value: "one"},
		{label: "two", value: "two"},
		{label: "three", value: "three"},
	})

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if app.quickOpen.selected != 2 {
		t.Fatalf("expected pgdown to clamp at last row, got %d", app.quickOpen.selected)
	}
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if app.quickOpen.selected != 2 {
		t.Fatalf("expected pgdown to stay at end, got %d", app.quickOpen.selected)
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	if app.quickOpen.selected != 0 {
		t.Fatalf("expected pgup to clamp at first row, got %d", app.quickOpen.selected)
	}
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	if app.quickOpen.selected != 0 {
		t.Fatalf("expected pgup to stay at start, got %d", app.quickOpen.selected)
	}
}

func TestAppSearchLineRendersQuickOpenRowsAndSelectionWindow(t *testing.T) {
	app := New(Config{})
	app.width = 100
	app.quickOpen = newQuickOpenState([]quickOpenItem{
		{label: "one", value: "one", detail: "d1", status: "workflow"},
		{label: "two", value: "two", detail: "d2", status: "workflow"},
		{label: "three", value: "three", detail: "d3", status: "workflow"},
		{label: "four", value: "four", detail: "d4", status: "workflow"},
		{label: "five", value: "five", detail: "d5", status: "workflow"},
		{label: "six", value: "six", detail: "d6", status: "workflow"},
		{label: "seven", value: "seven", detail: "d7", status: "workflow"},
	})

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	line := app.renderSearchLine()
	if !strings.Contains(line, "> six") {
		t.Fatalf("expected search line to include selected quick-open row, got %q", line)
	}
	if !strings.Contains(line, "... +1 more") {
		t.Fatalf("expected search line to include overflow marker for windowed rows, got %q", line)
	}
}

func TestAppViewKeepsInputVisibleWhenSearchOpen(t *testing.T) {
	app := New(Config{})
	updated, _ := app.Update(tea.WindowSizeMsg{Width: 100, Height: 26})
	ready, ok := updated.(*App)
	if !ok {
		t.Fatalf("window update returned %T", updated)
	}
	app = ready

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	view := app.View()
	if !strings.Contains(view, "Type a message") {
		t.Fatalf("expected input area to remain visible while search is open, got %q", view)
	}
}

func TestAppViewportHeightStableWhileSearchAndPaletteAreOpen(t *testing.T) {
	app := New(Config{})
	updated, _ := app.Update(tea.WindowSizeMsg{Width: 100, Height: 28})
	ready, ok := updated.(*App)
	if !ok {
		t.Fatalf("window update returned %T", updated)
	}
	app = ready

	idleHeight := app.viewport.Height
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	searchHeight := app.viewport.Height
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	if app.viewport.Height != searchHeight {
		t.Fatalf("expected stable viewport while quick-open query changes, got %d want %d", app.viewport.Height, searchHeight)
	}
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEsc})

	app.input.SetValue("/")
	app.syncSlashAutocomplete()
	app.recalcLayout()
	slashHeight := app.viewport.Height
	if slashHeight == idleHeight {
		t.Fatalf("expected palette-reserved viewport to differ from idle")
	}
	app.input.SetValue("/m")
	app.syncSlashAutocomplete()
	app.recalcLayout()
	if app.viewport.Height != slashHeight {
		t.Fatalf("expected stable viewport while slash palette is open, got %d want %d", app.viewport.Height, slashHeight)
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

	app.handleAgentEvent(types.AgentEvent{Type: types.AgentEventPermissionAsk, ToolUseID: "tool-1", ToolName: "bash", Turn: 1, ToolInput: json.RawMessage(`{"command":"go test ./...","workdir":"/repo"}`)})
	app.handleAgentEvent(types.AgentEvent{Type: types.AgentEventPermissionAsk, ToolUseID: "tool-2", ToolName: "read", Turn: 1, ToolInput: json.RawMessage(`{"file_path":"internal/tui/app.go","offset":"20","limit":10}`)})

	if app.state != statePermissionPrompt {
		t.Fatalf("expected permission prompt state, got %d", app.state)
	}
	if app.activePermissionToolUseID != "tool-1" {
		t.Fatalf("expected first queued tool to be active, got %q", app.activePermissionToolUseID)
	}
	if app.permission.queueIndex != 1 || app.permission.queueTotal != 2 {
		t.Fatalf("expected queue index 1/2, got %d/%d", app.permission.queueIndex, app.permission.queueTotal)
	}
	if app.permission.toolKind != "bash" {
		t.Fatalf("expected active permission tool kind metadata, got %q", app.permission.toolKind)
	}
	if len(app.permission.toolDetails) == 0 {
		t.Fatalf("expected active permission tool details, got %#v", app.permission.toolDetails)
	}
	if len(app.permission.queueNext) != 1 || !strings.Contains(app.permission.queueNext[0], "read") {
		t.Fatalf("expected queue preview for next action, got %#v", app.permission.queueNext)
	}
	if !strings.Contains(app.permission.queueNext[0], "[PENDING]") {
		t.Fatalf("expected queue preview status marker, got %#v", app.permission.queueNext)
	}
	if !strings.Contains(app.permission.View(), "next in queue") {
		t.Fatalf("expected permission view to render queue heading")
	}
	if !strings.Contains(app.permission.View(), "command details:") {
		t.Fatalf("expected permission prompt to show command detail panel")
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
	if app.permission.toolKind != "file" {
		t.Fatalf("expected active permission to switch to file kind, got %q", app.permission.toolKind)
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

func TestAppReferenceKeyboardRightAppliesAndLeftDismisses(t *testing.T) {
	app := New(Config{})
	app.input.SetValue("see @int")
	app.syncReferenceAutocomplete()
	if !app.refAuto.active {
		t.Fatalf("expected reference autocomplete to open")
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRight})
	if app.refAuto.active {
		t.Fatalf("expected right key to apply and close reference autocomplete")
	}
	if !strings.Contains(app.input.Value(), "@") {
		t.Fatalf("expected applied reference insertion, got %q", app.input.Value())
	}

	app.input.SetValue("see @int")
	app.syncReferenceAutocomplete()
	if !app.refAuto.active {
		t.Fatalf("expected reference autocomplete to reopen")
	}
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if app.refAuto.active {
		t.Fatalf("expected left key to dismiss reference autocomplete")
	}
}

func TestAppSearchEnterWithNoMatchesExitsSearchForQuickOpenAndHistory(t *testing.T) {
	app := New(Config{})

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.state != stateIdle {
		t.Fatalf("expected quick-open enter with no matches to exit search, got state=%d mode=%d", app.state, app.searchMode)
	}

	app.input.history = []string{"deploy release"}
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.state != stateIdle {
		t.Fatalf("expected history enter with no matches to exit search, got state=%d mode=%d", app.state, app.searchMode)
	}
}

func TestAppTimelineSearchSurfaceShowsReverseTraversalHints(t *testing.T) {
	app := New(Config{})
	app.width = 120
	app.addTimeline(timelineEntry{kind: timelineAssistant, text: "status ready"})

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlF})
	line := app.renderSearchLine()
	if !strings.Contains(line, "shift+tab/up prev") {
		t.Fatalf("expected timeline search footer to advertise shift+tab reverse, got %q", line)
	}
	if !strings.Contains(line, "jump:ctrl+n/ctrl+p") || !strings.Contains(line, "shift+tab") {
		t.Fatalf("expected timeline search jump hint to include shift+tab, got %q", line)
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
	if !strings.Contains(app.renderSearchLine(), "mode:history-search") {
		t.Fatalf("expected search surface to include history-search mode")
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
	app.width = 180
	app.height = 40
	app.recalcLayout()

	app.handleStreamEvent(types.StreamEvent{Type: types.StreamToolUseStart, ToolUseID: "tool-1", ToolName: "bash"})
	app.handleAgentEvent(types.AgentEvent{Type: types.AgentEventPermissionAsk, ToolUseID: "tool-1", ToolName: "bash", Turn: 1})

	bar := app.renderStatusBar()
	runtime := app.renderStatusRuntimePanes()
	if !strings.Contains(bar, "tools:1") && !strings.Contains(runtime, "tools=1") {
		t.Fatalf("expected status surfaces to include tool metric, bar=%q runtime=%q", bar, runtime)
	}
	if !strings.Contains(bar, "perm:1") && !strings.Contains(runtime, "permissions=1") {
		t.Fatalf("expected status surfaces to include permission metric, bar=%q runtime=%q", bar, runtime)
	}
	if app.cmdState.StatuslineTransitions == 0 {
		t.Fatalf("expected statusline transitions to increment")
	}
	if app.cmdState.StatuslineLastRenderMS < 0 {
		t.Fatalf("expected non-negative render duration, got %d", app.cmdState.StatuslineLastRenderMS)
	}
}

func TestAppStatuslineCompactSpacingAndOrder(t *testing.T) {
	app := New(Config{})
	app.width = 120
	app.height = 40
	app.recalcLayout()

	bar := app.renderStatusBar()
	if !strings.Contains(bar, "-/unknown  |  default  |  input chat") {
		t.Fatalf("expected compact statusline core fields with separators, got %q", bar)
	}
	if !strings.Contains(bar, "tok 0  cost $0.0000  n/a") {
		t.Fatalf("expected compact right-side budget and session block, got %q", bar)
	}
	if strings.Contains(bar, "render:") {
		t.Fatalf("expected telemetry fields hidden in idle compact view, got %q", bar)
	}
}

func TestAppStatuslineRuntimePaneHiddenInIdleAndShownForOverlayContext(t *testing.T) {
	app := New(Config{})
	app.width = 120
	app.height = 28
	app.recalcLayout()

	if pane := app.renderStatusRuntimePanes(); strings.TrimSpace(pane) != "" {
		t.Fatalf("expected runtime pane hidden in idle view, got %q", pane)
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	pane := stripANSIForTest(app.renderStatusRuntimePanes())
	if !strings.Contains(pane, "runtime: idle") {
		t.Fatalf("expected runtime pane when command palette overlay opens, got %q", pane)
	}
	if !strings.Contains(pane, "context: command /") {
		t.Fatalf("expected command context in runtime pane, got %q", pane)
	}
}

func TestAppViewportOffsetStaysAnchoredAcrossOverlayTransitions(t *testing.T) {
	app := New(Config{})
	updated, _ := app.Update(tea.WindowSizeMsg{Width: 120, Height: 32})
	ready, ok := updated.(*App)
	if !ok {
		t.Fatalf("window update returned %T", updated)
	}
	app = ready

	for i := 0; i < 80; i++ {
		app.addTimeline(timelineEntry{kind: timelineAssistant, text: strings.Repeat("line ", 20), turn: i + 1})
	}
	app.followTail = false
	app.viewport.SetYOffset(25)
	offsetBefore := app.viewport.YOffset

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	offsetPalette := app.viewport.YOffset
	if offsetPalette < 0 {
		t.Fatalf("expected non-negative offset while palette open, got %d", offsetPalette)
	}
	if offsetPalette > offsetBefore {
		t.Fatalf("expected anchored or clamped offset while palette opens, got %d want <= %d", offsetPalette, offsetBefore)
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	offsetAfter := app.viewport.YOffset
	if offsetAfter < 0 {
		t.Fatalf("expected non-negative offset after overlay closes, got %d", offsetAfter)
	}
	if absInt(offsetAfter-offsetPalette) > 2 {
		t.Fatalf("expected stable offset across overlay transition, got before=%d open=%d after=%d", offsetBefore, offsetPalette, offsetAfter)
	}
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
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
	if !strings.Contains(view, "input mode: slash") {
		t.Fatalf("expected slash-mode input hint in view, got %q", view)
	}
}

func TestAppSlashAutocompleteEscDownstreamInputSemantics(t *testing.T) {
	app := New(Config{})
	updated, _ := app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	ready, ok := updated.(*App)
	if !ok {
		t.Fatalf("window update returned %T", updated)
	}
	app = ready

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if !app.slashAutocomplete.isVisible() {
		t.Fatalf("expected slash autocomplete to open")
	}
	view := app.View()
	if !strings.Contains(view, "commands: /") {
		t.Fatalf("expected slash drawer in view")
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if app.slashAutocomplete.isVisible() {
		t.Fatalf("expected esc to dismiss slash drawer")
	}
	if got := app.input.Value(); got != "" && got != "/" {
		t.Fatalf("expected input to preserve or clear slash root after esc, got %q", got)
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if len(app.timeline) != 0 {
		t.Fatalf("expected enter on bare slash to avoid submission, timeline=%d", len(app.timeline))
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if got := app.input.Value(); got != "" {
		t.Fatalf("expected backspace to clear slash root, got %q", got)
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if len(app.timeline) != 0 {
		t.Fatalf("expected enter on empty input to remain no-op, timeline=%d", len(app.timeline))
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

func TestAppQuickOpenShiftTabSequenceAliasReversesSelection(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyDown})
	initial := app.quickOpen.selected
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'', '[', 'Z'}})
	if app.quickOpen.selected == initial {
		t.Fatalf("expected shift+tab sequence alias to reverse quick-open selection")
	}
}
