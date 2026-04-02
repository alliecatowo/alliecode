package mcp

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// TransportType identifies the client transport used for an MCP server.
type TransportType string

const (
	TransportStdio     TransportType = "stdio"
	TransportSSE       TransportType = "sse"
	TransportWebSocket TransportType = "websocket"
)

func normalizeTransportType(transport TransportType) TransportType {
	switch strings.ToLower(strings.TrimSpace(string(transport))) {
	case "", "stdio":
		return TransportStdio
	case "sse":
		return TransportSSE
	case "ws", "websocket":
		return TransportWebSocket
	default:
		return TransportType(strings.ToLower(strings.TrimSpace(string(transport))))
	}
}

// ErrTransportUnsupported is returned by no-op transport adapters.
var ErrTransportUnsupported = fmt.Errorf("mcp: transport is not supported")

type rpcTransport interface {
	WriteMessage(data []byte) error
	ReadMessage() ([]byte, error)
	Close() error
}

type stdioTransport struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
}

func newStdioTransport(serverCmd string, args []string) (*stdioTransport, error) {
	cmd := exec.Command(serverCmd, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("mcp: creating stdin pipe: %w", err)
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("mcp: creating stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("mcp: starting server %q: %w", serverCmd, err)
	}

	return &stdioTransport{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdoutPipe),
	}, nil
}

func (t *stdioTransport) WriteMessage(data []byte) error {
	_, err := t.stdin.Write(data)
	return err
}

func (t *stdioTransport) ReadMessage() ([]byte, error) {
	return t.stdout.ReadBytes('\n')
}

func (t *stdioTransport) Close() error {
	if t.stdin != nil {
		_ = t.stdin.Close()
	}
	if t.cmd != nil {
		return t.cmd.Wait()
	}
	return nil
}

type unsupportedTransport struct {
	transportType TransportType
}

func (t *unsupportedTransport) unsupportedErr() error {
	return fmt.Errorf("%w: %s", ErrTransportUnsupported, t.transportType)
}

func (t *unsupportedTransport) WriteMessage(_ []byte) error {
	return t.unsupportedErr()
}

func (t *unsupportedTransport) ReadMessage() ([]byte, error) {
	return nil, t.unsupportedErr()
}

func (t *unsupportedTransport) Close() error {
	return nil
}
