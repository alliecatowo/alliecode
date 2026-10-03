package commands

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStageGFamilyEEmitIntents(t *testing.T) {
	res := dispatchStageG(t, "/reload-plugins")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/export")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/extra-usage status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/rate-limit-options status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/pr-comments status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/web-setup status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/bridge-kick status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/brief status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)
}
