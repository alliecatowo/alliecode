package tasks

import "testing"

func TestNewTeamStatusSummary(t *testing.T) {
	s := NewTeamStatusSummary([]Team{{Name: "a", Status: TeamStatusActive, Members: []TeamMember{{Name: "dev1"}}}, {Name: "b", Status: TeamStatusArchived}})
	if s.Total != 2 || s.ByStatus[TeamStatusActive] != 1 {
		t.Fatalf("unexpected team summary: %+v", s)
	}
}
