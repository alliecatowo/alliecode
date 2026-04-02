package tasks

import "testing"

func TestTeamStoreLifecycle(t *testing.T) {
	store := NewTeamStore()
	members := []TeamMember{{Name: "team_lead", Role: "lead"}, {Name: "dev1", Role: "member"}}

	created, err := store.Create("alpha", "desc", members)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if created.Status != TeamStatusActive {
		t.Fatalf("status = %q, want %q", created.Status, TeamStatusActive)
	}
	if store.Active() != "alpha" {
		t.Fatalf("active team = %q, want alpha", store.Active())
	}

	msg, err := store.AddMessage("alpha", "team_lead", "dev1", "sync", "hello")
	if err != nil {
		t.Fatalf("AddMessage returned error: %v", err)
	}
	if msg.ID == "" {
		t.Fatalf("expected message id")
	}

	current, ok := store.Get("alpha")
	if !ok {
		t.Fatalf("expected team lookup to succeed")
	}
	if len(current.Messages) != 1 {
		t.Fatalf("expected one message, got %d", len(current.Messages))
	}

	newDesc := "updated"
	updated, ok := store.Update("alpha", TeamUpdateParams{Description: &newDesc})
	if !ok {
		t.Fatalf("expected update to succeed")
	}
	if updated.Description != "updated" {
		t.Fatalf("description = %q, want updated", updated.Description)
	}

	canceled, ok := store.Cancel("alpha", "done")
	if !ok {
		t.Fatalf("expected cancel to succeed")
	}
	if canceled.Status != TeamStatusCanceled {
		t.Fatalf("status = %q, want %q", canceled.Status, TeamStatusCanceled)
	}
	if store.Active() != "" {
		t.Fatalf("expected no active team after cancel")
	}

	stats := store.Stats()
	if stats.Total != 1 || stats.Active != 0 || stats.Canceled != 1 || stats.Archived != 0 {
		t.Fatalf("unexpected team stats: %+v", stats)
	}
}

func TestTeamStoreStatsWithActiveTeams(t *testing.T) {
	store := NewTeamStore()
	if _, err := store.Create("alpha", "", []TeamMember{{Name: "team_lead", Role: "lead"}}); err != nil {
		t.Fatalf("create alpha failed: %v", err)
	}
	if _, err := store.Create("beta", "", []TeamMember{{Name: "team_lead", Role: "lead"}}); err != nil {
		t.Fatalf("create beta failed: %v", err)
	}

	stats := store.Stats()
	if stats.Total != 2 || stats.Active != 2 || stats.Canceled != 0 || stats.Archived != 0 {
		t.Fatalf("unexpected team stats: %+v", stats)
	}
}

func TestTeamStoreArchiveLifecycle(t *testing.T) {
	store := NewTeamStore()
	if _, err := store.Create("alpha", "", []TeamMember{{Name: "team_lead", Role: "lead"}}); err != nil {
		t.Fatalf("create alpha failed: %v", err)
	}

	archived, ok := store.Archive("alpha", "idle")
	if !ok {
		t.Fatalf("expected archive to succeed")
	}
	if archived.Status != TeamStatusArchived {
		t.Fatalf("status = %q, want %q", archived.Status, TeamStatusArchived)
	}
	if store.Active() != "" {
		t.Fatalf("expected no active team after archive")
	}

	stats := store.Stats()
	if stats.Total != 1 || stats.Archived != 1 || stats.Active != 0 || stats.Canceled != 0 {
		t.Fatalf("unexpected team stats: %+v", stats)
	}
}
