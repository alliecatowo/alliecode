package state

import "testing"

func TestQueryOnboardingStates(t *testing.T) {
	items := []ProjectOnboardingState{
		{HasCompleted: true, LastStepKey: "claudemd", LastShownSessionID: "s1", LastEvaluationAtUnix: 10},
		{HasCompleted: false, LastStepKey: "workspace", LastShownSessionID: "s2", LastEvaluationAtUnix: 20},
	}
	out := QueryOnboardingStates(items, OnboardingQuery{CompletedOnly: true, StepContains: "claude", ShownSessionContains: "s1"})
	if len(out) != 1 || !out[0].HasCompleted {
		t.Fatalf("unexpected query result: %+v", out)
	}
}
