package commands

import (
	"context"
	"strings"
	"testing"
)

func TestWaveKLane4ProviderModelPermutationStatusAlignment(t *testing.T) {
	r := DefaultRegistry()
	state := &RuntimeState{ProviderName: "openai", Model: "gpt-4o-mini", ModelRef: "openai/gpt-4o-mini"}
	ctx := Context{State: state}

	steps := []string{
		"/login provider openai",
		"/model gpt-4o",
		"/provider set anthropic",
		"/model sonnet",
		"/logout",
		"/login provider anthropic",
		"/model opus",
	}
	for _, step := range steps {
		if _, err := r.Dispatch(context.Background(), ctx, step); err != nil {
			t.Fatalf("dispatch %q failed: %v", step, err)
		}
	}

	statusRes, err := r.Dispatch(context.Background(), ctx, "/status")
	if err != nil {
		t.Fatalf("/status failed: %v", err)
	}
	statuslineRes, err := r.Dispatch(context.Background(), ctx, "/statusline status")
	if err != nil {
		t.Fatalf("/statusline status failed: %v", err)
	}
	for _, want := range []string{"provider=anthropic", "model=claude-opus-4-20250514", "provider_ready=true", "logged_in=true"} {
		if !strings.Contains(statusRes.Message, want) {
			t.Fatalf("status output missing %q: %q", want, statusRes.Message)
		}
		if !strings.Contains(statuslineRes.Message, want) {
			t.Fatalf("statusline output missing %q: %q", want, statuslineRes.Message)
		}
	}
}

func TestWaveKLane4DoctorOutputsContainActionableRecovery(t *testing.T) {
	r := DefaultRegistry()
	state := &RuntimeState{ProviderName: "anthropic", Model: "claude-sonnet-4-20250514", ModelRef: "anthropic/claude-sonnet-4-20250514", LoggedIn: false}
	ctx := Context{State: state}

	providerDoctor, err := r.Dispatch(context.Background(), ctx, "/provider doctor")
	if err != nil {
		t.Fatalf("/provider doctor failed: %v", err)
	}
	if !strings.Contains(providerDoctor.Message, "quick_fix=") {
		t.Fatalf("provider doctor missing quick fix: %q", providerDoctor.Message)
	}

	modelDoctor, err := r.Dispatch(context.Background(), ctx, "/model doctor")
	if err != nil {
		t.Fatalf("/model doctor failed: %v", err)
	}
	if !strings.Contains(modelDoctor.Message, "quick_fix=") {
		t.Fatalf("model doctor missing quick fix: %q", modelDoctor.Message)
	}
}
