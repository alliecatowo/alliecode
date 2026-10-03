package commands

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStageGFamilyBEmitIntents(t *testing.T) {
	res := dispatchStageG(t, "/vim status")
	assertIntentKind(t, res, types.RenderIntentSummaryCard)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/voice status")
	assertIntentKind(t, res, types.RenderIntentSummaryCard)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/buddy help")
	assertIntentKind(t, res, types.RenderIntentOptionList)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/statusline status")
	assertIntentKind(t, res, types.RenderIntentSummaryCard)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/ide detect")
	assertIntentKind(t, res, types.RenderIntentSummaryCard)
	assertIntentKind(t, res, types.RenderIntentOptionList)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/keybindings status")
	assertIntentKind(t, res, types.RenderIntentSummaryCard)
	assertNoContractIntent(t, res)
}
