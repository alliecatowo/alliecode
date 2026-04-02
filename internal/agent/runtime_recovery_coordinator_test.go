package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestRecoveryCoordinatorRecoverPromptTooLong(t *testing.T) {
	a := New(Config{Provider: summaryProvider{}, MaxTurns: 1})
	a.messages = append(a.messages,
		types.NewTextMessage(types.RoleUser, strings.Repeat("u1 ", 9000)),
		types.NewTextMessage(types.RoleAssistant, strings.Repeat("a1 ", 9000)),
		types.NewTextMessage(types.RoleUser, strings.Repeat("u2 ", 9000)),
		types.NewTextMessage(types.RoleAssistant, strings.Repeat("a2 ", 9000)),
		types.NewTextMessage(types.RoleUser, strings.Repeat("u3 ", 9000)),
		types.NewTextMessage(types.RoleAssistant, strings.Repeat("a3 ", 9000)),
	)

	recovered, err := newRecoveryCoordinator(a).RecoverPromptTooLong(context.Background())
	if err != nil {
		t.Fatalf("RecoverPromptTooLong() err = %v", err)
	}
	if !recovered {
		t.Fatalf("expected recovery success")
	}
}
