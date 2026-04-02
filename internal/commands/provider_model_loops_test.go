package commands

import (
	"context"
	"strings"
	"testing"
)

func TestProviderAndModelLoopsSubcommands(t *testing.T) {
	state := &RuntimeState{}
	providerRes, err := NewProviderCommand().Execute(context.Background(), Context{State: state}, Invocation{Name: "provider", Args: []string{"loops"}})
	if err != nil {
		t.Fatalf("provider loops failed: %v", err)
	}
	if !strings.Contains(providerRes.Message, "PROVIDER_LOOPS") {
		t.Fatalf("unexpected provider loops output: %q", providerRes.Message)
	}
	modelRes, err := NewModelCommand().Execute(context.Background(), Context{State: state}, Invocation{Name: "model", Args: []string{"loops"}})
	if err != nil {
		t.Fatalf("model loops failed: %v", err)
	}
	if !strings.Contains(modelRes.Message, "MODEL_LOOPS") {
		t.Fatalf("unexpected model loops output: %q", modelRes.Message)
	}
}
