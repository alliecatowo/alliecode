package settings

import (
	"path/filepath"
	"testing"
	"time"
)

func TestManagerRuntimeHydrationSummaries(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)

	projectPath := filepath.Join(project, ".alliecode", "config.yaml")
	writeConfig(t, projectPath, "default_provider: openai\ndefault_model: gpt-4o-mini\nstartup:\n  hydration_mode: compat\n")

	manager := NewManager(time.Minute)
	scope := ScopeProject
	if _, _, err := manager.Load(LoadOptions{ProjectDir: project, Scope: &scope}); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	requireValidation := true
	out, err := manager.RuntimeHydrationSummaries(project, RuntimeHydrationQuery{Provider: "openai", HydrationMode: "compat", RequireValidation: &requireValidation, Limit: 5})
	if err != nil {
		t.Fatalf("RuntimeHydrationSummaries() error = %v", err)
	}
	if len(out) == 0 || out[0].HydrationMode != "compat" || !out[0].ValidationPassed {
		t.Fatalf("unexpected runtime hydration summaries: %+v", out)
	}
}
