package agent

import (
	"context"
	"sync"

	"github.com/alliecatowo/alliecode/internal/hooks"
)

type hookSpy struct {
	mu     sync.Mutex
	events []hooks.Event
}

func (h *hookSpy) Fire(_ context.Context, event hooks.Event, _ map[string]string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events = append(h.events, event)
	return nil
}

func (h *hookSpy) HasHooks(event hooks.Event) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, ev := range h.events {
		if ev == event {
			return true
		}
	}
	return true
}

func (h *hookSpy) seen(event hooks.Event) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, ev := range h.events {
		if ev == event {
			return true
		}
	}
	return false
}
