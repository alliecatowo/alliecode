package tui

import "testing"

func TestDetectTerminalCapabilitiesTrueColorAndHyperlinks(t *testing.T) {
	env := map[string]string{
		"TERM":         "xterm-256color",
		"TERM_PROGRAM": "iTerm.app",
		"COLORTERM":    "truecolor",
	}
	got := detectTerminalCapabilities(env)
	if !got.Hyperlinks {
		t.Fatalf("expected hyperlinks support")
	}
	if !got.TrueColor {
		t.Fatalf("expected truecolor support")
	}
	if !got.Mouse {
		t.Fatalf("expected mouse support")
	}
}

func TestDetectTerminalCapabilitiesDumbTerminal(t *testing.T) {
	env := map[string]string{
		"TERM": "dumb",
	}
	got := detectTerminalCapabilities(env)
	if got.Hyperlinks {
		t.Fatalf("expected no hyperlinks support")
	}
	if got.TrueColor {
		t.Fatalf("expected no truecolor support")
	}
	if got.Mouse {
		t.Fatalf("expected no mouse support")
	}
}

func TestSupportsHyperlinksFromVTEVersion(t *testing.T) {
	if !supportsHyperlinks("xterm", "", "6000") {
		t.Fatalf("expected xterm with VTE version to support hyperlinks")
	}
}

func TestSupportsTrueColorFromDirectTerm(t *testing.T) {
	if !supportsTrueColor("xterm-direct", "") {
		t.Fatalf("expected direct term to support truecolor")
	}
}
