package commands

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestLegacyOutputIntentsParsesContractPayload(t *testing.T) {
	msg := "VIM_STATUS\nenabled=true"
	intents := legacyOutputIntents(msg)
	if len(intents) != 1 {
		t.Fatalf("expected one detail intent, got %#v", intents)
	}
	intent := intents[0]
	if intent.Kind != types.RenderIntentDetailRows {
		t.Fatalf("expected detail rows intent kind, got %#v", intent.Kind)
	}
	if intent.Title != "Vim Status" {
		t.Fatalf("expected humanized title, got %q", intent.Title)
	}
	if len(intent.DetailRows) != 1 || intent.DetailRows[0].Label != "Enabled" || intent.DetailRows[0].Value != "true" {
		t.Fatalf("unexpected detail rows: %#v", intent.DetailRows)
	}
}
