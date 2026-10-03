package permissions

import "testing"

func TestEffectiveSandboxPolicySummaryWave6(t *testing.T) {
	s := EffectiveSandboxPolicySummary(SandboxSettings{Mode: "danger-full-access", WorkspaceLocked: true, ExcludedCommands: []string{"git push", "git push"}})
	if s.ModeAlias != "permissive" || !s.WorkspaceLocked || s.ExcludedCount != 1 {
		t.Fatalf("unexpected summary: %+v", s)
	}
}
