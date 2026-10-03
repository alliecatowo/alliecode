package commands

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStageGFamilyCEmitIntents(t *testing.T) {
	res := dispatchStageG(t, "/release-notes status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/install-github-app status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/install-slack-app status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/feedback status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/hooks status")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)

	res = dispatchStageG(t, "/sandbox check")
	assertIntentKind(t, res, types.RenderIntentDetailRows)
	assertNoContractIntent(t, res)
}
