package permissions

import "testing"

func TestClassifyAmbiguousParseDecisionEscapedOperator(t *testing.T) {
	classification := ClassifyCommandDetailed(`echo hi \\| cat`)
	decision, _, ok := ClassifyAmbiguousParseDecision(classification)
	if !ok || decision != "ask" {
		t.Fatalf("expected ambiguous escaped operator to ask, got decision=%q ok=%v", decision, ok)
	}
}
