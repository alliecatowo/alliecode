package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func keyMatches(msg tea.KeyMsg, combo string) bool {
	got := normalizeCombo(msg.String())
	for _, want := range comboVariants(combo) {
		if want == got {
			return true
		}
		if want == "shift+tab" && msg.Type == tea.KeyShiftTab {
			return true
		}
		if want == "alt+enter" && msg.Type == tea.KeyEnter && msg.Alt {
			return true
		}
		if want == "enter" && msg.Type == tea.KeyEnter {
			return true
		}
		if want == "backspace" && (msg.Type == tea.KeyBackspace || msg.Type == tea.KeyDelete) {
			return true
		}
	}
	return false
}

func comboVariants(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == '|'
	})
	if len(parts) == 0 {
		return []string{normalizeCombo(s)}
	}

	out := make([]string, 0, len(parts))
	for _, part := range parts {
		candidate := strings.TrimSpace(part)
		if candidate == "" {
			continue
		}
		if strings.Contains(candidate, ">") {
			steps := strings.Split(candidate, ">")
			candidate = strings.TrimSpace(steps[len(steps)-1])
		}
		norm := normalizeCombo(candidate)
		if norm != "" {
			out = append(out, norm)
		}
	}
	if len(out) == 0 {
		return []string{normalizeCombo(s)}
	}
	return out
}

func normalizeCombo(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return ""
	}
	s = normalizeTerminalSequenceAlias(s)
	parts := strings.Split(s, "+")

	ctrl := false
	alt := false
	shift := false
	key := ""

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		p = normalizeTerminalSequenceAlias(p)
		switch p {
		case "ctrl", "control", "cmd", "command":
			ctrl = true
		case "alt", "meta", "option":
			alt = true
		case "shift":
			shift = true
		case "s-enter", "shift+return", "shift+iso-enter":
			if key == "" {
				key = "enter"
			}
			shift = true
		case "esc":
			if key == "" {
				key = "escape"
			}
		case "return":
			if key == "" {
				key = "enter"
			}
		case "del":
			if key == "" {
				key = "delete"
			}
		case "pgup":
			if key == "" {
				key = "pageup"
			}
		case "pgdn", "pgdown":
			if key == "" {
				key = "pagedown"
			}
		case "spacebar":
			if key == "" {
				key = "space"
			}
		case "s-tab", "backtab", "iso-left-tab":
			if key == "" {
				key = "tab"
			}
			shift = true
		default:
			if key == "" {
				key = p
			}
		}
	}

	out := make([]string, 0, 4)
	if ctrl {
		out = append(out, "ctrl")
	}
	if alt {
		out = append(out, "alt")
	}
	if shift {
		out = append(out, "shift")
	}
	if key != "" {
		out = append(out, key)
	}
	return strings.Join(out, "+")
}

func normalizeTerminalSequenceAlias(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "\\x1b", "\x1b")
	if s == "" {
		return ""
	}

	switch s {
	case "\x1b[z", "\x1b[1;2z", "\x1b[9;2u", "\x1b[27;2;9~", "[z", "1;2z", "9;2u", "27;2;9~", "back-tab", "kcbt", "btab":
		return "shift+tab"
	case "\x1b[13;2u", "\x1b[27;2;13~", "13;2u", "27;2;13~":
		return "shift+enter"
	}

	return s
}
