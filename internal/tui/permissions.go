package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PermissionDecision represents the user's response to a permission prompt.
type PermissionDecision int

const (
	// PermissionUndecided means the user has not yet responded.
	PermissionUndecided PermissionDecision = iota
	// PermissionYes allows the tool call once.
	PermissionYes
	// PermissionNo denies the tool call.
	PermissionNo
	// PermissionAlways allows this tool without future prompts.
	PermissionAlways
)

// permissionDecisionMsg is sent when the user makes a permission decision.
type permissionDecisionMsg struct {
	decision PermissionDecision
}

// PermissionModel displays a permission prompt overlay for tool calls.
type PermissionModel struct {
	toolName    string
	description string
	toolKind    string
	toolDetails []string
	decision    PermissionDecision
	status      permissionStatus
	selected    PermissionDecision
	queueIndex  int
	queueTotal  int
	queueNext   []string
	queueStack  []string
	recent      []string
	width       int
}

// NewPermission creates a permission prompt for the given tool and description.
func NewPermission(toolName, description string) PermissionModel {
	return PermissionModel{
		toolName:    toolName,
		description: description,
		toolKind:    normalizePermissionToolType(toolName),
		decision:    PermissionUndecided,
		status:      permissionPending,
		selected:    PermissionYes,
		width:       60,
	}
}

func (m *PermissionModel) SetToolDetails(kind string, rows []string) {
	if strings.TrimSpace(kind) == "" {
		kind = normalizePermissionToolType(m.toolName)
	}
	m.toolKind = kind
	if len(rows) == 0 {
		m.toolDetails = nil
		return
	}
	m.toolDetails = append([]string(nil), rows...)
}

func (m *PermissionModel) SetQueueIndex(index, total int) {
	if total < 0 {
		total = 0
	}
	if index < 0 {
		index = 0
	}
	if total > 0 && index > total {
		index = total
	}
	m.queueIndex = index
	m.queueTotal = total
}

func (m *PermissionModel) SetQueuePreview(next []string) {
	if len(next) == 0 {
		m.queueNext = nil
		return
	}
	m.queueNext = append([]string(nil), next...)
}

func (m *PermissionModel) SetQueueStack(rows []string) {
	if len(rows) == 0 {
		m.queueStack = nil
		return
	}
	m.queueStack = append([]string(nil), rows...)
}

func (m *PermissionModel) SetRecentDecisions(rows []string) {
	if len(rows) == 0 {
		m.recent = nil
		return
	}
	m.recent = append([]string(nil), rows...)
}

type permissionPromptContext struct {
	kind    string
	summary string
	details []string
}

func permissionPromptDescription(toolName string, toolInput json.RawMessage) string {
	return permissionPromptContextFromInput(toolName, toolInput).summary
}

func permissionPromptContextFromInput(toolName string, toolInput json.RawMessage) permissionPromptContext {
	toolType := normalizePermissionToolType(toolName)
	parsed := decodePermissionInput(toolInput)

	switch toolType {
	case "bash":
		command := firstNonEmptyString(parsed, "command")
		workdir := firstNonEmptyString(parsed, "workdir", "cwd")
		timeout := firstNonEmptyValueString(parsed, "timeout", "timeoutMs")
		detailRows := make([]string, 0, 4)
		if command != "" {
			detailRows = append(detailRows, "command: "+truncateDisplayWidth(command, 72, "..."))
		}
		if workdir != "" {
			detailRows = append(detailRows, "workdir: "+truncateDisplayWidth(workdir, 72, "..."))
		}
		if timeout != "" {
			detailRows = append(detailRows, "timeout: "+truncateDisplayWidth(timeout, 24, "..."))
		}
		description := firstNonEmptyString(parsed, "description")
		if description != "" {
			detailRows = append(detailRows, "description: "+truncateDisplayWidth(description, 72, "..."))
		}
		if command == "" {
			if workdir != "" {
				return permissionPromptContext{kind: toolType, summary: "execute a shell command in: " + truncateDisplayWidth(workdir, 72, "..."), details: detailRows}
			}
			return permissionPromptContext{kind: toolType, summary: "execute a shell command", details: detailRows}
		}
		desc := "execute shell command: " + truncateDisplayWidth(command, 72, "...")
		if workdir != "" {
			desc += " (in " + truncateDisplayWidth(workdir, 32, "...") + ")"
		}
		if timeout != "" {
			desc += " timeout " + truncateDisplayWidth(timeout, 12, "...")
		}
		return permissionPromptContext{kind: toolType, summary: desc, details: detailRows}
	case "file":
		path := firstNonEmptyString(parsed, "file_path", "path")
		lineStart := firstNonEmptyString(parsed, "offset", "start_line")
		lineCount := firstNonEmptyValueString(parsed, "limit", "line_count")
		pattern := firstNonEmptyString(parsed, "pattern")
		include := firstNonEmptyString(parsed, "include")
		detailRows := make([]string, 0, 4)
		if path != "" {
			detailRows = append(detailRows, "path: "+truncateDisplayWidth(path, 72, "..."))
		}
		if pattern != "" {
			detailRows = append(detailRows, "pattern: "+truncateDisplayWidth(pattern, 72, "..."))
		}
		if include != "" {
			detailRows = append(detailRows, "include: "+truncateDisplayWidth(include, 48, "..."))
		}
		if lineStart != "" || lineCount != "" {
			rangeRow := "range: "
			if lineStart != "" {
				rangeRow += "from line " + lineStart
			} else {
				rangeRow += "from start"
			}
			if lineCount != "" {
				rangeRow += " (" + lineCount + " lines)"
			}
			detailRows = append(detailRows, rangeRow)
		}
		if path != "" {
			desc := "access local files at: " + truncateDisplayWidth(path, 72, "...")
			if lineStart != "" || lineCount != "" {
				desc += ""
				if lineStart != "" {
					desc += " from line " + lineStart
				}
				if lineCount != "" {
					desc += " (" + lineCount + " lines)"
				}
			}
			return permissionPromptContext{kind: toolType, summary: desc, details: detailRows}
		}
		if pattern != "" {
			return permissionPromptContext{kind: toolType, summary: fmt.Sprintf("search local files matching: %s", truncateDisplayWidth(pattern, 72, "...")), details: detailRows}
		}
		return permissionPromptContext{kind: toolType, summary: "access local files", details: detailRows}
	case "webfetch":
		url := firstNonEmptyString(parsed, "url")
		format := firstNonEmptyString(parsed, "format")
		timeout := firstNonEmptyValueString(parsed, "timeout")
		detailRows := make([]string, 0, 3)
		if url != "" {
			detailRows = append(detailRows, "url: "+truncateDisplayWidth(url, 72, "..."))
		}
		if format != "" {
			detailRows = append(detailRows, "format: "+truncateDisplayWidth(format, 24, "..."))
		}
		if timeout != "" {
			detailRows = append(detailRows, "timeout: "+truncateDisplayWidth(timeout, 24, "..."))
		}
		if url == "" {
			return permissionPromptContext{kind: toolType, summary: "fetch content from the web", details: detailRows}
		}
		return permissionPromptContext{kind: toolType, summary: "fetch content from: " + truncateDisplayWidth(url, 72, "..."), details: detailRows}
	default:
		if strings.TrimSpace(toolName) == "" {
			return permissionPromptContext{kind: toolType, summary: "run this tool call"}
		}
		return permissionPromptContext{kind: toolType, summary: "run " + toolName}
	}
}

func normalizePermissionToolType(toolName string) string {
	norm := strings.ToLower(strings.TrimSpace(toolName))
	switch norm {
	case "bash", "powershell":
		return "bash"
	case "read", "write", "edit", "glob", "grep", "notebookedit", "ls", "mcp_resource_read", "mcp_resource_list":
		return "file"
	case "webfetch", "websearch":
		return "webfetch"
	default:
		return "generic"
	}
}

func decodePermissionInput(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return nil
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil
	}
	return obj
}

func firstNonEmptyString(obj map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := obj[key]
		if !ok {
			continue
		}
		s, ok := value.(string)
		if !ok {
			continue
		}
		s = strings.TrimSpace(s)
		if s != "" {
			return s
		}
	}
	return ""
}

func firstNonEmptyValueString(obj map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := obj[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case string:
			typed = strings.TrimSpace(typed)
			if typed != "" {
				return typed
			}
		case json.Number:
			return typed.String()
		case float64:
			return fmt.Sprintf("%.0f", typed)
		case int:
			return itoa(typed)
		case int64:
			return fmt.Sprintf("%d", typed)
		}
	}
	return ""
}

// SetWidth sets the display width for the permission prompt.
func (m *PermissionModel) SetWidth(w int) {
	m.width = w
}

// Decision returns the user's current decision.
func (m PermissionModel) Decision() PermissionDecision {
	return m.decision
}

func (m *PermissionModel) SetStatus(status permissionStatus) {
	m.status = status
}

func (m *PermissionModel) selectedIndex() int {
	switch m.selected {
	case PermissionYes:
		return 0
	case PermissionNo:
		return 1
	case PermissionAlways:
		return 2
	default:
		return 0
	}
}

func (m *PermissionModel) setSelectedByIndex(index int) {
	if index < 0 {
		index = 2
	}
	if index > 2 {
		index = 0
	}
	switch index {
	case 0:
		m.selected = PermissionYes
	case 1:
		m.selected = PermissionNo
	default:
		m.selected = PermissionAlways
	}
}

func (m *PermissionModel) cycleSelection(dir int) {
	index := m.selectedIndex()
	if dir < 0 {
		index--
	} else {
		index++
	}
	m.setSelectedByIndex(index)
}

func (m PermissionModel) toolHintLines() []string {
	kind := m.toolKind
	if strings.TrimSpace(kind) == "" {
		kind = normalizePermissionToolType(m.toolName)
	}
	switch kind {
	case "bash":
		return []string{"risk: medium-high", "hint: review command and arguments before allowing", "hint: choose always only for trusted local workflows"}
	case "file":
		return []string{"risk: medium", "hint: verify file paths and globs are expected", "hint: deny if edits read/write outside intended scope"}
	case "webfetch":
		return []string{"risk: medium", "hint: verify destination host before allowing", "hint: deny unexpected network access"}
	default:
		return []string{"risk: unknown", "hint: review tool intent and input carefully"}
	}
}

func (m PermissionModel) renderAction(label, key string, decision PermissionDecision, keyStyle, optionStyle, selectedStyle lipgloss.Style) string {
	prefix := " "
	if m.selected == decision {
		prefix = ">"
	}
	text := prefix + " " + keyStyle.Render("["+key+"]") + " " + label
	if m.selected == decision {
		return selectedStyle.Render(text)
	}
	return optionStyle.Render(text)
}

// Init implements tea.Model.
func (m PermissionModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m PermissionModel) Update(msg tea.Msg) (PermissionModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		s := normalizeCombo(msg.String())
		switch {
		case keyMatches(msg, "esc") || keyMatches(msg, "ctrl+c"):
			m.selected = PermissionNo
			m.decision = PermissionNo
			return m, func() tea.Msg {
				return permissionDecisionMsg{decision: PermissionNo}
			}
		case keyMatches(msg, "left") || keyMatches(msg, "up") || keyMatches(msg, "shift+tab") || keyMatches(msg, "ctrl+p") || s == "k":
			m.cycleSelection(-1)
			return m, nil
		case keyMatches(msg, "right") || keyMatches(msg, "down") || keyMatches(msg, "tab") || keyMatches(msg, "ctrl+n") || s == "j":
			m.cycleSelection(1)
			return m, nil
		case keyMatches(msg, "home"):
			m.setSelectedByIndex(0)
			return m, nil
		case keyMatches(msg, "end"):
			m.setSelectedByIndex(2)
			return m, nil
		case keyMatches(msg, "enter") || s == "space":
			m.decision = m.selected
			return m, func() tea.Msg {
				return permissionDecisionMsg{decision: m.selected}
			}
		case s == "y":
			m.selected = PermissionYes
			m.decision = PermissionYes
			return m, func() tea.Msg {
				return permissionDecisionMsg{decision: PermissionYes}
			}
		case s == "n":
			m.selected = PermissionNo
			m.decision = PermissionNo
			return m, func() tea.Msg {
				return permissionDecisionMsg{decision: PermissionNo}
			}
		case s == "a":
			m.selected = PermissionAlways
			m.decision = PermissionAlways
			return m, func() tea.Msg {
				return permissionDecisionMsg{decision: PermissionAlways}
			}
		}
	}
	return m, nil
}

// View implements tea.Model.
func (m PermissionModel) View() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("208")).
		Padding(1, 2).
		Width(m.width)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("208"))

	toolStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212"))

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	optionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	selectedOptionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("231")).
		Background(lipgloss.Color("63"))

	detailHeadingStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("221"))

	statusStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("76"))

	if m.status == permissionDenied {
		statusStyle = statusStyle.Foreground(lipgloss.Color("196"))
	}
	if m.status == permissionAlwaysStatus {
		statusStyle = statusStyle.Foreground(lipgloss.Color("69"))
	}

	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("76"))

	title := titleStyle.Render("Permission Required")
	tool := "Allow " + toolStyle.Render(m.toolName) + " to:"
	desc := descStyle.Render(m.description)
	detailBlock := ""
	if len(m.toolDetails) > 0 {
		heading := "tool details:"
		switch m.toolKind {
		case "bash":
			heading = "command details:"
		case "file":
			heading = "file request details:"
		case "webfetch":
			heading = "network request details:"
		}
		rows := make([]string, 0, len(m.toolDetails)+1)
		rows = append(rows, detailHeadingStyle.Render(heading))
		for _, row := range m.toolDetails {
			rows = append(rows, optionStyle.Render("  - "+truncateDisplayWidth(row, m.width-10, "...")))
		}
		detailBlock = "\n" + strings.Join(rows, "\n")
	}
	options := strings.Join([]string{
		m.renderAction("allow once", "y", PermissionYes, keyStyle, optionStyle, selectedOptionStyle),
		m.renderAction("deny", "n", PermissionNo, keyStyle, optionStyle, selectedOptionStyle),
		m.renderAction("always allow", "a", PermissionAlways, keyStyle, optionStyle, selectedOptionStyle),
	}, "\n")
	actionsHint := optionStyle.Render("navigate: tab/shift+tab arrows j/k ctrl+n/ctrl+p home/end") + "\n" + optionStyle.Render("confirm: enter/space  direct: y n a  cancel: esc")
	status := optionStyle.Render("status: ") + statusStyle.Render(strings.ToUpper(string(m.status)))
	if m.queueTotal > 0 {
		remaining := m.queueTotal - m.queueIndex
		if remaining < 0 {
			remaining = 0
		}
		status += optionStyle.Render(fmt.Sprintf("  queue: %d/%d (%d waiting)", m.queueIndex, m.queueTotal, remaining))
	}
	queuePreview := ""
	if len(m.queueNext) > 0 {
		lines := make([]string, 0, len(m.queueNext)+1)
		lines = append(lines, optionStyle.Render("next in queue:"))
		for _, row := range m.queueNext {
			lines = append(lines, optionStyle.Render("  - "+row))
		}
		queuePreview = "\n" + strings.Join(lines, "\n")
	}
	stackPreview := ""
	if len(m.queueStack) > 0 {
		lines := make([]string, 0, len(m.queueStack)+1)
		lines = append(lines, optionStyle.Render("stack:"))
		for _, row := range m.queueStack {
			lines = append(lines, optionStyle.Render("  "+row))
		}
		stackPreview = "\n" + strings.Join(lines, "\n")
	}
	recentPreview := ""
	if len(m.recent) > 0 {
		lines := make([]string, 0, len(m.recent)+1)
		lines = append(lines, optionStyle.Render("recent decisions:"))
		for _, row := range m.recent {
			lines = append(lines, optionStyle.Render("  - "+row))
		}
		recentPreview = "\n" + strings.Join(lines, "\n")
	}
	hints := m.toolHintLines()
	hintBlock := ""
	if len(hints) > 0 {
		hintBlock = "\n" + optionStyle.Render(strings.Join(hints, "\n"))
	}

	content := title + "\n\n" + tool + "\n" + desc + detailBlock + hintBlock + "\n\n" + status + queuePreview + stackPreview + recentPreview + "\n" + options + "\n" + actionsHint

	return boxStyle.Render(content)
}
