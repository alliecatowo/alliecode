package commands

import (
	"strings"
	"testing"
)

func TestInteractivePanelModelIncludesCurrentAndSwitchTargets(t *testing.T) {
	panel := requirePanel(t, "model", &RuntimeState{ProviderName: "anthropic", Model: "claude-opus-4-20250514", ModelRef: "anthropic/claude-opus-4-20250514", LoggedIn: true})
	if len(panel.HeaderIntents) == 0 {
		t.Fatalf("expected header intents for model panel")
	}
	if panel.Items[0].ApplyInput != "/model" {
		t.Fatalf("expected first model item to inspect current selection, got %q", panel.Items[0].ApplyInput)
	}
	foundSwitch := false
	foundStructuredSwitch := false
	for _, item := range panel.Items {
		if item.ApplyInput == "/model anthropic/claude-opus-4-20250514" {
			foundSwitch = true
			if strings.TrimSpace(item.Section) != "" && len(item.PreviewIntents) > 0 {
				foundStructuredSwitch = true
			}
		}
	}
	if !foundSwitch {
		t.Fatalf("expected model panel to include current model switch target")
	}
	if !foundStructuredSwitch {
		t.Fatalf("expected structured preview for concrete model target")
	}
}
