package settings

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alliecatowo/alliecode/internal/config"
	"github.com/alliecatowo/alliecode/internal/state"
)

func TestManagerCacheHitAndMiss(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)

	projectPath := filepath.Join(project, ".alliecode", "config.yaml")
	writeConfig(t, projectPath, "default_provider: openai\ndefault_model: gpt-4o-mini\n")

	manager := NewManager(time.Minute)
	scope := ScopeProject

	first, firstHit, err := manager.Load(LoadOptions{ProjectDir: project, Scope: &scope})
	if err != nil {
		t.Fatalf("first load failed: %v", err)
	}
	if firstHit {
		t.Fatalf("expected first load to miss cache")
	}

	second, secondHit, err := manager.Load(LoadOptions{ProjectDir: project, Scope: &scope})
	if err != nil {
		t.Fatalf("second load failed: %v", err)
	}
	if !secondHit {
		t.Fatalf("expected second load to hit cache")
	}
	if first != second {
		t.Fatalf("expected cached config pointer to be reused")
	}
}

func TestManagerInvalidationAndFreshness(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)

	projectPath := filepath.Join(project, ".alliecode", "config.yaml")
	writeConfig(t, projectPath, "default_provider: openai\ndefault_model: gpt-4o-mini\n")

	manager := NewManager(time.Minute)
	scope := ScopeProject
	opts := LoadOptions{ProjectDir: project, Scope: &scope}

	if _, _, err := manager.Load(opts); err != nil {
		t.Fatalf("initial load failed: %v", err)
	}

	fresh, err := manager.IsFresh(opts)
	if err != nil {
		t.Fatalf("freshness check failed: %v", err)
	}
	if !fresh {
		t.Fatalf("expected entry to be fresh right after load")
	}

	if err := manager.Invalidate(opts); err != nil {
		t.Fatalf("invalidate failed: %v", err)
	}
	meta, err := manager.ReadCacheMetadata(project)
	if err != nil {
		t.Fatalf("ReadCacheMetadata() after invalidate error = %v", err)
	}
	if meta.Fresh {
		t.Fatalf("expected metadata freshness false after invalidate")
	}
	if meta.LastInvalidateCause != string(InvalidateReasonManual) {
		t.Fatalf("LastInvalidateCause = %q, want %q", meta.LastInvalidateCause, InvalidateReasonManual)
	}

	if _, hit, err := manager.Load(opts); err != nil {
		t.Fatalf("load after invalidation failed: %v", err)
	} else if hit {
		t.Fatalf("expected cache miss after explicit invalidation")
	}

	if _, _, err := manager.Load(opts); err != nil {
		t.Fatalf("reload before file mutation failed: %v", err)
	}
	writeConfig(t, projectPath, "default_provider: anthropic\ndefault_model: claude\n")

	fresh, err = manager.IsFresh(opts)
	if err != nil {
		t.Fatalf("freshness check after file mutation failed: %v", err)
	}
	if fresh {
		t.Fatalf("expected stale cache after config file mutation")
	}
}

func TestAggregateValidationDiagnostics(t *testing.T) {
	invalid := config.NewDefaultConfig()
	invalid.DefaultProvider = ""

	diags := AggregateValidationDiagnostics(map[Scope]*config.Config{
		ScopeGlobal:  invalid,
		ScopeProject: nil,
	})

	if len(diags) == 0 {
		t.Fatalf("expected validation diagnostics")
	}

	if !hasValidationDiagnostic(diags, ScopeProject, "config_nil", "configuration is nil") {
		t.Fatalf("expected nil-config diagnostic for project scope")
	}
	if !hasValidationDiagnosticContains(diags, ScopeGlobal, "config_invalid", "default_provider must not be empty") {
		t.Fatalf("expected validation error detail for global scope")
	}
}

func TestManagerCachePathAndMetadataRoundTrip(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)

	manager := NewManager(time.Minute)
	cachePath, err := manager.CacheFilePath(project)
	if err != nil {
		t.Fatalf("CacheFilePath() error = %v", err)
	}
	if filepath.Base(cachePath) != "settings-cache.json" {
		t.Fatalf("unexpected cache filename: %q", cachePath)
	}

	meta := state.SettingsCacheMetadata{LastScope: "project", CacheHitCount: 2}
	if err := manager.WriteCacheMetadata(project, meta); err != nil {
		t.Fatalf("WriteCacheMetadata() error = %v", err)
	}
	got, err := manager.ReadCacheMetadata(project)
	if err != nil {
		t.Fatalf("ReadCacheMetadata() error = %v", err)
	}
	if got.LastScope != "project" || got.CacheHitCount != 2 {
		t.Fatalf("metadata mismatch: %+v", got)
	}
}

func TestManagerLoadUpdatesCacheMetadata(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)

	projectPath := filepath.Join(project, ".alliecode", "config.yaml")
	writeConfig(t, projectPath, "default_provider: openai\ndefault_model: gpt-4o-mini\n")

	manager := NewManager(time.Minute)
	scope := ScopeProject
	opts := LoadOptions{ProjectDir: project, Scope: &scope}

	if _, hit, err := manager.Load(opts); err != nil {
		t.Fatalf("first load failed: %v", err)
	} else if hit {
		t.Fatalf("expected first load miss")
	}
	if _, hit, err := manager.Load(opts); err != nil {
		t.Fatalf("second load failed: %v", err)
	} else if !hit {
		t.Fatalf("expected second load hit")
	}

	meta, err := manager.ReadCacheMetadata(project)
	if err != nil {
		t.Fatalf("ReadCacheMetadata() error = %v", err)
	}
	if meta.CacheMissCount < 1 {
		t.Fatalf("expected at least one cache miss, got %+v", meta)
	}
	if meta.CacheHitCount < 1 {
		t.Fatalf("expected at least one cache hit, got %+v", meta)
	}
	if meta.LastLoadUnixNano == 0 || meta.LastHitUnixNano == 0 {
		t.Fatalf("expected load/hit timestamps to be populated, got %+v", meta)
	}
	if meta.LastScope != "project" {
		t.Fatalf("LastScope = %q, want project", meta.LastScope)
	}
	if !meta.Fresh {
		t.Fatalf("expected freshness true after successful loads, got %+v", meta)
	}
	if meta.LastProjectDir != filepath.Clean(project) {
		t.Fatalf("LastProjectDir = %q, want %q", meta.LastProjectDir, filepath.Clean(project))
	}
}

func TestManagerInvalidateWithReasonPersistsCause(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)

	projectPath := filepath.Join(project, ".alliecode", "config.yaml")
	writeConfig(t, projectPath, "default_provider: openai\ndefault_model: gpt-4o-mini\n")

	manager := NewManager(time.Minute)
	scope := ScopeProject
	opts := LoadOptions{ProjectDir: project, Scope: &scope}
	if _, _, err := manager.Load(opts); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if err := manager.InvalidateWithReason(opts, InvalidateReasonProjectSwitch); err != nil {
		t.Fatalf("InvalidateWithReason() error = %v", err)
	}

	meta, err := manager.ReadCacheMetadata(project)
	if err != nil {
		t.Fatalf("ReadCacheMetadata() error = %v", err)
	}
	if meta.LastInvalidateCause != string(InvalidateReasonProjectSwitch) {
		t.Fatalf("LastInvalidateCause = %q, want %q", meta.LastInvalidateCause, InvalidateReasonProjectSwitch)
	}
	if meta.LastInvalidateAt == 0 {
		t.Fatalf("expected LastInvalidateAt to be set")
	}
}

func TestResolveScopePaths(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)

	paths, err := ResolveScopePaths(LoadOptions{ProjectDir: project})
	if err != nil {
		t.Fatalf("ResolveScopePaths() error = %v", err)
	}
	if !paths.Layered || paths.Effective == "" || paths.GlobalPath == "" || paths.ProjectPath == "" {
		t.Fatalf("unexpected scope paths: %+v", paths)
	}
}

func TestManagerHydrateForRuntime(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)

	projectPath := filepath.Join(project, ".alliecode", "config.yaml")
	writeConfig(t, projectPath, "default_provider: openai\ndefault_model: gpt-4o-mini\n")

	manager := NewManager(time.Minute)
	cfg, _, err := manager.HydrateForRuntime(LoadOptions{ProjectDir: project}, "cli", "repl")
	if err != nil {
		t.Fatalf("HydrateForRuntime() error = %v", err)
	}
	if cfg.Runtime.Surface != "cli" || cfg.Runtime.CommandSurface != "repl" || cfg.Startup.HydrationMode == "" {
		t.Fatalf("unexpected hydrated runtime config: %+v", cfg)
	}
}

func TestManagerMetadataContainsFingerprintAndPaths(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("ALLIECODE_DEFAULT_PROVIDER", "openai")

	projectPath := filepath.Join(project, ".alliecode", "config.yaml")
	writeConfig(t, projectPath, "default_provider: openai\ndefault_model: gpt-4o-mini\n")

	manager := NewManager(time.Minute)
	scope := ScopeProject
	if _, _, err := manager.Load(LoadOptions{ProjectDir: project, Scope: &scope}); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	meta, err := manager.ReadCacheMetadata(project)
	if err != nil {
		t.Fatalf("ReadCacheMetadata() error = %v", err)
	}
	if meta.LastKnownEnvFingerprint == "" || meta.LastKnownProjectPath == "" || meta.LastKnownGlobalPath == "" {
		t.Fatalf("expected fingerprint and paths in metadata, got %+v", meta)
	}
}

func TestManagerSnapshotTimeline(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)

	projectPath := filepath.Join(project, ".alliecode", "config.yaml")
	writeConfig(t, projectPath, "default_provider: openai\ndefault_model: gpt-4o-mini\n")

	manager := NewManager(time.Minute)
	scope := ScopeProject
	if _, _, err := manager.Load(LoadOptions{ProjectDir: project, Scope: &scope}); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	out := manager.SnapshotTimeline(SnapshotQuery{Provider: "openai", ProjectContains: project, Limit: 5})
	if len(out) == 0 {
		t.Fatalf("expected timeline snapshots")
	}
}

func hasValidationDiagnostic(diags []ValidationDiagnostic, scope Scope, code, message string) bool {
	for _, diag := range diags {
		if diag.Scope == scope && diag.Code == code && diag.Message == message {
			return true
		}
	}
	return false
}

func hasValidationDiagnosticContains(diags []ValidationDiagnostic, scope Scope, code, messagePart string) bool {
	for _, diag := range diags {
		if diag.Scope == scope && diag.Code == code && strings.Contains(diag.Message, messagePart) {
			return true
		}
	}
	return false
}

func writeConfig(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}
}
