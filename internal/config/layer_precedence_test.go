package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildLayerPrecedenceAndQuery(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)
	writeConfigFile(t, filepath.Join(home, ".config", "alliecode", "config.yaml"), "default_provider: openai\ndefault_model: gpt-4o-mini\n")
	writeConfigFile(t, filepath.Join(project, ".alliecode", "config.yaml"), "default_provider: anthropic\ndefault_model: claude\nstartup:\n  hydration_mode: strict\n")

	items, err := BuildLayerPrecedence(project)
	if err != nil {
		t.Fatalf("BuildLayerPrecedence() error = %v", err)
	}
	out := QueryLayerPrecedence(items, LayerPrecedenceQuery{LayerNameContains: "effective", Provider: "anthropic", HydrationMode: "strict", Limit: 1})
	if len(out) != 1 || out[0].LayerName != "effective" {
		t.Fatalf("unexpected layer precedence output: %+v", out)
	}
}

func writeConfigFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}
