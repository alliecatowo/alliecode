package settings

import (
	"testing"
	"time"

	"github.com/alliecatowo/alliecode/internal/config"
)

func TestTimelineQueryExtendedFields(t *testing.T) {
	tl := NewTimeline()
	cfg := config.NewDefaultConfig()
	cfg.MigrationVersion = 3
	cfg.ProviderAllow = []string{"openai"}
	tl.Add(SnapshotFromConfig(ScopeProject, "/repo/a", cfg, time.Unix(100, 0).UTC()))
	out := tl.Query(SnapshotQuery{ProjectContains: "/repo", MinMigration: 2, LoadedAfterUnix: 90, RequireAllowList: true, Limit: 5})
	if len(out) != 1 || out[0].ProjectDir != "/repo/a" {
		t.Fatalf("unexpected timeline query output: %+v", out)
	}
	diag := LayerDiagnosticsFromSnapshot(out[0])
	if diag.HydrationMode != out[0].HydrationMode {
		t.Fatalf("unexpected layer diagnostics from snapshot: %+v", diag)
	}
}
