package tasks

import "testing"

func TestTeamStoreQueryAndSummary(t *testing.T) {
	store := NewTeamStore()
	_, _ = store.Create("a", "", []TeamMember{{Name: "dev1"}})
	_, _ = store.Create("b", "", []TeamMember{{Name: "dev2"}})
	items := store.Query(TeamQuery{Member: "dev1"})
	if len(items) != 1 || items[0].Name != "a" {
		t.Fatalf("unexpected queried teams: %+v", items)
	}
	s := store.QuerySummary(TeamQuery{Member: "dev1"})
	if s.Matched != 1 || s.ByStatus[TeamStatusActive] != 1 {
		t.Fatalf("unexpected query summary: %+v", s)
	}
}
