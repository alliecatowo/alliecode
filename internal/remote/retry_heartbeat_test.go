package remote

import (
	"testing"
	"time"
)

func TestBackoffMonotonicWithoutJitter(t *testing.T) {
	b := NewBackoff(RetryPolicy{
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     80 * time.Millisecond,
		Multiplier:   2,
		Jitter:       0,
	}, 1)

	got := []time.Duration{
		b.DelayForAttempt(0),
		b.DelayForAttempt(1),
		b.DelayForAttempt(2),
		b.DelayForAttempt(3),
		b.DelayForAttempt(4),
	}
	want := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		40 * time.Millisecond,
		80 * time.Millisecond,
		80 * time.Millisecond,
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("attempt %d delay = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestHeartbeatMonitorExpiry(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	h := NewHeartbeatMonitor(2*time.Second, start)

	h.Beat(start.Add(time.Second))
	if h.Expired(start.Add(2500 * time.Millisecond)) {
		t.Fatalf("Expired() true before timeout")
	}
	if !h.Expired(start.Add(4 * time.Second)) {
		t.Fatalf("Expired() false after timeout")
	}
}
