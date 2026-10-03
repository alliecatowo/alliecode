package commands

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStageGFamilyGEmitIntents(t *testing.T) {
	res := dispatchStageG(t, "/summary status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/reset-limits status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/oauth-refresh status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/ant-trace status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/autofix-pr status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/backfill-sessions status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/break-cache status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/bughunter status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/ctx-viz status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/debug-tool-call status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)
}
