package agent

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestRetryPlanIncludesPhaseTransition(t *testing.T) {
	plan := retryPlan("provider_error", 1, 2, types.AgentTurnPhaseProviderRequest, types.AgentTurnPhaseRetry, false)
	if plan.Kind != "provider_error" || plan.FromPhase != types.AgentTurnPhaseProviderRequest || plan.ToPhase != types.AgentTurnPhaseRetry {
		t.Fatalf("unexpected retry plan: %+v", plan)
	}
}
