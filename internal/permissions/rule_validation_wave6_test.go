package permissions

import "testing"

func TestValidateRuleRejectsInvalidSourceWave6(t *testing.T) {
	err := validateRule(Rule{Source: RuleSource("invalid"), Tool: "bash", Decision: DecisionAsk})
	if err == nil {
		t.Fatalf("expected invalid source error")
	}
}
