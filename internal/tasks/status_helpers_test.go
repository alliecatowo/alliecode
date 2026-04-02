package tasks

import "testing"

func TestNewStatusSummary(t *testing.T) {
	s := newStatusSummary([]Task{{ID: "a", Status: StatusRunning}, {ID: "b", Status: StatusCompleted, Owner: "dev"}})
	if s.Total != 2 || s.Terminal != 1 || s.ByOwner["dev"] != 1 {
		t.Fatalf("unexpected summary: %+v", s)
	}
}
