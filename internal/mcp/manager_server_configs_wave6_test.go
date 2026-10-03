package mcp

import "testing"

func TestManagerServerConfigsDeterministicWave6(t *testing.T) {
	m := &Manager{serverConfigs: map[string]ServerConfig{"b": {Name: "b"}, "a": {Name: "a"}}}
	out := m.ServerConfigs()
	if len(out) != 2 || out[0].Name != "a" || out[1].Name != "b" {
		t.Fatalf("unexpected order: %+v", out)
	}
}
