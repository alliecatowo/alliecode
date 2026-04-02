package history

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

const defaultHistoryDir = ".alliecode/history"

const (
	globalHistoryFileName  = "history.jsonl"
	defaultMaxHistoryItems = 100
)

// EventType identifies a stored history event.
type EventType string

const (
	EventSessionStart EventType = "session_start"
	EventMessage      EventType = "message"
	EventCustom       EventType = "custom"
)

// Event is a single append-only session history event.
type Event struct {
	Type           EventType        `json:"type"`
	SessionID      string           `json:"session_id"`
	Timestamp      time.Time        `json:"ts"`
	Display        string           `json:"display,omitempty"`
	Project        string           `json:"project,omitempty"`
	Message        *types.Message   `json:"message,omitempty"`
	Payload        *json.RawMessage `json:"payload,omitempty"`
	RuntimeSurface string           `json:"runtime_surface,omitempty"`
	CommandSurface string           `json:"command_surface,omitempty"`
	Provider       string           `json:"provider,omitempty"`
	Model          string           `json:"model,omitempty"`
	Tags           []string         `json:"tags,omitempty"`
}

// HistoryQuery controls global-history retrieval semantics.
type HistoryQuery struct {
	Project          string
	CurrentSessionID string
	MaxItems         int
	ContainsDisplay  string
	Type             EventType
	SessionID        string
}

// TimestampedHistoryEntry is a deduped display item with lazy-resolvable source.
type TimestampedHistoryEntry struct {
	Display   string    `json:"display"`
	Timestamp time.Time `json:"timestamp"`
	Event     Event     `json:"event"`
}

// SessionSummary contains compact session details for recent-session listing.
type SessionSummary struct {
	SessionID       string    `json:"session_id"`
	Path            string    `json:"path"`
	Project         string    `json:"project,omitempty"`
	FirstEvent      time.Time `json:"first_event"`
	LastEvent       time.Time `json:"last_event"`
	EventCount      int       `json:"event_count"`
	MessageCount    int       `json:"message_count"`
	Title           string    `json:"title,omitempty"`
	Summary         string    `json:"summary,omitempty"`
	RuntimeSurface  string    `json:"runtime_surface,omitempty"`
	CommandSurface  string    `json:"command_surface,omitempty"`
	Provider        string    `json:"provider,omitempty"`
	Model           string    `json:"model,omitempty"`
	UniquePrompts   int       `json:"unique_prompts,omitempty"`
	DistinctTagKeys int       `json:"distinct_tag_keys,omitempty"`
}

type SessionQuery struct {
	ProjectContains string
	SessionContains string
	Limit           int
}

// Store defines persistence and query operations for session history.
type Store interface {
	Append(Event) error
	AppendSessionStart(sessionID string) error
	AppendSessionStartForProject(sessionID, project string) error
	AppendMessage(sessionID string, msg types.Message) error
	AppendMessageForProject(sessionID, project string, msg types.Message) error
	AppendDisplay(sessionID, project, display string) error
	SessionEvents(sessionID string) ([]Event, error)
	RecentSessions(limit int) ([]SessionSummary, error)
	QueryRecentSessions(query SessionQuery) ([]SessionSummary, error)
	GetHistory(query HistoryQuery) ([]Event, error)
	GetTimestampedHistory(query HistoryQuery) ([]TimestampedHistoryEntry, error)
}

// JSONLStore persists each session as an append-only JSONL file.
type JSONLStore struct {
	rootDir string
	mu      sync.Mutex
}

// NewJSONLStore creates a history store rooted under rootDir.
func NewJSONLStore(rootDir string) (*JSONLStore, error) {
	dir := filepath.Join(rootDir, defaultHistoryDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create history directory: %w", err)
	}
	return &JSONLStore{rootDir: rootDir}, nil
}

// SessionPath returns the JSONL file path for a session.
func SessionPath(rootDir, sessionID string) string {
	return filepath.Join(rootDir, defaultHistoryDir, sessionID+".jsonl")
}

func GlobalPath(rootDir string) string {
	return filepath.Join(rootDir, defaultHistoryDir, globalHistoryFileName)
}

// AppendSessionStart appends a session_start event.
func (s *JSONLStore) AppendSessionStart(sessionID string) error {
	return s.Append(Event{Type: EventSessionStart, SessionID: sessionID})
}

// AppendSessionStartForProject appends a session_start event with project metadata.
func (s *JSONLStore) AppendSessionStartForProject(sessionID, project string) error {
	return s.Append(Event{Type: EventSessionStart, SessionID: sessionID, Project: project})
}

// AppendMessage appends a message event for a session.
func (s *JSONLStore) AppendMessage(sessionID string, msg types.Message) error {
	m := msg
	return s.Append(Event{Type: EventMessage, SessionID: sessionID, Display: msg.GetText(), Message: &m})
}

// AppendMessageForProject appends a message event with project metadata.
func (s *JSONLStore) AppendMessageForProject(sessionID, project string, msg types.Message) error {
	m := msg
	return s.Append(Event{Type: EventMessage, SessionID: sessionID, Project: project, Display: msg.GetText(), Message: &m})
}

// AppendDisplay appends a display-only event for prompt history retrieval.
func (s *JSONLStore) AppendDisplay(sessionID, project, display string) error {
	return s.Append(Event{Type: EventCustom, SessionID: sessionID, Project: project, Display: display})
}

// Append writes one event to the corresponding session JSONL file.
func (s *JSONLStore) Append(ev Event) error {
	if strings.TrimSpace(ev.SessionID) == "" {
		return fmt.Errorf("session_id is required")
	}
	if ev.Timestamp.IsZero() {
		ev.Timestamp = time.Now().UTC()
	}
	ev.Project = strings.TrimSpace(ev.Project)
	ev.Display = strings.TrimSpace(ev.Display)
	ev.RuntimeSurface = strings.TrimSpace(ev.RuntimeSurface)
	ev.CommandSurface = strings.TrimSpace(ev.CommandSurface)
	ev.Provider = strings.TrimSpace(ev.Provider)
	ev.Model = strings.TrimSpace(ev.Model)
	if len(ev.Tags) > 0 {
		ev.Tags = normalizeTags(ev.Tags)
	}

	path := SessionPath(s.rootDir, ev.SessionID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create session history directory: %w", err)
	}
	globalPath := GlobalPath(s.rootDir)

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := appendEvent(path, ev); err != nil {
		return err
	}
	if err := appendEvent(globalPath, ev); err != nil {
		return err
	}
	return nil
}

func appendEvent(path string, ev Event) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open history file: %w", err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(ev); err != nil {
		return fmt.Errorf("encode history event: %w", err)
	}
	return nil
}

// SessionEvents loads all events for one session in append order.
func (s *JSONLStore) SessionEvents(sessionID string) ([]Event, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, fmt.Errorf("session_id is required")
	}
	return loadSessionEvents(SessionPath(s.rootDir, sessionID))
}

// RecentSessions returns recent sessions sorted by latest event timestamp.
func (s *JSONLStore) RecentSessions(limit int) ([]SessionSummary, error) {
	return s.QueryRecentSessions(SessionQuery{Limit: limit})
}

func (s *JSONLStore) QueryRecentSessions(query SessionQuery) ([]SessionSummary, error) {
	dir := filepath.Join(s.rootDir, defaultHistoryDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read history directory: %w", err)
	}

	projectFilter := strings.ToLower(strings.TrimSpace(query.ProjectContains))
	sessionFilter := strings.ToLower(strings.TrimSpace(query.SessionContains))

	out := make([]SessionSummary, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".jsonl" {
			continue
		}
		if entry.Name() == globalHistoryFileName {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		events, err := loadSessionEvents(path)
		if err != nil {
			return nil, err
		}
		if len(events) == 0 {
			continue
		}

		summary := SessionSummary{
			SessionID:  strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())),
			Path:       path,
			FirstEvent: events[0].Timestamp,
			LastEvent:  events[len(events)-1].Timestamp,
			EventCount: len(events),
		}
		summary.Project = latestProject(events)
		summary.RuntimeSurface = latestRuntimeSurface(events)
		summary.CommandSurface = latestCommandSurface(events)
		summary.Provider = latestProvider(events)
		summary.Model = latestModel(events)
		summary.UniquePrompts = countUniqueDisplays(events)
		summary.DistinctTagKeys = countDistinctTags(events)
		for _, ev := range events {
			if ev.Type == EventMessage && ev.Message != nil {
				summary.MessageCount++
			}
		}
		if display := latestDisplay(events); display != "" {
			summary.Title = display
			summary.Summary = display
		}
		if events[len(events)-1].SessionID != "" {
			summary.SessionID = events[len(events)-1].SessionID
		}
		if projectFilter != "" && !strings.Contains(strings.ToLower(summary.Project), projectFilter) {
			continue
		}
		if sessionFilter != "" && !strings.Contains(strings.ToLower(summary.SessionID), sessionFilter) {
			continue
		}
		out = append(out, summary)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].LastEvent.After(out[j].LastEvent)
	})

	if query.Limit > 0 && len(out) > query.Limit {
		out = out[:query.Limit]
	}
	return out, nil
}

func (s *JSONLStore) QueryIndex(query HistoryQuery, indexQuery IndexQuery) ([]IndexRecord, error) {
	events, err := s.GetHistory(query)
	if err != nil {
		return nil, err
	}
	return BuildIndex(events, indexQuery), nil
}

// GetHistory returns project-filtered global history with current-session-first ordering.
func (s *JSONLStore) GetHistory(query HistoryQuery) ([]Event, error) {
	if strings.TrimSpace(query.Project) == "" {
		return nil, fmt.Errorf("project is required")
	}

	events, err := loadGlobalEventsReverse(GlobalPath(s.rootDir))
	if err != nil {
		return nil, err
	}

	maxItems := query.MaxItems
	if maxItems <= 0 {
		maxItems = defaultMaxHistoryItems
	}

	currentSession := make([]Event, 0, maxItems)
	otherSessions := make([]Event, 0, maxItems)
	seenInWindow := 0

	for _, ev := range events {
		if query.Type != "" && ev.Type != query.Type {
			continue
		}
		if query.SessionID != "" && ev.SessionID != query.SessionID {
			continue
		}
		if ev.Project != query.Project {
			continue
		}
		display := eventDisplay(ev)
		if display == "" {
			continue
		}
		if contains := strings.ToLower(strings.TrimSpace(query.ContainsDisplay)); contains != "" {
			if !strings.Contains(strings.ToLower(display), contains) {
				continue
			}
		}
		if display == "" {
			continue
		}

		seenInWindow++
		if query.CurrentSessionID != "" && ev.SessionID == query.CurrentSessionID {
			currentSession = append(currentSession, ev)
		} else {
			otherSessions = append(otherSessions, ev)
		}

		if seenInWindow >= maxItems {
			break
		}
	}

	out := make([]Event, 0, len(currentSession)+len(otherSessions))
	out = append(out, currentSession...)
	out = append(out, otherSessions...)
	if len(out) > maxItems {
		out = out[:maxItems]
	}
	return out, nil
}

// GetTimestampedHistory returns project-filtered, display-deduped history entries.
func (s *JSONLStore) GetTimestampedHistory(query HistoryQuery) ([]TimestampedHistoryEntry, error) {
	if strings.TrimSpace(query.Project) == "" {
		return nil, fmt.Errorf("project is required")
	}

	events, err := loadGlobalEventsReverse(GlobalPath(s.rootDir))
	if err != nil {
		return nil, err
	}

	maxItems := query.MaxItems
	if maxItems <= 0 {
		maxItems = defaultMaxHistoryItems
	}

	seen := make(map[string]struct{}, maxItems)
	out := make([]TimestampedHistoryEntry, 0, maxItems)
	for _, ev := range events {
		if query.Type != "" && ev.Type != query.Type {
			continue
		}
		if query.SessionID != "" && ev.SessionID != query.SessionID {
			continue
		}
		if ev.Project != query.Project {
			continue
		}
		display := eventDisplay(ev)
		if display == "" {
			continue
		}
		if contains := strings.ToLower(strings.TrimSpace(query.ContainsDisplay)); contains != "" {
			if !strings.Contains(strings.ToLower(display), contains) {
				continue
			}
		}
		if _, ok := seen[display]; ok {
			continue
		}
		seen[display] = struct{}{}
		out = append(out, TimestampedHistoryEntry{Display: display, Timestamp: ev.Timestamp, Event: ev})
		if len(out) >= maxItems {
			break
		}
	}

	return out, nil
}

func loadGlobalEventsReverse(path string) ([]Event, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open global history file %q: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lines := make([]string, 0, 128)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan global history file %q: %w", path, err)
	}

	out := make([]Event, 0, len(lines))
	for i := len(lines) - 1; i >= 0; i-- {
		var ev Event
		if err := json.Unmarshal([]byte(lines[i]), &ev); err != nil {
			continue
		}
		out = append(out, ev)
	}
	return out, nil
}

func eventDisplay(ev Event) string {
	if strings.TrimSpace(ev.Display) != "" {
		return ev.Display
	}
	if ev.Message != nil {
		return ev.Message.GetText()
	}
	return ""
}

func latestDisplay(events []Event) string {
	for i := len(events) - 1; i >= 0; i-- {
		display := strings.TrimSpace(eventDisplay(events[i]))
		if display != "" {
			return display
		}
	}
	return ""
}

func loadSessionEvents(path string) ([]Event, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open history file %q: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	out := make([]Event, 0, 64)
	line := 0
	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		var ev Event
		if err := json.Unmarshal([]byte(text), &ev); err != nil {
			return nil, fmt.Errorf("decode history event %q line %d: %w", path, line, err)
		}
		out = append(out, ev)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan history file %q: %w", path, err)
	}
	return out, nil
}

func normalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func latestProject(events []Event) string {
	for i := len(events) - 1; i >= 0; i-- {
		if p := strings.TrimSpace(events[i].Project); p != "" {
			return p
		}
	}
	return ""
}

func latestRuntimeSurface(events []Event) string {
	for i := len(events) - 1; i >= 0; i-- {
		if p := strings.TrimSpace(events[i].RuntimeSurface); p != "" {
			return p
		}
	}
	return ""
}

func latestCommandSurface(events []Event) string {
	for i := len(events) - 1; i >= 0; i-- {
		if p := strings.TrimSpace(events[i].CommandSurface); p != "" {
			return p
		}
	}
	return ""
}

func latestProvider(events []Event) string {
	for i := len(events) - 1; i >= 0; i-- {
		if p := strings.TrimSpace(events[i].Provider); p != "" {
			return p
		}
	}
	return ""
}

func latestModel(events []Event) string {
	for i := len(events) - 1; i >= 0; i-- {
		if p := strings.TrimSpace(events[i].Model); p != "" {
			return p
		}
	}
	return ""
}

func countUniqueDisplays(events []Event) int {
	seen := map[string]struct{}{}
	for _, ev := range events {
		display := strings.TrimSpace(eventDisplay(ev))
		if display == "" {
			continue
		}
		seen[display] = struct{}{}
	}
	return len(seen)
}

func countDistinctTags(events []Event) int {
	seen := map[string]struct{}{}
	for _, ev := range events {
		for _, tag := range ev.Tags {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			seen[tag] = struct{}{}
		}
	}
	return len(seen)
}
