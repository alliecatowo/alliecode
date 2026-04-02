package commands

import (
	"context"
	"strings"
	"testing"
)

func TestHelpIncludesPaletteContextAndDiagnosticsFields(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/help status")
	if err != nil {
		t.Fatalf("help dispatch failed: %v", err)
	}
	if !strings.Contains(res.Message, "palette_group=") || !strings.Contains(res.Message, "context_count=") || !strings.Contains(res.Message, "diagnostic_count=") {
		t.Fatalf("expected rich help metadata fields, got: %q", res.Message)
	}
}
