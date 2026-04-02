package tui

import "testing"

func TestDetectTerminalCapabilitiesDumbTerminalExt(t *testing.T) {
	caps := detectTerminalCapabilities(map[string]string{"TERM": "dumb"})
	if caps.Mouse {
		t.Fatalf("expected dumb terminal without mouse support")
	}
}
