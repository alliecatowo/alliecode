package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alliecatowo/alliecode/internal/agent"
	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/keybindings"
	"github.com/alliecatowo/alliecode/internal/permissions"
	"github.com/alliecatowo/alliecode/internal/providers"
	"github.com/alliecatowo/alliecode/internal/references"
	"github.com/alliecatowo/alliecode/internal/types"
)

type appState int

const (
	stateIdle appState = iota
	stateThinking
	stateStreaming
	statePermissionPrompt
	stateSearch
)

type streamEventMsg struct {
	event types.StreamEvent
}

type agentEventMsg struct {
	event types.AgentEvent
}

type agentDoneMsg struct{}

type agentErrorMsg struct {
	err error
}

type timelineKind string

const (
	timelineUser       timelineKind = "user"
	timelineAssistant  timelineKind = "assistant"
	timelineTool       timelineKind = "tool"
	timelinePermission timelineKind = "permission"
	timelineError      timelineKind = "error"
	maxToolPreviewLen               = 96
)

type toolProgressState string

const (
	toolProgressPending toolProgressState = "pending"
	toolProgressRunning toolProgressState = "running"
	toolProgressDone    toolProgressState = "done"
	toolProgressFailed  toolProgressState = "failed"
	toolProgressDenied  toolProgressState = "denied"
)

type permissionStatus string

const (
	permissionPending      permissionStatus = "pending"
	permissionApproved     permissionStatus = "approved"
	permissionDenied       permissionStatus = "denied"
	permissionAlwaysStatus permissionStatus = "always"
)

type timelineEntry struct {
	kind             timelineKind
	text             string
	toolName         string
	toolUseID        string
	toolSummary      string
	toolInputPreview string
	toolInputBytes   int
	toolState        toolProgressState
	permissionState  permissionStatus
	turn             int
	intents          []types.RenderIntent
}

type Config struct {
	Agent           *agent.Agent
	Version         string
	Debug           bool
	InitialModel    string
	InitialState    commands.RuntimeState
	ActiveSessionID string
}

type App struct {
	cfg Config

	input      InputModel
	viewport   viewport.Model
	spinner    SpinnerModel
	permission PermissionModel

	state     appState
	streamBuf strings.Builder
	width     int
	height    int
	ready     bool
	err       error
	model     string

	totalTokens int
	costUSD     float64
	turns       int

	timeline             []timelineEntry
	toolRows             map[string]int
	permissionRows       map[string]int
	visibleRows          []int
	visibleLines         []int
	timelineMatches      []timelineSearchMatch
	timelineVersion      int
	cachedTimeline       timelineCache
	followTail           bool
	lastRenderedLines    int
	prevViewportHeight   int
	statusPanelsHeight   int
	overlayPanelsHeight  int
	composerPanelsHeight int
	layoutRecalcInFlight bool
	anchorLockOffset     int
	anchorLockActive     bool
	streamCacheWidth     int
	streamCacheChars     int
	streamCacheBlock     string
	streamCacheLines     int

	searchQuery          string
	searchTimelineQuery  string
	searchQuickOpenQuery string
	searchHistoryQuery   string
	matchPos             int
	searchMode           searchMode
	quickOpen            quickOpenState
	history              historySearchState
	permDialog           permissionDialogState
	buddy                buddyState
	now                  func() time.Time
	recentRefs           []string
	openRefs             []string
	contextRefs          []string
	resolver             *references.Resolver
	refAuto              referenceAutocompleteState

	commands *commands.Registry
	cmdState commands.RuntimeState
	keySet   *keybindings.Set

	slashAutocomplete slashAutocompleteState
	commandPanel      commandPanelState

	activePermissionToolUseID string
	activePermissionQueueKey  string
	activePermissionTurn      int
	permissionQueueSeq        int
	permissionQueue           []permissionPromptRequest
	permissionHistory         []permissionDecisionRecord
	inputMode                 inputMode
	modalStack                modalStackState
	store                     tuiStore
}

type timelineCache struct {
	ready      bool
	version    int
	width      int
	query      string
	content    string
	visible    []int
	lineOffset []int
	matches    []timelineSearchMatch
	totalLines int
}

const (
	statusRuntimePanelStableHeight = 3
)

func (a *App) setState(next appState) {
	if a.stateValue() != next {
		a.cmdState.StatuslineTransitions++
	}
	stateChanged := a.stateValue() != next
	a.setStateValue(next)
	a.syncModalStack()
	a.syncInputMode()
	if stateChanged && a.ready {
		a.recalcLayout()
	}
}

func (a *App) noteToolRuntimeEvent() {
	a.cmdState.StatuslineToolCalls++
}

func (a *App) notePermissionRuntimeEvent() {
	a.cmdState.StatuslinePermissions++
}

func New(cfg Config) *App {
	model := strings.TrimSpace(cfg.InitialState.Model)
	if model == "" {
		model = cfg.InitialModel
	}
	if model == "" {
		model = "unknown"
	}
	cmdState := cfg.InitialState
	if strings.TrimSpace(cmdState.Model) == "" {
		cmdState.Model = model
	}
	cmdState.PermissionMode = permissions.ModeDefault
	cmdState.Agent = cfg.Agent
	if hasInitialRuntimeSelection(cmdState) {
		commands.HydrateRuntimeSelection(&cmdState)
	}
	if strings.TrimSpace(cfg.ActiveSessionID) == "" {
		cfg.ActiveSessionID = cmdState.SessionID
	}

	return &App{
		cfg:            cfg,
		input:          NewInput(),
		spinner:        NewSpinner("thinking"),
		state:          stateIdle,
		toolRows:       make(map[string]int),
		permissionRows: make(map[string]int),
		followTail:     true,
		searchMode:     searchModeTimeline,
		quickOpen: newQuickOpenState([]quickOpenItem{
			{label: "Search timeline", detail: "Filter visible conversation rows", status: "navigation", keywords: "find filter history", value: "search.timeline", hint: "ctrl+f"},
			{label: "Quick open", detail: "Browse workflows and common actions", status: "navigation", keywords: "palette actions", value: "search.quick_open", hint: "ctrl+o"},
			{label: "Change model", detail: "Stage /model command in input", status: "command", keywords: "switch model", value: "command.model", hint: "/model"},
			{label: "Set permissions auto", detail: "Stage /permissions auto command", status: "command", keywords: "permissions auto allow", value: "command.permissions_auto", hint: "/permissions"},
			{label: "Set permissions default", detail: "Stage /permissions default command", status: "command", keywords: "permissions default", value: "command.permissions_default", hint: "/permissions"},
			{label: "History search", detail: "Open scored prompt history picker", status: "history", keywords: "ctrl+r search prompt", value: "search.history", hint: "ctrl+r"},
			{label: "Reuse last prompt", detail: "Select latest history entry", status: "history", keywords: "repeat previous prompt", value: "history.last", hint: "ctrl+p"},
			{label: "Permission queue", detail: "Inspect live permission approvals", status: "workflow", keywords: "permission allow deny queue", value: "workflow.permission", hint: "runtime"},
			{label: "Reference picker", detail: "Use @path autocomplete with preview", status: "workflow", keywords: "references files context", value: "workflow.references", hint: "@"},
			{label: "Insert slash command", detail: "Open command palette from input", status: "workflow", keywords: "slash commands", value: "workflow.slash", hint: "/"},
		}),
		history:           newHistorySearchState(nil),
		permDialog:        newPermissionDialogState(),
		buddy:             newBuddyState(),
		now:               time.Now,
		resolver:          references.NewResolver(resolveBaseDir(cfg)),
		commands:          commands.DefaultRegistry(),
		keySet:            defaultKeybindingSet(),
		cmdState:          cmdState,
		slashAutocomplete: slashAutocompleteState{selected: -1},
		commandPanel:      commandPanelState{selected: -1},
		inputMode:         inputModeChat,
		store:             newTUIStore(),
	}
}

func hasInitialRuntimeSelection(state commands.RuntimeState) bool {
	if strings.TrimSpace(state.ProviderName) != "" || strings.TrimSpace(state.ModelRef) != "" {
		return true
	}
	if model := strings.TrimSpace(state.Model); model != "" && !strings.EqualFold(model, "unknown") {
		return true
	}
	if strings.TrimSpace(state.Runtime.ProviderName) != "" || strings.TrimSpace(state.Runtime.ModelRef) != "" {
		return true
	}
	if model := strings.TrimSpace(state.Runtime.Model); model != "" && !strings.EqualFold(model, "unknown") {
		return true
	}
	return false
}

func resolveBaseDir(cfg Config) string {
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return ""
}

func (a *App) Run() error {
	p := tea.NewProgram(a, tea.WithAltScreen())
	if a.cfg.Agent != nil {
		a.cfg.Agent.SetStreamCallback(MakeStreamCallback(p))
		a.cfg.Agent.SetEventCallback(MakeEventCallback(p))
	}
	_, err := p.Run()
	return err
}

func (a *App) Init() tea.Cmd {
	a.syncModalStack()
	a.syncInputMode()
	return tea.Batch(a.input.Init(), a.spinner.Init(), tea.EnterAltScreen)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	if isScrollActivityMsg(msg) {
		a.buddy = clearBuddyReaction(a.buddy)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.recalcLayout()
		a.ready = true
		return a, nil

	case tea.KeyMsg:
		switch {
		case keyMatches(msg, "ctrl+c"):
			return a, tea.Quit
		}

		if handled, cmd := a.handleModeKey(msg); handled {
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return a, tea.Batch(cmds...)
		}

	case submitMsg:
		if a.stateValue() != stateIdle {
			return a, nil
		}
		userText := strings.TrimSpace(msg.text)
		if userText == "" {
			return a, nil
		}
		a.captureRecentReferences(userText)
		if strings.HasPrefix(strings.TrimSpace(userText), "/") {
			res, err := a.commands.Dispatch(context.Background(), commands.Context{State: &a.cmdState}, userText)
			if err != nil {
				a.addTimeline(timelineEntry{kind: timelineError, text: err.Error()})
				return a, nil
			}
			if res.Handled && (strings.TrimSpace(res.Message) != "" || len(res.RenderIntents) > 0) {
				a.addTimeline(timelineEntry{kind: timelineAssistant, text: res.Message, intents: append([]types.RenderIntent(nil), res.RenderIntents...)})
			}
			return a, nil
		}

		selection := commands.RuntimeSelectionTruth(&a.cmdState)
		providerName := strings.TrimSpace(selection.ProviderName)
		modelName := strings.TrimSpace(selection.ModelName)
		if providerName == "" || modelName == "" {
			a.addTimeline(timelineEntry{kind: timelineError, text: "runtime selection is incomplete; next: run /provider set <name> then /model <provider>/<model>"})
			return a, nil
		}
		if !selection.ProviderReady && !strings.EqualFold(providerName, "ollama") {
			a.addTimeline(timelineEntry{kind: timelineError, text: fmt.Sprintf("provider %s needs login; next: run /login provider %s and retry", providerName, providerName)})
			return a, nil
		}
		if a.cfg.Agent == nil {
			a.addTimeline(timelineEntry{kind: timelineError, text: "agent is not initialized; next: run /status to inspect runtime, then restart session"})
			return a, nil
		}

		a.turns++
		a.addTimeline(timelineEntry{kind: timelineUser, text: userText, turn: a.turns})
		a.setState(stateThinking)
		a.spinner = NewSpinner("thinking")
		a.input.SetEnabled(false)

		return a, func() tea.Msg {
			err := a.cfg.Agent.Run(nil, userText)
			if err != nil {
				return agentErrorMsg{err: err}
			}
			return agentDoneMsg{}
		}

	case streamEventMsg:
		return a.handleStreamEvent(msg.event)

	case agentEventMsg:
		a.handleAgentEvent(msg.event)
		return a, nil

	case agentDoneMsg:
		if a.streamBuf.Len() > 0 {
			a.addTimeline(timelineEntry{kind: timelineAssistant, text: a.streamBuf.String(), turn: a.turns})
			a.streamBuf.Reset()
			a.resetStreamRenderCache()
		}
		a.updateUsage()
		a.setState(stateIdle)
		a.input.SetEnabled(true)
		a.input.Focus()
		return a, nil

	case agentErrorMsg:
		a.err = msg.err
		a.addTimeline(timelineEntry{kind: timelineError, text: msg.err.Error(), turn: a.turns})
		a.setState(stateIdle)
		a.input.SetEnabled(true)
		a.input.Focus()
		return a, nil

	case permissionDecisionMsg:
		if a.stateValue() != statePermissionPrompt || a.permDialog.stage != permissionDialogPrompt {
			return a, nil
		}
		a.permDialog = a.permDialog.transition(permissionDialogEventFromDecision(msg.decision))
		a.permission.SetStatus(a.permDialog.status)
		a.updatePermissionProgress(a.activePermissionToolUseID, a.permission.toolName, a.activePermissionTurn, a.permDialog.status)
		if a.permDialog.stage != permissionDialogResolved {
			return a, nil
		}
		a.setState(stateThinking)
		a.spinner = NewSpinner("executing")
		return a, nil
	}

	switch a.stateValue() {
	case stateIdle:
		var cmd tea.Cmd
		wasPanelVisible := a.commandPanel.active
		wasSlashVisible := a.slashAutocomplete.isVisible()
		wasModelPickerVisible := a.modelPickerActive()
		wasRefVisible := a.refAuto.active
		a.input, cmd = a.input.Update(msg)
		cmds = append(cmds, cmd)
		a.syncSlashAutocomplete()
		a.syncModelPicker()
		a.syncReferenceAutocomplete()
		a.syncInputMode()
		if wasPanelVisible != a.commandPanel.active || wasSlashVisible != a.slashAutocomplete.isVisible() || wasModelPickerVisible != a.modelPickerActive() || wasRefVisible != a.refAuto.active {
			a.recalcLayout()
		}
	case stateThinking:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	var cmd tea.Cmd
	a.viewport, cmd = a.viewport.Update(msg)
	cmds = append(cmds, cmd)
	a.followTail = a.viewport.AtBottom()

	return a, tea.Batch(cmds...)
}

func (a *App) View() string {
	if !a.ready {
		return "initializing alliecode..."
	}
	layout := a.composeMeasuredLayout(a.width)
	var sections []string
	sections = a.appendPanels(sections, layout.header.panels)
	sections = append(sections, a.viewport.View())
	sections = a.appendPanels(sections, layout.overlays.panels)
	sections = a.appendPanels(sections, layout.composer.panels)
	sections = a.appendPanels(sections, layout.buddy.panels)
	sections = a.appendPanels(sections, layout.status.panels)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (a *App) syncSlashAutocomplete() {
	if a.stateValue() != stateIdle {
		a.slashAutocomplete.clear()
		a.syncInputMode()
		return
	}
	if a.commandPanel.active {
		a.slashAutocomplete.clear()
		a.syncInputMode()
		return
	}
	if a.modelPickerActive() {
		a.slashAutocomplete.clear()
		a.syncInputMode()
		return
	}
	value := strings.TrimLeft(a.input.Value(), " \t")
	if !strings.HasPrefix(value, "/") {
		a.slashAutocomplete.clear()
		a.syncInputMode()
		return
	}
	tail := value[1:]
	if strings.ContainsAny(tail, " \t\n\r") {
		a.slashAutocomplete.clear()
		a.syncInputMode()
		return
	}
	query := strings.ToLower(strings.TrimSpace(tail))
	a.slashAutocomplete.setItems(query, a.commands.Suggestions(query))
	a.syncInputMode()
}

func (a *App) syncModelPicker() {
	if a.stateValue() != stateIdle {
		a.clearModelPicker()
		a.syncInputMode()
		return
	}
	if a.commandPanel.active {
		a.clearModelPicker()
		a.syncInputMode()
		return
	}
	value := strings.TrimLeft(a.input.Value(), " \t")
	normalized := strings.TrimSpace(value)
	if !shouldOpenModelPicker(value) {
		a.slashAutocomplete.modelPicker.dismissedInput = ""
		a.clearModelPicker()
		a.syncInputMode()
		return
	}
	if strings.EqualFold(a.slashAutocomplete.modelPicker.dismissedInput, normalized) {
		return
	}
	items := a.buildModelPickerItems()
	if len(items) == 0 {
		a.clearModelPicker()
		a.syncInputMode()
		return
	}
	prevKey := ""
	if item, ok := a.selectedModelPickerItem(); ok {
		prevKey = item.provider + "/" + item.model.Model
	}
	a.slashAutocomplete.modelPicker.active = true
	a.slashAutocomplete.modelPicker.items = items
	a.slashAutocomplete.modelPicker.offset = 0
	if prevKey != "" {
		for i, item := range items {
			if item.provider+"/"+item.model.Model == prevKey {
				a.slashAutocomplete.modelPicker.selected = i
				a.ensureModelPickerVisible(6)
				return
			}
		}
	}
	if current := strings.TrimSpace(a.cmdState.ModelRef); current != "" {
		for i, item := range items {
			if strings.EqualFold(item.provider+"/"+item.model.Model, current) {
				a.slashAutocomplete.modelPicker.selected = i
				a.ensureModelPickerVisible(6)
				return
			}
		}
	}
	if a.slashAutocomplete.modelPicker.selected < 0 || a.slashAutocomplete.modelPicker.selected >= len(items) {
		a.slashAutocomplete.modelPicker.selected = 0
	}
	a.ensureModelPickerVisible(6)
	a.slashAutocomplete.clear()
	a.syncInputMode()
}

func shouldOpenModelPicker(value string) bool {
	normalized := strings.TrimLeft(value, " \t")
	lower := strings.ToLower(normalized)
	if !strings.HasPrefix(lower, "/model") {
		return false
	}
	if len(normalized) <= len("/model") {
		return false
	}
	tail := normalized[len("/model"):]
	if strings.TrimSpace(tail) != "" {
		return false
	}
	return strings.ContainsAny(tail, " \t\n\r")
}

func (a *App) buildModelPickerItems() []modelPickerItem {
	providersList := providers.SupportedProviderNames()
	items := make([]modelPickerItem, 0, 16)
	for _, providerName := range providersList {
		ready := commands.ProviderReadyForState(providerName, &a.cmdState)
		models := providers.ListModelsByProvider(providerName)
		for _, model := range models {
			items = append(items, modelPickerItem{provider: providerName, model: model, ready: ready})
		}
	}
	return items
}

func (a *App) modelPickerActive() bool {
	return a.slashAutocomplete.modelPicker.active && len(a.slashAutocomplete.modelPicker.items) > 0
}

func (a *App) clearModelPicker() {
	a.slashAutocomplete.modelPicker.active = false
	a.slashAutocomplete.modelPicker.items = nil
	a.slashAutocomplete.modelPicker.selected = -1
	a.slashAutocomplete.modelPicker.offset = 0
	a.slashAutocomplete.modelPicker.showSlashHeader = false
	a.syncInputMode()
}

func (a *App) dismissModelPicker() {
	if !a.modelPickerActive() {
		return
	}
	a.slashAutocomplete.modelPicker.dismissedInput = strings.TrimSpace(strings.TrimLeft(a.input.Value(), " \t"))
	a.clearModelPicker()
}

func (a *App) selectedModelPickerItem() (modelPickerItem, bool) {
	if !a.modelPickerActive() {
		return modelPickerItem{}, false
	}
	selected := a.slashAutocomplete.modelPicker.selected
	if selected < 0 || selected >= len(a.slashAutocomplete.modelPicker.items) {
		return modelPickerItem{}, false
	}
	return a.slashAutocomplete.modelPicker.items[selected], true
}

func (a *App) moveModelPickerSelection(dir int) {
	if !a.modelPickerActive() {
		return
	}
	a.slashAutocomplete.modelPicker.selected = nextMatchPos(a.slashAutocomplete.modelPicker.selected, len(a.slashAutocomplete.modelPicker.items), dir)
	a.ensureModelPickerVisible(6)
}

func (a *App) pageModelPickerSelection(dir int) {
	if !a.modelPickerActive() {
		return
	}
	step := 5
	sign := 1
	if dir < 0 {
		sign = -1
	}
	next := a.slashAutocomplete.modelPicker.selected + sign*step
	if next < 0 {
		next = 0
	}
	if next >= len(a.slashAutocomplete.modelPicker.items) {
		next = len(a.slashAutocomplete.modelPicker.items) - 1
	}
	a.slashAutocomplete.modelPicker.selected = next
	a.ensureModelPickerVisible(6)
}

func (a *App) jumpModelPickerSelection(toEnd bool) {
	if !a.modelPickerActive() {
		return
	}
	if toEnd {
		a.slashAutocomplete.modelPicker.selected = len(a.slashAutocomplete.modelPicker.items) - 1
	} else {
		a.slashAutocomplete.modelPicker.selected = 0
	}
	a.ensureModelPickerVisible(6)
}

func (a *App) ensureModelPickerVisible(window int) {
	if !a.modelPickerActive() {
		a.slashAutocomplete.modelPicker.offset = 0
		return
	}
	if window <= 0 {
		window = 6
	}
	selected := a.slashAutocomplete.modelPicker.selected
	offset := a.slashAutocomplete.modelPicker.offset
	if selected < 0 {
		a.slashAutocomplete.modelPicker.offset = 0
		return
	}
	if selected < offset {
		a.slashAutocomplete.modelPicker.offset = selected
		return
	}
	if selected >= offset+window {
		a.slashAutocomplete.modelPicker.offset = selected - window + 1
	}
	maxOffset := len(a.slashAutocomplete.modelPicker.items) - window
	if maxOffset < 0 {
		maxOffset = 0
	}
	if a.slashAutocomplete.modelPicker.offset > maxOffset {
		a.slashAutocomplete.modelPicker.offset = maxOffset
	}
}

func (a *App) applyModelPickerSelection() (string, bool) {
	item, ok := a.selectedModelPickerItem()
	if !ok {
		return "", false
	}
	command := fmt.Sprintf("/model %s/%s", item.provider, item.model.Model)
	a.input.Reset()
	a.slashAutocomplete.modelPicker.dismissedInput = ""
	a.clearModelPicker()
	a.slashAutocomplete.clear()
	return command, true
}

func (a *App) renderModelPicker() string {
	if !a.modelPickerActive() {
		return ""
	}
	lines := []string{"model picker: /model", strings.Repeat("-", 20)}
	if a.slashAutocomplete.modelPicker.showSlashHeader {
		lines = append(lines, "commands: /model")
	}
	rowWidth := a.width - 2
	if rowWidth < 1 {
		rowWidth = 1
	}
	window := 6
	start := a.slashAutocomplete.modelPicker.offset
	if start < 0 || start >= len(a.slashAutocomplete.modelPicker.items) {
		start = 0
	}
	end := start + window
	if end > len(a.slashAutocomplete.modelPicker.items) {
		end = len(a.slashAutocomplete.modelPicker.items)
	}
	lastProvider := ""
	providerCounts := make(map[string]int, 8)
	for _, item := range a.slashAutocomplete.modelPicker.items {
		providerCounts[item.provider]++
	}
	for i := start; i < end; i++ {
		item := a.slashAutocomplete.modelPicker.items[i]
		if item.provider != lastProvider {
			providerState := "login"
			if item.ready {
				providerState = "ready"
			}
			lines = append(lines, fmt.Sprintf("  %s (%d) [%s]:", item.provider, providerCounts[item.provider], providerState))
			lastProvider = item.provider
		}
		prefix := "  "
		if i == a.slashAutocomplete.modelPicker.selected {
			prefix = "> "
		}
		marker := "[auth]"
		if item.ready {
			marker = "[ready]"
		}
		caps := item.model.CapabilitySummary()
		row := fmt.Sprintf("%s%-28s  %-7s  ctx:%-7d  caps:%s", prefix, item.provider+"/"+item.model.Model, marker, item.model.ContextWindow, caps)
		lines = append(lines, truncateDisplayWidth(row, rowWidth, "..."))
	}
	if selected, ok := a.selectedModelPickerItem(); ok {
		lines = append(lines, "")
		lines = append(lines, truncateDisplayWidth("selected: /model "+selected.provider+"/"+selected.model.Model, rowWidth, "..."))
		providerState := "requires login"
		if selected.ready {
			providerState = "provider ready"
		}
		lines = append(lines, truncateDisplayWidth("provider: "+selected.provider+" ("+providerState+")", rowWidth, "..."))
	}
	if len(a.slashAutocomplete.modelPicker.items) > window {
		remaining := len(a.slashAutocomplete.modelPicker.items) - end
		if remaining > 0 {
			lines = append(lines, fmt.Sprintf("  ... +%d more", remaining))
		}
	}
	lines = append(lines, "")
	lines = append(lines, truncateDisplayWidth("enter select+run  esc dismiss  up/down navigate  pgup/pgdown page  home/end jump", rowWidth, "..."))
	return slashAutocompleteStyle.Render(strings.Join(lines, "\n"))
}

func (a *App) applySlashAutocompleteSelection() (applied bool, immediate bool, submitText string) {
	if token, ok := exactSlashToken(a.input.Value()); ok {
		if cmd, exists := a.commands.Lookup(token); exists {
			usage := normalizeSlashUsage(cmd.Usage())
			if strings.EqualFold(token, "model") {
				a.input.SetValue("/model ")
				a.slashAutocomplete.modelPicker.showSlashHeader = true
				a.slashAutocomplete.clear()
				a.syncSlashAutocomplete()
				a.syncModelPicker()
				a.syncInputMode()
				if a.ready {
					a.recalcLayout()
				}
				return true, false, ""
			}
			if commands.SupportsInteractivePanel(token) {
				if applied, immediate, submitText := a.applyInteractiveCommandPanelSlashDefault(token); applied {
					a.slashAutocomplete.clear()
					a.syncInputMode()
					return true, immediate, submitText
				}
				if a.openInteractiveCommandPanel(token) {
					a.slashAutocomplete.clear()
					a.syncInputMode()
					if a.ready {
						a.recalcLayout()
					}
					return true, false, ""
				}
			}
			if slashUsageExpectsArgs(usage) {
				a.input.SetValue("/" + token + " ")
				a.slashAutocomplete.clear()
				a.syncSlashAutocomplete()
				a.syncInputMode()
				if a.ready {
					a.recalcLayout()
				}
				return true, false, ""
			}
			if isImmediateSlashCommandSafe(token) {
				a.input.Reset()
				a.slashAutocomplete.clear()
				a.syncInputMode()
				return true, true, "/" + token
			}
			a.input.SetValue("/" + token)
			a.slashAutocomplete.clear()
			a.syncSlashAutocomplete()
			a.syncInputMode()
			if a.ready {
				a.recalcLayout()
			}
			return true, false, ""
		}
	}
	item, ok := a.slashAutocomplete.selectedItem()
	query := strings.TrimSpace(strings.TrimPrefix(a.input.Value(), "/"))
	if query != "" {
		for _, candidate := range a.slashAutocomplete.items {
			name := strings.TrimSpace(candidate.Name)
			if strings.EqualFold(name, query) || strings.HasPrefix(strings.ToLower(name), strings.ToLower(query)) {
				item = candidate
				ok = true
				break
			}
		}
	}
	if !ok {
		return false, false, ""
	}
	if applied, immediate, submitText := a.applyInteractiveCommandPanelSlashDefault(item.Name); applied {
		a.slashAutocomplete.clear()
		a.syncInputMode()
		return true, immediate, submitText
	}
	if commands.SupportsInteractivePanel(item.Name) {
		if a.openInteractiveCommandPanel(item.Name) {
			a.slashAutocomplete.clear()
			a.syncInputMode()
			if a.ready {
				a.recalcLayout()
			}
			return true, false, ""
		}
	}
	if strings.EqualFold(item.Name, "model") || slashUsageExpectsArgs(item.Usage) {
		cmdText := "/" + item.Name + " "
		if strings.EqualFold(item.Name, "model") {
			a.slashAutocomplete.modelPicker.showSlashHeader = true
		}
		a.input.SetValue(cmdText)
		a.slashAutocomplete.clear()
		a.syncSlashAutocomplete()
		if strings.EqualFold(item.Name, "model") {
			a.syncModelPicker()
		}
		a.syncInputMode()
		if a.ready {
			a.recalcLayout()
		}
		return true, false, ""
	}
	if !slashUsageExpectsArgs(item.Usage) && isImmediateSlashCommandSafe(item.Name) {
		a.input.Reset()
		a.slashAutocomplete.clear()
		a.syncInputMode()
		return true, true, "/" + item.Name
	}
	cmdText := "/" + item.Name
	a.input.SetValue(cmdText)
	a.slashAutocomplete.clear()
	a.syncSlashAutocomplete()
	if strings.EqualFold(item.Name, "model") {
		a.syncModelPicker()
	}
	a.syncInputMode()
	if a.ready {
		a.recalcLayout()
	}
	return true, false, ""
}

func (a *App) renderSlashAutocomplete() string {
	header := "commands: /"
	if q := strings.TrimSpace(a.slashAutocomplete.query); q != "" {
		header += q
	}
	rowWidth := a.width - 2
	if rowWidth < 1 {
		rowWidth = 1
	}
	lines := []string{header, panelDivider(rowWidth)}
	selected, ok := a.slashAutocomplete.selectedItem()
	if !ok && len(a.slashAutocomplete.items) > 0 {
		selected = a.slashAutocomplete.items[0]
		ok = true
	}
	if ok {
		meta := strings.TrimSpace(selected.MatchReason)
		if meta == "" {
			meta = "browse"
		}
		row := fmt.Sprintf("> /%s  (%s)", selected.Name, meta)
		lines = append(lines, truncateDisplayWidth(row, rowWidth, "..."))
		lines = append(lines, truncateDisplayWidth("preview: /"+selected.Name, rowWidth, "..."))
		lines = append(lines, truncateDisplayWidth(fmt.Sprintf("selection: %d/%d", a.slashAutocomplete.selected+1, len(a.slashAutocomplete.items)), rowWidth, "..."))
		if usage := strings.TrimSpace(normalizeSlashUsage(selected.Usage)); usage != "" {
			lines = append(lines, truncateDisplayWidth("usage: "+usage, rowWidth, "..."))
		}
	} else {
		lines = append(lines, "  no matching commands")
	}
	if len(a.slashAutocomplete.items) > 1 {
		lines = append(lines, truncateDisplayWidth(fmt.Sprintf("results: %d", len(a.slashAutocomplete.items)), rowWidth, "..."))
	}
	lines = append(lines, truncateDisplayWidth(drawerNavHelpSlashRef, rowWidth, "..."))
	if len(lines) > slashPanelMaxLines {
		lines = lines[:slashPanelMaxLines]
	}
	return slashAutocompleteStyle.Render(strings.Join(lines, "\n"))
}

func slashSuggestionSection(reason string) string {
	norm := strings.ToLower(strings.TrimSpace(reason))
	switch {
	case norm == "exact" || strings.Contains(norm, "exact"):
		return "Best Match"
	case strings.Contains(norm, "prefix"):
		return "Prefix Matches"
	case strings.Contains(norm, "contains") || strings.Contains(norm, "token"):
		return "Contains"
	case strings.Contains(norm, "fuzzy"):
		return "Fuzzy"
	default:
		return "Browse"
	}
}

func (a *App) renderCommandContextHint() string {
	value := strings.TrimSpace(a.input.Value())
	if !strings.HasPrefix(value, "/") {
		return ""
	}
	inv, err := commands.Parse(value)
	if err != nil {
		trimmed := strings.TrimSpace(strings.TrimPrefix(value, "/"))
		if trimmed == "" {
			return ""
		}
		parts := strings.Fields(trimmed)
		if len(parts) == 0 {
			return ""
		}
		inv.Name = strings.ToLower(strings.TrimSpace(parts[0]))
	}
	cmd, ok := a.commands.Lookup(inv.Name)
	if !ok {
		return ""
	}
	usage := strings.TrimSpace(cmd.Usage())
	if usage == "" {
		return ""
	}
	desc := strings.TrimSpace(cmd.Description())
	width := a.width - 2
	if width < 20 {
		width = 20
	}
	line := "hint: " + usage
	if desc != "" {
		line += "  -  " + desc
	}
	if slashUsageExpectsArgs(usage) {
		line += "  (add args, then enter)"
	} else if token, ok := exactSlashToken(value); ok && token == strings.ToLower(strings.TrimSpace(inv.Name)) && isImmediateSlashCommandSafe(token) {
		line += "  (enter runs now)"
	} else if isImmediateSlashCommandSafe(inv.Name) {
		line += "  (enter runs now)"
	} else {
		line += "  (enter to run)"
	}
	return slashAutocompleteStyle.Render(truncateDisplayWidth(line, width, "..."))
}

func slashUsageExpectsArgs(usage string) bool {
	usage = normalizeSlashUsage(usage)
	if usage == "" {
		return false
	}
	parts := strings.Fields(usage)
	if len(parts) <= 1 {
		return false
	}
	tail := strings.TrimSpace(strings.TrimPrefix(usage, parts[0]))
	if tail == "" {
		return false
	}
	if strings.Contains(tail, "<") {
		return true
	}
	if strings.Contains(tail, "[") {
		return true
	}
	normTail := strings.ReplaceAll(tail, "|", " ")
	for _, token := range strings.Fields(normTail) {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if strings.HasPrefix(token, "[") && strings.HasSuffix(token, "]") {
			continue
		}
		if strings.HasPrefix(token, "(") && strings.HasSuffix(token, ")") {
			continue
		}
		return true
	}
	return false
}

func normalizeSlashUsage(usage string) string {
	usage = strings.TrimSpace(usage)
	lower := strings.ToLower(usage)
	if strings.HasPrefix(lower, "usage:") {
		usage = strings.TrimSpace(usage[len("usage:"):])
	}
	return usage
}

func isImmediateSlashCommandSafe(name string) bool {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "exit", "logout", "clear", "reset-limits", "upgrade":
		return false
	default:
		return true
	}
}

func exactSlashToken(inputValue string) (string, bool) {
	value := strings.TrimLeft(inputValue, " \t")
	if !strings.HasPrefix(value, "/") {
		return "", false
	}
	tail := strings.TrimSpace(strings.TrimPrefix(value, "/"))
	if tail == "" || strings.ContainsAny(tail, " \t\n\r") {
		return "", false
	}
	return strings.ToLower(tail), true
}

func (a *App) handleStreamEvent(ev types.StreamEvent) (tea.Model, tea.Cmd) {
	switch ev.Type {
	case types.StreamStart:
		a.setState(stateStreaming)
		a.streamBuf.Reset()
		a.resetStreamRenderCache()

	case types.StreamContentDelta:
		a.streamBuf.WriteString(ev.Delta)
		a.refreshViewport()

	case types.StreamThinkingDelta:
		if a.cfg.Debug {
			a.streamBuf.WriteString(ev.Delta)
			a.refreshViewport()
		}

	case types.StreamToolUseStart:
		a.noteToolRuntimeEvent()
		a.captureToolReferences(ev.ToolName, ev.Input)
		toolSummary := summarizeToolInput(ev.ToolName, ev.Input)
		idx := a.addTimeline(timelineEntry{
			kind:             timelineTool,
			toolName:         ev.ToolName,
			toolUseID:        ev.ToolUseID,
			toolSummary:      toolSummary,
			toolInputPreview: truncateDisplayWidth(toolSummary, maxToolPreviewLen, "..."),
			toolInputBytes:   len(toolSummary),
			toolState:        toolProgressRunning,
			turn:             a.turns,
		})
		if ev.ToolUseID != "" {
			a.toolRows[ev.ToolUseID] = idx
		}

	case types.StreamToolUseDelta:
		a.updateToolProgress(ev.ToolUseID, ev.Delta, toolProgressRunning)

	case types.StreamToolUseDone:
		a.updateToolProgress(ev.ToolUseID, "", toolProgressDone)

	case types.StreamContentDone:
		if a.streamBuf.Len() > 0 {
			a.addTimeline(timelineEntry{kind: timelineAssistant, text: a.streamBuf.String(), turn: a.turns})
			a.streamBuf.Reset()
			a.resetStreamRenderCache()
		}
		a.setState(stateThinking)

	case types.StreamMessageDone:
		if ev.Usage != nil {
			a.totalTokens += ev.Usage.InputTokens + ev.Usage.OutputTokens
		}

	case types.StreamError:
		errText := "stream error"
		if ev.Error != nil {
			errText = ev.Error.Error()
		}
		a.addTimeline(timelineEntry{kind: timelineError, text: errText, turn: a.turns})
	}

	return a, nil
}

func (a *App) handleAgentEvent(ev types.AgentEvent) {
	switch ev.Type {
	case types.AgentEventPermissionAsk:
		a.notePermissionRuntimeEvent()
		queueKey := strings.TrimSpace(ev.ToolUseID)
		if queueKey == "" {
			a.permissionQueueSeq++
			queueKey = fmt.Sprintf("permission-%d", a.permissionQueueSeq)
		}
		request := permissionPromptRequest{
			queueKey:  queueKey,
			toolUseID: ev.ToolUseID,
			toolName:  ev.ToolName,
			turn:      ev.Turn,
			status:    permissionPending,
		}
		promptContext := permissionPromptContextFromInput(ev.ToolName, ev.ToolInput)
		request.description = promptContext.summary
		request.toolKind = promptContext.kind
		request.toolDetails = append([]string(nil), promptContext.details...)
		idx := a.addTimeline(timelineEntry{
			kind:            timelinePermission,
			toolName:        ev.ToolName,
			toolUseID:       ev.ToolUseID,
			permissionState: permissionPending,
			turn:            ev.Turn,
		})
		request.timelineRow = idx
		a.permissionQueue = append(a.permissionQueue, request)
		if ev.ToolUseID != "" {
			a.permissionRows[ev.ToolUseID] = idx
		}
		a.ensurePermissionPromptVisible()
	case types.AgentEventPermissionResult:
		a.notePermissionRuntimeEvent()
		selectedDecision := a.permDialog.decision
		resolvedActive := false
		if strings.TrimSpace(ev.ToolUseID) != "" {
			resolvedActive = ev.ToolUseID == a.activePermissionToolUseID
		} else if strings.TrimSpace(a.activePermissionQueueKey) != "" {
			resolvedActive = true
		} else if a.stateValue() == statePermissionPrompt && a.permDialog.stage == permissionDialogPrompt {
			resolvedActive = true
		}
		state := permissionDenied
		toolState := toolProgressDenied
		decision := PermissionNo
		if ev.PermissionDecision == types.AgentPermissionAllow {
			state = permissionApproved
			if resolvedActive && selectedDecision == PermissionAlways {
				state = permissionAlwaysStatus
				decision = PermissionAlways
			} else {
				decision = PermissionYes
			}
			toolState = toolProgressRunning
		} else if resolvedActive {
			decision = selectedDecision
		}
		if decision == PermissionUndecided {
			if state == permissionDenied {
				decision = PermissionNo
			} else {
				decision = PermissionYes
			}
		}
		for i := range a.permissionQueue {
			request := &a.permissionQueue[i]
			if strings.TrimSpace(ev.ToolUseID) != "" {
				if request.toolUseID != ev.ToolUseID {
					continue
				}
			} else if strings.TrimSpace(a.activePermissionQueueKey) != "" {
				if request.queueKey != a.activePermissionQueueKey {
					continue
				}
			} else {
				continue
			}
			request.status = state
			break
		}
		a.updatePermissionProgress(ev.ToolUseID, ev.ToolName, ev.Turn, state)
		a.permissionHistory = append(a.permissionHistory, permissionDecisionRecord{
			ToolName:  ev.ToolName,
			Decision:  decision,
			Status:    state,
			Turn:      ev.Turn,
			QueueKey:  a.activePermissionQueueKey,
			Timestamp: time.Now().Unix(),
		})
		if len(a.permissionHistory) > 32 {
			a.permissionHistory = a.permissionHistory[len(a.permissionHistory)-32:]
		}
		a.updateToolProgress(ev.ToolUseID, "", toolState)
		if resolvedActive {
			a.dequeueActivePermissionPrompt()
			a.permDialog = a.permDialog.transition(permissionDialogReset)
		}
		a.ensurePermissionPromptVisible()
		if a.stateValue() != statePermissionPrompt {
			a.setState(stateThinking)
			a.spinner = NewSpinner("executing")
		}
	case types.AgentEventToolEnd:
		if strings.TrimSpace(ev.ToolError) != "" {
			a.updateToolProgress(ev.ToolUseID, ev.ToolError, toolProgressFailed)
		}
	}
}

func (a *App) addTimeline(entry timelineEntry) int {
	a.buddy = updateBuddyState(a.buddy, entry, a.now())
	a.timeline = append(a.timeline, entry)
	a.timelineVersion++
	a.refreshViewport()
	return len(a.timeline) - 1
}

func (a *App) updateToolProgress(toolUseID, delta string, state toolProgressState) {
	idx, ok := a.toolRows[toolUseID]
	if !ok || idx < 0 || idx >= len(a.timeline) {
		return
	}
	row := a.timeline[idx]
	if state != "" {
		row.toolState = state
	}
	if strings.TrimSpace(delta) != "" {
		row.toolInputBytes += len(strings.TrimSpace(delta))
		row.toolInputPreview = buildToolPreview(row.toolInputPreview, delta, maxToolPreviewLen)
	}
	a.timeline[idx] = row
	a.timelineVersion++
	a.refreshViewport()
}

func summarizeToolInput(toolName string, input json.RawMessage) string {
	name := strings.TrimSpace(strings.ToLower(toolName))
	if len(input) == 0 {
		return ""
	}
	payload := map[string]any{}
	if err := json.Unmarshal(input, &payload); err != nil {
		text := strings.TrimSpace(string(input))
		return truncateDisplayWidth(text, 80, "...")
	}
	switch name {
	case "bash":
		return summarizeToolFields(payload, []string{"command", "description"})
	case "read":
		return summarizeToolFields(payload, []string{"filePath", "offset", "limit"})
	case "grep":
		return summarizeToolFields(payload, []string{"pattern", "include", "path"})
	case "glob":
		return summarizeToolFields(payload, []string{"pattern", "path"})
	case "write":
		return summarizeToolFields(payload, []string{"filePath"})
	case "edit":
		return summarizeToolFields(payload, []string{"filePath", "oldString", "newString"})
	default:
		return summarizeToolFields(payload, []string{"description", "command", "filePath", "pattern", "url"})
	}
}

func summarizeToolFields(payload map[string]any, keys []string) string {
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		value, ok := payload[key]
		if !ok {
			continue
		}
		s := stringifyToolFieldValue(value)
		if strings.TrimSpace(s) == "" {
			continue
		}
		s = strings.Join(strings.Fields(s), " ")
		s = truncateDisplayWidth(s, 40, "...")
		parts = append(parts, key+"="+s)
	}
	if len(parts) > 0 {
		return strings.Join(parts, " | ")
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return truncateDisplayWidth(strings.Join(strings.Fields(string(b)), " "), 80, "...")
}

func stringifyToolFieldValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case float64:
		if typed == float64(int64(typed)) {
			return fmt.Sprintf("%d", int64(typed))
		}
		return fmt.Sprintf("%g", typed)
	case bool:
		if typed {
			return "true"
		}
		return "false"
	default:
		b, err := json.Marshal(typed)
		if err != nil {
			return ""
		}
		return string(b)
	}
}

func (a *App) updatePermissionProgress(toolUseID, toolName string, turn int, state permissionStatus) {
	if strings.TrimSpace(toolUseID) == "" {
		for i := len(a.timeline) - 1; i >= 0; i-- {
			if a.timeline[i].kind != timelinePermission {
				continue
			}
			row := a.timeline[i]
			row.permissionState = state
			a.timeline[i] = row
			a.timelineVersion++
			a.refreshViewport()
			return
		}
		return
	}

	idx, ok := a.permissionRows[toolUseID]
	if !ok || idx < 0 || idx >= len(a.timeline) {
		if strings.TrimSpace(toolName) == "" {
			toolName = a.permission.toolName
		}
		if turn <= 0 {
			turn = a.activePermissionTurn
		}
		idx = a.addTimeline(timelineEntry{
			kind:            timelinePermission,
			toolName:        toolName,
			toolUseID:       toolUseID,
			permissionState: state,
			turn:            turn,
		})
		a.permissionRows[toolUseID] = idx
		return
	}
	row := a.timeline[idx]
	row.permissionState = state
	a.timeline[idx] = row
	a.timelineVersion++
	a.refreshViewport()
}

func (a *App) timelineFilterQuery() string {
	if !a.timelineSearchActive() {
		return ""
	}
	return strings.TrimSpace(a.searchQueryValue())
}

func (a *App) timelineFrame(width int) (string, []int, []int, []timelineSearchMatch, int) {
	query := a.timelineFilterQuery()
	if a.cachedTimeline.ready && a.cachedTimeline.version == a.timelineVersion && a.cachedTimeline.width == width && a.cachedTimeline.query == query {
		return a.cachedTimeline.content, a.cachedTimeline.visible, a.cachedTimeline.lineOffset, a.cachedTimeline.matches, a.cachedTimeline.totalLines
	}
	content, visible, lineOffsets, matches, totalLines := renderTimeline(a.timeline, query, width)
	a.cachedTimeline = timelineCache{
		ready:      true,
		version:    a.timelineVersion,
		width:      width,
		query:      query,
		content:    content,
		visible:    visible,
		lineOffset: lineOffsets,
		matches:    matches,
		totalLines: totalLines,
	}
	return content, visible, lineOffsets, matches, totalLines
}

func (a *App) resetStreamRenderCache() {
	a.streamCacheWidth = 0
	a.streamCacheChars = 0
	a.streamCacheBlock = ""
	a.streamCacheLines = 0
}

func (a *App) streamRenderBlock(width int) (string, int) {
	chars := a.streamBuf.Len()
	if chars <= 0 {
		a.resetStreamRenderCache()
		return "", 0
	}
	if a.streamCacheWidth == width && a.streamCacheChars == chars && a.streamCacheBlock != "" {
		return a.streamCacheBlock, a.streamCacheLines
	}
	block := assistantLabelStyle.Render("AI") + "\n" + assistantTextStyle.Render(a.streamBuf.String())
	lines := visualLineCount(block, width)
	a.streamCacheWidth = width
	a.streamCacheChars = chars
	a.streamCacheBlock = block
	a.streamCacheLines = lines
	return block, lines
}

func (a *App) captureAnchorLock() {
	if !a.ready {
		a.anchorLockActive = false
		return
	}
	if a.followTail || a.viewport.AtBottom() {
		a.anchorLockActive = false
		return
	}
	a.anchorLockOffset = a.viewport.YOffset
	a.anchorLockActive = true
}

func (a *App) refreshViewport() {
	if a.ready && !a.layoutRecalcInFlight {
		if a.syncPanelLayoutIfNeeded() {
			return
		}
	}
	width := a.viewport.Width
	if width <= 0 {
		width = 1
	}
	content, visible, lineOffsets, matches, totalLines := a.timelineFrame(width)
	a.visibleRows = visible
	a.visibleLines = lineOffsets
	a.timelineMatches = matches
	prevYOffset := a.viewport.YOffset
	prevAtBottom := a.viewport.AtBottom()
	prevViewportHeight := a.prevViewportHeight
	if prevViewportHeight <= 0 {
		prevViewportHeight = a.viewport.Height
	}
	a.prevViewportHeight = a.viewport.Height

	if streamBlock, streamLines := a.streamRenderBlock(width); streamLines > 0 {
		if content != "" {
			content += "\n\n"
			totalLines += 2
		}
		content += streamBlock
		totalLines += streamLines
	}

	a.viewport.SetContent(content)

	if a.timelineSearchActive() && len(a.timelineMatches) > 0 && a.searchQueryValue() != "" && a.searchMatchPosValue() >= 0 && a.searchMatchPosValue() < len(a.timelineMatches) {
		a.viewport.SetYOffset(clampYOffset(a.timelineMatches[a.searchMatchPosValue()].lineOffset, totalLines, a.viewport.Height))
		a.anchorLockActive = false
		a.lastRenderedLines = totalLines
		return
	}

	if a.anchorLockActive {
		a.viewport.SetYOffset(clampYOffset(a.anchorLockOffset, totalLines, a.viewport.Height))
		a.anchorLockActive = false
		a.lastRenderedLines = totalLines
		return
	}

	offset, toBottom := resolveScrollAnchor(prevYOffset, prevAtBottom, a.followTail, a.lastRenderedLines, totalLines, prevViewportHeight, a.viewport.Height)
	a.lastRenderedLines = totalLines
	if toBottom {
		a.viewport.GotoBottom()
		return
	}
	a.viewport.SetYOffset(clampYOffset(offset, totalLines, a.viewport.Height))
}

func (a *App) syncPanelLayoutIfNeeded() bool {
	layout := a.composeMeasuredLayout(a.width)
	if layout.status.height == a.statusPanelsHeight && layout.overlays.height == a.overlayPanelsHeight && layout.composer.height == a.composerPanelsHeight {
		return false
	}
	a.recalcLayout()
	return true
}

func (a *App) recalcLayout() {
	a.layoutRecalcInFlight = true
	defer func() {
		a.layoutRecalcInFlight = false
	}()
	layout := a.composeMeasuredLayout(a.width)

	reserved := layout.reserved
	vpHeight := a.height - reserved
	if vpHeight < 1 {
		vpHeight = 1
	}

	prevViewportHeight := a.viewport.Height
	if !a.ready {
		a.viewport = viewport.New(a.width, vpHeight)
		a.viewport.YPosition = 0
	} else {
		a.viewport.Width = a.width
		a.viewport.Height = vpHeight
	}
	a.prevViewportHeight = prevViewportHeight
	a.input.SetWidth(a.width)
	a.permission.SetWidth(a.width - 4)
	a.statusPanelsHeight = layout.status.height
	a.overlayPanelsHeight = layout.overlays.height
	a.composerPanelsHeight = layout.composer.height
	a.refreshViewport()
}

func (a *App) statusRuntimePanelHeight() int {
	return a.panelStackHeight([]panelSurface{{key: "status-runtime", content: a.renderStatusRuntimePanes(), minLines: 1, maxLines: statusRuntimePanelStableHeight}}, a.width)
}

func (a *App) activityPanelHeight() int {
	return a.panelStackHeight(a.activityPanels(), a.width)
}

func (a *App) renderStatusBar() string {
	start := time.Now()
	defer func() {
		a.cmdState.StatuslineLastRenderMS = int(time.Since(start).Milliseconds())
	}()
	if a.width < 100 {
		return a.renderStatusBarLegacy()
	}
	return a.renderStatusBarCompact()
}

func (a *App) renderStatusBarCompact() string {
	leftParts := a.primaryStatusParts()
	if a.showContextualStatusDetails() {
		leftParts = append(leftParts, a.primaryStatusInlineDetails()...)
	}
	left := strings.Join(leftParts, "  |  ")
	sessionID := a.cfg.ActiveSessionID
	if strings.TrimSpace(sessionID) == "" {
		sessionID = "n/a"
	}
	right := fmt.Sprintf("tok %d  cost $%.4f  %s", a.totalTokens, a.costUSD, sessionID)
	return renderStatusLineWithRight(left, right, a.width)
}

func (a *App) renderStatusBarLegacy() string {
	leftParts := a.primaryStatusParts()
	if a.showContextualStatusDetails() {
		leftParts = append(leftParts, a.primaryStatusInlineDetails()...)
	}
	if a.viewport.Height > 0 && a.stateValue() != stateIdle {
		leftParts = append(leftParts, fmt.Sprintf("vp %d/%d", a.viewport.YOffset, a.viewport.Height))
	}
	left := strings.Join(leftParts, "  ")
	sessionID := a.cfg.ActiveSessionID
	if strings.TrimSpace(sessionID) == "" {
		sessionID = "n/a"
	}
	right := fmt.Sprintf("tok %d  cost $%.4f  %s", a.totalTokens, a.costUSD, sessionID)
	return renderStatusLineWithRight(left, right, a.width)
}

func (a *App) renderStreamingLine() string {
	return streamingStyle.Render("streaming...")
}

func (a *App) renderStatusHints() string {
	if !a.shouldShowStatusHints() {
		return ""
	}
	hints := make([]string, 0, 4)
	switch a.inputModeValue() {
	case inputModeSearch, inputModeQuickOpen, inputModeHistorySearch:
		hints = append(hints, "search active")
	case inputModeCommandPanel:
		hints = append(hints, "command panel")
	case inputModeSlash:
		hints = append(hints, "command palette")
	case inputModeModelPicker:
		hints = append(hints, "model picker")
	case inputModeReference:
		hints = append(hints, "reference palette")
	case inputModePermission:
		hints = append(hints, "permission review")
	default:
		if providerName := a.statusProviderLabel(); providerName != "-" {
			hints = append(hints, "provider "+providerName)
		}
	}
	if context := strings.TrimSpace(a.activeContextHint()); context != "" {
		hints = append(hints, context)
	}
	hints = append(hints, "input mode: "+a.inputModeValue().label())
	if len(hints) > 3 {
		hints = hints[:3]
	}
	if len(hints) == 0 {
		return ""
	}
	line := "hints: " + strings.Join(hints, " | ")
	width := a.width - 2
	if width < 1 {
		width = 1
	}
	return statusSubtleStyle.Render(" " + truncateDisplayWidth(line, width, "...") + " ")
}

func (a *App) renderStatusRuntimePanes() string {
	if !a.shouldShowStatusRuntimePane() {
		return ""
	}
	width := a.width - 2
	if width < 1 {
		width = 1
	}
	panes := make([]string, 0, 3)
	showRuntime := a.runtimeContextVisible() || a.cmdState.StatuslineToolCalls > 0 || a.cmdState.StatuslinePermissions > 0 || strings.TrimSpace(a.searchTimelineQuery) != "" || strings.TrimSpace(a.searchQuickOpenQuery) != "" || strings.TrimSpace(a.searchHistoryQuery) != ""
	if showRuntime {
		runtime := fmt.Sprintf("runtime: %s  runtime: state=%s transitions=%d tools=%d permissions=%d render=%dms", a.stateLabel(), a.stateLabel(), a.cmdState.StatuslineTransitions, a.cmdState.StatuslineToolCalls, a.cmdState.StatuslinePermissions, a.cmdState.StatuslineLastRenderMS)
		panes = append(panes, truncateDisplayWidth(runtime, width, "..."))
	}

	context := a.activeContextHint()
	if strings.TrimSpace(context) != "" {
		panes = append(panes, truncateDisplayWidth("context: "+context, width, "..."))
	}
	if strings.TrimSpace(a.searchTimelineQuery) != "" || strings.TrimSpace(a.searchQuickOpenQuery) != "" || strings.TrimSpace(a.searchHistoryQuery) != "" {
		searches := fmt.Sprintf("searches: timeline=%q quick-open=%q history=%q", strings.TrimSpace(a.searchTimelineQuery), strings.TrimSpace(a.searchQuickOpenQuery), strings.TrimSpace(a.searchHistoryQuery))
		panes = append(panes, truncateDisplayWidth(searches, width, "..."))
	}
	if len(panes) == 0 {
		return ""
	}
	return statusBarStyle.Render(" " + strings.Join(panes, "\n ") + " ")
}

func (a *App) shouldShowStatusHints() bool {
	return true
}

func (a *App) shouldShowStatusHintsPanel() bool {
	if a.stateValue() == stateThinking || a.stateValue() == stateStreaming || a.stateValue() == statePermissionPrompt {
		return true
	}
	if a.activeModalSurface() != modalSurfaceNone {
		return true
	}
	if strings.TrimSpace(a.input.Value()) != "" {
		return true
	}
	if strings.Contains(strings.ToLower(strings.TrimSpace(a.activeContextHint())), "needs login") {
		return true
	}
	if strings.TrimSpace(a.buddy.bubble) != "" && a.now().Before(a.buddy.bubbleUntil) {
		return true
	}
	return false
}

func (a *App) shouldShowStatusRuntimePane() bool {
	if a.stateValue() == stateThinking || a.stateValue() == stateStreaming || a.stateValue() == statePermissionPrompt {
		return true
	}
	if a.activeModalSurface() != modalSurfaceNone {
		return true
	}
	if a.cmdState.StatuslineToolCalls > 0 || a.cmdState.StatuslinePermissions > 0 {
		return true
	}
	if strings.TrimSpace(a.searchTimelineQuery) != "" || strings.TrimSpace(a.searchQuickOpenQuery) != "" || strings.TrimSpace(a.searchHistoryQuery) != "" {
		return true
	}
	return false
}

func (a *App) shouldShowStatusSecondaryLine() bool {
	if !a.shouldShowStatusHintsPanel() && !a.shouldShowStatusRuntimePane() {
		return false
	}
	if strings.TrimSpace(a.activeContextHint()) != "" {
		return true
	}
	if a.stateValue() == stateSearch || a.stateValue() == statePermissionPrompt {
		return true
	}
	if a.commandPanel.active || a.refAuto.active || a.slashAutocomplete.isVisible() || a.modelPickerActive() {
		return true
	}
	if a.turns > 0 && (a.cmdState.StatuslineToolCalls > 0 || a.cmdState.StatuslinePermissions > 0) {
		return true
	}
	return false
}

func (a *App) showContextualStatusDetails() bool {
	return a.shouldShowStatusHintsPanel() || a.shouldShowStatusRuntimePane()
}

func (a *App) stateLabel() string {
	switch a.stateValue() {
	case stateThinking:
		return "thinking"
	case stateStreaming:
		return "streaming"
	case statePermissionPrompt:
		return "permission"
	case stateSearch:
		return "search"
	default:
		return "idle"
	}
}

func (a *App) runtimeContextVisible() bool {
	return a.inputModeValue() != inputModeChat
}

func (a *App) statusProviderLabel() string {
	providerName := strings.TrimSpace(commands.RuntimeSelectionTruth(&a.cmdState).ProviderName)
	if providerName == "" {
		return "-"
	}
	return providerName
}

func (a *App) statusModelLabel() string {
	modelName := strings.TrimSpace(commands.RuntimeSelectionTruth(&a.cmdState).ModelName)
	if modelName == "" {
		return "unknown"
	}
	return modelName
}

func (a *App) renderBuddySpriteBlock() string {
	return buddySpriteStyle.Render(renderBuddySprite(a.buddy, a.width, a.now()))
}

func (a *App) renderBuddyBubbleLine() string {
	bubble, fading := renderBuddyBubble(a.buddy, a.width, a.now())
	if strings.TrimSpace(bubble) == "" {
		return buddyBubbleStyle.Render(" ")
	}
	if fading {
		return buddyBubbleFadeStyle.Render(bubble)
	}
	return buddyBubbleStyle.Render(bubble)
}

func (a *App) shouldRenderBuddyFull() bool {
	if strings.TrimSpace(a.buddy.bubble) != "" {
		return true
	}
	if a.stateValue() == stateThinking || a.stateValue() == stateStreaming || a.stateValue() == statePermissionPrompt {
		return true
	}
	return a.activeModalSurface() != modalSurfaceNone
}

func (a *App) renderBuddyCompactLine() string {
	face := buddyCompactStyle.Render(renderBuddySprite(a.buddy, 1, a.now()))
	bubble, fading := renderBuddyBubble(a.buddy, max(24, a.width/3), a.now())
	bubble = strings.TrimSpace(bubble)
	if bubble == "" {
		return face
	}
	if fading {
		bubble = buddyBubbleFadeStyle.Render(bubble)
	} else {
		bubble = buddyBubbleStyle.Render(bubble)
	}
	return face + "  " + bubble
}

func isScrollActivityMsg(msg tea.Msg) bool {
	switch v := msg.(type) {
	case tea.MouseMsg:
		return v.String() == "wheel up" || v.String() == "wheel down"
	case tea.KeyMsg:
		s := strings.ToLower(strings.TrimSpace(v.String()))
		scrollKeys := []string{"up", "down", "pgup", "pgdown", "home", "end", "ctrl+u", "ctrl+d", "k", "j", "g", "shift+g"}
		for _, key := range scrollKeys {
			if s == key {
				return true
			}
		}
	}
	return false
}

func (a *App) updateUsage() {
	usage := a.cfg.Agent.Usage()
	a.totalTokens = usage.InputTokens + usage.OutputTokens
	a.costUSD = a.cfg.Agent.TotalCost()
}

func (a *App) startTimelineSearch() {
	if a.stateValue() == statePermissionPrompt {
		return
	}
	a.captureAnchorLock()
	a.setSearchModeValue(searchModeTimeline)
	a.setSearchQueryValue("")
	a.setState(stateSearch)
	a.closeInteractiveCommandPanel()
	a.clearModelPicker()
	a.slashAutocomplete.clear()
	a.setSearchMatchPosValue(0)
	if a.ready {
		a.recalcLayout()
		return
	}
	a.refreshViewport()
}

func (a *App) startQuickOpen() {
	if a.stateValue() == statePermissionPrompt {
		return
	}
	a.captureAnchorLock()
	a.setSearchModeValue(searchModeQuickOpen)
	a.closeInteractiveCommandPanel()
	a.clearModelPicker()
	a.slashAutocomplete.clear()
	a.setSearchQueryValue("")
	a.quickOpen.setQuery("")
	a.setState(stateSearch)
	a.setSearchMatchPosValue(0)
	if a.ready {
		a.recalcLayout()
		return
	}
	a.refreshViewport()
}

func (a *App) startHistorySearch() {
	if a.stateValue() == statePermissionPrompt {
		return
	}
	a.captureAnchorLock()
	a.setSearchModeValue(searchModeHistory)
	a.closeInteractiveCommandPanel()
	a.clearModelPicker()
	a.slashAutocomplete.clear()
	a.history = newHistorySearchState(historyEntriesFromInput(a.input.history))
	a.setSearchQueryValue("")
	a.history.setQuery("")
	a.setState(stateSearch)
	a.setSearchMatchPosValue(0)
	if a.ready {
		a.recalcLayout()
		return
	}
	a.refreshViewport()
}

func (a *App) shouldExitSearchOnEnter() bool {
	switch a.searchModeValue() {
	case searchModeQuickOpen, searchModeHistory:
		return a.activeSearchMatchCount() > 0
	default:
		return true
	}
}

func (a *App) exitSearch() {
	a.captureAnchorLock()
	a.setState(stateIdle)
	a.setSearchModeValue(searchModeTimeline)
	a.setSearchQueryValue("")
	a.setSearchMatchPosValue(0)
	a.syncSearchHelpers()
	if a.ready {
		a.recalcLayout()
		return
	}
	a.refreshViewport()
}

func (a *App) applySearchSelection() (bool, tea.Cmd) {
	switch a.searchModeValue() {
	case searchModeQuickOpen:
		item, ok := a.quickOpen.selectedItem()
		if !ok {
			return false, nil
		}
		switch item.value {
		case "search.timeline":
			a.captureAnchorLock()
			a.setSearchModeValue(searchModeTimeline)
			a.setSearchQueryValue("")
			a.setSearchMatchPosValue(0)
			a.syncSearchHelpers()
			a.refreshViewport()
			return true, nil
		case "search.history":
			a.startHistorySearch()
			return true, nil
		case "search.quick_open":
			a.captureAnchorLock()
			a.setSearchModeValue(searchModeQuickOpen)
			a.setSearchQueryValue("")
			a.setSearchMatchPosValue(0)
			a.syncSearchHelpers()
			a.refreshViewport()
			return true, nil
		case "command.model":
			a.exitSearch()
			a.stageSearchInput("/model ")
			return true, nil
		case "command.permissions_auto":
			a.exitSearch()
			return true, func() tea.Msg { return submitMsg{text: "/permissions auto"} }
		case "command.permissions_default":
			a.exitSearch()
			return true, func() tea.Msg { return submitMsg{text: "/permissions default"} }
		case "workflow.permission":
			a.exitSearch()
			a.addTimeline(timelineEntry{kind: timelineAssistant, text: "permission queue: approvals are shown live while tools run"})
			return true, nil
		case "workflow.references":
			a.exitSearch()
			a.stageSearchInput("@")
			return true, nil
		case "workflow.slash":
			a.exitSearch()
			a.stageSearchInput("/")
			return true, nil
		case "history.last":
			entry, ok := a.historyLatestEntry()
			if !ok {
				return false, nil
			}
			a.exitSearch()
			a.input.SetValue(entry.text)
			return true, nil
		default:
			a.exitSearch()
			return true, nil
		}
	case searchModeHistory:
		entry, ok := a.history.selectedEntry()
		if !ok {
			return false, nil
		}
		a.exitSearch()
		a.input.SetValue(entry.text)
		return true, nil
	default:
		return false, nil
	}
}

func (a *App) stageSearchInput(value string) {
	a.input.SetValue(value)
	a.syncSlashAutocomplete()
	a.syncModelPicker()
	a.syncReferenceAutocomplete()
	a.syncInputMode()
	if a.ready {
		a.recalcLayout()
	}
}

func (a *App) historyLatestEntry() (historySearchEntry, bool) {
	entries := historyEntriesFromInput(a.input.history)
	if len(entries) == 0 {
		return historySearchEntry{}, false
	}
	return entries[len(entries)-1], true
}

func (a *App) advanceSearchSelection(dir int) {
	switch a.searchModeValue() {
	case searchModeQuickOpen:
		a.quickOpen.moveSelection(dir)
	case searchModeHistory:
		a.history.moveSelection(dir)
	default:
		a.jumpTimelineMatch(dir)
	}
}

func (a *App) pageSearchSelection(dir int) {
	step := 5
	if step < 1 {
		step = 1
	}
	sign := 1
	if dir < 0 {
		sign = -1
	}
	switch a.searchModeValue() {
	case searchModeQuickOpen:
		if len(a.quickOpen.visible) == 0 {
			a.quickOpen.selected = -1
			return
		}
		next := a.quickOpen.selected + sign*step
		if next < 0 {
			next = 0
		}
		if next >= len(a.quickOpen.visible) {
			next = len(a.quickOpen.visible) - 1
		}
		a.quickOpen.selected = next
		a.quickOpen.rememberSelection()
		a.quickOpen.ensureVisible(5)
	case searchModeHistory:
		if len(a.history.matches) == 0 {
			a.history.selected = -1
			return
		}
		next := a.history.selected + sign*step
		if next < 0 {
			next = 0
		}
		if next >= len(a.history.matches) {
			next = len(a.history.matches) - 1
		}
		a.history.selected = next
		a.history.rememberSelection()
	default:
		if len(a.timelineMatches) == 0 {
			return
		}
		next := a.matchPos + sign*step
		if next < 0 {
			next = 0
		}
		if next >= len(a.timelineMatches) {
			next = len(a.timelineMatches) - 1
		}
		a.matchPos = next
		if a.matchPos >= 0 && a.matchPos < len(a.timelineMatches) {
			a.viewport.SetYOffset(clampYOffset(a.timelineMatches[a.matchPos].lineOffset, a.lastRenderedLines, a.viewport.Height))
		}
	}
}

func (a *App) jumpSearchSelection(toEnd bool) {
	switch a.searchModeValue() {
	case searchModeQuickOpen:
		if len(a.quickOpen.visible) == 0 {
			a.quickOpen.selected = -1
			return
		}
		if toEnd {
			a.quickOpen.selected = len(a.quickOpen.visible) - 1
		} else {
			a.quickOpen.selected = 0
		}
	case searchModeHistory:
		if len(a.history.matches) == 0 {
			a.history.selected = -1
			return
		}
		if toEnd {
			a.history.selected = len(a.history.matches) - 1
		} else {
			a.history.selected = 0
		}
	default:
		if len(a.timelineMatches) == 0 {
			return
		}
		if toEnd {
			a.matchPos = len(a.timelineMatches) - 1
		} else {
			a.setSearchMatchPosValue(0)
		}
		if a.matchPos >= 0 && a.matchPos < len(a.timelineMatches) {
			a.viewport.SetYOffset(clampYOffset(a.timelineMatches[a.matchPos].lineOffset, a.lastRenderedLines, a.viewport.Height))
		}
	}
}

func (a *App) jumpTimelineMatch(dir int) {
	if len(a.timelineMatches) == 0 || strings.TrimSpace(a.searchQuery) == "" {
		return
	}
	a.matchPos = nextMatchPos(a.matchPos, len(a.timelineMatches), dir)
	if a.matchPos >= 0 && a.matchPos < len(a.timelineMatches) {
		a.viewport.SetYOffset(clampYOffset(a.timelineMatches[a.matchPos].lineOffset, a.lastRenderedLines, a.viewport.Height))
	}
}

func (a *App) activeSearchMatchCount() int {
	switch a.searchModeValue() {
	case searchModeQuickOpen:
		return len(a.quickOpen.visible)
	case searchModeHistory:
		return len(a.history.matches)
	default:
		return len(a.timelineMatches)
	}
}

func (a *App) syncSearchHelpers() {
	trimmed := a.searchQueryValue()
	switch a.searchModeValue() {
	case searchModeQuickOpen:
		a.setSearchQuickOpenQueryValue(trimmed)
	case searchModeHistory:
		a.setSearchHistoryQueryValue(trimmed)
	default:
		a.setSearchTimelineQueryValue(trimmed)
	}
	switch a.searchModeValue() {
	case searchModeQuickOpen:
		a.quickOpen.setQuery(a.searchQueryValue())
	case searchModeHistory:
		a.history.setQuery(a.searchQueryValue())
	}
}

func (a *App) renderPermissionHistorySummary() string {
	if len(a.permissionHistory) == 0 {
		return "permission history: no decisions yet"
	}
	latest := a.permissionHistory[len(a.permissionHistory)-1]
	decision := "allow"
	switch latest.Decision {
	case PermissionNo:
		decision = "deny"
	case PermissionAlways:
		decision = "always"
	}
	return fmt.Sprintf("permission history: %d entries, latest %s %s (turn %d)", len(a.permissionHistory), latest.ToolName, decision, latest.Turn)
}

func (a *App) activeSearchSelectionSummary() string {
	switch a.searchModeValue() {
	case searchModeQuickOpen:
		item, ok := a.quickOpen.selectedItem()
		if !ok {
			return "-"
		}
		return item.label
	case searchModeHistory:
		entry, ok := a.history.selectedEntry()
		if !ok {
			return "-"
		}
		return truncateDisplayWidth(strings.TrimSpace(entry.text), 28, "...")
	default:
		if len(a.timelineMatches) == 0 || strings.TrimSpace(a.searchQuery) == "" {
			return "-"
		}
		if a.matchPos < 0 || a.matchPos >= len(a.timelineMatches) {
			return "-"
		}
		match := a.timelineMatches[a.matchPos]
		return fmt.Sprintf("%d/%d row:%d hit:%d/%d", a.matchPos+1, len(a.timelineMatches), match.rowIndex+1, match.occurrence+1, max(1, match.rowMatches))
	}
}

func MakeStreamCallback(p *tea.Program) func(types.StreamEvent) {
	return func(ev types.StreamEvent) {
		p.Send(streamEventMsg{event: ev})
	}
}

func MakeEventCallback(p *tea.Program) func(types.AgentEvent) {
	return func(ev types.AgentEvent) {
		p.Send(agentEventMsg{event: ev})
	}
}

var (
	userLabelStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))

	userTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).PaddingLeft(2)

	assistantLabelStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))

	assistantTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).PaddingLeft(2)

	toolStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).PaddingLeft(2)

	permissionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).PaddingLeft(2)

	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)

	headerBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("254")).Background(lipgloss.Color("238")).Bold(true)

	statusBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("254")).Background(lipgloss.Color("237")).Bold(true)

	statusSubtleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Background(lipgloss.Color("236"))

	streamingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("189")).Italic(true)

	searchStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("153"))

	slashAutocompleteStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("189"))

	referenceAutocompleteStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("186"))

	searchHighlightStyle = lipgloss.NewStyle().Background(lipgloss.Color("58")).Foreground(lipgloss.Color("230")).Bold(true)

	buddySpriteStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("151"))

	buddyBubbleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("223")).Italic(true)

	buddyBubbleFadeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true)

	buddyCompactStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("151"))
)
