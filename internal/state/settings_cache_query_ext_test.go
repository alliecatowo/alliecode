package state

import "testing"

func TestQuerySettingsCacheMetadataMinFreshHitStreak(t *testing.T) {
	out := QuerySettingsCacheMetadata([]SettingsCacheMetadata{
		{LastProjectDir: "/repo/a", ConsecutiveFreshHitCount: 3},
		{LastProjectDir: "/repo/b", ConsecutiveFreshHitCount: 1},
	}, SettingsCacheQuery{MinFreshHitStreak: 2})
	if len(out) != 1 || out[0].LastProjectDir != "/repo/a" {
		t.Fatalf("unexpected filtered cache metadata: %+v", out)
	}
}
