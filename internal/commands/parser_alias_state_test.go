package commands

import (
	"context"
	"strings"
	"testing"
)

func TestAliasStateCanRunDiagnosticsArgument(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/state diagnostics")
	if err != nil {
		t.Fatalf("state diagnostics alias failed: %v", err)
	}
	if !strings.Contains(res.Message, "STATUS_DIAGNOSTICS") {
		t.Fatalf("unexpected diagnostics message via alias: %q", res.Message)
	}
}
