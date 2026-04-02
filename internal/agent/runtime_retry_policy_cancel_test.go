package agent

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestEvaluateContinuationRecoveryCanceled(t *testing.T) {
	action := evaluateContinuationRecovery(types.StopCanceled, types.NewTextMessage(types.RoleAssistant, "partial output"), 1, 2)
	if !action.Canceled || action.ShouldRetry {
		t.Fatalf("unexpected canceled recovery action: %+v", action)
	}
}
