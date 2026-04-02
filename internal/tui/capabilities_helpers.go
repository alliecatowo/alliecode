package tui

import "strings"

type terminalCapabilities struct {
	Hyperlinks bool
	TrueColor  bool
	Mouse      bool
}

func detectTerminalCapabilities(env map[string]string) terminalCapabilities {
	term := strings.ToLower(strings.TrimSpace(env["TERM"]))
	program := strings.ToLower(strings.TrimSpace(env["TERM_PROGRAM"]))
	colorTerm := strings.ToLower(strings.TrimSpace(env["COLORTERM"]))
	vteVersion := strings.TrimSpace(env["VTE_VERSION"])

	return terminalCapabilities{
		Hyperlinks: supportsHyperlinks(term, program, vteVersion),
		TrueColor:  supportsTrueColor(term, colorTerm),
		Mouse:      supportsMouse(term),
	}
}

func supportsHyperlinks(term, termProgram, vteVersion string) bool {
	if strings.Contains(termProgram, "wezterm") || strings.Contains(termProgram, "iterm") {
		return true
	}
	if strings.Contains(term, "kitty") || strings.Contains(term, "wezterm") {
		return true
	}
	if strings.HasPrefix(term, "xterm") && strings.TrimSpace(vteVersion) != "" {
		return true
	}
	if strings.Contains(term, "vte") || strings.Contains(term, "foot") || strings.Contains(term, "alacritty") {
		return true
	}
	return false
}

func supportsTrueColor(term, colorTerm string) bool {
	if strings.Contains(colorTerm, "truecolor") || strings.Contains(colorTerm, "24bit") {
		return true
	}
	if strings.Contains(term, "truecolor") || strings.Contains(term, "24bit") {
		return true
	}
	if strings.Contains(term, "direct") {
		return true
	}
	return false
}

func supportsMouse(term string) bool {
	if term == "" || term == "dumb" {
		return false
	}
	if strings.Contains(term, "xterm") || strings.Contains(term, "screen") || strings.Contains(term, "tmux") {
		return true
	}
	if strings.Contains(term, "rxvt") || strings.Contains(term, "linux") || strings.Contains(term, "kitty") {
		return true
	}
	if strings.Contains(term, "wezterm") || strings.Contains(term, "foot") || strings.Contains(term, "alacritty") {
		return true
	}
	return false
}
