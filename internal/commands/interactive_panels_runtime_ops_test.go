package commands

import "testing"

func TestInteractivePanelsRuntimeOpsAvailable(t *testing.T) {
	state := &RuntimeState{
		Branches:         []string{"main"},
		ActiveBranch:     "main",
		DiffMode:         "working",
		ContextFiles:     []string{"README.md"},
		OutputStyle:      "system",
		OutputFormat:     "default",
		MemoryEntries:    []string{"remember this"},
		PrivacyTelemetry: true,
	}
	for _, name := range []string{"provider", "branch", "diff", "files", "memory", "theme", "output-style", "privacy-settings", "upgrade", "resume", "plan"} {
		if !SupportsInteractivePanel(name) {
			t.Fatalf("expected interactive panel support for %q", name)
		}
		panel := requirePanel(t, name, state)
		if len(panel.HeaderIntents) == 0 {
			t.Fatalf("expected header intents for %q", name)
		}
	}
}

func TestInteractivePanelsWorkflowsOpsFamiliesIncludeHeaderIntents(t *testing.T) {
	state := &RuntimeState{
		Tasks:              []string{"stabilize command drawer", "add panel tests"},
		TasksCompleted:     3,
		ConfigValues:       map[string]string{"env.API_KEY": "redacted", "env.MODE": "test"},
		Issues:             []IssueRecord{{ID: "ISSUE-1", Title: "Panel drift", Status: "open", Provider: "github"}},
		WorkflowRuns:       []WorkflowRun{{Name: "lint", Status: "running", Provider: "github", LastRun: "2026-04-02T00:00:00Z"}},
		ProactiveEnabled:   true,
		ProactiveRules:     []string{"queue-review"},
		AssistantMode:      "plan",
		AssistantSessionID: "assistant-7",
		ShareLinks:         []ShareLink{{ID: "share-1", Scope: "session", Visibility: "private", URL: "https://share.example.invalid/share-1"}},
		CompactMode:        "auto",
		CompactCount:       2,
	}
	for _, name := range []string{"tasks", "env", "issue", "workflows", "proactive", "assistant", "share", "compact"} {
		if !SupportsInteractivePanel(name) {
			t.Fatalf("expected interactive panel support for %q", name)
		}
		panel := requirePanel(t, name, state)
		if len(panel.HeaderIntents) == 0 {
			t.Fatalf("expected header intents for %q", name)
		}
	}
}
