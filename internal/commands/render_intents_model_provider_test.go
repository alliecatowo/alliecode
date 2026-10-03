package commands

import (
	"context"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestModelCommandEmitsRenderIntents(t *testing.T) {
	state := &RuntimeState{ProviderName: "openai", Model: "gpt-4o", ModelRef: "openai/gpt-4o", LoggedIn: true}
	res, err := NewModelCommand().Execute(context.Background(), Context{State: state}, Invocation{Name: "model"})
	if err != nil {
		t.Fatalf("model status failed: %v", err)
	}
	if len(res.RenderIntents) == 0 || res.RenderIntents[0].Kind != types.RenderIntentSummaryCard {
		t.Fatalf("expected summary card render intent, got %#v", res.RenderIntents)
	}
}

func TestProviderCommandListEmitsTableIntent(t *testing.T) {
	state := &RuntimeState{}
	res, err := NewProviderCommand().Execute(context.Background(), Context{State: state}, Invocation{Name: "provider", Args: []string{"list"}})
	if err != nil {
		t.Fatalf("provider list failed: %v", err)
	}
	if len(res.RenderIntents) == 0 || res.RenderIntents[0].Kind != types.RenderIntentTable {
		t.Fatalf("expected table render intent, got %#v", res.RenderIntents)
	}
}
