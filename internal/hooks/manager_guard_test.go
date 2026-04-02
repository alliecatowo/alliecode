package hooks

import "testing"

func TestNewManagerSkipsInvalidHookConfigs(t *testing.T) {
	m := NewManager([]HookConfig{
		{Event: "", Command: "echo nope"},
		{Event: string(EventPreTool), Command: ""},
		{Event: string(EventPreTool), Command: "echo ok"},
		{Event: "not_real", Command: "echo nope"},
	})
	if !m.HasHooks(EventPreTool) {
		t.Fatalf("expected only valid pre_tool hook to remain")
	}
	if m.HasHooks(Event("not_real")) {
		t.Fatalf("did not expect unknown event hook to be registered")
	}
}
