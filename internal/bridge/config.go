package bridge

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const defaultHeartbeatInterval = 30 * time.Second

// BridgeConfig controls bridge feature enablement and runtime behavior.
type BridgeConfig struct {
	Enabled           bool          `json:"enabled"`
	Poll              PollConfig    `json:"poll"`
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`
	ReconnectBackoff  BackoffPolicy `json:"reconnect_backoff"`
}

// DefaultBridgeConfig returns the package defaults.
func DefaultBridgeConfig() BridgeConfig {
	return BridgeConfig{
		Enabled:           false,
		Poll:              PollConfig{},
		HeartbeatInterval: defaultHeartbeatInterval,
		ReconnectBackoff:  BackoffPolicy{},
	}
}

// ValidateBridgeConfig applies defaults and validates bridge config.
func ValidateBridgeConfig(cfg BridgeConfig) (BridgeConfig, error) {
	poll, err := NormalizePollConfig(cfg.Poll)
	if err != nil {
		return BridgeConfig{}, fmt.Errorf("poll config: %w", err)
	}
	backoff, err := NormalizeBackoffPolicy(cfg.ReconnectBackoff)
	if err != nil {
		return BridgeConfig{}, fmt.Errorf("reconnect backoff: %w", err)
	}
	if cfg.HeartbeatInterval < 0 {
		return BridgeConfig{}, errors.New("heartbeat interval cannot be negative")
	}
	if cfg.HeartbeatInterval == 0 {
		cfg.HeartbeatInterval = defaultHeartbeatInterval
	}
	cfg.Poll = poll
	cfg.ReconnectBackoff = backoff
	return cfg, nil
}

// ConfigStore persists bridge config at one file path.
type ConfigStore struct {
	mu   sync.Mutex
	path string
}

// NewConfigStore builds a file-backed bridge config store.
func NewConfigStore(path string) *ConfigStore {
	return &ConfigStore{path: path}
}

// Load reads persisted config. Missing files return defaults.
func (s *ConfigStore) Load() (BridgeConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.path == "" {
		return ValidateBridgeConfig(DefaultBridgeConfig())
	}

	raw, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return ValidateBridgeConfig(DefaultBridgeConfig())
	}
	if err != nil {
		return BridgeConfig{}, fmt.Errorf("read bridge config: %w", err)
	}
	if len(raw) == 0 {
		return ValidateBridgeConfig(DefaultBridgeConfig())
	}

	cfg := DefaultBridgeConfig()
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return BridgeConfig{}, fmt.Errorf("parse bridge config: %w", err)
	}
	return ValidateBridgeConfig(cfg)
}

// Save validates and persists config atomically.
func (s *ConfigStore) Save(cfg BridgeConfig) (BridgeConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	validated, err := ValidateBridgeConfig(cfg)
	if err != nil {
		return BridgeConfig{}, err
	}
	if s.path == "" {
		return validated, nil
	}

	raw, err := json.MarshalIndent(validated, "", "  ")
	if err != nil {
		return BridgeConfig{}, fmt.Errorf("marshal bridge config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return BridgeConfig{}, fmt.Errorf("prepare config directory: %w", err)
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return BridgeConfig{}, fmt.Errorf("write bridge config: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return BridgeConfig{}, fmt.Errorf("replace bridge config: %w", err)
	}
	return validated, nil
}

// SetEnabled toggles bridge enablement and persists the change.
func (s *ConfigStore) SetEnabled(enabled bool) (BridgeConfig, error) {
	current, err := s.Load()
	if err != nil {
		return BridgeConfig{}, err
	}
	current.Enabled = enabled
	return s.Save(current)
}
