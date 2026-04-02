package session

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Repository struct {
	rootDir string
}

type Query struct {
	ProjectContains string
	SessionContains string
	Provider        string
	Model           string
	RuntimeSurface  string
	CommandSurface  string
	LastRole        string
	MinMessages     int
	Limit           int
}

func NewRepository(rootDir string) *Repository {
	return &Repository{rootDir: rootDir}
}

func (r *Repository) Create(sessionID string, opts SessionStartOptions) (*Store, error) {
	return CreateWithOptions(r.rootDir, sessionID, opts)
}

func (r *Repository) Open(sessionID string) (*Store, error) {
	return OpenByID(r.rootDir, sessionID)
}

func (r *Repository) Load(sessionID string) (*Transcript, error) {
	return LoadByID(r.rootDir, sessionID)
}

func (r *Repository) Query(query Query) ([]Transcript, error) {
	dir := filepath.Join(r.rootDir, defaultSessionsDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read sessions directory: %w", err)
	}

	projectContains := strings.ToLower(strings.TrimSpace(query.ProjectContains))
	sessionContains := strings.ToLower(strings.TrimSpace(query.SessionContains))
	provider := strings.ToLower(strings.TrimSpace(query.Provider))
	model := strings.ToLower(strings.TrimSpace(query.Model))
	runtimeSurface := strings.ToLower(strings.TrimSpace(query.RuntimeSurface))
	commandSurface := strings.ToLower(strings.TrimSpace(query.CommandSurface))
	lastRole := strings.ToLower(strings.TrimSpace(query.LastRole))

	out := make([]Transcript, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".jsonl" {
			continue
		}
		sessionID := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		if sessionContains != "" && !strings.Contains(strings.ToLower(sessionID), sessionContains) {
			continue
		}
		transcript, err := LoadMetadataByPath(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		if projectContains != "" && !strings.Contains(strings.ToLower(strings.TrimSpace(transcript.ProjectPath)), projectContains) {
			continue
		}
		if provider != "" && strings.ToLower(strings.TrimSpace(transcript.Provider)) != provider {
			continue
		}
		if model != "" && strings.ToLower(strings.TrimSpace(transcript.Model)) != model {
			continue
		}
		if runtimeSurface != "" && strings.ToLower(strings.TrimSpace(transcript.RuntimeSurface)) != runtimeSurface {
			continue
		}
		if commandSurface != "" && strings.ToLower(strings.TrimSpace(transcript.CommandSurface)) != commandSurface {
			continue
		}
		if lastRole != "" && strings.ToLower(strings.TrimSpace(transcript.LastMessageRole)) != lastRole {
			continue
		}
		if query.MinMessages > 0 && transcript.MessageCount < query.MinMessages {
			continue
		}
		out = append(out, *transcript)
	}

	sort.SliceStable(out, func(i, j int) bool {
		if !out[i].LastEventAt.Equal(out[j].LastEventAt) {
			return out[i].LastEventAt.After(out[j].LastEventAt)
		}
		return out[i].SessionID < out[j].SessionID
	})

	if query.Limit > 0 && len(out) > query.Limit {
		out = out[:query.Limit]
	}
	return out, nil
}

func (r *Repository) QuerySnapshots(query Query, snapshotQuery SnapshotQuery) ([]Snapshot, error) {
	items, err := r.Query(query)
	if err != nil {
		return nil, err
	}
	return BuildSnapshots(items, snapshotQuery), nil
}
