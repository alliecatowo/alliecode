package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestCompactionCoordinatorAttemptForced(t *testing.T) {
	a := New(Config{Provider: summaryProvider{}, MaxTurns: 1})
	a.messages = append(a.messages,
		types.NewTextMessage(types.RoleUser, strings.Repeat("u1 ", 9000)),
		types.NewTextMessage(types.RoleAssistant, strings.Repeat("a1 ", 9000)),
		types.NewTextMessage(types.RoleUser, strings.Repeat("u2 ", 9000)),
		types.NewTextMessage(types.RoleAssistant, strings.Repeat("a2 ", 9000)),
		types.NewTextMessage(types.RoleUser, strings.Repeat("u3 ", 9000)),
		types.NewTextMessage(types.RoleAssistant, strings.Repeat("a3 ", 9000)),
	)

	applied, err := newCompactionCoordinator(a).Attempt(context.Background(), true)
	if err != nil {
		t.Fatalf("Attempt() err = %v", err)
	}
	if !applied {
		t.Fatalf("expected forced compaction to apply")
	}
}
