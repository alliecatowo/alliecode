package cost

import (
	"sort"
	"sync"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

const oneMillion = 1_000_000.0

// UsageRecord captures token usage and optional direct cost for one model call.
type UsageRecord struct {
	SessionID    string      `json:"session_id"`
	Provider     string      `json:"provider,omitempty"`
	Model        string      `json:"model,omitempty"`
	Usage        types.Usage `json:"usage"`
	CostUSD      float64     `json:"cost_usd,omitempty"`
	TimestampUTC time.Time   `json:"ts"`
}

// SessionTotal contains aggregate cost and token counts for a session.
type SessionTotal struct {
	SessionID    string  `json:"session_id"`
	Calls        int     `json:"calls"`
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	CacheHits    int     `json:"cache_hits"`
	EstimatedUSD float64 `json:"estimated_usd"`
}

// Summary returns totals across all sessions.
type Summary struct {
	Sessions     int     `json:"sessions"`
	Calls        int     `json:"calls"`
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	CacheHits    int     `json:"cache_hits"`
	EstimatedUSD float64 `json:"estimated_usd"`
}

// SessionCostsReader provides read APIs for session and global cost totals.
type SessionCostsReader interface {
	SessionTotal(sessionID string) SessionTotal
	Summary() Summary
	RecentSessionTotals(limit int) []SessionTotal
}

// Accountant accumulates per-session token and currency usage.
type Accountant struct {
	mu       sync.RWMutex
	sessions map[string]SessionTotal
	updated  map[string]time.Time
}

// NewAccountant creates an in-memory cost accumulator.
func NewAccountant() *Accountant {
	return &Accountant{
		sessions: map[string]SessionTotal{},
		updated:  map[string]time.Time{},
	}
}

// EstimateCostUSD estimates cost from usage and model pricing.
func EstimateCostUSD(usage types.Usage, model types.Model) float64 {
	inputCost := (float64(usage.InputTokens) / oneMillion) * model.CostPerMInput
	outputCost := (float64(usage.OutputTokens) / oneMillion) * model.CostPerMOutput
	return inputCost + outputCost
}

// AddUsage records one usage record and updates per-session totals.
func (a *Accountant) AddUsage(record UsageRecord, model *types.Model) SessionTotal {
	a.mu.Lock()
	defer a.mu.Unlock()

	total := a.sessions[record.SessionID]
	total.SessionID = record.SessionID
	total.Calls++
	total.InputTokens += record.Usage.InputTokens
	total.OutputTokens += record.Usage.OutputTokens
	total.CacheHits += record.Usage.CacheHits

	if record.CostUSD > 0 {
		total.EstimatedUSD += record.CostUSD
	} else if model != nil {
		total.EstimatedUSD += EstimateCostUSD(record.Usage, *model)
	}

	a.sessions[record.SessionID] = total
	ts := record.TimestampUTC
	if ts.IsZero() {
		ts = time.Now().UTC()
	}
	a.updated[record.SessionID] = ts
	return total
}

// SessionTotal returns aggregate totals for one session.
func (a *Accountant) SessionTotal(sessionID string) SessionTotal {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.sessions[sessionID]
}

// Summary returns global totals across all sessions.
func (a *Accountant) Summary() Summary {
	a.mu.RLock()
	defer a.mu.RUnlock()

	out := Summary{Sessions: len(a.sessions)}
	for _, total := range a.sessions {
		out.Calls += total.Calls
		out.InputTokens += total.InputTokens
		out.OutputTokens += total.OutputTokens
		out.CacheHits += total.CacheHits
		out.EstimatedUSD += total.EstimatedUSD
	}
	return out
}

// RecentSessionTotals returns recent session totals sorted by update time.
func (a *Accountant) RecentSessionTotals(limit int) []SessionTotal {
	a.mu.RLock()
	defer a.mu.RUnlock()

	type entry struct {
		total SessionTotal
		ts    time.Time
	}
	entries := make([]entry, 0, len(a.sessions))
	for sessionID, total := range a.sessions {
		entries = append(entries, entry{total: total, ts: a.updated[sessionID]})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ts.After(entries[j].ts)
	})

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	out := make([]SessionTotal, 0, len(entries))
	for _, entry := range entries {
		out = append(out, entry.total)
	}
	return out
}
