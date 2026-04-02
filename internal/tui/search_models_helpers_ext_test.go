package tui

import "testing"

func TestQuickOpenSectionLabelFallback(t *testing.T) {
	if got := quickOpenSectionLabel("custom"); got != "Custom" {
		t.Fatalf("expected title-cased custom section, got %q", got)
	}
}
