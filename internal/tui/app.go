package tui

import (
	"context"
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
	toolInputPreview string
	toolState        toolProgressState
	permissionState  permissionStatus
	turn             int
}

type Config struct {
	Agent           *agent.Agent
	Version         string
	Debug           bool
	InitialModel    string
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

	model       string
	totalTokens int
	costUSD     float64
	turns       int

	timeline          []timelineEntry
	toolRows          map[string]int
	permissionRows    map[string]int
	visibleRows       []int
	visibleLines      []int
	followTail        bool
	lastRenderedLines int

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
	resolver             *references.Resolver
	refAuto              referenceAutocompleteState

	commands *commands.Registry
	cmdState commands.RuntimeState
	keySet   *keybindings.Set

	slashAutocomplete slashAutocompleteState

	activePermissionToolUseID string
	activePermissionQueueKey  string
	activePermissionTurn      int
	permissionQueueSeq        int
	permissionQueue           []permissionPromptRequest
	permissionHistory         []permissionDecisionRecord
}

func (a *App) setState(next appState) {
	if a.state != next {
		a.cmdState.StatuslineTransitions++
	}
	a.state = next
}

func (a *App) noteToolRuntimeEvent() {
	a.cmdState.StatuslineToolCalls++
}

func (a *App) notePermissionRuntimeEvent() {
	a.cmdState.StatuslinePermissions++
}

func New(cfg Config) *App {
	model := cfg.InitialModel
	if model == "" {
		model = "unknown"
	}

	return &App{
		cfg:            cfg,
		input:          NewInput(),
		spinner:        NewSpinner("thinking"),
		state:          stateIdle,
		model:          model,
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
		history:    newHistorySearchState(nil),
		permDialog: newPermissionDialogState(),
		buddy:      newBuddyState(),
		now:        time.Now,
		resolver:   references.NewResolver(resolveBaseDir(cfg)),
		commands:   commands.DefaultRegistry(),
		keySet:     defaultKeybindingSet(),
		cmdState: commands.RuntimeState{
			Model:          model,
			PermissionMode: permissions.ModeDefault,
			Agent:          cfg.Agent,
		},
		slashAutocomplete: slashAutocompleteState{selected: -1},
	}
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

		if a.state == stateIdle && a.refAuto.active {
			switch {
			case keyMatches(msg, "ctrl+n") || keyMatches(msg, "down") || keyMatches(msg, "tab"):
				a.advanceReferenceSelection(1)
				return a, nil
			case keyMatches(msg, "ctrl+p") || keyMatches(msg, "up") || keyMatches(msg, "shift+tab"):
				a.advanceReferenceSelection(-1)
				return a, nil
			case keyMatches(msg, "pgdown") || keyMatches(msg, "pagedown"):
				a.pageReferenceSelection(1)
				return a, nil
			case keyMatches(msg, "pgup") || keyMatches(msg, "pageup"):
				a.pageReferenceSelection(-1)
				return a, nil
			case keyMatches(msg, "home"):
				a.jumpReferenceSelection(false)
				return a, nil
			case keyMatches(msg, "end"):
				a.jumpReferenceSelection(true)
				return a, nil
			case keyMatches(msg, "enter"):
				if a.applyReferenceSelection() {
					return a, nil
				}
			case keyMatches(msg, "esc"):
				a.clearReferenceAutocomplete()
				return a, nil
			}
		}

		if a.state == statePermissionPrompt {
			var cmd tea.Cmd
			a.permission, cmd = a.permission.Update(msg)
			cmds = append(cmds, cmd)
			return a, tea.Batch(cmds...)
		}

		if a.state == stateSearch {
			switch {
			case keyMatches(msg, "ctrl+f"):
				a.startTimelineSearch()
				return a, nil
			case keyMatches(msg, "ctrl+o"):
				a.startQuickOpen()
				return a, nil
			case keyMatches(msg, "ctrl+r"):
				a.startHistorySearch()
				return a, nil
			case keyMatches(msg, "ctrl+n") || keyMatches(msg, "down") || keyMatches(msg, "tab"):
				a.advanceSearchSelection(1)
				return a, nil
			case keyMatches(msg, "ctrl+p") || keyMatches(msg, "up") || keyMatches(msg, "shift+tab"):
				a.advanceSearchSelection(-1)
				return a, nil
			case keyMatches(msg, "pgdown") || keyMatches(msg, "pagedown"):
				a.pageSearchSelection(1)
				return a, nil
			case keyMatches(msg, "pgup") || keyMatches(msg, "pageup"):
				a.pageSearchSelection(-1)
				return a, nil
			case keyMatches(msg, "home"):
				a.jumpSearchSelection(false)
				return a, nil
			case keyMatches(msg, "end"):
				a.jumpSearchSelection(true)
				return a, nil
			}
			a.updateSearchInput(msg)
			return a, nil
		}

		if a.state == stateIdle && a.slashAutocomplete.isVisible() {
			switch {
			case keyMatches(msg, "ctrl+n") || keyMatches(msg, "down") || keyMatches(msg, "tab"):
				a.slashAutocomplete.moveSelection(1)
				return a, nil
			case keyMatches(msg, "ctrl+p") || keyMatches(msg, "up") || keyMatches(msg, "shift+tab"):
				a.slashAutocomplete.moveSelection(-1)
				return a, nil
			case keyMatches(msg, "pgdown") || keyMatches(msg, "pagedown"):
				a.slashAutocomplete.pageSelection(1)
				return a, nil
			case keyMatches(msg, "pgup") || keyMatches(msg, "pageup"):
				a.slashAutocomplete.pageSelection(-1)
				return a, nil
			case keyMatches(msg, "home"):
				a.slashAutocomplete.jumpSelection(false)
				return a, nil
			case keyMatches(msg, "end"):
				a.slashAutocomplete.jumpSelection(true)
				return a, nil
			case keyMatches(msg, "esc"):
				a.slashAutocomplete.clear()
				return a, nil
			case keyMatches(msg, "enter"):
				if a.applySlashAutocompleteSelection() {
					return a, nil
				}
			}
		}

		switch {
		case keyMatches(msg, "ctrl+f"):
			a.startTimelineSearch()
			return a, nil
		case keyMatches(msg, "ctrl+o"):
			a.startQuickOpen()
			return a, nil
		case keyMatches(msg, "ctrl+r"):
			a.startHistorySearch()
			return a, nil
		}

	case submitMsg:
		if a.state != stateIdle {
			return a, nil
		}
		userText := msg.text
		a.captureRecentReferences(userText)
		if strings.HasPrefix(strings.TrimSpace(userText), "/") {
			res, err := a.commands.Dispatch(context.Background(), commands.Context{State: &a.cmdState}, userText)
			if err != nil {
				a.addTimeline(timelineEntry{kind: timelineError, text: err.Error()})
				return a, nil
			}
			if res.Handled && strings.TrimSpace(res.Message) != "" {
				a.addTimeline(timelineEntry{kind: timelineAssistant, text: res.Message})
			}
			a.model = a.cmdState.Model
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
		if a.state != statePermissionPrompt || a.permDialog.stage != permissionDialogPrompt {
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

	switch a.state {
	case stateIdle:
		var cmd tea.Cmd
		wasSlashVisible := a.slashAutocomplete.isVisible()
		wasRefVisible := a.refAuto.active
		a.input, cmd = a.input.Update(msg)
		cmds = append(cmds, cmd)
		a.syncSlashAutocomplete()
		a.syncReferenceAutocomplete()
		if wasSlashVisible != a.slashAutocomplete.isVisible() || wasRefVisible != a.refAuto.active {
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

	var sections []string
	sections = append(sections, a.viewport.View())

	switch a.state {
	case stateThinking:
		sections = append(sections, a.spinner.View())
	case stateStreaming:
		sections = append(sections, a.renderStreamingLine())
	case statePermissionPrompt:
		sections = append(sections, a.permission.View())
	case stateSearch:
		sections = append(sections, a.renderSearchLine())
	}

	sections = append(sections, a.renderBuddySpriteBlock())
	sections = append(sections, a.renderBuddyBubbleLine())
	sections = append(sections, a.renderStatusBar())
	if hints := a.renderStatusHints(); strings.TrimSpace(hints) != "" {
		sections = append(sections, hints)
	}
	if panes := a.renderStatusRuntimePanes(); strings.TrimSpace(panes) != "" {
		sections = append(sections, panes)
	}

	if a.state == stateIdle {
		if a.slashAutocomplete.isVisible() {
			sections = append(sections, a.renderSlashAutocomplete())
		} else if a.refAuto.active {
			sections = append(sections, a.renderReferenceAutocomplete())
		}
		if hint := a.renderCommandContextHint(); strings.TrimSpace(hint) != "" {
			sections = append(sections, hint)
		}
		if hint := strings.TrimSpace(a.input.Hint()); hint != "" {
			sections = append(sections, slashAutocompleteStyle.Render("input: "+hint))
		}
		sections = append(sections, a.input.View())
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (a *App) syncSlashAutocomplete() {
	if a.state != stateIdle {
		a.slashAutocomplete.clear()
		return
	}
	value := a.input.Value()
	if !strings.HasPrefix(value, "/") {
		a.slashAutocomplete.clear()
		return
	}
	tail := value[1:]
	if strings.Contains(tail, " ") {
		a.slashAutocomplete.clear()
		return
	}
	query := strings.ToLower(strings.TrimSpace(tail))
	a.slashAutocomplete.setItems(query, a.commands.Suggestions(query))
}

func (a *App) applySlashAutocompleteSelection() bool {
	item, ok := a.slashAutocomplete.selectedItem()
	if !ok {
		return false
	}
	a.input.SetValue("/" + item.Name + " ")
	a.slashAutocomplete.clear()
	return true
}

func (a *App) renderSlashAutocomplete() string {
	lines := []string{"commands:"}
	rowWidth := a.width - 2
	if rowWidth < 1 {
		rowWidth = 1
	}
	window := 6
	start := a.slashAutocomplete.offset
	if start < 0 {
		start = 0
	}
	if start >= len(a.slashAutocomplete.items) {
		start = 0
	}
	end := start + window
	if end > len(a.slashAutocomplete.items) {
		end = len(a.slashAutocomplete.items)
	}
	lastSection := ""
	sectionCounts := make(map[string]int, 6)
	for _, item := range a.slashAutocomplete.items {
		sectionCounts[slashSuggestionSection(strings.TrimSpace(item.MatchReason))]++
	}
	for i := start; i < end; i++ {
		item := a.slashAutocomplete.items[i]
		if i >= len(a.slashAutocomplete.items) {
			break
		}
		prefix := "  "
		if i == a.slashAutocomplete.selected {
			prefix = "> "
		}
		reason := strings.TrimSpace(item.MatchReason)
		if reason == "" {
			reason = "browse"
		}
		section := slashSuggestionSection(reason)
		if section != lastSection {
			lines = append(lines, "  "+section+" ("+itoa(sectionCounts[section])+"):")
			lastSection = section
		}
		row := fmt.Sprintf("%s/%s - %s [%s]", prefix, item.Name, strings.TrimSpace(item.Description), reason)
		lines = append(lines, truncateDisplayWidth(row, rowWidth, "..."))
	}
	if selected, ok := a.slashAutocomplete.selectedItem(); ok {
		lines = append(lines, "")
		lines = append(lines, truncateDisplayWidth("preview: /"+selected.Name, rowWidth, "..."))
		usage := strings.TrimSpace(selected.Usage)
		if usage != "" {
			lines = append(lines, truncateDisplayWidth("usage: "+usage, rowWidth, "..."))
		}
		if reason := strings.TrimSpace(selected.MatchReason); reason != "" {
			lines = append(lines, truncateDisplayWidth("match: "+reason, rowWidth, "..."))
		}
		if len(selected.Aliases) > 0 {
			aliases := make([]string, 0, len(selected.Aliases))
			for _, alias := range selected.Aliases {
				alias = strings.TrimSpace(alias)
				if alias == "" {
					continue
				}
				aliases = append(aliases, "/"+alias)
			}
			if len(aliases) > 0 {
				lines = append(lines, truncateDisplayWidth("aliases: "+strings.Join(aliases, ", "), rowWidth, "..."))
			}
		}
	}
	if len(a.slashAutocomplete.items) > window {
		remaining := len(a.slashAutocomplete.items) - end
		if remaining > 0 {
			lines = append(lines, fmt.Sprintf("  ... +%d more", remaining))
		}
	}
	lines = append(lines, "")
	lines = append(lines, truncateDisplayWidth("tab/enter apply  esc dismiss  up/down navigate  pgup/pgdown page  home/end jump", rowWidth, "..."))
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
	return slashAutocompleteStyle.Render(truncateDisplayWidth(line, width, "..."))
}

func (a *App) handleStreamEvent(ev types.StreamEvent) (tea.Model, tea.Cmd) {
	switch ev.Type {
	case types.StreamStart:
		a.setState(stateStreaming)
		a.streamBuf.Reset()

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
		idx := a.addTimeline(timelineEntry{
			kind:      timelineTool,
			toolName:  ev.ToolName,
			toolUseID: ev.ToolUseID,
			toolState: toolProgressRunning,
			turn:      a.turns,
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
			queueKey:    queueKey,
			toolUseID:   ev.ToolUseID,
			toolName:    ev.ToolName,
			turn:        ev.Turn,
			description: permissionPromptDescription(ev.ToolName, ev.ToolInput),
		}
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
		decision := a.permDialog.decision
		resolvedActive := false
		if strings.TrimSpace(ev.ToolUseID) != "" {
			resolvedActive = ev.ToolUseID == a.activePermissionToolUseID
		} else if strings.TrimSpace(a.activePermissionQueueKey) != "" {
			resolvedActive = true
		} else if a.state == statePermissionPrompt && a.permDialog.stage == permissionDialogPrompt {
			resolvedActive = true
		}
		state := permissionDenied
		toolState := toolProgressDenied
		if ev.PermissionDecision == types.AgentPermissionAllow {
			state = permissionApproved
			if resolvedActive && decision == PermissionAlways {
				state = permissionAlwaysStatus
			}
			toolState = toolProgressRunning
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
		if a.state != statePermissionPrompt {
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
		row.toolInputPreview = buildToolPreview(row.toolInputPreview, delta, maxToolPreviewLen)
	}
	a.timeline[idx] = row
	a.refreshViewport()
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
	a.refreshViewport()
}

func (a *App) refreshViewport() {
	width := a.viewport.Width
	if width <= 0 {
		width = 1
	}
	content, visible, lineOffsets, totalLines := renderTimeline(a.timeline, a.searchQuery, width)
	a.visibleRows = visible
	a.visibleLines = lineOffsets
	prevYOffset := a.viewport.YOffset
	prevAtBottom := a.viewport.AtBottom()

	if a.streamBuf.Len() > 0 {
		if content != "" {
			content += "\n\n"
			totalLines += 2
		}
		streamBlock := assistantLabelStyle.Render("AI") + "\n" + assistantTextStyle.Render(a.streamBuf.String())
		content += streamBlock
		totalLines += visualLineCount(streamBlock, width)
	}

	a.viewport.SetContent(content)

	if len(a.visibleRows) > 0 && a.searchQuery != "" && a.matchPos >= 0 && a.matchPos < len(a.visibleLines) {
		a.viewport.SetYOffset(a.visibleLines[a.matchPos])
		return
	}

	offset, toBottom := resolveScrollAnchor(prevYOffset, prevAtBottom, a.followTail, a.lastRenderedLines, totalLines, a.viewport.Height)
	a.lastRenderedLines = totalLines
	if toBottom {
		a.viewport.GotoBottom()
		return
	}
	a.viewport.SetYOffset(offset)
}

func (a *App) recalcLayout() {
	statusBarHeight := 1
	statusRuntimeHeight := 2
	inputHeight := 3
	activityHeight := 1
	if a.state == stateSearch {
		activityHeight = 2
	}
	buddyHeight := buddyPanelHeight(a.width)

	reserved := statusBarHeight + statusRuntimeHeight + inputHeight + activityHeight + buddyHeight + 2
	vpHeight := a.height - reserved
	if vpHeight < 1 {
		vpHeight = 1
	}

	if !a.ready {
		a.viewport = viewport.New(a.width, vpHeight)
		a.viewport.YPosition = 0
	} else {
		a.viewport.Width = a.width
		a.viewport.Height = vpHeight
	}
	a.input.SetWidth(a.width)
	a.permission.SetWidth(a.width - 4)
	a.refreshViewport()
}

func (a *App) renderStatusBar() string {
	start := time.Now()
	defer func() {
		a.cmdState.StatuslineLastRenderMS = int(time.Since(start).Milliseconds())
	}()

	leftParts := []string{fmt.Sprintf("model:%s", a.model), fmt.Sprintf("turns:%d", a.turns), fmt.Sprintf("mode:%s", permissionModeLabel(a.cmdState.PermissionMode))}
	leftParts = append(leftParts,
		fmt.Sprintf("render:%dms", a.cmdState.StatuslineLastRenderMS),
		fmt.Sprintf("tools:%d", a.cmdState.StatuslineToolCalls),
		fmt.Sprintf("perm:%d", a.cmdState.StatuslinePermissions),
	)
	if a.state == stateSearch {
		leftParts = append(leftParts, fmt.Sprintf("search:%s", a.searchMode.label()))
		leftParts = append(leftParts, fmt.Sprintf("matches:%d", a.activeSearchMatchCount()))
		if sel := a.activeSearchSelectionSummary(); strings.TrimSpace(sel) != "" {
			leftParts = append(leftParts, "sel:"+sel)
		}
	}
	if a.slashAutocomplete.isVisible() {
		leftParts = append(leftParts, fmt.Sprintf("slash:%d", len(a.slashAutocomplete.items)))
	}
	if a.refAuto.active {
		leftParts = append(leftParts, fmt.Sprintf("refs:%d", len(a.refAuto.suggestions)))
	}
	if a.state == statePermissionPrompt {
		leftParts = append(leftParts, fmt.Sprintf("queue:%d", len(a.permissionQueue)))
	}
	if len(a.permissionHistory) > 0 {
		leftParts = append(leftParts, fmt.Sprintf("perm_hist:%d", len(a.permissionHistory)))
	}
	if a.state != stateIdle && a.viewport.Height > 0 {
		leftParts = append(leftParts, fmt.Sprintf("vp:%d/%d", a.viewport.YOffset, a.viewport.Height))
	}
	left := statusBarStyle.Render(" " + strings.Join(leftParts, " ") + "")
	sessionID := a.cfg.ActiveSessionID
	if strings.TrimSpace(sessionID) == "" {
		sessionID = "n/a"
	}
	right := statusBarStyle.Render(fmt.Sprintf(" budget:$%.4f tok:%d sid:%s ", a.costUSD, a.totalTokens, sessionID))

	gap := a.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}
	return left + strings.Repeat(" ", gap) + right
}

func (a *App) renderStreamingLine() string {
	return streamingStyle.Render("streaming...")
}

func (a *App) renderStatusHints() string {
	hints := make([]string, 0, 3)
	if a.state == stateSearch {
		hints = append(hints, "search active")
	}
	if a.slashAutocomplete.isVisible() {
		hints = append(hints, "command palette")
	} else if a.refAuto.active {
		hints = append(hints, "reference palette")
	}
	if a.state == statePermissionPrompt {
		hints = append(hints, "permission review")
	}
	if len(hints) == 0 {
		return ""
	}
	line := "hints: " + strings.Join(hints, " | ")
	return statusBarStyle.Render(" " + truncateDisplayWidth(line, a.width-1, "...") + "")
}

func (a *App) renderStatusRuntimePanes() string {
	width := a.width - 1
	if width < 20 {
		width = 20
	}
	panes := make([]string, 0, 3)
	runtime := fmt.Sprintf("runtime: state=%s transitions=%d tools=%d permissions=%d", a.stateLabel(), a.cmdState.StatuslineTransitions, a.cmdState.StatuslineToolCalls, a.cmdState.StatuslinePermissions)
	panes = append(panes, truncateDisplayWidth(runtime, width, "..."))

	context := a.activeContextHint()
	if strings.TrimSpace(context) != "" {
		panes = append(panes, truncateDisplayWidth("context: "+context, width, "..."))
	}
	if strings.TrimSpace(a.searchTimelineQuery) != "" || strings.TrimSpace(a.searchQuickOpenQuery) != "" || strings.TrimSpace(a.searchHistoryQuery) != "" {
		searches := fmt.Sprintf("searches: timeline=%q quick-open=%q history=%q", strings.TrimSpace(a.searchTimelineQuery), strings.TrimSpace(a.searchQuickOpenQuery), strings.TrimSpace(a.searchHistoryQuery))
		panes = append(panes, truncateDisplayWidth(searches, width, "..."))
	}
	return statusBarStyle.Render(" " + strings.Join(panes, "\n ") + "")
}

func (a *App) stateLabel() string {
	switch a.state {
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

func (a *App) activeContextHint() string {
	if a.state == statePermissionPrompt {
		return a.renderPermissionHistorySummary()
	}
	if a.slashAutocomplete.isVisible() {
		if item, ok := a.slashAutocomplete.selectedItem(); ok {
			return "command /" + item.Name + " selected"
		}
		return "command palette open"
	}
	if a.refAuto.active {
		if item, ok := a.selectedReferenceSuggestion(); ok {
			return "reference @" + item.Path
		}
		return "reference palette open"
	}
	if a.state == stateSearch {
		return "search mode " + a.searchMode.label() + " selection=" + a.activeSearchSelectionSummary()
	}
	if hint := strings.TrimSpace(a.input.Hint()); hint != "" {
		return "input " + hint
	}
	return ""
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
	if a.state == statePermissionPrompt {
		return
	}
	a.searchMode = searchModeTimeline
	a.searchQuery = ""
	a.setState(stateSearch)
	a.slashAutocomplete.clear()
	a.matchPos = 0
	a.refreshViewport()
}

func (a *App) startQuickOpen() {
	if a.state == statePermissionPrompt {
		return
	}
	a.searchMode = searchModeQuickOpen
	a.slashAutocomplete.clear()
	a.searchQuery = ""
	a.quickOpen.setQuery("")
	a.setState(stateSearch)
	a.matchPos = 0
	a.refreshViewport()
}

func (a *App) startHistorySearch() {
	if a.state == statePermissionPrompt {
		return
	}
	a.searchMode = searchModeHistory
	a.slashAutocomplete.clear()
	a.history = newHistorySearchState(historyEntriesFromInput(a.input.history))
	a.searchQuery = ""
	a.history.setQuery("")
	a.setState(stateSearch)
	a.matchPos = 0
	a.refreshViewport()
}

func (a *App) shouldExitSearchOnEnter() bool {
	switch a.searchMode {
	case searchModeQuickOpen, searchModeHistory:
		return a.activeSearchMatchCount() > 0
	default:
		return true
	}
}

func (a *App) exitSearch() {
	a.setState(stateIdle)
	a.searchMode = searchModeTimeline
	a.searchQuery = ""
	a.matchPos = 0
	a.syncSearchHelpers()
	a.refreshViewport()
}

func (a *App) applySearchSelection() bool {
	switch a.searchMode {
	case searchModeQuickOpen:
		item, ok := a.quickOpen.selectedItem()
		if !ok {
			return false
		}
		switch item.value {
		case "search.timeline":
			a.searchMode = searchModeTimeline
			a.searchQuery = ""
			a.matchPos = 0
			a.syncSearchHelpers()
			a.refreshViewport()
			return true
		case "search.history":
			a.startHistorySearch()
			return true
		case "search.quick_open":
			a.searchMode = searchModeQuickOpen
			a.searchQuery = ""
			a.matchPos = 0
			a.syncSearchHelpers()
			a.refreshViewport()
			return true
		case "command.model":
			a.exitSearch()
			a.input.SetValue("/model ")
			return true
		case "command.permissions_auto":
			a.exitSearch()
			a.input.SetValue("/permissions auto ")
			return true
		case "command.permissions_default":
			a.exitSearch()
			a.input.SetValue("/permissions default ")
			return true
		case "workflow.permission":
			a.exitSearch()
			a.addTimeline(timelineEntry{kind: timelineAssistant, text: "permission queue: approvals are shown live while tools run"})
			return true
		case "workflow.references":
			a.exitSearch()
			a.input.SetValue("@")
			return true
		case "workflow.slash":
			a.exitSearch()
			a.input.SetValue("/")
			return true
		case "history.last":
			entry, ok := a.historyLatestEntry()
			if !ok {
				return false
			}
			a.exitSearch()
			a.input.SetValue(entry.text)
			return true
		default:
			a.exitSearch()
			return true
		}
	case searchModeHistory:
		entry, ok := a.history.selectedEntry()
		if !ok {
			return false
		}
		a.exitSearch()
		a.input.SetValue(entry.text)
		return true
	default:
		return false
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
	switch a.searchMode {
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
	for i := 0; i < step; i++ {
		a.advanceSearchSelection(dir)
	}
}

func (a *App) jumpSearchSelection(toEnd bool) {
	switch a.searchMode {
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
		if len(a.visibleRows) == 0 {
			return
		}
		if toEnd {
			a.matchPos = len(a.visibleRows) - 1
		} else {
			a.matchPos = 0
		}
		if a.matchPos >= 0 && a.matchPos < len(a.visibleLines) {
			a.viewport.SetYOffset(a.visibleLines[a.matchPos])
		}
	}
}

func (a *App) jumpTimelineMatch(dir int) {
	if len(a.visibleRows) == 0 || strings.TrimSpace(a.searchQuery) == "" {
		return
	}
	a.matchPos = nextMatchPos(a.matchPos, len(a.visibleRows), dir)
	if a.matchPos >= 0 && a.matchPos < len(a.visibleLines) {
		a.viewport.SetYOffset(a.visibleLines[a.matchPos])
	}
}

func (a *App) activeSearchMatchCount() int {
	switch a.searchMode {
	case searchModeQuickOpen:
		return len(a.quickOpen.visible)
	case searchModeHistory:
		return len(a.history.matches)
	default:
		return len(a.visibleRows)
	}
}

func (a *App) syncSearchHelpers() {
	trimmed := a.searchQuery
	switch a.searchMode {
	case searchModeQuickOpen:
		a.searchQuickOpenQuery = trimmed
	case searchModeHistory:
		a.searchHistoryQuery = trimmed
	default:
		a.searchTimelineQuery = trimmed
	}
	switch a.searchMode {
	case searchModeQuickOpen:
		a.quickOpen.setQuery(a.searchQuery)
	case searchModeHistory:
		a.history.setQuery(a.searchQuery)
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
	switch a.searchMode {
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
		if len(a.visibleRows) == 0 || strings.TrimSpace(a.searchQuery) == "" {
			return "-"
		}
		if a.matchPos < 0 || a.matchPos >= len(a.visibleRows) {
			return "-"
		}
		return fmt.Sprintf("%d/%d", a.matchPos+1, len(a.visibleRows))
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

	statusBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Background(lipgloss.Color("236"))

	streamingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true)

	searchStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))

	slashAutocompleteStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("111"))

	referenceAutocompleteStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("150"))

	searchHighlightStyle = lipgloss.NewStyle().Background(lipgloss.Color("58")).Foreground(lipgloss.Color("230")).Bold(true)

	buddySpriteStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("151"))

	buddyBubbleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("223")).Italic(true)

	buddyBubbleFadeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true)
)
