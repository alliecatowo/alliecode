package types

import "testing"

func TestRenderIntentContractHasContentAndClone(t *testing.T) {
	intent := RenderIntent{Kind: RenderIntentContract, ContractTag: "VIM_STATUS", Contract: []RenderContractEntry{{Key: "enabled", Value: "true"}}}
	if !intent.HasContent() {
		t.Fatalf("expected contract intent to have content")
	}
	clone := intent.Clone()
	if clone.Kind != RenderIntentContract {
		t.Fatalf("expected clone kind %q, got %q", RenderIntentContract, clone.Kind)
	}
	if clone.ContractTag != "VIM_STATUS" {
		t.Fatalf("expected contract tag to clone, got %q", clone.ContractTag)
	}
	if len(clone.Contract) != 1 || clone.Contract[0].Key != "enabled" || clone.Contract[0].Value != "true" {
		t.Fatalf("expected contract entries to clone, got %#v", clone.Contract)
	}
	clone.Contract[0].Value = "false"
	if intent.Contract[0].Value != "true" {
		t.Fatalf("expected deep clone of contract entries")
	}
}
