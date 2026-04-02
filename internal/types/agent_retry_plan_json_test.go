package types

import (
	"encoding/json"
	"testing"
)

func TestAgentRetryPlanJSONIncludesPhaseFields(t *testing.T) {
	plan := AgentRetryPlan{Kind: "provider_error", Attempt: 1, Max: 2, FromPhase: AgentTurnPhaseProviderRequest, ToPhase: AgentTurnPhaseRetry}
	b, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("Marshal() err = %v", err)
	}
	s := string(b)
	if s == "{}" || !containsJSONKey(s, "from_phase") || !containsJSONKey(s, "to_phase") {
		t.Fatalf("unexpected json: %s", s)
	}
}

func containsJSONKey(s, key string) bool {
	return len(s) > 0 && jsonContains(s, "\""+key+"\"")
}

func jsonContains(s, needle string) bool {
	for i := 0; i+len(needle) <= len(s); i++ {
		if s[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
