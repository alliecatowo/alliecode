package migrations

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
)

type Migration struct {
	Version     string
	Description string
	Run         func() error
}

type LedgerEntry struct {
	Version     string    `json:"version"`
	Description string    `json:"description,omitempty"`
	AppliedAt   time.Time `json:"applied_at"`
	DurationMs  int64     `json:"duration_ms,omitempty"`
	Status      string    `json:"status,omitempty"`
}

type RunDiagnostics struct {
	AppliedVersions []string
	SkippedVersions []string
	FailedVersion   string
	DurationMs      int64
	AppliedCount    int
	SkippedCount    int
}

type RunReport struct {
	AppliedCount    int      `json:"applied_count"`
	SkippedCount    int      `json:"skipped_count"`
	FailedVersion   string   `json:"failed_version,omitempty"`
	AppliedVersions []string `json:"applied_versions,omitempty"`
	SkippedVersions []string `json:"skipped_versions,omitempty"`
	DurationMs      int64    `json:"duration_ms"`
	LastApplied     string   `json:"last_applied,omitempty"`
	HasFailure      bool     `json:"has_failure"`
}

type Ledger struct {
	path string
	mu   sync.Mutex
}

func NewLedger(path string) *Ledger {
	return &Ledger{path: path}
}

func (l *Ledger) Entries() ([]LedgerEntry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.readLocked()
}

func (l *Ledger) Has(version string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	entries, err := l.readLocked()
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		if entry.Version == version {
			return true, nil
		}
	}
	return false, nil
}

func (l *Ledger) Record(version, description string) error {
	return l.RecordDetailed(LedgerEntry{
		Version:     version,
		Description: description,
		AppliedAt:   time.Now().UTC(),
		Status:      "applied",
	})
}

func (l *Ledger) RecordDetailed(entry LedgerEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if strings.TrimSpace(l.path) == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(l.path), 0o755); err != nil {
		return err
	}

	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	entry.Version = strings.TrimSpace(entry.Version)
	entry.Description = strings.TrimSpace(entry.Description)
	entry.Status = strings.TrimSpace(entry.Status)
	if entry.Status == "" {
		entry.Status = "applied"
	}
	if entry.AppliedAt.IsZero() {
		entry.AppliedAt = time.Now().UTC()
	}
	if entry.DurationMs < 0 {
		entry.DurationMs = 0
	}
	return json.NewEncoder(f).Encode(entry)
}

func (l *Ledger) readLocked() ([]LedgerEntry, error) {
	if strings.TrimSpace(l.path) == "" {
		return nil, nil
	}
	f, err := os.Open(l.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	entries := make([]LedgerEntry, 0, 16)
	scanner := bufio.NewScanner(f)
	line := 0
	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		var entry LedgerEntry
		if err := json.Unmarshal([]byte(text), &entry); err != nil {
			return nil, fmt.Errorf("decode migration ledger line %d: %w", line, err)
		}
		entry.Version = strings.TrimSpace(entry.Version)
		entry.Description = strings.TrimSpace(entry.Description)
		entry.Status = strings.TrimSpace(entry.Status)
		if entry.Status == "" {
			entry.Status = "applied"
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func (l *Ledger) LastVersion() (string, error) {
	entries, err := l.Entries()
	if err != nil {
		return "", err
	}
	if len(entries) == 0 {
		return "", nil
	}
	return entries[len(entries)-1].Version, nil
}

func (l *Ledger) Query(prefix string, limit int) ([]LedgerEntry, error) {
	entries, err := l.Entries()
	if err != nil {
		return nil, err
	}
	prefix = strings.TrimSpace(prefix)
	if prefix != "" {
		filtered := make([]LedgerEntry, 0, len(entries))
		for _, entry := range entries {
			if strings.HasPrefix(entry.Version, prefix) {
				filtered = append(filtered, entry)
			}
		}
		entries = filtered
	}
	if limit > 0 && len(entries) > limit {
		entries = entries[len(entries)-limit:]
	}
	return entries, nil
}

type Runner struct {
	ledger *Ledger
}

func NewRunner(ledger *Ledger) *Runner {
	return &Runner{ledger: ledger}
}

func (r *Runner) RunAll(migrations []Migration) ([]string, error) {
	applied, _, err := r.RunAllDetailed(migrations)
	return applied, err
}

func (r *Runner) RunAllDetailed(migrations []Migration) ([]string, RunDiagnostics, error) {
	if r.ledger == nil {
		return nil, RunDiagnostics{}, fmt.Errorf("runner requires ledger")
	}
	startedAt := time.Now()

	ordered := make([]Migration, len(migrations))
	copy(ordered, migrations)
	sort.SliceStable(ordered, func(i, j int) bool {
		return strings.TrimSpace(ordered[i].Version) < strings.TrimSpace(ordered[j].Version)
	})

	seen := make(map[string]struct{}, len(migrations))
	applied := make([]string, 0, len(migrations))
	skipped := make([]string, 0, len(migrations))

	for _, migration := range ordered {
		version := strings.TrimSpace(migration.Version)
		if version == "" {
			return nil, RunDiagnostics{}, fmt.Errorf("migration version is required")
		}
		if migration.Run == nil {
			return nil, RunDiagnostics{}, fmt.Errorf("migration %q has nil Run function", version)
		}
		if _, ok := seen[version]; ok {
			return nil, RunDiagnostics{}, fmt.Errorf("duplicate migration version %q", version)
		}
		seen[version] = struct{}{}

		has, err := r.ledger.Has(version)
		if err != nil {
			return nil, RunDiagnostics{}, err
		}
		if has {
			skipped = append(skipped, version)
			continue
		}

		runStart := time.Now()
		if err := migration.Run(); err != nil {
			diag := RunDiagnostics{
				AppliedVersions: applied,
				SkippedVersions: skipped,
				FailedVersion:   version,
				DurationMs:      time.Since(startedAt).Milliseconds(),
				AppliedCount:    len(applied),
				SkippedCount:    len(skipped),
			}
			return nil, diag, fmt.Errorf("run migration %q: %w", version, err)
		}
		if err := r.ledger.RecordDetailed(LedgerEntry{
			Version:     version,
			Description: migration.Description,
			AppliedAt:   time.Now().UTC(),
			DurationMs:  time.Since(runStart).Milliseconds(),
			Status:      "applied",
		}); err != nil {
			return nil, RunDiagnostics{}, fmt.Errorf("record migration %q: %w", version, err)
		}
		applied = append(applied, version)
	}

	diag := RunDiagnostics{
		AppliedVersions: applied,
		SkippedVersions: skipped,
		DurationMs:      time.Since(startedAt).Milliseconds(),
		AppliedCount:    len(applied),
		SkippedCount:    len(skipped),
	}
	return applied, diag, nil
}

func BuildRunReport(diag RunDiagnostics) RunReport {
	report := RunReport{
		AppliedCount:    diag.AppliedCount,
		SkippedCount:    diag.SkippedCount,
		FailedVersion:   strings.TrimSpace(diag.FailedVersion),
		AppliedVersions: append([]string(nil), diag.AppliedVersions...),
		SkippedVersions: append([]string(nil), diag.SkippedVersions...),
		DurationMs:      diag.DurationMs,
	}
	if len(diag.AppliedVersions) > 0 {
		report.LastApplied = diag.AppliedVersions[len(diag.AppliedVersions)-1]
	}
	report.HasFailure = report.FailedVersion != ""
	return report
}
