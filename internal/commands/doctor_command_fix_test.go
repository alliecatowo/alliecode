package commands

import (
	"context"
	"strings"
	"testing"
)

func TestDoctorCommandFixPlan(t *testing.T) {
	cmd := NewDoctorCommand()
	state := &RuntimeState{}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "doctor", Args: []string{"fix"}})
	if err != nil {
		t.Fatalf("doctor fix failed: %v", err)
	}
	if !strings.Contains(res.Message, "DOCTOR_FIX") || !strings.Contains(res.Message, "fix.1=") {
		t.Fatalf("unexpected doctor fix message: %q", res.Message)
	}
}
