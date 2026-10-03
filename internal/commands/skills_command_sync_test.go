package commands

import (
	"context"
	"strings"
	"testing"
)

func TestSkillsSync(t *testing.T) {
	cmd := NewSkillsCommand()
	state := &RuntimeState{Skills: []string{"build", "build", "review"}, SkillsSources: map[string]string{"build": "state"}, SkillsOrigins: map[string]string{"build": "manual"}, SkillsEnabled: map[string]bool{"build": true}}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "skills", Args: []string{"sync"}})
	if err != nil || !strings.Contains(res.Message, "SKILLS_SYNC") {
		t.Fatalf("skills sync failed: %v %q", err, res.Message)
	}
	if len(state.Skills) != 2 {
		t.Fatalf("expected deduped skills, got %+v", state.Skills)
	}
	if state.SkillsLastSyncSource == "" {
		t.Fatalf("expected sync source metadata")
	}
	if state.SkillsSources == nil || state.SkillsOrigins == nil || state.SkillsEnabled == nil {
		t.Fatalf("expected skills metadata maps to be populated")
	}
}
