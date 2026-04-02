package commands

import (
	"context"
	"strings"
	"testing"
)

func TestHistoryDoctorIncludesQuickFix(t *testing.T) {
	cmd := NewHistoryCommand()
	res, err := cmd.Execute(context.Background(), Context{State: &RuntimeState{}}, Invocation{Name: "history", Args: []string{"doctor"}})
	if err != nil {
		t.Fatalf("history doctor failed: %v", err)
	}
	if !strings.Contains(res.Message, "HISTORY_DOCTOR") || !strings.Contains(res.Message, "quick_fix=") {
		t.Fatalf("unexpected history doctor output: %q", res.Message)
	}
}
