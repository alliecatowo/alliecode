package commands

import (
	"context"
	"testing"
)

func TestAliasAuthDispatchesLogin(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/auth status")
	if err != nil {
		t.Fatalf("auth alias failed: %v", err)
	}
	if got := res.Message; len(got) < len("LOGIN_STATUS") || got[:len("LOGIN_STATUS")] != "LOGIN_STATUS" {
		t.Fatalf("unexpected auth alias output: %q", got)
	}
}

func TestAliasMemDispatchesMemory(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/mem status")
	if err != nil {
		t.Fatalf("mem alias failed: %v", err)
	}
	if got := res.Message; len(got) < len("MEMORY_STATUS") || got[:len("MEMORY_STATUS")] != "MEMORY_STATUS" {
		t.Fatalf("unexpected mem alias output: %q", got)
	}
}

func TestAliasStatDispatchesStats(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/stat")
	if err != nil {
		t.Fatalf("stat alias failed: %v", err)
	}
	if got := res.Message; len(got) < len("STATS") || got[:len("STATS")] != "STATS" {
		t.Fatalf("unexpected stat alias output: %q", got)
	}
}

func TestAliasPrivacyDispatchesPrivacySettings(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/privacy status")
	if err != nil {
		t.Fatalf("privacy alias failed: %v", err)
	}
	if got := res.Message; len(got) < len("PRIVACY_SETTINGS") || got[:len("PRIVACY_SETTINGS")] != "PRIVACY_SETTINGS" {
		t.Fatalf("unexpected privacy alias output: %q", got)
	}
}
