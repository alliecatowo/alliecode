package state

import "testing"

func TestQuerySettingsCacheMetadata(t *testing.T) {
	fresh := true
	items := []SettingsCacheMetadata{
		{LastProjectDir: "/repo/a", LastScope: "project", Fresh: true, CacheHitCount: 2, ConsecutiveFreshHitCount: 2, InvalidateCount: 2, LastPersistedAtUnixNano: 10},
		{LastProjectDir: "/repo/b", LastScope: "project", Fresh: false, CacheHitCount: 0, LastPersistedAtUnixNano: 20},
	}
	out := QuerySettingsCacheMetadata(items, SettingsCacheQuery{ProjectContains: "/repo/a", Fresh: &fresh, MinHitCount: 1, MinInvalidateCount: 2, MinFreshHitStreak: 2})
	if len(out) != 1 || out[0].LastProjectDir != "/repo/a" {
		t.Fatalf("unexpected query result: %+v", out)
	}
}
