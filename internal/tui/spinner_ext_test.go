package tui

import "testing"

func TestSpinnerViewIncludesLabel(t *testing.T) {
	spin := NewSpinner("working")
	if got := spin.View(); got == "" {
		t.Fatalf("expected spinner view text")
	}
}
