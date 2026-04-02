package tui

import (
	"strings"
	"testing"
	"time"
)

func TestRenderBuddySpriteNarrowAndFull(t *testing.T) {
	state := newBuddyState()
	now := time.Unix(100, 0)

	narrow := renderBuddySprite(state, 40, now)
	if strings.Contains(narrow, "\n") {
		t.Fatalf("expected narrow one-line buddy, got %q", narrow)
	}

	full := renderBuddySprite(state, 120, now)
	if visualLineCount(full, 120) < 5 {
		t.Fatalf("expected multi-line full sprite at wide width, got %q", full)
	}
}

func TestRenderBuddyBubbleVisibilityAndFade(t *testing.T) {
	now := time.Unix(100, 0)
	state := buddyState{bubble: "nice", bubbleUntil: now.Add(10 * time.Second), bubbleFadeAt: now.Add(7 * time.Second)}
	if got, _ := renderBuddyBubble(state, 20, now); got != "" {
		t.Fatalf("expected hidden bubble on narrow width, got %q", got)
	}
	if got, fading := renderBuddyBubble(state, 40, now); got == "" || fading {
		t.Fatalf("expected visible bubble on wide width")
	}

	if got, fading := renderBuddyBubble(state, 40, now.Add(8*time.Second)); got == "" || !fading {
		t.Fatalf("expected visible fading bubble near expiry, got %q fading=%t", got, fading)
	}

	if got, _ := renderBuddyBubble(state, 40, now.Add(11*time.Second)); got != "" {
		t.Fatalf("expected hidden bubble when ttl expired, got %q", got)
	}
}

func TestUpdateBuddyStateHooksUserAndAssistantMessages(t *testing.T) {
	state := newBuddyState()
	now := time.Unix(100, 0)
	state = updateBuddyState(state, timelineEntry{kind: timelineUser, text: "buddy hello"}, now)
	if state.bubble != "hi hi" {
		t.Fatalf("expected user reaction bubble, got %q", state.bubble)
	}
	if !state.bubbleUntil.Equal(now.Add(10 * time.Second)) {
		t.Fatalf("expected user bubble ttl window, got %v", state.bubbleUntil)
	}

	state = updateBuddyState(state, timelineEntry{kind: timelineAssistant, text: "done"}, now.Add(1*time.Second))
	if state.bubble != "nice" {
		t.Fatalf("expected assistant reaction bubble, got %q", state.bubble)
	}
	if !state.bubbleUntil.Equal(now.Add(11 * time.Second)) {
		t.Fatalf("expected assistant ttl reset, got %v", state.bubbleUntil)
	}
}

func TestUpdateBuddyStatePetBurstAndMuteClear(t *testing.T) {
	state := newBuddyState()
	now := time.Unix(200, 0)
	state = updateBuddyState(state, timelineEntry{kind: timelineAssistant, text: "BUDDY_PET\nstatus=pet\nhatched=true\nmuted=false\npet_count=1"}, now)
	if !state.petBurstUntil.Equal(now.Add(2500 * time.Millisecond)) {
		t.Fatalf("expected pet burst window, got %v", state.petBurstUntil)
	}
	if state.bubble != "purr purr" {
		t.Fatalf("expected pet reaction, got %q", state.bubble)
	}

	state = updateBuddyState(state, timelineEntry{kind: timelineAssistant, text: "BUDDY_MUTE\nmuted=true"}, now.Add(1*time.Second))
	if state.bubble != "" {
		t.Fatalf("expected mute to clear reaction, got %q", state.bubble)
	}
}

func TestBuddyReactionFromTimelineAssistantEvents(t *testing.T) {
	state := newBuddyState()
	if got, ok := buddyReactionFromTimeline(state.companion, timelineEntry{kind: timelineAssistant, text: "BUDDY_HATCH\nstatus=hatched\nhatched=true"}); !ok || got != "hi hi" {
		t.Fatalf("expected hatch reaction, got %q ok=%t", got, ok)
	}
	if got, ok := buddyReactionFromTimeline(state.companion, timelineEntry{kind: timelineAssistant, text: "BUDDY_UNMUTE\nmuted=false"}); !ok || got == "" {
		t.Fatalf("expected unmute reaction, got %q ok=%t", got, ok)
	}
}
