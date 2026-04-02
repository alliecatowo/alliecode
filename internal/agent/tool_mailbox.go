package agent

import (
	"fmt"
	"sync"

	"github.com/alliecatowo/alliecode/internal/types"
)

type toolExecutionEnvelope struct {
	events  []types.AgentEvent
	message types.Message
}

type toolEventMailbox struct {
	mu       sync.Mutex
	entries  []toolExecutionEnvelope
	received []bool
	err      error
}

func newToolEventMailbox(total int) *toolEventMailbox {
	return &toolEventMailbox{
		entries:  make([]toolExecutionEnvelope, total),
		received: make([]bool, total),
	}
}

func (m *toolEventMailbox) Enqueue(index int, envelope toolExecutionEnvelope) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return
	}
	if index < 0 || index >= len(m.entries) {
		m.err = fmt.Errorf("tool mailbox index out of range: index=%d total=%d", index, len(m.entries))
		return
	}
	if m.received[index] {
		m.err = fmt.Errorf("tool mailbox duplicate enqueue for index %d", index)
		return
	}

	m.entries[index] = envelope
	m.received[index] = true

	for _, ok := range m.received {
		if !ok {
			return
		}
	}
}

func (m *toolEventMailbox) DrainInOrder() ([]toolExecutionEnvelope, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return nil, m.err
	}

	for i, ok := range m.received {
		if !ok {
			return nil, fmt.Errorf("tool mailbox missing envelope at index %d", i)
		}
	}

	out := make([]toolExecutionEnvelope, len(m.entries))
	copy(out, m.entries)
	return out, nil
}
