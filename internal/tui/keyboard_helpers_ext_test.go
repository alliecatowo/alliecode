package tui

import "testing"

func TestComboVariantsCommaSeparator(t *testing.T) {
	variants := comboVariants("ctrl+g, ctrl+f")
	if len(variants) != 2 {
		t.Fatalf("expected two combo variants, got %#v", variants)
	}
}
