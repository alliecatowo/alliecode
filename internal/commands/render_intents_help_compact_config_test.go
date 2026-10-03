package commands

import (
	"context"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestHelpCommandEmitsOptionListIntents(t *testing.T) {
	cmd := NewHelpCommand(DefaultRegistry())
	res, err := cmd.Execute(context.Background(), Context{}, Invocation{Name: "help", Args: []string{"model"}})
	if err != nil {
		t.Fatalf("help failed: %v", err)
	}
	if len(res.RenderIntents) < 2 {
		t.Fatalf("expected multiple help intents, got %#v", res.RenderIntents)
	}
	if res.RenderIntents[1].Kind != types.RenderIntentOptionList {
		t.Fatalf("expected option list intent, got %#v", res.RenderIntents)
	}
}

func TestCompactCommandEmitsActionListIntents(t *testing.T) {
	cmd := NewCompactCommand()
	res, err := cmd.Execute(context.Background(), Context{State: &RuntimeState{}}, Invocation{Name: "compact", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("compact status failed: %v", err)
	}
	found := false
	for _, intent := range res.RenderIntents {
		if intent.Kind == types.RenderIntentActionList {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected action list intent, got %#v", res.RenderIntents)
	}
}

func TestConfigCommandsEmitStructuredIntents(t *testing.T) {
	cmd := NewConfigCommand()
	state := &RuntimeState{ConfigValues: map[string]string{"settings.output-style": "human", "settings.transport": "local"}}
	showRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "config", Args: []string{"show"}})
	if err != nil {
		t.Fatalf("config show failed: %v", err)
	}
	if len(showRes.RenderIntents) < 2 || showRes.RenderIntents[1].Kind != types.RenderIntentOptionList {
		t.Fatalf("expected config show option list intent, got %#v", showRes.RenderIntents)
	}
	repairRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "config", Args: []string{"repair", "safe"}})
	if err != nil {
		t.Fatalf("config repair failed: %v", err)
	}
	if len(repairRes.RenderIntents) < 2 || repairRes.RenderIntents[1].Kind != types.RenderIntentDetailRows {
		t.Fatalf("expected config repair detail rows intent, got %#v", repairRes.RenderIntents)
	}
	doctorRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "config", Args: []string{"doctor"}})
	if err != nil {
		t.Fatalf("config doctor failed: %v", err)
	}
	if len(doctorRes.RenderIntents) < 2 || doctorRes.RenderIntents[1].Kind != types.RenderIntentActionList {
		t.Fatalf("expected config doctor action list intent, got %#v", doctorRes.RenderIntents)
	}
}
