package tasks

import "testing"

func TestApplyTaskQuerySummaryMetadata(t *testing.T) {
	items := []Task{{ID: "a", Status: StatusRunning, Owner: "dev1"}, {ID: "b", Status: StatusCompleted, Owner: "dev1"}, {ID: "c", Status: StatusFailed, Owner: "dev2"}}
	includeTerminal := false
	out, summary := ApplyTaskQuery(items, TaskQuery{Owner: "DEV1", Statuses: []Status{StatusRunning, StatusCompleted}, IncludeTerminal: &includeTerminal, Limit: 5})
	if len(out) != 1 || out[0].ID != "a" {
		t.Fatalf("unexpected filtered tasks: %+v", out)
	}
	if summary.Scanned != 3 || summary.FilteredOut != 2 {
		t.Fatalf("unexpected scan/filter counts: %+v", summary)
	}
	if !summary.HasStatusFilter || summary.IncludeTerminal {
		t.Fatalf("unexpected status/include flags: %+v", summary)
	}
	if summary.RequestedLimit != 5 || summary.OwnerFilter != "dev1" {
		t.Fatalf("unexpected requested limit/owner filter: %+v", summary)
	}
}

func TestApplyTeamQuerySummaryMetadata(t *testing.T) {
	teams := []Team{
		{Name: "alpha", Status: TeamStatusActive, Members: []TeamMember{{Name: "dev1"}}},
		{Name: "beta", Status: TeamStatusArchived, Members: []TeamMember{{Name: "dev2"}}},
		{Name: "gamma", Status: TeamStatusCanceled, Members: []TeamMember{{Name: "dev1"}}},
	}
	out, summary := ApplyTeamQuery(teams, TeamQuery{Member: "DEV1", Statuses: []TeamStatus{TeamStatusActive, TeamStatusCanceled}, Limit: 1})
	if len(out) != 1 || out[0].Name != "alpha" {
		t.Fatalf("unexpected filtered teams: %+v", out)
	}
	if summary.Scanned != 3 || summary.Matched != 2 || summary.FilteredOut != 1 || !summary.Truncated {
		t.Fatalf("unexpected summary counts: %+v", summary)
	}
	if summary.MemberFilter != "dev1" {
		t.Fatalf("member filter mismatch: %+v", summary)
	}
}
