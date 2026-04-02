package permissions

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type StoredRule struct {
	ID string `json:"id"`
	Rule
}

type ruleStoreFile struct {
	Rules []StoredRule `json:"rules"`
}

type SandboxSettings struct {
	Mode             string   `json:"mode"`
	WorkspaceLocked  bool     `json:"workspace_locked"`
	ExcludedCommands []string `json:"excluded_commands,omitempty"`
}

type SandboxPolicySummary struct {
	Mode            string
	ModeAlias       string
	WorkspaceLocked bool
	ExcludedCount   int
}

const RuleSourcePrecedence = "policy>user>project>session"

type sandboxSettingsFile struct {
	Sandbox SandboxSettings `json:"sandbox"`
}

type RuleStore struct {
	path string
	mu   sync.Mutex
}

type SandboxSettingsStore struct {
	path string
	mu   sync.Mutex
}

func NewRuleStore(path string) *RuleStore {
	return &RuleStore{path: path}
}

func NewSandboxSettingsStore(path string) *SandboxSettingsStore {
	return &SandboxSettingsStore{path: path}
}

func DefaultSandboxSettings() SandboxSettings {
	return SandboxSettings{Mode: "workspace-write"}
}

func NormalizeSandboxSettings(settings SandboxSettings) SandboxSettings {
	settings.Mode = strings.ToLower(strings.TrimSpace(settings.Mode))
	if settings.Mode == "" {
		settings.Mode = "workspace-write"
	}
	if settings.Mode != "read-only" && settings.Mode != "workspace-write" && settings.Mode != "danger-full-access" {
		settings.Mode = "workspace-write"
	}
	settings.ExcludedCommands = uniqueSortedNonEmpty(settings.ExcludedCommands)
	return settings
}

func SandboxModeAlias(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "read-only":
		return "restrictive"
	case "danger-full-access":
		return "permissive"
	default:
		return "balanced"
	}
}

func EffectiveSandboxPolicySummary(settings SandboxSettings) SandboxPolicySummary {
	settings = NormalizeSandboxSettings(settings)
	return SandboxPolicySummary{
		Mode:            settings.Mode,
		ModeAlias:       SandboxModeAlias(settings.Mode),
		WorkspaceLocked: settings.WorkspaceLocked,
		ExcludedCount:   len(settings.ExcludedCommands),
	}
}

func (s *SandboxSettingsStore) Load() (SandboxSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	loaded, err := s.readLocked()
	if err != nil {
		return SandboxSettings{}, err
	}
	return loaded, nil
}

func (s *SandboxSettingsStore) Save(settings SandboxSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	settings = NormalizeSandboxSettings(settings)
	return s.writeLocked(settings)
}

func (s *SandboxSettingsStore) readLocked() (SandboxSettings, error) {
	if strings.TrimSpace(s.path) == "" {
		return DefaultSandboxSettings(), nil
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultSandboxSettings(), nil
		}
		return SandboxSettings{}, err
	}
	if len(data) == 0 {
		return DefaultSandboxSettings(), nil
	}
	var file sandboxSettingsFile
	if err := json.Unmarshal(data, &file); err != nil {
		return SandboxSettings{}, err
	}
	return NormalizeSandboxSettings(file.Sandbox), nil
}

func (s *SandboxSettingsStore) writeLocked(settings SandboxSettings) error {
	if strings.TrimSpace(s.path) == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(sandboxSettingsFile{Sandbox: settings}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, payload, 0o644)
}

func (s *RuleStore) List() ([]StoredRule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	loaded, err := s.readLocked()
	if err != nil {
		return nil, err
	}
	sortStoredRules(loaded)
	out := make([]StoredRule, len(loaded))
	copy(out, loaded)
	return out, nil
}

func (s *RuleStore) Create(rule Rule) (StoredRule, error) {
	rule = normalizeRule(rule)
	if err := validateRule(rule); err != nil {
		return StoredRule{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	loaded, err := s.readLocked()
	if err != nil {
		return StoredRule{}, err
	}
	if hasDuplicateRule(loaded, "", rule) {
		return StoredRule{}, fmt.Errorf("duplicate rule already exists")
	}

	entry := StoredRule{ID: newRuleID(), Rule: rule}
	loaded = append(loaded, entry)
	sortStoredRules(loaded)
	if err := s.writeLocked(loaded); err != nil {
		return StoredRule{}, err
	}
	return entry, nil
}

func (s *RuleStore) Get(id string) (StoredRule, bool, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return StoredRule{}, false, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	loaded, err := s.readLocked()
	if err != nil {
		return StoredRule{}, false, err
	}
	for _, entry := range loaded {
		if entry.ID == id {
			return entry, true, nil
		}
	}
	return StoredRule{}, false, nil
}

func (s *RuleStore) Update(id string, rule Rule) (StoredRule, bool, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return StoredRule{}, false, fmt.Errorf("rule id cannot be empty")
	}

	rule = normalizeRule(rule)
	if err := validateRule(rule); err != nil {
		return StoredRule{}, false, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	loaded, err := s.readLocked()
	if err != nil {
		return StoredRule{}, false, err
	}
	if hasDuplicateRule(loaded, id, rule) {
		return StoredRule{}, false, fmt.Errorf("duplicate rule already exists")
	}

	for i := range loaded {
		if loaded[i].ID != id {
			continue
		}
		loaded[i].Rule = rule
		sortStoredRules(loaded)
		if err := s.writeLocked(loaded); err != nil {
			return StoredRule{}, false, err
		}
		for _, entry := range loaded {
			if entry.ID == id {
				return entry, true, nil
			}
		}
		break
	}

	return StoredRule{}, false, nil
}

func (s *RuleStore) Delete(id string) (bool, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return false, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	loaded, err := s.readLocked()
	if err != nil {
		return false, err
	}

	for i := range loaded {
		if loaded[i].ID != id {
			continue
		}
		next := append(loaded[:i], loaded[i+1:]...)
		if err := s.writeLocked(next); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func validateRule(rule Rule) error {
	source := normalizeRuleSource(rule.Source)
	if source != RuleSourcePolicy && source != RuleSourceUser && source != RuleSourceProject && source != RuleSourceSession {
		return fmt.Errorf("invalid source %q", rule.Source)
	}
	if strings.TrimSpace(rule.Tool) == "" && strings.TrimSpace(rule.FileGlob) == "" && strings.TrimSpace(rule.BashRegex) == "" {
		return fmt.Errorf("rule must declare at least one matcher")
	}
	if rule.Tool != "" {
		if err := validateMatcher(rule.Tool, matchModeGlob); err != nil {
			return fmt.Errorf("invalid tool matcher: %w", err)
		}
	}
	if rule.FileGlob != "" {
		if err := validateMatcher(rule.FileGlob, matchModeGlob); err != nil {
			return fmt.Errorf("invalid file_glob matcher: %w", err)
		}
	}
	if rule.BashRegex != "" {
		if err := validateMatcher(rule.BashRegex, matchModeRegex); err != nil {
			return fmt.Errorf("invalid bash_regex matcher: %w", err)
		}
	}
	if rule.Decision < DecisionAllow || rule.Decision > DecisionAsk {
		return fmt.Errorf("invalid decision value %d", rule.Decision)
	}
	return nil
}

func (s *RuleStore) readLocked() ([]StoredRule, error) {
	if strings.TrimSpace(s.path) == "" {
		return nil, nil
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	var f ruleStoreFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	if len(f.Rules) == 0 {
		return nil, nil
	}
	idSeen := make(map[string]struct{}, len(f.Rules))
	for i := range f.Rules {
		f.Rules[i].Rule = normalizeRule(f.Rules[i].Rule)
		f.Rules[i].ID = strings.TrimSpace(f.Rules[i].ID)
		if f.Rules[i].ID == "" {
			f.Rules[i].ID = newRuleID()
		}
		if _, exists := idSeen[f.Rules[i].ID]; exists {
			return nil, fmt.Errorf("duplicate rule id %q", f.Rules[i].ID)
		}
		idSeen[f.Rules[i].ID] = struct{}{}
	}
	return f.Rules, nil
}

func (s *RuleStore) writeLocked(rules []StoredRule) error {
	if strings.TrimSpace(s.path) == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(ruleStoreFile{Rules: rules}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, payload, 0o644)
}

type DenialEntry struct {
	Timestamp time.Time       `json:"ts"`
	Tool      string          `json:"tool"`
	Input     json.RawMessage `json:"input,omitempty"`
	Reason    string          `json:"reason,omitempty"`
}

type DenialsLedger struct {
	path string
	mu   sync.Mutex
}

type DenialQuery struct {
	Tool            string
	ReasonContains  string
	CommandContains string
	Since           time.Time
	Until           time.Time
	SortBy          string
	Desc            bool
	Limit           int
}

func NewDenialsLedger(path string) *DenialsLedger {
	return &DenialsLedger{path: path}
}

func (l *DenialsLedger) Append(entry DenialEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if strings.TrimSpace(l.path) == "" {
		return nil
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}
	if err := os.MkdirAll(filepath.Dir(l.path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(entry)
}

func (l *DenialsLedger) List() ([]DenialEntry, error) {
	return l.Query(DenialQuery{SortBy: "ts", Desc: true})
}

func (l *DenialsLedger) Query(query DenialQuery) ([]DenialEntry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

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

	scanner := bufio.NewScanner(f)
	entries := make([]DenialEntry, 0, 32)
	line := 0
	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		var entry DenialEntry
		if err := json.Unmarshal([]byte(text), &entry); err != nil {
			return nil, fmt.Errorf("decode denial entry line %d: %w", line, err)
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	filtered := filterDenials(entries, query)
	sortDenials(filtered, query)
	if query.Limit > 0 && len(filtered) > query.Limit {
		filtered = filtered[:query.Limit]
	}
	return filtered, nil
}

func normalizeRule(rule Rule) Rule {
	rule.Source = normalizeRuleSource(rule.Source)
	rule.Tool = strings.TrimSpace(rule.Tool)
	rule.FileGlob = strings.TrimSpace(rule.FileGlob)
	rule.BashRegex = strings.TrimSpace(rule.BashRegex)
	return rule
}

func hasDuplicateRule(existing []StoredRule, excludeID string, candidate Rule) bool {
	for _, entry := range existing {
		if entry.ID == excludeID {
			continue
		}
		if rulesEqual(entry.Rule, candidate) {
			return true
		}
	}
	return false
}

func rulesEqual(a, b Rule) bool {
	a = normalizeRule(a)
	b = normalizeRule(b)
	return a.Source == b.Source && a.Tool == b.Tool && a.FileGlob == b.FileGlob && a.BashRegex == b.BashRegex && a.Decision == b.Decision
}

func sortStoredRules(rules []StoredRule) {
	sort.SliceStable(rules, func(i, j int) bool {
		if rules[i].ID != rules[j].ID {
			return rules[i].ID < rules[j].ID
		}
		if rules[i].Source != rules[j].Source {
			return rules[i].Source < rules[j].Source
		}
		if rules[i].Tool != rules[j].Tool {
			return rules[i].Tool < rules[j].Tool
		}
		if rules[i].FileGlob != rules[j].FileGlob {
			return rules[i].FileGlob < rules[j].FileGlob
		}
		if rules[i].BashRegex != rules[j].BashRegex {
			return rules[i].BashRegex < rules[j].BashRegex
		}
		return rules[i].Decision < rules[j].Decision
	})
}

func filterDenials(entries []DenialEntry, query DenialQuery) []DenialEntry {
	if len(entries) == 0 {
		return nil
	}
	tool := strings.ToLower(strings.TrimSpace(query.Tool))
	reasonPart := strings.ToLower(strings.TrimSpace(query.ReasonContains))
	commandPart := strings.ToLower(strings.TrimSpace(query.CommandContains))
	out := make([]DenialEntry, 0, len(entries))
	for _, entry := range entries {
		if tool != "" && strings.ToLower(strings.TrimSpace(entry.Tool)) != tool {
			continue
		}
		if !query.Since.IsZero() && entry.Timestamp.Before(query.Since) {
			continue
		}
		if !query.Until.IsZero() && entry.Timestamp.After(query.Until) {
			continue
		}
		if reasonPart != "" && !strings.Contains(strings.ToLower(entry.Reason), reasonPart) {
			continue
		}
		if commandPart != "" {
			command := strings.ToLower(extractDenialCommand(entry.Input))
			if !strings.Contains(command, commandPart) {
				continue
			}
		}
		out = append(out, entry)
	}
	return out
}

func sortDenials(entries []DenialEntry, query DenialQuery) {
	if len(entries) == 0 {
		return
	}
	sortBy := strings.ToLower(strings.TrimSpace(query.SortBy))
	if sortBy == "" {
		sortBy = "ts"
	}
	desc := query.Desc
	if strings.TrimSpace(query.SortBy) == "" {
		desc = true
	}
	sort.SliceStable(entries, func(i, j int) bool {
		left := entries[i]
		right := entries[j]

		cmp := 0
		switch sortBy {
		case "tool":
			cmp = strings.Compare(strings.ToLower(left.Tool), strings.ToLower(right.Tool))
		case "reason":
			cmp = strings.Compare(strings.ToLower(left.Reason), strings.ToLower(right.Reason))
		case "command":
			cmp = strings.Compare(strings.ToLower(extractDenialCommand(left.Input)), strings.ToLower(extractDenialCommand(right.Input)))
		default:
			if left.Timestamp.Before(right.Timestamp) {
				cmp = -1
			} else if left.Timestamp.After(right.Timestamp) {
				cmp = 1
			}
		}

		if cmp == 0 {
			if left.Timestamp.Before(right.Timestamp) {
				cmp = -1
			} else if left.Timestamp.After(right.Timestamp) {
				cmp = 1
			} else if left.Tool != right.Tool {
				cmp = strings.Compare(strings.ToLower(left.Tool), strings.ToLower(right.Tool))
			} else if left.Reason != right.Reason {
				cmp = strings.Compare(strings.ToLower(left.Reason), strings.ToLower(right.Reason))
			} else {
				cmp = strings.Compare(strings.ToLower(extractDenialCommand(left.Input)), strings.ToLower(extractDenialCommand(right.Input)))
			}
		}

		if desc {
			return cmp > 0
		}
		return cmp < 0
	})
}

func extractDenialCommand(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var payload struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}
	return strings.TrimSpace(payload.Command)
}

func newRuleID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("rule-%d", time.Now().UTC().UnixNano())
	}
	return "rule-" + hex.EncodeToString(buf)
}

func uniqueSortedNonEmpty(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	uniq := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		uniq[value] = struct{}{}
	}
	out := make([]string, 0, len(uniq))
	for value := range uniq {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
