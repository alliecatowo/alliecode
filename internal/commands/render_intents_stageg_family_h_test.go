package commands

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStageGFamilyHAuthIntents(t *testing.T) {
	res := dispatchStageG(t, "/login status")
	assertIntentKind(t, res, types.RenderIntentSummaryCard)

	res = dispatchStageG(t, "/logout status")
	assertIntentKind(t, res, types.RenderIntentSummaryCard)
}
