package commands

import (
	"context"
	"strings"
	"testing"
)

func TestModelRepair(t *testing.T) {
	cmd := NewModelCommand()
	state := &RuntimeState{ProviderName: "openai"}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "model", Args: []string{"repair", "openai/gpt-4o"}})
	if err != nil || !strings.Contains(res.Message, "MODEL_REPAIR") {
		t.Fatalf("model repair failed: %v %q", err, res.Message)
	}
	if !strings.Contains(res.Message, "provider_ready=") {
		t.Fatalf("expected provider readiness in repair output: %q", res.Message)
	}
}
