package integration_test

import (
	"context"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
)

func TestBuddyCommandLifecycleFlow(t *testing.T) {
	registry := commands.DefaultRegistry()
	state := &commands.RuntimeState{}
	ctx := commands.Context{State: state}

	petBeforeHatch, err := registry.Dispatch(context.Background(), ctx, "/buddy pet")
	if err != nil {
		t.Fatalf("dispatch /buddy pet before hatch error = %v", err)
	}
	if !petBeforeHatch.Handled {
		t.Fatalf("/buddy pet before hatch should be handled")
	}
	if got := petBeforeHatch.Message; got != "BUDDY_PET\nstatus=not_hatched\nhatched=false\npet_count=0" {
		t.Fatalf("unexpected /buddy pet before hatch message = %q", got)
	}
	if state.BuddyHatched {
		t.Fatalf("BuddyHatched = true, want false before hatch")
	}

	hatchRes, err := registry.Dispatch(context.Background(), ctx, "/buddy hatch")
	if err != nil {
		t.Fatalf("dispatch /buddy hatch error = %v", err)
	}
	if got := hatchRes.Message; got != "BUDDY_HATCH\nstatus=hatched\nhatched=true" {
		t.Fatalf("unexpected /buddy hatch message = %q", got)
	}
	if !state.BuddyHatched {
		t.Fatalf("BuddyHatched = false, want true after hatch")
	}
	if state.BuddyPetCount != 0 {
		t.Fatalf("BuddyPetCount = %d, want 0 after hatch", state.BuddyPetCount)
	}

	statusRes, err := registry.Dispatch(context.Background(), ctx, "/buddy status")
	if err != nil {
		t.Fatalf("dispatch /buddy status error = %v", err)
	}
	if got := statusRes.Message; got != "BUDDY_STATUS\nstate=hatched\nhatched=true\nmuted=false\npet_count=0" {
		t.Fatalf("unexpected /buddy status message = %q", got)
	}

	petRes, err := registry.Dispatch(context.Background(), ctx, "/buddy pet")
	if err != nil {
		t.Fatalf("dispatch /buddy pet error = %v", err)
	}
	if got := petRes.Message; got != "BUDDY_PET\nstatus=pet\nhatched=true\nmuted=false\npet_count=1" {
		t.Fatalf("unexpected /buddy pet message = %q", got)
	}
	if state.BuddyPetCount != 1 {
		t.Fatalf("BuddyPetCount = %d, want 1", state.BuddyPetCount)
	}

	muteRes, err := registry.Dispatch(context.Background(), ctx, "/buddy mute")
	if err != nil {
		t.Fatalf("dispatch /buddy mute error = %v", err)
	}
	if got := muteRes.Message; got != "BUDDY_MUTE\nmuted=true" {
		t.Fatalf("unexpected /buddy mute message = %q", got)
	}
	if !state.BuddyMuted {
		t.Fatalf("BuddyMuted = false, want true after /buddy mute")
	}

	mutedPetRes, err := registry.Dispatch(context.Background(), ctx, "/buddy pet")
	if err != nil {
		t.Fatalf("dispatch /buddy pet while muted error = %v", err)
	}
	if got := mutedPetRes.Message; got != "BUDDY_PET\nstatus=pet_muted\nhatched=true\nmuted=true\npet_count=2" {
		t.Fatalf("unexpected /buddy pet while muted message = %q", got)
	}
	if state.BuddyPetCount != 2 {
		t.Fatalf("BuddyPetCount = %d, want 2 after muted pet", state.BuddyPetCount)
	}

	unmuteRes, err := registry.Dispatch(context.Background(), ctx, "/buddy unmute")
	if err != nil {
		t.Fatalf("dispatch /buddy unmute error = %v", err)
	}
	if got := unmuteRes.Message; got != "BUDDY_UNMUTE\nmuted=false" {
		t.Fatalf("unexpected /buddy unmute message = %q", got)
	}
	if state.BuddyMuted {
		t.Fatalf("BuddyMuted = true, want false after /buddy unmute")
	}

	postUnmutePetRes, err := registry.Dispatch(context.Background(), ctx, "/buddy pet")
	if err != nil {
		t.Fatalf("dispatch /buddy pet after unmute error = %v", err)
	}
	if got := postUnmutePetRes.Message; got != "BUDDY_PET\nstatus=pet\nhatched=true\nmuted=false\npet_count=3" {
		t.Fatalf("unexpected /buddy pet after unmute message = %q", got)
	}
	if state.BuddyPetCount != 3 {
		t.Fatalf("BuddyPetCount = %d, want 3 after unmuted pet", state.BuddyPetCount)
	}
}
