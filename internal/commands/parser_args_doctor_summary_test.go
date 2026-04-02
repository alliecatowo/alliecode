package commands

import "testing"

func TestParseDoctorSummaryArgument(t *testing.T) {
	inv, err := Parse(`/doctor summary`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "doctor" || len(inv.Args) != 1 || inv.Args[0] != "summary" {
		t.Fatalf("unexpected invocation: %#v", inv)
	}
}
