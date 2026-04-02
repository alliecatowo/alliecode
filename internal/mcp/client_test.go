package mcp

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestNewClientWithConfigUnsupportedTransportsReturnDeterministicError(t *testing.T) {
	tests := []struct {
		name      string
		transport TransportType
	}{
		{name: "sse", transport: TransportSSE},
		{name: "websocket", transport: TransportWebSocket},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewClientWithConfig(ClientConfig{Transport: tc.transport, URL: "https://example.invalid"})
			if err != nil {
				t.Fatalf("NewClientWithConfig() error = %v", err)
			}
			t.Cleanup(func() {
				_ = client.Close()
			})

			err = client.Initialize(context.Background())
			if err == nil {
				t.Fatalf("expected Initialize() to fail for %q", tc.transport)
			}
			if !errors.Is(err, ErrTransportUnsupported) {
				t.Fatalf("Initialize() error = %v, want ErrTransportUnsupported", err)
			}
			if !strings.Contains(err.Error(), string(tc.transport)) {
				t.Fatalf("Initialize() error = %q, expected transport name", err.Error())
			}
		})
	}
}

func TestNewClientWithConfigRejectsUnknownTransport(t *testing.T) {
	_, err := NewClientWithConfig(ClientConfig{Transport: TransportType("http")})
	if err == nil {
		t.Fatalf("expected unknown transport error")
	}
	if !strings.Contains(err.Error(), "unknown transport") {
		t.Fatalf("unexpected error: %v", err)
	}
}
