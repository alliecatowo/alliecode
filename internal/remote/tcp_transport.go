package remote

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

var (
	errTCPTransportClosed       = errors.New("tcp transport closed")
	errTCPTransportUnauthorized = errors.New("tcp transport unauthorized")
	errTCPHandshakeInvalid      = errors.New("tcp handshake invalid")
	errTCPAlreadyConnected      = errors.New("session already connected")
	errTCPSessionMismatch       = errors.New("tcp handshake session mismatch")
)

type TCPTransportMode string

const (
	TCPTransportModeHost   TCPTransportMode = "host"
	TCPTransportModeClient TCPTransportMode = "client"
)

type TCPTransportConfig struct {
	Mode             TCPTransportMode
	Addr             string
	Token            string
	DialTimeout      time.Duration
	HandshakeTimeout time.Duration
}

type tcpHandshake struct {
	SessionID string `json:"session_id"`
	Token     string `json:"token"`
}

type tcpHandshakeAck struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

type tcpSessionState struct {
	conn   net.Conn
	reader *bufio.Reader
	queue  []RemoteMessage
	closed bool
	err    error
}

type TCPTransport struct {
	mode             TCPTransportMode
	addr             string
	token            string
	dialTimeout      time.Duration
	handshakeTimeout time.Duration

	listener net.Listener

	mu       sync.Mutex
	sessions map[string]*tcpSessionState
}

func NewTCPTransport(cfg TCPTransportConfig) (*TCPTransport, error) {
	if cfg.Mode != TCPTransportModeHost && cfg.Mode != TCPTransportModeClient {
		return nil, errors.New("tcp mode must be host or client")
	}
	if cfg.Token == "" {
		return nil, errors.New("token cannot be empty")
	}
	if cfg.Mode == TCPTransportModeClient && cfg.Addr == "" {
		return nil, errors.New("addr cannot be empty in client mode")
	}
	if cfg.DialTimeout <= 0 {
		cfg.DialTimeout = 3 * time.Second
	}
	if cfg.HandshakeTimeout <= 0 {
		cfg.HandshakeTimeout = 2 * time.Second
	}

	t := &TCPTransport{
		mode:             cfg.Mode,
		addr:             cfg.Addr,
		token:            cfg.Token,
		dialTimeout:      cfg.DialTimeout,
		handshakeTimeout: cfg.HandshakeTimeout,
		sessions:         make(map[string]*tcpSessionState),
	}

	if cfg.Mode == TCPTransportModeHost {
		ln, err := net.Listen("tcp", cfg.Addr)
		if err != nil {
			return nil, fmt.Errorf("listen tcp: %w", err)
		}
		t.listener = ln
		t.addr = ln.Addr().String()
	}

	return t, nil
}

func (t *TCPTransport) Addr() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.addr
}

func (t *TCPTransport) Connect(ctx context.Context, sessionID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if sessionID == "" {
		return errors.New("sessionID cannot be empty")
	}

	t.mu.Lock()
	if st, ok := t.sessions[sessionID]; ok && !st.closed {
		t.mu.Unlock()
		return errTCPAlreadyConnected
	}
	t.mu.Unlock()

	var (
		conn net.Conn
		err  error
	)
	if t.mode == TCPTransportModeHost {
		conn, err = t.acceptWithContext(ctx)
	} else {
		conn, err = t.dialWithContext(ctx)
	}
	if err != nil {
		return err
	}

	if t.mode == TCPTransportModeHost {
		reader := bufio.NewReader(conn)
		err = t.acceptHandshake(conn, reader, sessionID)
		if err == nil {
			st := &tcpSessionState{conn: conn, reader: reader, queue: make([]RemoteMessage, 0)}
			t.mu.Lock()
			t.sessions[sessionID] = st
			t.mu.Unlock()
			go t.readLoop(sessionID, st)
			return nil
		}
	} else {
		reader := bufio.NewReader(conn)
		err = t.sendHandshake(conn, reader, sessionID)
		if err == nil {
			st := &tcpSessionState{conn: conn, reader: reader, queue: make([]RemoteMessage, 0)}
			t.mu.Lock()
			t.sessions[sessionID] = st
			t.mu.Unlock()
			go t.readLoop(sessionID, st)
			return nil
		}
	}
	if err != nil {
		_ = conn.Close()
		return err
	}

	_ = conn.Close()
	return errTCPHandshakeInvalid
}

func (t *TCPTransport) Send(ctx context.Context, msg RemoteMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := msg.Validate(); err != nil {
		return err
	}

	t.mu.Lock()
	st, ok := t.sessions[msg.SessionID]
	if !ok || st.closed {
		t.mu.Unlock()
		return errSessionNotConnected
	}
	conn := st.conn
	t.mu.Unlock()

	if err := conn.SetWriteDeadline(time.Now().Add(t.handshakeTimeout)); err != nil {
		return err
	}
	enc := json.NewEncoder(conn)
	if err := enc.Encode(msg); err != nil {
		if errors.Is(err, io.EOF) {
			return errSessionNotConnected
		}
		return err
	}
	_ = conn.SetWriteDeadline(time.Time{})
	return nil
}

func (t *TCPTransport) Receive(ctx context.Context, sessionID string, cfg PollConfig) ([]RemoteMessage, error) {
	if sessionID == "" {
		return nil, errors.New("sessionID cannot be empty")
	}
	normalized, err := NormalizePollConfig(cfg)
	if err != nil {
		return nil, err
	}

	deadline := time.Now().Add(normalized.WaitTimeout)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		t.mu.Lock()
		st, ok := t.sessions[sessionID]
		if !ok || st.closed {
			t.mu.Unlock()
			return nil, errSessionNotConnected
		}
		if len(st.queue) > 0 {
			n := normalized.BatchSize
			if n > len(st.queue) {
				n = len(st.queue)
			}
			out := make([]RemoteMessage, n)
			copy(out, st.queue[:n])
			remaining := make([]RemoteMessage, len(st.queue)-n)
			copy(remaining, st.queue[n:])
			st.queue = remaining
			t.mu.Unlock()
			return out, nil
		}
		if st.err != nil {
			recvErr := st.err
			t.mu.Unlock()
			return nil, recvErr
		}
		t.mu.Unlock()

		if time.Now().After(deadline) {
			return nil, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(normalized.Interval):
		}
	}
}

func (t *TCPTransport) Close(sessionID string) error {
	if sessionID == "" {
		return errors.New("sessionID cannot be empty")
	}

	t.mu.Lock()
	st, ok := t.sessions[sessionID]
	if !ok {
		t.mu.Unlock()
		return errSessionNotConnected
	}
	delete(t.sessions, sessionID)
	st.closed = true
	t.mu.Unlock()

	if err := st.conn.Close(); err != nil {
		return err
	}

	return nil
}

func (t *TCPTransport) Shutdown() error {
	t.mu.Lock()
	listener := t.listener
	t.listener = nil
	sessions := make([]*tcpSessionState, 0, len(t.sessions))
	for sessionID, st := range t.sessions {
		st.closed = true
		sessions = append(sessions, st)
		delete(t.sessions, sessionID)
	}
	t.mu.Unlock()

	for _, st := range sessions {
		_ = st.conn.Close()
	}
	if listener != nil {
		if err := listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			return err
		}
	}
	return nil
}

func (t *TCPTransport) acceptWithContext(ctx context.Context) (net.Conn, error) {
	t.mu.Lock()
	ln := t.listener
	t.mu.Unlock()
	if ln == nil {
		return nil, errTCPTransportClosed
	}

	type acceptResult struct {
		conn net.Conn
		err  error
	}
	ch := make(chan acceptResult, 1)
	go func() {
		conn, err := ln.Accept()
		ch <- acceptResult{conn: conn, err: err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		if res.err != nil {
			if errors.Is(res.err, net.ErrClosed) {
				return nil, errTCPTransportClosed
			}
			return nil, res.err
		}
		return res.conn, nil
	}
}

func (t *TCPTransport) dialWithContext(ctx context.Context) (net.Conn, error) {
	dialCtx, cancel := context.WithTimeout(ctx, t.dialTimeout)
	defer cancel()

	d := net.Dialer{}
	conn, err := d.DialContext(dialCtx, "tcp", t.addr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (t *TCPTransport) acceptHandshake(conn net.Conn, reader *bufio.Reader, sessionID string) error {
	if err := conn.SetDeadline(time.Now().Add(t.handshakeTimeout)); err != nil {
		return err
	}

	dec := json.NewDecoder(reader)
	var hs tcpHandshake
	if err := dec.Decode(&hs); err != nil {
		_ = t.writeHandshakeAck(conn, false, errTCPHandshakeInvalid.Error())
		return errTCPHandshakeInvalid
	}

	if hs.SessionID != sessionID {
		_ = t.writeHandshakeAck(conn, false, errTCPSessionMismatch.Error())
		return errTCPSessionMismatch
	}
	if hs.Token != t.token {
		_ = t.writeHandshakeAck(conn, false, errTCPTransportUnauthorized.Error())
		return errTCPTransportUnauthorized
	}

	if err := t.writeHandshakeAck(conn, true, ""); err != nil {
		return err
	}
	if err := conn.SetDeadline(time.Time{}); err != nil {
		return err
	}
	return nil
}

func (t *TCPTransport) sendHandshake(conn net.Conn, reader *bufio.Reader, sessionID string) error {
	if err := conn.SetDeadline(time.Now().Add(t.handshakeTimeout)); err != nil {
		return err
	}

	enc := json.NewEncoder(conn)
	if err := enc.Encode(tcpHandshake{SessionID: sessionID, Token: t.token}); err != nil {
		return err
	}

	dec := json.NewDecoder(reader)
	var ack tcpHandshakeAck
	if err := dec.Decode(&ack); err != nil {
		return errTCPHandshakeInvalid
	}
	if !ack.OK {
		if ack.Error == errTCPTransportUnauthorized.Error() {
			return errTCPTransportUnauthorized
		}
		if ack.Error == errTCPSessionMismatch.Error() {
			return errTCPSessionMismatch
		}
		if ack.Error == "" {
			return errTCPHandshakeInvalid
		}
		return fmt.Errorf("%w: %s", errTCPHandshakeInvalid, ack.Error)
	}

	if err := conn.SetDeadline(time.Time{}); err != nil {
		return err
	}
	return nil
}

func (t *TCPTransport) writeHandshakeAck(conn net.Conn, ok bool, errMsg string) error {
	enc := json.NewEncoder(conn)
	return enc.Encode(tcpHandshakeAck{OK: ok, Error: errMsg})
}

func (t *TCPTransport) readLoop(sessionID string, st *tcpSessionState) {
	dec := json.NewDecoder(st.reader)
	for {
		var msg RemoteMessage
		if err := dec.Decode(&msg); err != nil {
			t.mu.Lock()
			defer t.mu.Unlock()
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				st.err = errSessionNotConnected
			} else {
				st.err = err
			}
			st.closed = true
			delete(t.sessions, sessionID)
			return
		}
		if err := msg.Validate(); err != nil {
			t.mu.Lock()
			st.err = err
			st.closed = true
			delete(t.sessions, sessionID)
			t.mu.Unlock()
			return
		}
		t.mu.Lock()
		if st.closed {
			t.mu.Unlock()
			return
		}
		st.queue = append(st.queue, msg)
		t.mu.Unlock()
	}
}
