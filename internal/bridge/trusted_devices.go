package bridge

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// TrustedDevice represents one approved remote bridge device.
type TrustedDevice struct {
	DeviceID   string    `json:"device_id"`
	Label      string    `json:"label,omitempty"`
	AddedAt    time.Time `json:"added_at"`
	LastSeenAt time.Time `json:"last_seen_at,omitempty"`
}

// TrustedDeviceRegistry stores trusted devices and persists optionally.
type TrustedDeviceRegistry struct {
	mu      sync.RWMutex
	devices map[string]TrustedDevice
	path    string
}

// NewTrustedDeviceRegistry creates registry and loads persisted devices if path is set.
func NewTrustedDeviceRegistry(path string) (*TrustedDeviceRegistry, error) {
	r := &TrustedDeviceRegistry{
		devices: make(map[string]TrustedDevice),
		path:    path,
	}
	if err := r.load(); err != nil {
		return nil, err
	}
	return r, nil
}

// Add inserts or updates one trusted device.
func (r *TrustedDeviceRegistry) Add(device TrustedDevice) (TrustedDevice, error) {
	normalizedID := strings.TrimSpace(device.DeviceID)
	if normalizedID == "" {
		return TrustedDevice{}, errors.New("device ID cannot be empty")
	}
	device.DeviceID = normalizedID
	if device.AddedAt.IsZero() {
		device.AddedAt = time.Now().UTC()
	}

	r.mu.Lock()
	r.devices[device.DeviceID] = device
	r.mu.Unlock()

	if err := r.save(); err != nil {
		return TrustedDevice{}, err
	}
	return device, nil
}

// Revoke removes one device by ID.
func (r *TrustedDeviceRegistry) Revoke(deviceID string) error {
	normalizedID := strings.TrimSpace(deviceID)
	if normalizedID == "" {
		return errors.New("device ID cannot be empty")
	}
	r.mu.Lock()
	delete(r.devices, normalizedID)
	r.mu.Unlock()
	return r.save()
}

// List returns all trusted devices sorted by device ID.
func (r *TrustedDeviceRegistry) List() []TrustedDevice {
	r.mu.RLock()
	out := make([]TrustedDevice, 0, len(r.devices))
	for _, device := range r.devices {
		out = append(out, device)
	}
	r.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool {
		return out[i].DeviceID < out[j].DeviceID
	})
	return out
}

// IsTrusted reports whether device exists in registry.
func (r *TrustedDeviceRegistry) IsTrusted(deviceID string) bool {
	normalizedID := strings.TrimSpace(deviceID)
	if normalizedID == "" {
		return false
	}
	r.mu.RLock()
	_, ok := r.devices[normalizedID]
	r.mu.RUnlock()
	return ok
}

func (r *TrustedDeviceRegistry) load() error {
	if r.path == "" {
		return nil
	}
	raw, err := os.ReadFile(r.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read trusted devices: %w", err)
	}
	if len(raw) == 0 {
		return nil
	}
	var devices []TrustedDevice
	if err := json.Unmarshal(raw, &devices); err != nil {
		return fmt.Errorf("parse trusted devices: %w", err)
	}
	for _, device := range devices {
		normalizedID := strings.TrimSpace(device.DeviceID)
		if normalizedID == "" {
			continue
		}
		device.DeviceID = normalizedID
		r.devices[normalizedID] = device
	}
	return nil
}

func (r *TrustedDeviceRegistry) save() error {
	if r.path == "" {
		return nil
	}
	r.mu.RLock()
	devices := make([]TrustedDevice, 0, len(r.devices))
	for _, device := range r.devices {
		devices = append(devices, device)
	}
	r.mu.RUnlock()
	sort.Slice(devices, func(i, j int) bool {
		return devices[i].DeviceID < devices[j].DeviceID
	})

	raw, err := json.MarshalIndent(devices, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal trusted devices: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(r.path), 0o755); err != nil {
		return fmt.Errorf("prepare trusted device directory: %w", err)
	}
	tmp := r.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return fmt.Errorf("write trusted devices: %w", err)
	}
	if err := os.Rename(tmp, r.path); err != nil {
		return fmt.Errorf("replace trusted devices: %w", err)
	}
	return nil
}
