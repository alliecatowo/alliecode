package commands

import (
	"context"
	"strings"
	"testing"
)

func TestDoctorFixIncludesWave5LoopAreas(t *testing.T) {
	cmd := NewDoctorCommand()
	res, err := cmd.Execute(context.Background(), Context{State: &RuntimeState{}}, Invocation{Name: "doctor", Args: []string{"fix"}})
	if err != nil {
		t.Fatalf("doctor fix failed: %v", err)
	}
	for _, area := range []string{"history", "session", "mcp"} {
		if !strings.Contains(res.Message, "area="+area) {
			t.Fatalf("expected area %q in doctor fix output: %q", area, res.Message)
		}
	}
}
