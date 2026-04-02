package commands

import (
	"context"
	"strings"
	"testing"
)

func TestProviderRepairDefaultsToOllama(t *testing.T) {
	cmd := NewProviderCommand()
	state := &RuntimeState{}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "provider", Args: []string{"repair"}})
	if err != nil || !strings.Contains(res.Message, "PROVIDER_REPAIR") {
		t.Fatalf("provider repair failed: %v %q", err, res.Message)
	}
}
