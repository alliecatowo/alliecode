package tui

import "testing"

func TestBuddyMuteEventFalseForUserRows(t *testing.T) {
	if buddyMuteEvent(timelineEntry{kind: timelineUser, text: "BUDDY_MUTE"}) {
		t.Fatalf("expected mute event only for assistant rows")
	}
}
