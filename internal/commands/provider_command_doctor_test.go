package commands

import (
	"context"
	"strings"
	"testing"
)

func TestProviderDoctor(t *testing.T) {
	cmd := NewProviderCommand()
	state := &RuntimeState{ProviderName: "openai", Model: "openai/gpt-4o"}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "provider", Args: []string{"doctor"}})
	if err != nil || !strings.Contains(res.Message, "PROVIDER_DOCTOR") {
		t.Fatalf("provider doctor failed: %v %q", err, res.Message)
	}
	if !strings.Contains(res.Message, "quick_fix=") {
		t.Fatalf("expected quick fix in provider doctor output: %q", res.Message)
	}
}
