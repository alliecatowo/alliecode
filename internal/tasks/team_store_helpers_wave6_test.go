package tasks

import "testing"

func TestTeamStoreHelperQueriesWave6(t *testing.T) {
	store := NewTeamStore()
	team, err := store.Create("wave6-team", "", nil)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if _, err := store.AddMessage(team.Name, "a", "b", "sum", "msg"); err != nil {
		t.Fatalf("add message failed: %v", err)
	}
	count, ok := store.MessageCount(team.Name)
	if !ok || count != 1 {
		t.Fatalf("unexpected message count: ok=%v count=%d", ok, count)
	}
	ev, ok := store.LastEvent(team.Name)
	if !ok || ev.Type == "" {
		t.Fatalf("expected last event")
	}
}
