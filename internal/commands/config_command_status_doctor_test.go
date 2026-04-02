package commands

import (
	"context"
	"strings"
	"testing"
)

func TestConfigStatusAndDoctor(t *testing.T) {
	cmd := NewConfigCommand()
	state := &RuntimeState{ConfigValues: map[string]string{"settings.output-style": "human"}}
	statusRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "config", Args: []string{"status"}})
	if err != nil {
		t.Fatalf("config status failed: %v", err)
	}
	if !strings.Contains(statusRes.Message, "CONFIG_STATUS") {
		t.Fatalf("unexpected config status: %q", statusRes.Message)
	}
	doctorRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "config", Args: []string{"doctor"}})
	if err != nil {
		t.Fatalf("config doctor failed: %v", err)
	}
	if !strings.Contains(doctorRes.Message, "CONFIG_DOCTOR") || !strings.Contains(doctorRes.Message, "quick_fix=") {
		t.Fatalf("unexpected config doctor: %q", doctorRes.Message)
	}
}
