package fixtures

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/session"
	"github.com/alliecatowo/alliecode/internal/types"
)

type SessionFixture struct {
	RootDir   string
	SessionID string
	Store     *session.Store
}

func NewSessionFixture(t *testing.T, sessionID string) SessionFixture {
	t.Helper()
	root := t.TempDir()
	store, err := session.Create(root, sessionID)
	if err != nil {
		t.Fatalf("create session store: %v", err)
	}
	return SessionFixture{RootDir: root, SessionID: sessionID, Store: store}
}

func SeedMessages(t *testing.T, store *session.Store, messages ...types.Message) {
	t.Helper()
	for _, msg := range messages {
		if err := store.AppendMessage(msg); err != nil {
			t.Fatalf("append message: %v", err)
		}
	}
}
