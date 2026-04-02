package session

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

const (
	defaultSessionsDir = ".alliecode/sessions"
	eventSessionStart  = "session_start"
	eventMessageAppend = "message_append"
	eventCompaction    = "compaction_boundary"
)

// Transcript is the reconstructed state of a session from event replay.
type Transcript struct {
	SessionID       string
	Path            string
	Messages        []types.Message
	Boundaries      []types.CompactionBoundary
	ProjectPath     string
	RuntimeSurface  string
	CommandSurface  string
	Provider        string
	Model           string
	EventCount      int
	MessageCount    int
	FirstEventAt    time.Time
	LastEventAt     time.Time
	LastMessageRole string
}

// Store appends session events to a JSONL transcript.
type Store struct {
	path      string
	sessionID string
	mu        sync.Mutex
}

type event struct {
	Type           string                    `json:"type"`
	Timestamp      time.Time                 `json:"ts"`
	SessionID      string                    `json:"session_id,omitempty"`
	Message        *types.Message            `json:"message,omitempty"`
	Boundary       *types.CompactionBoundary `json:"boundary,omitempty"`
	ProjectPath    string                    `json:"project_path,omitempty"`
	RuntimeSurface string                    `json:"runtime_surface,omitempty"`
	CommandSurface string                    `json:"command_surface,omitempty"`
	Provider       string                    `json:"provider,omitempty"`
	Model          string                    `json:"model,omitempty"`
}

type SessionStartOptions struct {
	ProjectPath    string
	RuntimeSurface string
	CommandSurface string
	Provider       string
	Model          string
}

// SessionPath returns the transcript path for a session ID.
func SessionPath(rootDir, sessionID string) string {
	return filepath.Join(rootDir, defaultSessionsDir, sessionID+".jsonl")
}

// Create initializes a new persistent session transcript.
func Create(rootDir, sessionID string) (*Store, error) {
	return CreateWithOptions(rootDir, sessionID, SessionStartOptions{})
}

func CreateWithOptions(rootDir, sessionID string, opts SessionStartOptions) (*Store, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, fmt.Errorf("session ID cannot be empty")
	}
	path := SessionPath(rootDir, sessionID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create sessions directory: %w", err)
	}
	if _, err := os.Stat(path); err == nil {
		return nil, fmt.Errorf("session already exists: %s", path)
	}
	store := &Store{path: path, sessionID: sessionID}
	if err := store.append(event{
		Type:           eventSessionStart,
		SessionID:      sessionID,
		ProjectPath:    strings.TrimSpace(opts.ProjectPath),
		RuntimeSurface: strings.TrimSpace(opts.RuntimeSurface),
		CommandSurface: strings.TrimSpace(opts.CommandSurface),
		Provider:       strings.TrimSpace(opts.Provider),
		Model:          strings.TrimSpace(opts.Model),
	}); err != nil {
		return nil, err
	}
	return store, nil
}

// OpenByPath opens an existing transcript for appends.
func OpenByPath(path string, sessionID string) (*Store, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("open session transcript: %w", err)
	}
	return &Store{path: path, sessionID: sessionID}, nil
}

// OpenByID opens an existing transcript by session ID.
func OpenByID(rootDir, sessionID string) (*Store, error) {
	return OpenByPath(SessionPath(rootDir, sessionID), sessionID)
}

// SessionID returns the store session ID.
func (s *Store) SessionID() string {
	return s.sessionID
}

// Path returns the JSONL path for this store.
func (s *Store) Path() string {
	return s.path
}

// AppendMessage persists a new message as an append-only event.
func (s *Store) AppendMessage(msg types.Message) error {
	copyMsg := msg
	return s.append(event{Type: eventMessageAppend, Message: &copyMsg})
}

// AppendCompactionBoundary persists a transcript compaction operation.
func (s *Store) AppendCompactionBoundary(boundary types.CompactionBoundary) error {
	copyBoundary := boundary
	return s.append(event{Type: eventCompaction, Boundary: &copyBoundary})
}

func (s *Store) append(ev event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ev.Timestamp = time.Now().UTC()
	if ev.SessionID == "" {
		ev.SessionID = s.sessionID
	}

	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open transcript for append: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	if err := enc.Encode(ev); err != nil {
		return fmt.Errorf("append transcript event: %w", err)
	}

	return nil
}

// LoadByID loads and replays a session transcript by session ID.
func LoadByID(rootDir, sessionID string) (*Transcript, error) {
	path := SessionPath(rootDir, sessionID)
	return LoadByPath(path)
}

// LoadByPath loads and replays a session transcript by explicit path.
func LoadByPath(path string) (*Transcript, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open transcript: %w", err)
	}
	defer f.Close()

	out := &Transcript{Path: path}
	scanner := bufio.NewScanner(f)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var ev event
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			return nil, fmt.Errorf("decode transcript line %d: %w", lineNo, err)
		}

		out.EventCount++
		if out.FirstEventAt.IsZero() || ev.Timestamp.Before(out.FirstEventAt) {
			out.FirstEventAt = ev.Timestamp
		}
		if ev.Timestamp.After(out.LastEventAt) {
			out.LastEventAt = ev.Timestamp
		}
		if p := strings.TrimSpace(ev.ProjectPath); p != "" {
			out.ProjectPath = p
		}
		if p := strings.TrimSpace(ev.RuntimeSurface); p != "" {
			out.RuntimeSurface = p
		}
		if p := strings.TrimSpace(ev.CommandSurface); p != "" {
			out.CommandSurface = p
		}
		if p := strings.TrimSpace(ev.Provider); p != "" {
			out.Provider = p
		}
		if p := strings.TrimSpace(ev.Model); p != "" {
			out.Model = p
		}

		switch ev.Type {
		case eventSessionStart:
			if ev.SessionID != "" {
				out.SessionID = ev.SessionID
			}

		case eventMessageAppend:
			if ev.Message == nil {
				return nil, fmt.Errorf("transcript line %d: missing message", lineNo)
			}
			out.Messages = append(out.Messages, *ev.Message)
			out.MessageCount++
			out.LastMessageRole = string(ev.Message.Role)

		case eventCompaction:
			if ev.Boundary == nil {
				return nil, fmt.Errorf("transcript line %d: missing compaction boundary", lineNo)
			}
			next, err := ApplyCompactionBoundary(out.Messages, *ev.Boundary)
			if err != nil {
				return nil, fmt.Errorf("transcript line %d: %w", lineNo, err)
			}
			out.Messages = next
			out.Boundaries = append(out.Boundaries, *ev.Boundary)

		default:
			// Ignore unknown events for forward compatibility.
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan transcript: %w", err)
	}

	if out.SessionID == "" {
		base := filepath.Base(path)
		out.SessionID = strings.TrimSuffix(base, filepath.Ext(base))
	}

	return out, nil
}

func LoadMetadataByPath(path string) (*Transcript, error) {
	tr, err := LoadByPath(path)
	if err != nil {
		return nil, err
	}
	tr.Messages = nil
	tr.Boundaries = nil
	return tr, nil
}

// ApplyCompactionBoundary applies one compaction boundary to a message slice.
func ApplyCompactionBoundary(messages []types.Message, boundary types.CompactionBoundary) ([]types.Message, error) {
	if boundary.StartIndex < 0 || boundary.EndIndex < boundary.StartIndex || boundary.EndIndex > len(messages) {
		return nil, fmt.Errorf("invalid compaction boundary [%d:%d] for transcript len=%d", boundary.StartIndex, boundary.EndIndex, len(messages))
	}

	next := make([]types.Message, 0, len(messages)-(boundary.EndIndex-boundary.StartIndex)+len(boundary.Replacement))
	next = append(next, messages[:boundary.StartIndex]...)
	next = append(next, boundary.Replacement...)
	next = append(next, messages[boundary.EndIndex:]...)

	return next, nil
}
