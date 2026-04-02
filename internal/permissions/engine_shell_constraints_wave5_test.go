package permissions

import (
	"encoding/json"
	"testing"
)

func TestEngineChecksBashDynamicRedirectionConstraints(t *testing.T) {
	engine := NewEngine(ModeAuto, nil)
	in, _ := json.Marshal(map[string]any{"command": "echo hi > $TARGET", "workdir": "/tmp"})
	out := engine.CheckDetailed("bash", in)
	if out.Decision != DecisionAsk {
		t.Fatalf("expected ask for dynamic redirection, got %+v", out)
	}
}
