package commands

import "testing"

func TestInteractivePanelMCPIncludesServerScopedActions(t *testing.T) {
	panel := requirePanel(t, "mcp", &RuntimeState{MCPConnections: map[string]bool{"github": true, "linear": false}})
	if len(panel.HeaderIntents) == 0 {
		t.Fatalf("expected header intents for MCP panel")
	}
	want := map[string]bool{"/mcp status github": false, "/mcp disconnect github": false, "/mcp status linear": false, "/mcp connect linear": false}
	for _, item := range panel.Items {
		if _, ok := want[item.ApplyInput]; ok {
			want[item.ApplyInput] = true
		}
	}
	for apply, ok := range want {
		if !ok {
			t.Fatalf("expected MCP panel action %q", apply)
		}
	}
	for _, item := range panel.Items {
		if item.Section == "Servers" && len(item.PreviewIntents) == 0 {
			t.Fatalf("expected structured server preview for %q", item.ApplyInput)
		}
	}
}
