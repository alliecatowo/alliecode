package types

import "testing"

func TestStageHContractDetectorRequiresUpperHeader(t *testing.T) {
	if LooksLikeStructuredContract("status_report\nprovider=openai") {
		t.Fatalf("expected lowercase header to be treated as plain text")
	}
}
