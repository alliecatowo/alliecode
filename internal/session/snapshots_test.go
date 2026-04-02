package session

import (
	"testing"
	"time"
)

func TestBuildSessionSnapshots(t *testing.T) {
	items := []Transcript{
		{SessionID: "s1", ProjectPath: "/repo/a", Provider: "openai", RuntimeSurface: "cli", MessageCount: 3, LastMessageRole: "assistant", LastEventAt: time.Unix(10, 0).UTC()},
		{SessionID: "s2", ProjectPath: "/repo/b", Provider: "anthropic", RuntimeSurface: "ide", MessageCount: 1, LastMessageRole: "user", LastEventAt: time.Unix(20, 0).UTC()},
	}
	out := BuildSnapshots(items, SnapshotQuery{ProjectContains: "/repo/a", Provider: "openai", LastRole: "assistant", MinMessages: 2, Limit: 10})
	if len(out) != 1 || out[0].SessionID != "s1" {
		t.Fatalf("unexpected snapshot output: %+v", out)
	}
}
