package tasks

import "testing"

func TestApplyTeamQuery(t *testing.T) {
	teams := []Team{{Name: "alpha", Status: TeamStatusActive, Members: []TeamMember{{Name: "dev1"}}}, {Name: "beta", Status: TeamStatusArchived, Members: []TeamMember{{Name: "dev2"}}}}
	filtered, summary := ApplyTeamQuery(teams, TeamQuery{Member: "dev1", Limit: 1})
	if len(filtered) != 1 || filtered[0].Name != "alpha" {
		t.Fatalf("unexpected teams: %+v", filtered)
	}
	if summary.Matched != 1 || summary.Returned != 1 {
		t.Fatalf("unexpected team summary: %+v", summary)
	}
}
