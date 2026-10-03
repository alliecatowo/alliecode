package commands

import "testing"

func TestInteractivePanelSkillsIncludesActionableRows(t *testing.T) {
	state := &RuntimeState{
		Skills:              []string{"openclaw-status", "openclaw-config"},
		SkillsSources:       map[string]string{"openclaw-status": "plugin", "openclaw-config": "file"},
		SkillsOrigins:       map[string]string{"openclaw-status": "plugin:openclaw", "openclaw-config": "/workspace/.alliecode/skills/openclaw-config.md"},
		SkillsEnabled:       map[string]bool{"openclaw-status": true, "openclaw-config": false},
		SkillsConflictCount: 1,
		SkillsViewCount:     3,
	}
	panel := requirePanel(t, "skills", state)
	if panel.Title != "skills panel: /skills" {
		t.Fatalf("unexpected title: %q", panel.Title)
	}
	want := map[string]bool{
		"/skills list":                   false,
		"/skills status":                 false,
		"/skills doctor":                 false,
		"/skills sync":                   false,
		"/skills repair sync":            false,
		"/skills repair dedupe":          false,
		"/skills remove openclaw-status": false,
	}
	for _, item := range panel.Items {
		if _, ok := want[item.ApplyInput]; ok {
			want[item.ApplyInput] = true
		}
		if item.Section == "Skills" && len(item.PreviewIntents) == 0 {
			t.Fatalf("expected preview intents for skills row %q", item.Key)
		}
	}
	for cmd, ok := range want {
		if !ok {
			t.Fatalf("missing skills panel action %q", cmd)
		}
	}
}

func TestInteractivePanelSkillsStagesAddWhenEmpty(t *testing.T) {
	panel := requirePanel(t, "skills", &RuntimeState{})
	foundStage := false
	for _, item := range panel.Items {
		if item.ApplyInput == "/skills add " {
			foundStage = true
			if item.ApplyMode != PanelApplyStage {
				t.Fatalf("expected staged add mode, got %q", item.ApplyMode)
			}
		}
	}
	if !foundStage {
		t.Fatalf("expected staged add action for empty skills inventory")
	}
}
