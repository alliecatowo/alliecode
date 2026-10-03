package commands

import "testing"

func TestInteractivePanelsWorkflowsOpsBuildIntentfulRows(t *testing.T) {
	state := &RuntimeState{
		Tasks:              []string{"stabilize slash drawer", "panelize issue flows"},
		TasksCompleted:     2,
		ConfigValues:       map[string]string{"env.API_KEY": "redacted", "env.MODE": "test"},
		Issues:             []IssueRecord{{ID: "ISSUE-1", Title: "Broken panel flow", Status: "open", Provider: "github", Assignee: "allie"}},
		WorkflowRuns:       []WorkflowRun{{Name: "lint", Status: "running", Provider: "github", LastRun: "2026-04-02T00:00:00Z"}},
		ProactiveEnabled:   true,
		ProactiveRules:     []string{"queue-review"},
		AssistantMode:      "plan",
		AssistantSessionID: "assistant-9",
		ShareLinks:         []ShareLink{{ID: "share-1", Scope: "session", Visibility: "private", URL: "https://share.example.invalid/share-1"}},
		CompactMode:        "auto",
		CompactCount:       1,
	}
	for _, name := range []string{"tasks", "env", "issue", "workflows", "proactive", "assistant", "share", "compact"} {
		panel := requirePanel(t, name, state)
		if len(panel.HeaderIntents) == 0 {
			t.Fatalf("expected header intents for %q", name)
		}
		for _, item := range panel.Items {
			if item.ApplyInput == "" {
				t.Fatalf("expected apply input for %q item %+v", name, item)
			}
		}
	}
}
