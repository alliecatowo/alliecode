package commands

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStageGFamilyFEmitIntents(t *testing.T) {
	res := dispatchStageG(t, "/commit suggest")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/commit-push-pr doctor")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/init-verifiers status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/insights status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/passes status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/rename status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/stickers status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/think-back status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/thinkback-play status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/teleport status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)
}
