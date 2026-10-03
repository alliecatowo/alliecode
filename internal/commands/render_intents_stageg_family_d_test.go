package commands

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStageGFamilyDEmitIntents(t *testing.T) {
	res := dispatchStageG(t, "/tasks list")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/advisor status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/btw hello")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/chrome status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/color")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/desktop status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/mobile status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/fast status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/effort status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)
}
