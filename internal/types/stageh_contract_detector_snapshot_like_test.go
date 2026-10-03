package types

import "testing"

func TestStageHContractDetectorSnapshotLikePayload(t *testing.T) {
	text := "DOCTOR_REPORT\nstatus=warn\nsection.1=auth"
	if !LooksLikeStructuredContract(text) {
		t.Fatalf("expected report payload to be classified as contract")
	}
}
