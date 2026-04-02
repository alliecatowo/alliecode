package commands

import (
	"context"
	"strings"
	"testing"
)

func TestModelDoctor(t *testing.T) {
	cmd := NewModelCommand()
	state := &RuntimeState{ProviderName: "openai"}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "model", Args: []string{"doctor"}})
	if err != nil || !strings.Contains(res.Message, "MODEL_DOCTOR") {
		t.Fatalf("model doctor failed: %v %q", err, res.Message)
	}
}
