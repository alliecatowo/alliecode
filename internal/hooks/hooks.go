// Package hooks implements the event hook system for AllieCode.
// Hooks run shell commands in response to agent events (pre/post tool, chat, etc.).
package hooks

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const maxHookTimeout = 5 * time.Minute
const minHookTimeout = 250 * time.Millisecond

// Event identifies when a hook should fire.
type Event string

const (
	EventPreTool    Event = "pre_tool"
	EventPostTool   Event = "post_tool"
	EventPreChat    Event = "pre_chat"
	EventPostChat   Event = "post_chat"
	EventTurnStart  Event = "turn_start"
	EventTurnEnd    Event = "turn_end"
	EventRetry      Event = "retry"
	EventCompaction Event = "compaction"
	EventResume     Event = "resume"
	EventStop       Event = "stop"
)

// Runner is the hook execution interface used by the agent loop.
type Runner interface {
	Fire(ctx context.Context, event Event, data map[string]string) error
	HasHooks(event Event) bool
}

// HookConfig defines a hook in the configuration.
type HookConfig struct {
	Event          string `yaml:"event"`
	Command        string `yaml:"command"`
	Timeout        int    `yaml:"timeout"` // milliseconds, 0 = default (30s)
	MaxPayloadSize int    `yaml:"max_payload_size"`
	RetryOnce      bool   `yaml:"retry_once"`
}

// Hook is a resolved hook ready to fire.
type Hook struct {
	Event          Event
	Command        string
	Timeout        time.Duration
	MaxPayloadSize int
	RetryOnce      bool
}

// Manager manages and fires event hooks.
type Manager struct {
	hooks []Hook
}

// NewManager creates a hook manager from configuration.
func NewManager(configs []HookConfig) *Manager {
	hooks := make([]Hook, 0, len(configs))
	for _, cfg := range configs {
		if strings.TrimSpace(cfg.Command) == "" {
			continue
		}
		event := Event(strings.TrimSpace(cfg.Event))
		if !isKnownEvent(event) {
			continue
		}
		timeout := 30 * time.Second
		if cfg.Timeout > 0 {
			timeout = time.Duration(cfg.Timeout) * time.Millisecond
		}
		if timeout < minHookTimeout {
			timeout = minHookTimeout
		}
		if timeout > maxHookTimeout {
			timeout = maxHookTimeout
		}
		maxPayloadSize := cfg.MaxPayloadSize
		if maxPayloadSize <= 0 {
			maxPayloadSize = defaultMaxPayloadBytes
		}
		hooks = append(hooks, Hook{
			Event:          event,
			Command:        cfg.Command,
			Timeout:        timeout,
			MaxPayloadSize: maxPayloadSize,
			RetryOnce:      cfg.RetryOnce,
		})
	}
	return &Manager{hooks: hooks}
}

// Fire runs all hooks matching the given event.
// Data is passed as environment variables with the ALLIECODE_ prefix.
func (m *Manager) Fire(ctx context.Context, event Event, data map[string]string) error {
	for _, hook := range m.hooks {
		if hook.Event != event {
			continue
		}

		hookCtx, cancel := context.WithTimeout(ctx, hook.Timeout)
		err := runHookWithOptions(hookCtx, hook, data)
		cancel()

		if err != nil {
			return fmt.Errorf("hook %q for event %s: %w", hook.Command, event, err)
		}
	}
	return nil
}

// runHook executes a single hook command with environment variables.
func runHook(ctx context.Context, event Event, command string, data map[string]string) error {
	return runHookWithOptions(ctx, Hook{Event: event, Command: command, Timeout: 30 * time.Second, MaxPayloadSize: defaultMaxPayloadBytes}, data)
}

func runHookWithOptions(ctx context.Context, hook Hook, data map[string]string) error {
	attemptErr := runHookOnce(ctx, hook, data)
	if attemptErr == nil {
		return nil
	}
	if !hook.RetryOnce || !canRetryHookSafely(hook.Event, attemptErr) {
		return attemptErr
	}
	return runHookOnce(ctx, hook, data)
}

func runHookOnce(ctx context.Context, hook Hook, data map[string]string) error {
	command := hook.Command
	if strings.ContainsRune(command, '\x00') {
		return fmt.Errorf("invalid hook command")
	}
	cmd := exec.CommandContext(ctx, "sh", "-c", command)

	// Inherit current environment
	cmd.Env = os.Environ()

	// Add hook-specific environment variables
	payload := BuildPayloadGuarded(hook.Event, data, PayloadGuard{MaxTotalBytes: hook.MaxPayloadSize})
	cmd.Env = append(cmd.Env, "ALLIECODE_HOOK_SCHEMA=v2")
	for key, value := range payload {
		cmd.Env = append(cmd.Env, fmt.Sprintf("ALLIECODE_%s=%s", key, value))
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func canRetryHookSafely(event Event, err error) bool {
	if errors.Is(err, context.Canceled) {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		switch event {
		case EventPreChat, EventPostChat, EventTurnStart, EventTurnEnd, EventRetry, EventCompaction, EventResume, EventStop:
			return true
		default:
			return false
		}
	}
	return false
}

// HasHooks returns true if there are any hooks registered for the given event.
func (m *Manager) HasHooks(event Event) bool {
	for _, hook := range m.hooks {
		if hook.Event == event {
			return true
		}
	}
	return false
}

func isKnownEvent(event Event) bool {
	switch event {
	case EventPreTool, EventPostTool, EventPreChat, EventPostChat, EventTurnStart, EventTurnEnd, EventRetry, EventCompaction, EventResume, EventStop:
		return true
	default:
		return false
	}
}
