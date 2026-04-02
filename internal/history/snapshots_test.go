package history

import (
	"testing"
	"time"
)

func TestBuildSnapshots(t *testing.T) {
	events := []Event{
		{Type: EventCustom, SessionID: "s1", Project: "/repo/a", Display: "deploy", Provider: "openai", Timestamp: time.Unix(20, 0).UTC()},
		{Type: EventCustom, SessionID: "s2", Project: "/repo/b", Display: "test", Provider: "anthropic", Timestamp: time.Unix(10, 0).UTC()},
	}
	out := BuildSnapshots(events, SnapshotQuery{ProjectContains: "/repo/a", Provider: "openai", Limit: 2})
	if len(out) != 1 || out[0].SessionID != "s1" {
		t.Fatalf("unexpected snapshots output: %+v", out)
	}
}
