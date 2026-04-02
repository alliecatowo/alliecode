package state

import "testing"

func TestBuildSettingsCacheSnapshots(t *testing.T) {
	items := []SettingsCacheMetadata{
		{LastProjectDir: "/repo/a", LastScope: "project", LastKnownLayeringMode: "layered", Fresh: true, InvalidateCount: 2, LastPersistedAtUnixNano: 20},
		{LastProjectDir: "/repo/b", LastScope: "global", LastKnownLayeringMode: "scoped", Fresh: false, InvalidateCount: 1, LastLoadUnixNano: 11, LastHitUnixNano: 12, LastKnownEnvFingerprint: "abc", LastError: "broken", LastPersistedAtUnixNano: 30},
	}
	out := BuildSettingsCacheSnapshots(items, SettingsCacheSnapshotQuery{Scope: "global", HasError: boolPtrCache(true), Limit: 10})
	if len(out) != 1 || out[0].ProjectDir != "/repo/b" || out[0].EnvFingerprint != "abc" {
		t.Fatalf("unexpected settings cache snapshots: %+v", out)
	}
}

func boolPtrCache(v bool) *bool { return &v }
