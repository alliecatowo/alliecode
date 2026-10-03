package tui

import (
	"strings"
	"testing"
)

func TestModalStackSignatureIncludesActiveLayers(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.startQuickOpen()
	sig := app.modalStackSignature()
	if !strings.Contains(sig, "search") {
		t.Fatalf("expected signature to include search, got %q", sig)
	}
}
