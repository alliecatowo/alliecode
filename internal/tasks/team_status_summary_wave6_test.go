package tasks

import (
	"testing"
	"time"
)

func TestTeamStatusSummaryRecentTeamWave6(t *testing.T) {
	now := time.Now().UTC()
	s := NewTeamStatusSummary([]Team{{Name: "a", Status: TeamStatusActive, UpdatedAt: now}, {Name: "b", Status: TeamStatusArchived, UpdatedAt: now.Add(time.Second)}})
	if s.RecentTeam != "b" || s.Total != 2 {
		t.Fatalf("unexpected team summary: %+v", s)
	}
}
