package commands

import "testing"

func TestInteractivePanelStatusIncludesOverviewAndDiagnostics(t *testing.T) {
	panel := requirePanel(t, "status", &RuntimeState{ProviderName: "openai", Model: "gpt-4o-mini"})
	if len(panel.HeaderIntents) == 0 {
		t.Fatalf("expected header intents for status panel")
	}
	if panel.Items[0].ApplyInput != "/status" {
		t.Fatalf("expected first status action to be /status, got %q", panel.Items[0].ApplyInput)
	}
	if panel.Items[1].ApplyInput != "/status diagnostics" {
		t.Fatalf("expected second status action to be /status diagnostics, got %q", panel.Items[1].ApplyInput)
	}
	if panel.Items[0].Section != "Overview" || len(panel.Items[0].PreviewIntents) == 0 {
		t.Fatalf("expected structured overview preview, got %#v", panel.Items[0])
	}
}
