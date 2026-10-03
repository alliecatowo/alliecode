package commands

import (
	"context"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStatusCommandEmitsSummaryIntent(t *testing.T) {
	state := &RuntimeState{Model: "gpt-4o", ProviderName: "openai"}
	res, err := NewStatusCommand().Execute(context.Background(), Context{State: state}, Invocation{Name: "status"})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if len(res.RenderIntents) == 0 || res.RenderIntents[0].Kind != types.RenderIntentSummaryCard {
		t.Fatalf("expected summary card intent, got %#v", res.RenderIntents)
	}
}

func TestDoctorCommandFixEmitsChecklistIntent(t *testing.T) {
	state := &RuntimeState{}
	res, err := NewDoctorCommand().Execute(context.Background(), Context{State: state}, Invocation{Name: "doctor", Args: []string{"fix"}})
	if err != nil {
		t.Fatalf("doctor fix failed: %v", err)
	}
	foundChecklist := false
	for _, intent := range res.RenderIntents {
		if intent.Kind == types.RenderIntentChecklist {
			foundChecklist = true
		}
	}
	if !foundChecklist {
		t.Fatalf("expected checklist intent, got %#v", res.RenderIntents)
	}
}
