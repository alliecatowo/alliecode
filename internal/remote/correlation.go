package remote

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// RequestLifecycleState captures request/response/cancel progression.
type RequestLifecycleState string

const (
	RequestLifecycleStatePending   RequestLifecycleState = "pending"
	RequestLifecycleStateResponded RequestLifecycleState = "responded"
	RequestLifecycleStateCanceled  RequestLifecycleState = "canceled"
)

// RequestSnapshot is one immutable request lifecycle snapshot.
type RequestSnapshot struct {
	RequestID      string                `json:"request_id"`
	Method         string                `json:"method,omitempty"`
	State          RequestLifecycleState `json:"state"`
	Attempts       int                   `json:"attempts"`
	LastSeq        int64                 `json:"last_seq"`
	LastUpdatedAt  time.Time             `json:"last_updated_at"`
	LastError      string                `json:"last_error,omitempty"`
	OrphanTerminal bool                  `json:"orphan_terminal,omitempty"`
}

// RequestStats reports aggregate request lifecycle diagnostics.
type RequestStats struct {
	Pending       int `json:"pending"`
	Responded     int `json:"responded"`
	Canceled      int `json:"canceled"`
	Orphaned      int `json:"orphaned"`
	RetryAttempts int `json:"retry_attempts"`
	ReplayDrops   int `json:"replay_drops"`
}

type requestRecord struct {
	RequestSnapshot
}

// RequestCorrelator tracks request/response/cancel transitions.
//
// It is retry-safe:
// - repeated request messages for a pending request are treated as retries
// - replayed older messages are dropped deterministically
// - terminal state conflicts are rejected deterministically
type RequestCorrelator struct {
	mu      sync.Mutex
	records map[string]requestRecord
	stats   RequestStats
}

// NewRequestCorrelator returns an empty request correlator.
func NewRequestCorrelator() *RequestCorrelator {
	return &RequestCorrelator{records: make(map[string]requestRecord)}
}

// Observe updates correlation state from one protocol message.
func (c *RequestCorrelator) Observe(msg RemoteMessage) error {
	if c == nil {
		return errors.New("request correlator cannot be nil")
	}

	switch msg.Kind {
	case MessageKindRequest:
		var payload ActionRequestPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return fmt.Errorf("decode request payload: %w", err)
		}
		return c.observeRequest(payload, msg.Seq, msg.SentAt)
	case MessageKindResponse:
		var payload ActionResponsePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return fmt.Errorf("decode response payload: %w", err)
		}
		return c.observeResponse(payload, msg.Seq, msg.SentAt)
	case MessageKindCancel:
		var payload ActionCancelPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return fmt.Errorf("decode cancel payload: %w", err)
		}
		return c.observeCancel(payload, msg.Seq, msg.SentAt)
	default:
		return nil
	}
}

func (c *RequestCorrelator) observeRequest(payload ActionRequestPayload, seq int64, sentAt time.Time) error {
	requestID := strings.TrimSpace(payload.RequestID)
	if requestID == "" {
		return errors.New("request payload request_id cannot be empty")
	}
	method := strings.TrimSpace(payload.Method)
	if method == "" {
		return errors.New("request payload method cannot be empty")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	rec, exists := c.records[requestID]
	if !exists {
		rec = requestRecord{RequestSnapshot: RequestSnapshot{
			RequestID:     requestID,
			Method:        method,
			State:         RequestLifecycleStatePending,
			Attempts:      1,
			LastSeq:       seq,
			LastUpdatedAt: normalizeSentAt(sentAt),
		}}
		c.records[requestID] = rec
		c.stats.Pending++
		return nil
	}

	if seq <= rec.LastSeq {
		c.stats.ReplayDrops++
		return nil
	}

	if rec.Method != "" && rec.Method != method {
		return fmt.Errorf("request %q method mismatch: have %q got %q", requestID, rec.Method, method)
	}

	switch rec.State {
	case RequestLifecycleStatePending:
		rec.Attempts++
		rec.LastSeq = seq
		rec.LastUpdatedAt = normalizeSentAt(sentAt)
		rec.LastError = ""
		rec.OrphanTerminal = false
		c.records[requestID] = rec
		c.stats.RetryAttempts++
		return nil
	case RequestLifecycleStateResponded, RequestLifecycleStateCanceled:
		return fmt.Errorf("request %q already terminal in state %q", requestID, rec.State)
	default:
		return fmt.Errorf("request %q has invalid state %q", requestID, rec.State)
	}
}

func (c *RequestCorrelator) observeResponse(payload ActionResponsePayload, seq int64, sentAt time.Time) error {
	requestID := strings.TrimSpace(payload.RequestID)
	if requestID == "" {
		return errors.New("response payload request_id cannot be empty")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	rec, exists := c.records[requestID]
	if !exists {
		c.records[requestID] = requestRecord{RequestSnapshot: RequestSnapshot{
			RequestID:      requestID,
			State:          RequestLifecycleStateResponded,
			Attempts:       0,
			LastSeq:        seq,
			LastUpdatedAt:  normalizeSentAt(sentAt),
			LastError:      strings.TrimSpace(payload.Error),
			OrphanTerminal: true,
		}}
		c.stats.Responded++
		c.stats.Orphaned++
		return nil
	}

	if seq <= rec.LastSeq {
		c.stats.ReplayDrops++
		return nil
	}

	switch rec.State {
	case RequestLifecycleStatePending:
		c.stats.Pending--
		c.stats.Responded++
		rec.State = RequestLifecycleStateResponded
		rec.LastSeq = seq
		rec.LastUpdatedAt = normalizeSentAt(sentAt)
		rec.LastError = strings.TrimSpace(payload.Error)
		rec.OrphanTerminal = false
		c.records[requestID] = rec
		return nil
	case RequestLifecycleStateResponded:
		rec.LastSeq = seq
		rec.LastUpdatedAt = normalizeSentAt(sentAt)
		if errText := strings.TrimSpace(payload.Error); errText != "" {
			rec.LastError = errText
		}
		c.records[requestID] = rec
		return nil
	case RequestLifecycleStateCanceled:
		return fmt.Errorf("request %q already canceled before response", requestID)
	default:
		return fmt.Errorf("request %q has invalid state %q", requestID, rec.State)
	}
}

func (c *RequestCorrelator) observeCancel(payload ActionCancelPayload, seq int64, sentAt time.Time) error {
	requestID := strings.TrimSpace(payload.RequestID)
	if requestID == "" {
		return errors.New("cancel payload request_id cannot be empty")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	rec, exists := c.records[requestID]
	if !exists {
		c.records[requestID] = requestRecord{RequestSnapshot: RequestSnapshot{
			RequestID:      requestID,
			State:          RequestLifecycleStateCanceled,
			Attempts:       0,
			LastSeq:        seq,
			LastUpdatedAt:  normalizeSentAt(sentAt),
			LastError:      strings.TrimSpace(payload.Reason),
			OrphanTerminal: true,
		}}
		c.stats.Canceled++
		c.stats.Orphaned++
		return nil
	}

	if seq <= rec.LastSeq {
		c.stats.ReplayDrops++
		return nil
	}

	switch rec.State {
	case RequestLifecycleStatePending:
		c.stats.Pending--
		c.stats.Canceled++
		rec.State = RequestLifecycleStateCanceled
		rec.LastSeq = seq
		rec.LastUpdatedAt = normalizeSentAt(sentAt)
		rec.LastError = strings.TrimSpace(payload.Reason)
		rec.OrphanTerminal = false
		c.records[requestID] = rec
		return nil
	case RequestLifecycleStateCanceled:
		rec.LastSeq = seq
		rec.LastUpdatedAt = normalizeSentAt(sentAt)
		if reason := strings.TrimSpace(payload.Reason); reason != "" {
			rec.LastError = reason
		}
		c.records[requestID] = rec
		return nil
	case RequestLifecycleStateResponded:
		return fmt.Errorf("request %q already responded before cancel", requestID)
	default:
		return fmt.Errorf("request %q has invalid state %q", requestID, rec.State)
	}
}

func normalizeSentAt(ts time.Time) time.Time {
	if ts.IsZero() {
		return time.Now().UTC()
	}
	return ts.UTC()
}

// Snapshot returns lifecycle details for one request ID.
func (c *RequestCorrelator) Snapshot(requestID string) (RequestSnapshot, bool) {
	if c == nil {
		return RequestSnapshot{}, false
	}
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return RequestSnapshot{}, false
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	rec, ok := c.records[requestID]
	if !ok {
		return RequestSnapshot{}, false
	}
	return rec.RequestSnapshot, true
}

// Stats returns aggregate request lifecycle diagnostics.
func (c *RequestCorrelator) Stats() RequestStats {
	if c == nil {
		return RequestStats{}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.stats
}

// PendingRequestIDs returns stable-sorted pending request IDs.
func (c *RequestCorrelator) PendingRequestIDs() []string {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]string, 0, c.stats.Pending)
	for requestID, rec := range c.records {
		if rec.State == RequestLifecycleStatePending {
			out = append(out, requestID)
		}
	}
	sort.Strings(out)
	return out
}
