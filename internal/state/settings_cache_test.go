package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSettingsCacheMetadataReadWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings-cache.json")
	want := SettingsCacheMetadata{
		CacheFilePath:            "/tmp/.alliecode/settings-cache.json",
		LastProjectDir:           "/tmp/repo",
		LastScope:                "project",
		LastLoadUnixNano:         101,
		LastHitUnixNano:          102,
		CacheHitCount:            3,
		CacheMissCount:           1,
		LastInvalidateCause:      "source_changed",
		LastInvalidateAt:         103,
		Fresh:                    true,
		LastKnownGlobalPath:      "/home/u/.config/alliecode/config.yaml",
		LastKnownProjectPath:     "/tmp/repo/.alliecode/config.yaml",
		LastKnownEnvFingerprint:  "abc",
		LastKnownLayeringMode:    "layered",
		InvalidateCount:          2,
		ConsecutiveFreshHitCount: 3,
		ConsecutiveMissCount:     1,
		LastPersistedAtUnixNano:  104,
		LastError:                "none",
		LastErrorAtUnixNano:      105,
	}
	if err := WriteSettingsCacheMetadata(path, want); err != nil {
		t.Fatalf("WriteSettingsCacheMetadata() error = %v", err)
	}
	got, err := ReadSettingsCacheMetadata(path)
	if err != nil {
		t.Fatalf("ReadSettingsCacheMetadata() error = %v", err)
	}
	if got != want {
		t.Fatalf("metadata mismatch: got %+v want %+v", got, want)
	}
}

func TestSettingsCacheMetadataMissingDefaults(t *testing.T) {
	got, err := ReadSettingsCacheMetadata(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("ReadSettingsCacheMetadata() error = %v", err)
	}
	if got != (SettingsCacheMetadata{}) {
		t.Fatalf("default metadata mismatch: got %+v", got)
	}
}

func TestSettingsCacheMetadataRejectsInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings-cache.json")
	if err := os.WriteFile(path, []byte("{"), stateFilePerm); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	_, err := ReadSettingsCacheMetadata(path)
	if err == nil || !strings.Contains(err.Error(), "decode settings cache metadata") {
		t.Fatalf("expected decode error, got %v", err)
	}
}

func TestSettingsCacheMetadataNormalizesValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings-cache.json")
	if err := WriteSettingsCacheMetadata(path, SettingsCacheMetadata{
		CacheFilePath:       " /tmp/cache ",
		LastProjectDir:      " /tmp/repo ",
		LastScope:           " project ",
		LastLoadUnixNano:    -1,
		LastHitUnixNano:     -2,
		CacheHitCount:       -3,
		CacheMissCount:      -4,
		LastInvalidateCause: " source_changed ",
		LastInvalidateAt:    -5,
	}); err != nil {
		t.Fatalf("WriteSettingsCacheMetadata() error = %v", err)
	}
	got, err := ReadSettingsCacheMetadata(path)
	if err != nil {
		t.Fatalf("ReadSettingsCacheMetadata() error = %v", err)
	}
	if got.CacheFilePath != "/tmp/cache" || got.LastProjectDir != "/tmp/repo" || got.LastScope != "project" {
		t.Fatalf("trim mismatch: %+v", got)
	}
	if got.LastLoadUnixNano != 0 || got.LastHitUnixNano != 0 || got.CacheHitCount != 0 || got.CacheMissCount != 0 || got.LastInvalidateAt != 0 {
		t.Fatalf("expected clamped values, got %+v", got)
	}
	if got.LastInvalidateCause != "source_changed" {
		t.Fatalf("expected trimmed invalidate cause, got %+v", got)
	}
}
