package state

import "testing"

func TestBuildSessionMetadataIndex(t *testing.T) {
	registry := SessionMetadataRegistry{Sessions: map[string]SessionMetadata{
		"s1": {ProjectPath: "/repo/a", UpdatedAt: 20, MessageCount: 5, EventCount: 7, LastRole: "assistant", RuntimeSurface: "cli"},
		"s2": {ProjectPath: "/repo/b", UpdatedAt: 10, MessageCount: 1, EventCount: 2, LastRole: "user", RuntimeSurface: "ide"},
	}}
	out := BuildSessionMetadataIndex(registry, SessionMetadataIndexQuery{ProjectContains: "/repo/a", Role: "assistant", MinMessages: 2, Limit: 10})
	if len(out) != 1 || out[0].SessionID != "s1" || out[0].MessageEventRatio != "5/7" {
		t.Fatalf("unexpected session metadata index output: %+v", out)
	}
}
