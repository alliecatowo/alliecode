package commands

import (
	"context"
	"strings"
	"testing"
)

func TestSkillsDoctorRepairAndSync(t *testing.T) {
	cmd := NewSkillsCommand()
	state := &RuntimeState{Skills: []string{"lint", "lint"}, SkillsSources: map[string]string{"lint": "state"}, SkillsOrigins: map[string]string{"lint": "manual"}, SkillsEnabled: map[string]bool{"lint": true}}
	syncRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "skills", Args: []string{"sync"}})
	if err != nil {
		t.Fatalf("skills sync failed: %v", err)
	}
	if !strings.Contains(syncRes.Message, "SKILLS_SYNC") {
		t.Fatalf("unexpected skills sync: %q", syncRes.Message)
	}
	doctorRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "skills", Args: []string{"doctor"}})
	if err != nil {
		t.Fatalf("skills doctor failed: %v", err)
	}
	if !strings.Contains(doctorRes.Message, "SKILLS_DOCTOR") || !strings.Contains(doctorRes.Message, "enabled=") || !strings.Contains(doctorRes.Message, "conflicts=") {
		t.Fatalf("unexpected skills doctor: %q", doctorRes.Message)
	}
	repairRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "skills", Args: []string{"repair", "dedupe"}})
	if err != nil {
		t.Fatalf("skills repair failed: %v", err)
	}
	if !strings.Contains(repairRes.Message, "SKILLS_REPAIR") || !strings.Contains(repairRes.Message, "conflicts=") {
		t.Fatalf("unexpected skills repair: %q", repairRes.Message)
	}
}
