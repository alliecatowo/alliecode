package commands

import (
	"strings"
	"testing"
)

func TestInteractivePanelProviderIncludesProviderActionsAndIntentPreviews(t *testing.T) {
	state := &RuntimeState{ProviderName: "openai", Model: "gpt-4o-mini", ModelRef: "openai/gpt-4o-mini", LoggedIn: true, ProviderReady: true}
	panel := requirePanel(t, "provider", state)

	if panel.Command != "provider" {
		t.Fatalf("expected provider command panel, got %q", panel.Command)
	}
	if !strings.Contains(panel.Title, "/provider") {
		t.Fatalf("expected provider title, got %q", panel.Title)
	}
	if len(panel.HeaderIntents) == 0 {
		t.Fatalf("expected provider panel header intents")
	}
	if len(panel.Items) < 4 {
		t.Fatalf("expected provider panel items, got %d", len(panel.Items))
	}

	foundSet := false
	foundModels := false
	for _, item := range panel.Items {
		if strings.TrimSpace(item.ApplyInput) == "" {
			t.Fatalf("expected provider panel item with apply input: %+v", item)
		}
		if len(item.PreviewIntents) == 0 {
			t.Fatalf("expected intent-backed preview for provider item %q", item.Key)
		}
		if strings.HasPrefix(item.ApplyInput, "/provider set ") {
			foundSet = true
		}
		if strings.HasPrefix(item.ApplyInput, "/provider models ") {
			foundModels = true
		}
	}
	if !foundSet {
		t.Fatalf("expected provider panel to include provider set actions")
	}
	if !foundModels {
		t.Fatalf("expected provider panel to include provider models actions")
	}
}
