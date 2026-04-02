package commands

import (
	"context"
	"strings"
	"testing"
)

func TestModelListByProvider(t *testing.T) {
	cmd := NewModelCommand()
	state := &RuntimeState{ProviderName: "openai"}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "model", Args: []string{"list", "openai"}})
	if err != nil || !strings.Contains(res.Message, "MODEL_LIST") {
		t.Fatalf("model list failed: %v %q", err, res.Message)
	}
}
