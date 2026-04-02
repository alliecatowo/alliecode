package tasks

import "testing"

func TestTeamStoreStatusSummary(t *testing.T) {
	store := NewTeamStore()
	if _, err := store.Create("a", "", []TeamMember{{Name: "team_lead", Role: "lead"}}); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	s := store.StatusSummary()
	if s.Total != 1 || s.ByStatus[TeamStatusActive] != 1 {
		t.Fatalf("unexpected summary: %+v", s)
	}
}
