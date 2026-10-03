package commands

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestRuntimeSelectionTruthPrefersHydratedRuntimeSnapshot(t *testing.T) {
	state := &RuntimeState{
		ProviderName: "openai",
		Model:        "gpt-4o-mini",
		ModelRef:     "openai/gpt-4o-mini",
		LoggedIn:     true,
		AuthProvider: "openai",
		Runtime:      types.AgentRuntimeSnapshot{ProviderName: "anthropic", Model: "claude-opus-4-20250514", ModelRef: "anthropic/claude-opus-4-20250514"},
	}

	selection := RuntimeSelectionTruth(state)
	if selection.ProviderName != "anthropic" || selection.ModelName != "claude-opus-4-20250514" || selection.ModelRef != "anthropic/claude-opus-4-20250514" {
		t.Fatalf("unexpected runtime selection: %+v", selection)
	}
	if selection.ProviderReady {
		t.Fatalf("expected readiness false when auth provider mismatches, got %+v", selection)
	}
}

func TestRuntimeSelectionNextActionReflectsReadiness(t *testing.T) {
	state := &RuntimeState{ProviderName: "anthropic", Model: "claude-opus-4-20250514", LoggedIn: false}
	if got := RuntimeSelectionNextAction(state); got != "/login provider anthropic" {
		t.Fatalf("RuntimeSelectionNextAction() = %q, want login hint", got)
	}
}

func TestRuntimeSelectionTruthParsesProviderFromModelRefWhenSplitFieldsDrift(t *testing.T) {
	state := &RuntimeState{
		ProviderName: "openai",
		Model:        "claude-opus-4-20250514",
		ModelRef:     "anthropic/claude-opus-4-20250514",
		LoggedIn:     true,
		AuthProvider: "anthropic",
	}

	selection := RuntimeSelectionTruth(state)
	if selection.ProviderName != "anthropic" || selection.ModelName != "claude-opus-4-20250514" || selection.ModelRef != "anthropic/claude-opus-4-20250514" {
		t.Fatalf("unexpected parsed selection: %+v", selection)
	}
	if !selection.ProviderReady {
		t.Fatalf("expected provider ready for matching auth provider: %+v", selection)
	}
}

func TestRuntimeSelectionTruthNormalizesRuntimeTupleWhenModelContainsProviderPrefix(t *testing.T) {
	state := &RuntimeState{
		ProviderName: "openai",
		Model:        "gpt-4o-mini",
		Runtime:      types.AgentRuntimeSnapshot{ProviderName: "openai", Model: "anthropic/claude-sonnet-4-20250514"},
	}

	selection := RuntimeSelectionTruth(state)
	if selection.ProviderName != "anthropic" || selection.ModelName != "claude-sonnet-4-20250514" || selection.ModelRef != "anthropic/claude-sonnet-4-20250514" {
		t.Fatalf("unexpected normalized runtime selection: %+v", selection)
	}
}

func TestHydrateRuntimeSelectionBackfillsCanonicalModelRefFromModelTuple(t *testing.T) {
	state := &RuntimeState{
		ProviderName: "openai",
		Model:        "anthropic/claude-opus-4-20250514",
		LoggedIn:     true,
		AuthProvider: "anthropic",
	}
	HydrateRuntimeSelection(state)

	if state.ProviderName != "anthropic" {
		t.Fatalf("ProviderName = %q, want anthropic", state.ProviderName)
	}
	if state.Model != "claude-opus-4-20250514" {
		t.Fatalf("Model = %q, want claude-opus-4-20250514", state.Model)
	}
	if state.ModelRef != "anthropic/claude-opus-4-20250514" {
		t.Fatalf("ModelRef = %q, want anthropic/claude-opus-4-20250514", state.ModelRef)
	}
	if state.Runtime.ModelRef != "anthropic/claude-opus-4-20250514" {
		t.Fatalf("Runtime.ModelRef = %q, want anthropic/claude-opus-4-20250514", state.Runtime.ModelRef)
	}
}
