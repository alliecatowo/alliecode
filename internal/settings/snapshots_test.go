package settings

import (
	"testing"
	"time"

	"github.com/alliecatowo/alliecode/internal/config"
)

func TestTimelineAddAndQuery(t *testing.T) {
	tl := NewTimeline()
	cfgA := config.NewDefaultConfig()
	cfgA.DefaultProvider = "openai"
	cfgA.DefaultModel = "gpt-4o-mini"
	cfgB := config.NewDefaultConfig()
	cfgB.DefaultProvider = "anthropic"
	cfgB.DefaultModel = "claude"

	tl.Add(SnapshotFromConfig(ScopeProject, "/repo/a", cfgA, time.Unix(1, 0).UTC()))
	tl.Add(SnapshotFromConfig(ScopeProject, "/repo/b", cfgB, time.Unix(2, 0).UTC()))

	out := tl.Query(SnapshotQuery{Provider: "anthropic", ProjectContains: "/repo/b", Limit: 5})
	if len(out) != 1 || out[0].ProjectDir != "/repo/b" {
		t.Fatalf("unexpected query result: %+v", out)
	}
}
