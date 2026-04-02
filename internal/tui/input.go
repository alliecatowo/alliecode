package tui

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// submitMsg is emitted when the user presses Enter to submit input.
type submitMsg struct {
	text string
}

// InputModel wraps a textarea with history and submit-on-enter behavior.
type InputModel struct {
	textarea textarea.Model
	history  []string
	histIdx  int // -1 means not browsing history
	draft    string
	focused  bool
	lastHint string
}

// NewInput creates a new input area ready for user text entry.
func NewInput() InputModel {
	ta := textarea.New()
	ta.Placeholder = "Type a message... (Enter to send, Shift+Enter for newline)"
	ta.CharLimit = 0
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.Focus()

	return InputModel{
		textarea: ta,
		histIdx:  -1,
		focused:  true,
	}
}

// Value returns the current text in the input area.
func (m InputModel) Value() string {
	return m.textarea.Value()
}

// Reset clears the input area.
func (m *InputModel) Reset() {
	m.textarea.Reset()
	m.histIdx = -1
	m.draft = ""
}

// SetEnabled toggles whether the input area accepts input.
func (m *InputModel) SetEnabled(enabled bool) {
	if enabled {
		m.textarea.Focus()
		m.focused = true
	} else {
		m.textarea.Blur()
		m.focused = false
	}
}

// Focus gives the input area focus.
func (m *InputModel) Focus() {
	m.textarea.Focus()
	m.focused = true
}

// Blur removes focus from the input area.
func (m *InputModel) Blur() {
	m.textarea.Blur()
	m.focused = false
}

// SetWidth updates the input area's width.
func (m *InputModel) SetWidth(w int) {
	m.textarea.SetWidth(w)
}

// SetValue replaces the input contents.
func (m *InputModel) SetValue(text string) {
	m.textarea.SetValue(text)
	m.lastHint = inferInputHint(text)
}

// CursorEnd moves cursor to end of input.
func (m *InputModel) CursorEnd() {
	m.textarea.CursorEnd()
}

// Init implements tea.Model.
func (m InputModel) Init() tea.Cmd {
	return textarea.Blink
}

// Update implements tea.Model.
func (m InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case keyMatches(msg, "ctrl+j"):
			m.textarea, _ = m.textarea.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\n'}})
			m.lastHint = inferInputHint(m.textarea.Value())
			return m, nil
		case keyMatches(msg, "ctrl+w") || keyMatches(msg, "alt+backspace"):
			m.textarea.SetValue(deleteTrailingWord(m.textarea.Value()))
			m.lastHint = inferInputHint(m.textarea.Value())
			return m, nil
		case keyMatches(msg, "ctrl+u"):
			m.textarea.SetValue("")
			m.lastHint = inferInputHint("")
			return m, nil
		case keyMatches(msg, "ctrl+a"):
			m.textarea.CursorStart()
			return m, nil
		case keyMatches(msg, "ctrl+e"):
			m.textarea.CursorEnd()
			return m, nil
		case keyMatches(msg, "enter"):
			// Shift+Enter inserts a newline (handled by textarea).
			// Plain Enter submits.
			if keyMatches(msg, "alt+enter") {
				break // let textarea handle alt+enter as newline
			}
			text := m.textarea.Value()
			if text == "" {
				return m, nil
			}
			m.history = append(m.history, text)
			m.histIdx = -1
			m.draft = ""
			m.textarea.Reset()
			m.lastHint = ""
			return m, func() tea.Msg { return submitMsg{text: text} }

		case keyMatches(msg, "up") || keyMatches(msg, "ctrl+p"):
			// If the cursor is on the first line, browse history.
			row := m.textarea.Line()
			if row == 0 && len(m.history) > 0 {
				if m.histIdx == -1 {
					m.draft = m.textarea.Value()
					m.histIdx = len(m.history) - 1
				} else if m.histIdx > 0 {
					m.histIdx--
				}
				m.textarea.SetValue(m.history[m.histIdx])
				return m, nil
			}

		case keyMatches(msg, "down") || keyMatches(msg, "ctrl+n"):
			// Browse history forward.
			if m.histIdx >= 0 {
				if m.histIdx < len(m.history)-1 {
					m.histIdx++
					m.textarea.SetValue(m.history[m.histIdx])
				} else {
					m.histIdx = -1
					m.textarea.SetValue(m.draft)
				}
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	m.lastHint = inferInputHint(m.textarea.Value())
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m InputModel) Hint() string {
	return m.lastHint
}

func inferInputHint(text string) string {
	trimmed := strings.TrimSpace(text)
	if strings.HasPrefix(trimmed, "/") {
		if strings.Count(trimmed, " ") == 0 {
			return "slash command mode"
		}
		return "slash arguments"
	}
	if strings.Contains(trimmed, "@") {
		return "reference mode"
	}
	if trimmed == "" {
		return "chat"
	}
	return "message"
}

func deleteTrailingWord(text string) string {
	runes := []rune(text)
	if len(runes) == 0 {
		return ""
	}
	i := len(runes) - 1
	for i >= 0 && unicode.IsSpace(runes[i]) {
		i--
	}
	for i >= 0 && !unicode.IsSpace(runes[i]) {
		i--
	}
	return strings.TrimRightFunc(string(runes[:i+1]), unicode.IsSpace)
}

// View implements tea.Model.
func (m InputModel) View() string {
	return m.textarea.View()
}
