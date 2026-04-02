package remote

import (
	"context"
	"testing"
	"time"
)

func TestInMemoryTransportSendReceive(t *testing.T) {
	tr := NewInMemoryTransport()
	ctx := context.Background()

	if err := tr.Connect(ctx, "sid-1"); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	base := time.Now().UTC()
	for i := int64(1); i <= 3; i++ {
		msg := RemoteMessage{SessionID: "sid-1", Kind: MessageKindProviderEvent, Name: "content_delta", Seq: i, SentAt: base}
		if err := tr.Send(ctx, msg); err != nil {
			t.Fatalf("Send(%d) error = %v", i, err)
		}
	}

	got, err := tr.Receive(ctx, "sid-1", PollConfig{BatchSize: 2, WaitTimeout: 5 * time.Millisecond})
	if err != nil {
		t.Fatalf("Receive(batch1) error = %v", err)
	}
	if len(got) != 2 || got[0].Seq != 1 || got[1].Seq != 2 {
		t.Fatalf("batch1 mismatch: %+v", got)
	}

	got, err = tr.Receive(ctx, "sid-1", PollConfig{BatchSize: 2, WaitTimeout: 5 * time.Millisecond})
	if err != nil {
		t.Fatalf("Receive(batch2) error = %v", err)
	}
	if len(got) != 1 || got[0].Seq != 3 {
		t.Fatalf("batch2 mismatch: %+v", got)
	}
}

func TestNormalizePollConfigDefaultsAndValidation(t *testing.T) {
	cfg, err := NormalizePollConfig(PollConfig{})
	if err != nil {
		t.Fatalf("NormalizePollConfig() error = %v", err)
	}
	if cfg.Interval <= 0 || cfg.BatchSize <= 0 || cfg.WaitTimeout <= 0 {
		t.Fatalf("defaults not applied: %+v", cfg)
	}

	if _, err := NormalizePollConfig(PollConfig{WaitTimeout: -1}); err == nil {
		t.Fatalf("expected negative wait timeout validation error")
	}
}
