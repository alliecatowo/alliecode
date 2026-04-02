package types

import (
	"encoding/json"
	"testing"
)

func TestAgentReplayCursorIncludesLabel(t *testing.T) {
	c := AgentReplayCursor{Start: 1, End: 2, Size: 2, Label: "turn_end"}
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("Marshal() err = %v", err)
	}
	if string(b) == "{}" {
		t.Fatalf("unexpected empty json")
	}
}
