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
		switch p {
		case "ctrl", "control", "cmd", "command":
			ctrl = true
		case "alt", "meta", "option":
			alt = true
		case "shift":
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
