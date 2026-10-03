package commands

import (
	"context"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func dispatchStageG(t *testing.T, input string) Result {
	t.Helper()
	registry := DefaultRegistry()
	res, err := registry.Dispatch(context.Background(), Context{State: stageGState()}, input)
	if err != nil {
		t.Fatalf("dispatch %q failed: %v", input, err)
	}
	if len(res.RenderIntents) == 0 {
		t.Fatalf("dispatch %q returned no render intents", input)
	}
	for _, intent := range res.RenderIntents {
		if !intent.HasContent() {
			t.Fatalf("dispatch %q returned empty intent: %#v", input, intent)
		}
	}
	return res
}

func assertIntentKind(t *testing.T, res Result, kind types.RenderIntentKind) {
	t.Helper()
	for _, intent := range res.RenderIntents {
		if intent.Kind == kind {
			return
		}
	}
	t.Fatalf("missing intent kind %q in %#v", kind, res.RenderIntents)
}

func assertNoContractIntent(t *testing.T, res Result) {
	t.Helper()
	for _, intent := range res.RenderIntents {
		if intent.Kind == types.RenderIntentContract {
			t.Fatalf("unexpected contract intent in %#v", res.RenderIntents)
		}
	}
}
