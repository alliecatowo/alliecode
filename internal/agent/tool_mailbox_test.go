package agent

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestToolEventMailboxDrainInOrder(t *testing.T) {
	mailbox := newToolEventMailbox(2)
	mailbox.Enqueue(1, toolExecutionEnvelope{message: types.NewToolResultMessage("tool-2", "second", false)})
	mailbox.Enqueue(0, toolExecutionEnvelope{message: types.NewToolResultMessage("tool-1", "first", false)})

	envelopes, err := mailbox.DrainInOrder()
	if err != nil {
		t.Fatalf("DrainInOrder() error = %v", err)
	}
	if got := envelopes[0].message.Content[0].ForToolUseID; got != "tool-1" {
		t.Fatalf("first envelope tool id = %q, want %q", got, "tool-1")
	}
	if got := envelopes[1].message.Content[0].ForToolUseID; got != "tool-2" {
		t.Fatalf("second envelope tool id = %q, want %q", got, "tool-2")
	}
}

func TestToolEventMailboxRejectsDuplicateEnqueue(t *testing.T) {
	mailbox := newToolEventMailbox(1)
	mailbox.Enqueue(0, toolExecutionEnvelope{message: types.NewToolResultMessage("tool-1", "first", false)})
	mailbox.Enqueue(0, toolExecutionEnvelope{message: types.NewToolResultMessage("tool-1", "second", false)})

	_, err := mailbox.DrainInOrder()
	if err == nil {
		t.Fatalf("expected duplicate enqueue error")
	}
}
