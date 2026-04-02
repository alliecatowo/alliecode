package hooks

import (
	"testing"
	"time"
)

func TestNewManagerAppliesMinimumTimeout(t *testing.T) {
	m := NewManager([]HookConfig{{Event: string(EventPreChat), Command: "true", Timeout: 1}})
	if len(m.hooks) != 1 {
		t.Fatalf("expected one hook")
	}
	if m.hooks[0].Timeout < minHookTimeout {
		t.Fatalf("timeout = %s, want >= %s", m.hooks[0].Timeout, minHookTimeout)
	}
}

func TestNewManagerStoresPayloadLimitAndRetryFlag(t *testing.T) {
	m := NewManager([]HookConfig{{Event: string(EventRetry), Command: "true", MaxPayloadSize: 1234, RetryOnce: true}})
	if len(m.hooks) != 1 {
		t.Fatalf("expected one hook")
	}
	if m.hooks[0].MaxPayloadSize != 1234 || !m.hooks[0].RetryOnce {
		t.Fatalf("unexpected hook config: %+v", m.hooks[0])
	}
	if m.hooks[0].Timeout <= 0 || m.hooks[0].Timeout > 5*time.Minute {
		t.Fatalf("unexpected timeout: %s", m.hooks[0].Timeout)
	}
}
