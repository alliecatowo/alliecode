package commands

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStageGFamilyAEmitIntents(t *testing.T) {
	res := dispatchStageG(t, "/init")
	assertIntentKind(t, res, types.RenderIntentSummaryCard)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/review status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/rewind status")
	assertIntentKind(t, res, types.RenderIntentSummaryCard)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/tag status")
	assertIntentKind(t, res, types.RenderIntentSummaryCard)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/remote-env status")
	assertIntentKind(t, res, types.RenderIntentSummaryCard)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/security-review status")
	assertIntentKind(t, res, types.RenderIntentSummaryCard)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/add-dir .")
	assertIntentKind(t, res, types.RenderIntentSummaryCard)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/agents list")
	assertIntentKind(t, res, types.RenderIntentTable)
	assertNoContractIntent(t, res)
}
