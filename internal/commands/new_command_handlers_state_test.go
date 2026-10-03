package commands

import (
	"context"
	"strings"
	"testing"
)

func TestBridgeCommandStateTransitions(t *testing.T) {
	r := DefaultRegistry()
	state := &RuntimeState{}
	ctx := Context{State: state}
	if _, err := r.Dispatch(context.Background(), ctx, "/bridge on"); err != nil {
		t.Fatalf("bridge on failed: %v", err)
	}
	if !state.BridgeEnabled || state.BridgeTransitions == 0 {
		t.Fatalf("expected bridge enabled transition, got %+v", state)
	}
	if _, err := r.Dispatch(context.Background(), ctx, "/bridge off"); err != nil {
		t.Fatalf("bridge off failed: %v", err)
	}
	if state.BridgeEnabled {
		t.Fatalf("expected bridge to be disabled")
	}
}

func TestBackfillSessionsAndCacheBreakMutateState(t *testing.T) {
	r := DefaultRegistry()
	state := &RuntimeState{}
	ctx := Context{State: state}
	if _, err := r.Dispatch(context.Background(), ctx, "/backfill-sessions run 7"); err != nil {
		t.Fatalf("backfill failed: %v", err)
	}
	if state.BackfillRuns != 1 || state.BackfillLastCount != 7 || state.BackfillLastAction != "run" {
		t.Fatalf("unexpected backfill state: %+v", state)
	}
	if _, err := r.Dispatch(context.Background(), ctx, "/break-cache models"); err != nil {
		t.Fatalf("break-cache failed: %v", err)
	}
	if state.CacheBreakCount != 1 || state.CacheBreakLastScope != "models" || state.CacheBreakScopes["models"] != 1 {
		t.Fatalf("unexpected break-cache state: %+v", state)
	}
}

func TestAutofixPerfAndUltraplanMutateState(t *testing.T) {
	r := DefaultRegistry()
	state := &RuntimeState{}
	ctx := Context{State: state}
	if _, err := r.Dispatch(context.Background(), ctx, "/autofix-pr plan pr-9"); err != nil {
		t.Fatalf("autofix-pr plan failed: %v", err)
	}
	if state.AutofixPRCount != 1 || state.AutofixPRLastAction != "plan" || state.AutofixPRLastRef != "pr-9" {
		t.Fatalf("unexpected autofix state: %+v", state)
	}
	if _, err := r.Dispatch(context.Background(), ctx, "/perf-issue open cpu spike"); err != nil {
		t.Fatalf("perf-issue open failed: %v", err)
	}
	if state.PerfIssueCount != 1 || state.PerfIssueLastTitle != "cpu spike" {
		t.Fatalf("unexpected perf issue state: %+v", state)
	}
	if _, err := r.Dispatch(context.Background(), ctx, "/ultraplan run release"); err != nil {
		t.Fatalf("ultraplan run failed: %v", err)
	}
	if state.UltraplanCount != 1 || state.UltraplanLastAction != "run" || state.UltraplanLastTarget != "release" {
		t.Fatalf("unexpected ultraplan state: %+v", state)
	}
}

func TestRemoteSetupAndSandboxToggleOutputs(t *testing.T) {
	r := DefaultRegistry()
	state := &RuntimeState{}
	ctx := Context{State: state}
	res, err := r.Dispatch(context.Background(), ctx, "/remote-setup connect")
	if err != nil {
		t.Fatalf("remote-setup connect failed: %v", err)
	}
	if !strings.Contains(res.Message, "REMOTE_SETUP_APPLY") || !state.WebSetupConnected {
		t.Fatalf("unexpected remote-setup result/state: %q %+v", res.Message, state)
	}
	res, err = r.Dispatch(context.Background(), ctx, "/sandbox-toggle toggle")
	if err != nil {
		t.Fatalf("sandbox-toggle toggle failed: %v", err)
	}
	if !strings.Contains(res.Message, "SANDBOX_TOGGLE_SET") {
		t.Fatalf("unexpected sandbox-toggle output: %q", res.Message)
	}
}
