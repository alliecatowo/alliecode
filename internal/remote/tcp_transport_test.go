package remote

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTCPTransportHandshakeSuccessAndSendReceive(t *testing.T) {
	host, err := NewTCPTransport(TCPTransportConfig{
		Mode:  TCPTransportModeHost,
		Addr:  "127.0.0.1:0",
		Token: "secret-token",
	})
	if err != nil {
		t.Fatalf("NewTCPTransport(host) error = %v", err)
	}
	defer func() { _ = host.Shutdown() }()

	client, err := NewTCPTransport(TCPTransportConfig{
		Mode:  TCPTransportModeClient,
		Addr:  host.Addr(),
		Token: "secret-token",
	})
	if err != nil {
		t.Fatalf("NewTCPTransport(client) error = %v", err)
	}
	defer func() { _ = client.Shutdown() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	hostErrCh := make(chan error, 1)
	go func() {
		hostErrCh <- host.Connect(ctx, "sid-1")
	}()

	if err := client.Connect(ctx, "sid-1"); err != nil {
		t.Fatalf("client Connect() error = %v", err)
	}
	if err := <-hostErrCh; err != nil {
		t.Fatalf("host Connect() error = %v", err)
	}

	now := time.Now().UTC()
	fromClient := RemoteMessage{SessionID: "sid-1", Kind: MessageKindProviderEvent, Name: "delta", Seq: 1, SentAt: now}
	fromHost := RemoteMessage{SessionID: "sid-1", Kind: MessageKindToolEvent, Name: "tool", Seq: 2, SentAt: now.Add(time.Millisecond)}

	if err := client.Send(ctx, fromClient); err != nil {
		t.Fatalf("client Send() error = %v", err)
	}
	hostRecv, err := host.Receive(ctx, "sid-1", PollConfig{WaitTimeout: 250 * time.Millisecond})
	if err != nil {
		t.Fatalf("host Receive() error = %v", err)
	}
	if len(hostRecv) != 1 || hostRecv[0].Seq != 1 {
		t.Fatalf("host receive mismatch: %+v", hostRecv)
	}

	if err := host.Send(ctx, fromHost); err != nil {
		t.Fatalf("host Send() error = %v", err)
	}
	clientRecv, err := client.Receive(ctx, "sid-1", PollConfig{WaitTimeout: 250 * time.Millisecond})
	if err != nil {
		t.Fatalf("client Receive() error = %v", err)
	}
	if len(clientRecv) != 1 || clientRecv[0].Seq != 2 {
		t.Fatalf("client receive mismatch: %+v", clientRecv)
	}
}

func TestTCPTransportHandshakeFailureUnauthorized(t *testing.T) {
	host, err := NewTCPTransport(TCPTransportConfig{
		Mode:  TCPTransportModeHost,
		Addr:  "127.0.0.1:0",
		Token: "correct-token",
	})
	if err != nil {
		t.Fatalf("NewTCPTransport(host) error = %v", err)
	}
	defer func() { _ = host.Shutdown() }()

	client, err := NewTCPTransport(TCPTransportConfig{
		Mode:  TCPTransportModeClient,
		Addr:  host.Addr(),
		Token: "wrong-token",
	})
	if err != nil {
		t.Fatalf("NewTCPTransport(client) error = %v", err)
	}
	defer func() { _ = client.Shutdown() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	hostErrCh := make(chan error, 1)
	go func() {
		hostErrCh <- host.Connect(ctx, "sid-unauth")
	}()

	if err := client.Connect(ctx, "sid-unauth"); !errors.Is(err, errTCPTransportUnauthorized) {
		t.Fatalf("client Connect() error = %v, want unauthorized", err)
	}
	if err := <-hostErrCh; !errors.Is(err, errTCPTransportUnauthorized) {
		t.Fatalf("host Connect() error = %v, want unauthorized", err)
	}
}
