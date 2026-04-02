package commands

import (
	"context"
	"testing"
)

func TestAliasSigninAndSignoutDispatchAuthCommands(t *testing.T) {
	r := DefaultRegistry()
	state := &RuntimeState{}
	if _, err := r.Dispatch(context.Background(), Context{State: state}, "/signin provider openai"); err != nil {
		t.Fatalf("signin alias failed: %v", err)
	}
	if _, err := r.Dispatch(context.Background(), Context{State: state}, "/signout"); err != nil {
		t.Fatalf("signout alias failed: %v", err)
	}
}
