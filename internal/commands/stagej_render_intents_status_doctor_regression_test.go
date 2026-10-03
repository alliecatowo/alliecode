package commands

import (
	"context"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStageJStatusAndDoctorEmitTypedNonContractIntents(t *testing.T) {
	tests := []struct {
		input string
		want  types.RenderIntentKind
	}{
		{input: "/status", want: types.RenderIntentSummaryCard},
		{input: "/doctor fix", want: types.RenderIntentChecklist},
	}

	registry := DefaultRegistry()
	for _, tc := range tests {
		res, err := registry.Dispatch(context.Background(), Context{State: stageJRuntimeState()}, tc.input)
		if err != nil {
			t.Fatalf("dispatch %q failed: %v", tc.input, err)
		}
		if !containsStageJIntentKind(res.RenderIntents, tc.want) {
			t.Fatalf("expected %q in intents for %q, got %#v", tc.want, tc.input, res.RenderIntents)
		}
		for _, intent := range res.RenderIntents {
			if intent.Kind == types.RenderIntentContract {
				t.Fatalf("unexpected contract intent for %q: %#v", tc.input, res.RenderIntents)
			}
		}
	}
}

func containsStageJIntentKind(intents []types.RenderIntent, kind types.RenderIntentKind) bool {
	for _, intent := range intents {
		if intent.Kind == kind {
			return true
		}
	}
	return false
}
