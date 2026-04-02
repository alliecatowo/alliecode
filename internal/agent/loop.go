// Package agent implements AllieCode's core agent loop.
// It orchestrates the conversation between the user, the LLM provider, and tools.
package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alliecatowo/alliecode/internal/hooks"
	"github.com/alliecatowo/alliecode/internal/permissions"
	"github.com/alliecatowo/alliecode/internal/providers/streamnorm"
	"github.com/alliecatowo/alliecode/internal/references"
	"github.com/alliecatowo/alliecode/internal/session"
	"github.com/alliecatowo/alliecode/internal/types"
)

// StreamCallback is a function that receives streaming events from the provider.
// The TUI registers a callback to render tokens as they arrive.
type StreamCallback func(types.StreamEvent)

// EventCallback receives structured loop progress events.
type EventCallback func(types.AgentEvent)

// PermissionRequest describes one tool permission check request.
type PermissionRequest struct {
	Turn     int
	ToolUse  types.ContentBlock
	Tool     types.Tool
	ToolCtx  types.ToolContext
	Decision permissions.Decision
}

// PermissionResolver is called when a permission decision requires user input.
type PermissionResolver func(ctx context.Context, req PermissionRequest) permissions.Decision

// Config holds all configuration for the agent loop.
type Config struct {
	Provider     types.Provider
	Tools        []types.Tool
	Model        string
	SystemPrompt string
	MaxTurns     int
	MaxBudgetUSD float64
	WorkingDir   string
	Debug        bool

	// ContextWindow is the model's context window size in tokens.
	// Used for auto-compaction decisions. Defaults to 200000 if zero.
	ContextWindow int

	// MaxTokens is the max output tokens per request. Defaults to 16384 if zero.
	MaxTokens int

	// MaxTokenBudget caps cumulative token usage (input+output). 0 disables it.
	MaxTokenBudget int

	// EnableCheckpoints enables git-based checkpointing before destructive tool calls.
	EnableCheckpoints bool

	// PermissionChecker centralizes permission checks before tool execution.
	PermissionChecker permissions.Checker

	// ResolvePermission handles DecisionAsk outcomes from PermissionChecker.
	ResolvePermission PermissionResolver

	// Hooks integrates loop lifecycle hook execution.
	Hooks hooks.Runner

	// SessionStore enables append-only persistent transcripts.
	SessionStore *session.Store
}

// Agent is the core agent loop that mediates between the user, LLM, and tools.
type Agent struct {
	config   Config
	messages []types.Message
	usage    types.Usage
	toolMap  map[string]types.Tool
	session  *session.Store

	streamCB StreamCallback
	eventCB  EventCallback
	mu       sync.RWMutex // guards messages, usage

	// costPerMInput and costPerMOutput for budget tracking.
	costPerMInput  float64
	costPerMOutput float64
	totalCostUSD   float64

	runtimeMu       sync.RWMutex
	runtimeTurns    int
	runtimeTurnIdx  int
	runtimePhase    types.AgentTurnPhase
	runtimeInflight int
	runtimeLastStop types.AgentStopReason
	runtimeTasks    map[string]string
	runtimeTeams    map[string]string

	eventMu         sync.RWMutex
	eventSeq        uint64
	eventHistory    []types.AgentEvent
	eventHistoryCap int
	lifecycleState  types.AgentLifecycleState
}

const toolMaxAttempts = 2

// New creates a new Agent with the given configuration.
func New(config Config) *Agent {
	if config.MaxTurns == 0 {
		config.MaxTurns = 100
	}
	if config.ContextWindow == 0 {
		config.ContextWindow = 200000
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 16384
	}

	toolMap := make(map[string]types.Tool, len(config.Tools))
	for _, t := range config.Tools {
		toolMap[t.Name()] = t
	}

	return &Agent{
		config:  config,
		toolMap: toolMap,
		session: config.SessionStore,
		// Default cost estimates (Sonnet-class pricing). Overridden if model info is available.
		costPerMInput:   3.0,
		costPerMOutput:  15.0,
		runtimeTasks:    make(map[string]string),
		runtimeTeams:    make(map[string]string),
		eventHistoryCap: 256,
		lifecycleState:  types.AgentLifecycleIdle,
	}
}

// SetStreamCallback registers a callback that receives every stream event.
func (a *Agent) SetStreamCallback(cb StreamCallback) {
	a.streamCB = cb
}

// SetEventCallback registers a callback that receives structured agent loop events.
func (a *Agent) SetEventCallback(cb EventCallback) {
	a.eventCB = cb
}

// SetSessionStore configures persistent transcript storage for this agent.
func (a *Agent) SetSessionStore(store *session.Store) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.session = store
	a.config.SessionStore = store
}

// ResumeFromSessionID loads transcript state by session ID and resumes appends.
func (a *Agent) ResumeFromSessionID(rootDir, sessionID string) error {
	transcript, err := session.LoadByID(rootDir, sessionID)
	if err != nil {
		return err
	}
	store, err := session.OpenByID(rootDir, transcript.SessionID)
	if err != nil {
		return err
	}

	a.mu.Lock()
	a.messages = append([]types.Message(nil), transcript.Messages...)
	a.session = store
	a.config.SessionStore = store
	a.mu.Unlock()

	a.emit(types.AgentEvent{
		Type:                   types.AgentEventSessionResumed,
		SessionID:              transcript.SessionID,
		SessionPath:            transcript.Path,
		SessionMessageCount:    len(transcript.Messages),
		SessionBoundaryCount:   len(transcript.Boundaries),
		SessionResumedFromPath: false,
		Runtime:                a.RuntimeSnapshot(),
	})
	a.emitTransition(0, types.AgentLifecycleIdle, types.AgentTransitionResumed)
	a.emit(types.AgentEvent{Type: types.AgentEventReplayCheckpoint, Replay: a.ReplayCursor()})

	return nil
}

// ResumeFromSessionPath loads transcript state by path and resumes appends.
func (a *Agent) ResumeFromSessionPath(path string) error {
	transcript, err := session.LoadByPath(path)
	if err != nil {
		return err
	}
	store, err := session.OpenByPath(path, transcript.SessionID)
	if err != nil {
		return err
	}

	a.mu.Lock()
	a.messages = append([]types.Message(nil), transcript.Messages...)
	a.session = store
	a.config.SessionStore = store
	a.mu.Unlock()

	a.emit(types.AgentEvent{
		Type:                   types.AgentEventSessionResumed,
		SessionID:              transcript.SessionID,
		SessionPath:            transcript.Path,
		SessionMessageCount:    len(transcript.Messages),
		SessionBoundaryCount:   len(transcript.Boundaries),
		SessionResumedFromPath: true,
		Runtime:                a.RuntimeSnapshot(),
	})
	a.emitTransition(0, types.AgentLifecycleIdle, types.AgentTransitionResumed)
	a.emit(types.AgentEvent{Type: types.AgentEventReplayCheckpoint, Replay: a.ReplayCursor()})

	return nil
}

// Messages returns a copy of the conversation history.
func (a *Agent) Messages() []types.Message {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]types.Message, len(a.messages))
	copy(out, a.messages)
	return out
}

// Usage returns the cumulative token usage across all turns.
func (a *Agent) Usage() types.Usage {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.usage
}

// TotalCost returns the estimated total cost in USD so far.
func (a *Agent) TotalCost() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.totalCostUSD
}

// RuntimeSnapshot returns loop-runtime counters for statusline/status views.
func (a *Agent) RuntimeSnapshot() types.AgentRuntimeSnapshot {
	a.runtimeMu.RLock()
	defer a.runtimeMu.RUnlock()
	out := types.AgentRuntimeSnapshot{
		Turns:          a.runtimeTurns,
		TurnIndex:      a.runtimeTurnIdx,
		Phase:          a.runtimePhase,
		ToolInflight:   a.runtimeInflight,
		LastStopReason: a.runtimeLastStop,
	}
	for _, status := range a.runtimeTasks {
		out.TasksTotal++
		switch status {
		case "running":
			out.TasksRunning++
		case "completed":
			out.TasksCompleted++
		}
	}
	for _, status := range a.runtimeTeams {
		out.TeamsTotal++
		if status == "active" {
			out.TeamsActive++
		}
	}
	return out
}

// Run executes a single user turn: sends the user message, loops through any
// tool calls until the model produces a final text response or hits limits.
func (a *Agent) Run(ctx context.Context, userMessage string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	resolvedRefs := references.NewResolver(a.config.WorkingDir).Resolve(userMessage)
	a.emit(types.AgentEvent{
		Type:               types.AgentEventInputProcessed,
		Input:              userMessage,
		ResolvedReferences: toAgentReferences(resolvedRefs),
	})
	a.emitTransition(0, types.AgentLifecycleTurnStarting, types.AgentTransitionInputAccepted)

	if err := a.appendMessage(types.NewTextMessage(types.RoleUser, userMessage)); err != nil {
		return fmt.Errorf("persist user message: %w", err)
	}

	return a.agentLoop(ctx)
}

func toAgentReferences(refs []references.ResolvedReference) []types.AgentReference {
	if len(refs) == 0 {
		return nil
	}
	out := make([]types.AgentReference, 0, len(refs))
	for _, ref := range refs {
		entry := types.AgentReference{
			Raw:           ref.Raw,
			Path:          ref.Path,
			CanonicalPath: ref.CanonicalPath,
			Start:         ref.Start,
			End:           ref.End,
			Exists:        ref.Exists,
			ResourceType:  ref.ResourceType,
		}
		if ref.HasLine {
			entry.Line = ref.Line
		}
		out = append(out, entry)
	}
	return out
}

// RunLoop runs the interactive agent loop. It reads user input via the provided
// inputFn and loops until the context is cancelled or inputFn returns an error.
func (a *Agent) RunLoop(ctx context.Context, inputFn func() (string, error)) error {
	for {
		select {
		case <-ctx.Done():
			a.emitStop(types.AgentStopCanceled, "", ctx.Err().Error(), 0)
			return ctx.Err()
		default:
		}

		input, err := inputFn()
		if err != nil {
			a.emitStop(types.AgentStopInputError, "", err.Error(), 0)
			return err
		}
		if input == "" {
			continue
		}

		if err := a.Run(ctx, input); err != nil {
			return err
		}
	}
}

// agentLoop is the core loop: send to provider, handle tool calls, repeat.
func (a *Agent) agentLoop(ctx context.Context) error {
	const maxContinuationRecoveries = defaultMaxContinuationRecoveries
	continuationRecoveries := 0
	planner := newTurnPlanner(a)

	for turn := 0; turn < a.config.MaxTurns; turn++ {
		turnNumber := turn + 1
		a.emitTurnPhase(turnNumber, types.AgentTurnPhaseInit, "", "turn_start", false, "")
		if err := planner.BeginTurn(ctx, turnNumber); err != nil {
			return err
		}
		a.emitReplayCheckpoint(turnNumber, "turn_begin")

		a.emitTurnPhase(turnNumber, types.AgentTurnPhaseProviderRequest, types.AgentTurnPhaseCompactionCheck, "provider_request", false, "")
		a.emitTransition(turnNumber, types.AgentLifecycleProviderRequest, types.AgentTransitionProviderCall)

		// Build the chat request.
		a.mu.RLock()
		msgs := make([]types.Message, len(a.messages))
		copy(msgs, a.messages)
		a.mu.RUnlock()

		req := types.ChatRequest{
			Messages:  msgs,
			Tools:     types.ToToolDefs(a.config.Tools),
			System:    a.config.SystemPrompt,
			Model:     a.config.Model,
			MaxTokens: a.config.MaxTokens,
		}

		if a.config.Hooks != nil {
			if err := a.config.Hooks.Fire(ctx, hooks.EventPreChat, map[string]string{
				"MODEL": req.Model,
				"TURN":  strconv.Itoa(turnNumber),
			}); err != nil {
				a.emitStop(types.AgentStopHookError, "", err.Error(), turnNumber)
				return fmt.Errorf("pre-chat hook failed on turn %d: %w", turnNumber, err)
			}
		}

		// Send request to provider via streaming.
		assistantMsg, stopReason, err := a.doChat(ctx, req, turnNumber)
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				a.emitTurnPhase(turnNumber, types.AgentTurnPhaseRetry, types.AgentTurnPhaseProviderStream, "retry_canceled", true, "context_canceled")
				details := err.Error()
				if ctx.Err() != nil {
					details = ctx.Err().Error()
				}
				a.emitStop(types.AgentStopCanceled, types.StopCanceled, details, turnNumber)
				if ctx.Err() != nil {
					return ctx.Err()
				}
				return err
			}
			if recovered, recoverErr := a.recoverFromProviderFailure(ctx, err); recoverErr != nil {
				a.emitStop(types.AgentStopPersistenceError, "", recoverErr.Error(), turnNumber)
				return recoverErr
			} else if recovered {
				a.emitTurnPhase(turnNumber, types.AgentTurnPhaseRetry, types.AgentTurnPhaseProviderRequest, "retry_provider_error", false, "provider_error")
				a.emitTransition(turnNumber, types.AgentLifecycleRetrying, types.AgentTransitionRetryProviderError)
				a.emit(types.AgentEvent{Type: types.AgentEventRetry, Turn: turnNumber, RetryKind: "provider_error", RetryAttempt: 1, RetryMax: 1, RetryPlan: retryPlan("provider_error", 1, 1, types.AgentTurnPhaseProviderRequest, types.AgentTurnPhaseRetry, false), Details: err.Error(), Runtime: a.RuntimeSnapshot(), ReplayLabel: "retry_provider_error"})
				a.emitReplayCheckpoint(turnNumber, "retry_provider_error")
				continue
			}
			a.emitStop(types.AgentStopProviderError, "", err.Error(), turnNumber)
			return fmt.Errorf("provider chat error on turn %d: %w", turnNumber, err)
		}

		if a.config.Hooks != nil {
			if err := a.config.Hooks.Fire(ctx, hooks.EventPostChat, map[string]string{
				"MODEL":       req.Model,
				"TURN":        strconv.Itoa(turnNumber),
				"STOP_REASON": string(stopReason),
			}); err != nil {
				a.emitStop(types.AgentStopHookError, stopReason, err.Error(), turnNumber)
				return fmt.Errorf("post-chat hook failed on turn %d: %w", turnNumber, err)
			}
		}

		if err := a.appendMessage(assistantMsg); err != nil {
			a.emitStop(types.AgentStopPersistenceError, stopReason, err.Error(), turnNumber)
			return fmt.Errorf("persist assistant message: %w", err)
		}

		action := evaluateContinuationRecovery(stopReason, assistantMsg, continuationRecoveries, maxContinuationRecoveries)
		if action.ShouldRetry {
			if err := a.appendMessage(types.NewTextMessage(types.RoleUser, maxOutputRecoveryMessage())); err != nil {
				a.emitStop(types.AgentStopPersistenceError, stopReason, err.Error(), turnNumber)
				return fmt.Errorf("persist continuation prompt: %w", err)
			}
			continuationRecoveries = action.NextRecoveryCount
			a.emitTurnPhase(turnNumber, types.AgentTurnPhaseRetry, types.AgentTurnPhaseProviderStream, "retry_max_tokens", false, "max_output_tokens")
			a.emitTransition(turnNumber, types.AgentLifecycleRetrying, types.AgentTransitionRetryMaxTokens)
			a.emit(types.AgentEvent{Type: types.AgentEventRetry, Turn: turnNumber, RetryKind: "max_output_tokens", RetryAttempt: continuationRecoveries, RetryMax: maxContinuationRecoveries, RetryPlan: retryPlan("max_output_tokens", continuationRecoveries, maxContinuationRecoveries, types.AgentTurnPhaseProviderStream, types.AgentTurnPhaseRetry, false), Runtime: a.RuntimeSnapshot(), ReplayLabel: "retry_max_tokens"})
			a.emitReplayCheckpoint(turnNumber, "retry_max_tokens")
			continue
		}
		if action.Canceled {
			a.emitTurnPhase(turnNumber, types.AgentTurnPhaseRetry, types.AgentTurnPhaseProviderStream, "retry_canceled", true, "context_canceled")
			a.emitTransition(turnNumber, types.AgentLifecycleRetrying, types.AgentTransitionRetryCanceled)
			a.emit(types.AgentEvent{Type: types.AgentEventRetry, Turn: turnNumber, RetryKind: "canceled", RetryAttempt: continuationRecoveries, RetryMax: maxContinuationRecoveries, RetryPlan: retryPlan("canceled", continuationRecoveries, maxContinuationRecoveries, types.AgentTurnPhaseProviderStream, types.AgentTurnPhaseRetry, true), Runtime: a.RuntimeSnapshot(), ReplayLabel: "retry_canceled"})
			a.emitReplayCheckpoint(turnNumber, "retry_canceled")
		}

		if stopReason == types.StopContextLimit {
			recovered, recoverErr := a.recoverFromPromptTooLongStop(ctx)
			if recoverErr != nil {
				a.emitStop(types.AgentStopPersistenceError, stopReason, recoverErr.Error(), turnNumber)
				return recoverErr
			}
			if recovered {
				a.emitTurnPhase(turnNumber, types.AgentTurnPhaseRetry, types.AgentTurnPhaseProviderRequest, "retry_prompt_too_long", false, "prompt_too_long")
				a.emitTransition(turnNumber, types.AgentLifecycleRetrying, types.AgentTransitionRetryPromptTooLong)
				a.emit(types.AgentEvent{Type: types.AgentEventRetry, Turn: turnNumber, RetryKind: "prompt_too_long", RetryAttempt: 1, RetryMax: 1, RetryPlan: retryPlan("prompt_too_long", 1, 1, types.AgentTurnPhaseProviderRequest, types.AgentTurnPhaseRetry, false), Runtime: a.RuntimeSnapshot(), ReplayLabel: "retry_prompt_too_long"})
				a.emitReplayCheckpoint(turnNumber, "retry_prompt_too_long")
				continuationRecoveries = 0
				continue
			}
		}

		// If the model stopped for a final non-tool reason, we're done.
		if stopReason == types.StopEndTurn || stopReason == types.StopMaxTokens || stopReason == types.StopStopSequence || stopReason == types.StopContextLimit || stopReason == types.StopContentFilter || stopReason == types.StopRefusal || stopReason == types.StopCanceled {
			stop := mapProviderStopReason(stopReason)
			details := ""
			if stopReason == types.StopMaxTokens && continuationRecoveries >= maxContinuationRecoveries && a.shouldContinueAfterMaxTokens(assistantMsg) {
				stop = types.AgentStopMaxTokensExhausted
				details = fmt.Sprintf("continuation_recoveries=%d", continuationRecoveries)
			}
			continuationRecoveries = 0
			a.emitStop(stop, stopReason, details, turnNumber)
			return nil
		}

		// stop_reason == tool_use: execute tool calls and loop.
		if stopReason == types.StopToolUse {
			a.emitTurnPhase(turnNumber, types.AgentTurnPhaseToolExecution, types.AgentTurnPhaseProviderStream, "tool_execution", false, "")
			a.emitTransition(turnNumber, types.AgentLifecycleToolExecution, types.AgentTransitionToolUseDetected)
			toolUses := assistantMsg.GetToolUses()
			if len(toolUses) == 0 {
				// Model said tool_use but no tool blocks — treat as end_turn.
				a.emitStop(types.AgentStopToolUseMalformed, stopReason, "tool_use without tool blocks", turnNumber)
				return nil
			}

			resultMessages, err := a.executeTools(ctx, turnNumber, toolUses)
			if err != nil {
				if ctx.Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					details := err.Error()
					if ctx.Err() != nil {
						details = ctx.Err().Error()
					}
					a.emitStop(types.AgentStopCanceled, types.StopCanceled, details, turnNumber)
					if ctx.Err() != nil {
						return ctx.Err()
					}
					return err
				}
				a.emitStop(types.AgentStopToolExecutionError, stopReason, err.Error(), turnNumber)
				return fmt.Errorf("tool execution error on turn %d: %w", turnNumber, err)
			}

			if err := a.appendMessages(resultMessages); err != nil {
				a.emitStop(types.AgentStopPersistenceError, stopReason, err.Error(), turnNumber)
				return fmt.Errorf("persist tool result messages: %w", err)
			}
			a.emitTransition(turnNumber, types.AgentLifecycleProviderRequest, types.AgentTransitionToolExecutionEnd)
			a.emitTurnPhase(turnNumber, types.AgentTurnPhaseProviderRequest, types.AgentTurnPhaseToolExecution, "tools_complete", false, "")

			continue
		}

		// Unknown stop reason — treat as done.
		if a.config.Debug {
			log.Printf("[agent] unknown stop_reason %q, ending loop", stopReason)
		}
		a.emitStop(types.AgentStopUnknown, stopReason, string(stopReason), turnNumber)
		return nil
	}

	a.emitStop(types.AgentStopMaxTurns, "", fmt.Sprintf("reached %d turns", a.config.MaxTurns), a.config.MaxTurns)
	return fmt.Errorf("agent loop reached max turns (%d)", a.config.MaxTurns)
}

func (a *Agent) shouldContinueAfterMaxTokens(msg types.Message) bool {
	if len(msg.GetToolUses()) > 0 {
		return false
	}
	text := strings.TrimSpace(msg.GetText())
	if len(text) < 80 {
		return false
	}
	last := text[len(text)-1]
	return last != '.' && last != '!' && last != '?'
}

func (a *Agent) recoverFromProviderFailure(ctx context.Context, err error) (bool, error) {
	if !isProviderPromptTooLongError(err) {
		return false, nil
	}

	return a.recoverFromPromptTooLongStop(ctx)
}

func (a *Agent) recoverFromPromptTooLongStop(ctx context.Context) (bool, error) {
	return newRecoveryCoordinator(a).RecoverPromptTooLong(ctx)
}

func maxOutputRecoveryMessage() string {
	return "Output token limit hit. Resume directly from where you left off. No apology and no recap."
}

func promptTooLongRecoveryMessage() string {
	return "The previous attempt exceeded the context window. Continue using the compacted conversation context."
}

func looksLikeTooLongError(err error) bool {
	return isProviderPromptTooLongError(err)
}

func isProviderPromptTooLongError(err error) bool {
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "maximum context length") || strings.Contains(msg, "context length exceeded") {
		return true
	}
	if strings.Contains(msg, "prompt is too long") || strings.Contains(msg, "too many tokens") {
		return true
	}
	if strings.Contains(msg, "max output") || strings.Contains(msg, "output token limit") {
		return true
	}
	return strings.Contains(msg, "model_context_window_exceeded")
}

func (a *Agent) attemptCompaction(ctx context.Context, force bool) (bool, error) {
	return newCompactionCoordinator(a).Attempt(ctx, force)
}

func (a *Agent) attemptCompactionLegacy(ctx context.Context, force bool) (bool, error) {
	a.mu.Lock()
	beforeCount := len(a.messages)
	a.mu.Unlock()

	a.emitTransition(0, types.AgentLifecycleCompacting, types.AgentTransitionCompactionCheck)
	a.emit(types.AgentEvent{Type: types.AgentEventCompactionStart, CompactionForced: force, CompactionBeforeCount: beforeCount, Runtime: a.RuntimeSnapshot()})

	a.mu.Lock()
	if !force && !ShouldCompact(a.messages, a.config.ContextWindow) {
		a.mu.Unlock()
		a.emitTransition(0, types.AgentLifecycleProviderRequest, types.AgentTransitionCompactionSkipped)
		a.emit(types.AgentEvent{Type: types.AgentEventCompactionEnd, CompactionForced: force, CompactionApplied: false, CompactionBeforeCount: beforeCount, CompactionAfterCount: beforeCount, Runtime: a.RuntimeSnapshot()})
		return false, nil
	}

	compacted, boundary, err := CompactMessages(ctx, a.messages, a.config.Provider, a.config.Model)
	if err != nil {
		a.mu.Unlock()
		if a.config.Debug {
			log.Printf("[agent] compaction failed: %v", err)
		}
		a.emitTransition(0, types.AgentLifecycleProviderRequest, types.AgentTransitionCompactionSkipped)
		a.emit(types.AgentEvent{Type: types.AgentEventCompactionEnd, CompactionForced: force, CompactionApplied: false, CompactionBeforeCount: beforeCount, CompactionAfterCount: beforeCount, Details: err.Error(), Runtime: a.RuntimeSnapshot()})
		return false, nil
	}

	if boundary != nil && a.session != nil {
		if persistErr := a.session.AppendCompactionBoundary(*boundary); persistErr != nil {
			a.mu.Unlock()
			return false, fmt.Errorf("persist compaction boundary: %w", persistErr)
		}
	}

	currentCount := len(a.messages)
	if len(compacted) == currentCount {
		a.mu.Unlock()
		a.emitTransition(0, types.AgentLifecycleProviderRequest, types.AgentTransitionCompactionSkipped)
		a.emit(types.AgentEvent{Type: types.AgentEventCompactionEnd, CompactionForced: force, CompactionApplied: false, CompactionBeforeCount: beforeCount, CompactionAfterCount: currentCount, Runtime: a.RuntimeSnapshot()})
		return false, nil
	}

	a.messages = compacted
	afterCount := len(compacted)
	a.mu.Unlock()

	a.emitTransition(0, types.AgentLifecycleProviderRequest, types.AgentTransitionCompactionApplied)
	a.emit(types.AgentEvent{Type: types.AgentEventCompactionEnd, CompactionForced: force, CompactionApplied: true, CompactionBeforeCount: beforeCount, CompactionAfterCount: afterCount, Runtime: a.RuntimeSnapshot()})
	return true, nil
}

func (a *Agent) checkBudgets(turn int) error {
	if a.config.MaxBudgetUSD > 0 {
		a.mu.RLock()
		cost := a.totalCostUSD
		a.mu.RUnlock()
		if cost >= a.config.MaxBudgetUSD {
			details := fmt.Sprintf("$%.4f >= $%.4f", cost, a.config.MaxBudgetUSD)
			a.emitStop(types.AgentStopBudgetUSD, "", details, turn)
			return fmt.Errorf("budget limit reached: %s", details)
		}
	}

	if a.config.MaxTokenBudget > 0 {
		a.mu.RLock()
		tokens := a.usage.InputTokens + a.usage.OutputTokens
		a.mu.RUnlock()
		if tokens >= a.config.MaxTokenBudget {
			details := fmt.Sprintf("%d >= %d", tokens, a.config.MaxTokenBudget)
			a.emitStop(types.AgentStopBudgetToken, "", details, turn)
			return fmt.Errorf("token budget reached: %s", details)
		}
	}

	return nil
}

// doChat sends a streaming request and accumulates the assistant response.
// Returns the assembled assistant message, stop reason, and any error.
func (a *Agent) doChat(ctx context.Context, req types.ChatRequest, turn int) (types.Message, types.StopReason, error) {
	a.emitTurnPhase(turn, types.AgentTurnPhaseProviderStream, types.AgentTurnPhaseProviderRequest, "provider_stream", false, "")
	a.emitTransition(turn, types.AgentLifecycleProviderStream, types.AgentTransitionProviderCall)
	stream, err := a.config.Provider.Chat(ctx, req)
	if err != nil {
		return types.Message{}, "", err
	}

	type usageState struct {
		hasCumulative         bool
		lastCumulative        types.Usage
		incrementalBeforeCumu types.Usage
		incrementalAfterCumu  types.Usage
	}

	usage := usageState{}
	accumulateEventUsage := func(event types.StreamEvent) {
		if event.Usage == nil {
			return
		}
		if !event.UsageCumulative {
			if !usage.hasCumulative {
				usage.incrementalBeforeCumu.InputTokens += event.Usage.InputTokens
				usage.incrementalBeforeCumu.OutputTokens += event.Usage.OutputTokens
				usage.incrementalBeforeCumu.CacheHits += event.Usage.CacheHits
			} else {
				usage.incrementalAfterCumu.InputTokens += event.Usage.InputTokens
				usage.incrementalAfterCumu.OutputTokens += event.Usage.OutputTokens
				usage.incrementalAfterCumu.CacheHits += event.Usage.CacheHits
			}
			a.accumulateUsage(event.Usage)
			return
		}

		delta := *event.Usage
		if usage.hasCumulative {
			delta.InputTokens -= usage.lastCumulative.InputTokens
			delta.OutputTokens -= usage.lastCumulative.OutputTokens
			delta.CacheHits -= usage.lastCumulative.CacheHits
			delta.InputTokens -= usage.incrementalAfterCumu.InputTokens
			delta.OutputTokens -= usage.incrementalAfterCumu.OutputTokens
			delta.CacheHits -= usage.incrementalAfterCumu.CacheHits
		} else {
			delta.InputTokens -= usage.incrementalBeforeCumu.InputTokens
			delta.OutputTokens -= usage.incrementalBeforeCumu.OutputTokens
			delta.CacheHits -= usage.incrementalBeforeCumu.CacheHits
		}
		if delta.InputTokens < 0 {
			delta.InputTokens = 0
		}
		if delta.OutputTokens < 0 {
			delta.OutputTokens = 0
		}
		if delta.CacheHits < 0 {
			delta.CacheHits = 0
		}
		a.accumulateUsage(&delta)
		usage.hasCumulative = true
		usage.lastCumulative = *event.Usage
		usage.incrementalAfterCumu = types.Usage{}
	}

	var (
		blocks     []types.ContentBlock
		curText    string
		curToolID  string
		curToolNm  string
		curToolBuf []byte
		stopReason types.StopReason
	)
	normalizer := streamnorm.New(a.config.Provider.Name(), req.Model)

	flushText := func() {
		if curText != "" {
			blocks = append(blocks, types.ContentBlock{
				Type: types.ContentText,
				Text: curText,
			})
			curText = ""
		}
	}

	flushTool := func() {
		if curToolID != "" {
			blocks = append(blocks, types.ContentBlock{
				Type:      types.ContentToolUse,
				ToolUseID: curToolID,
				ToolName:  curToolNm,
				Input:     json.RawMessage(curToolBuf),
			})
			curToolID = ""
			curToolNm = ""
			curToolBuf = nil
		}
	}

	for raw := range stream {
		normalizedEvents := normalizer.Normalize(raw)
		for _, event := range normalizedEvents {
			a.emitTransition(turn, types.AgentLifecycleProviderStream, types.AgentTransitionProviderChunk)
			if event.StopReason != "" {
				stopReason = event.StopReason
			}

			if a.streamCB != nil {
				a.streamCB(event)
			}

			switch event.Type {
			case types.StreamContentDelta:
				flushTool()
				curText += event.Delta
				a.emit(types.AgentEvent{
					Type:           types.AgentEventAssistantChunk,
					Turn:           turn,
					AssistantChunk: event.Delta,
				})

			case types.StreamContentDone:
				if curToolID != "" {
					flushTool()
				} else {
					flushText()
				}

			case types.StreamToolUseStart:
				flushText() // flush any pending text before tool
				flushTool()
				curToolID = event.ToolUseID
				curToolNm = event.ToolName
				curToolBuf = nil

			case types.StreamToolUseDelta:
				if curToolID == "" && event.ToolUseID != "" {
					curToolID = event.ToolUseID
					curToolNm = event.ToolName
				}
				if len(event.Input) > 0 {
					curToolBuf = append(curToolBuf, event.Input...)
				}
				curToolBuf = append(curToolBuf, []byte(event.Delta)...)

			case types.StreamToolUseDone:
				if event.ToolUseID != "" && curToolID != "" && event.ToolUseID != curToolID {
					flushTool()
					curToolID = event.ToolUseID
					curToolNm = event.ToolName
				}
				flushTool()

			case types.StreamThinkingDelta:
				// Accumulate thinking as a thinking content block.
				blocks = append(blocks, types.ContentBlock{
					Type: types.ContentThinking,
					Text: event.Delta,
				})

			case types.StreamMessageDone:
				if event.Message != nil {
					// Use the provider-assembled message if available.
					accumulateEventUsage(event)
					if stopReason == "" {
						stopReason = inferStopReason(*event.Message)
					}
					return *event.Message, stopReason, nil
				}

			case types.StreamError:
				return types.Message{}, "", event.Error
			}

			accumulateEventUsage(event)
		}
	}
	if ctx.Err() != nil {
		return types.Message{}, types.StopCanceled, ctx.Err()
	}

	// Flush any remaining content.
	flushText()
	flushTool()

	msg := types.Message{
		Role:    types.RoleAssistant,
		Content: blocks,
	}

	if stopReason == "" {
		stopReason = inferStopReason(msg)
	}

	return msg, stopReason, nil
}

func inferStopReason(msg types.Message) types.StopReason {
	for _, b := range msg.Content {
		if b.Type == types.ContentToolUse {
			return types.StopToolUse
		}
	}
	return types.StopEndTurn
}

// executeTools runs all tool uses from an assistant message.
// Concurrency-safe tools are run in parallel; others run sequentially.
func (a *Agent) executeTools(ctx context.Context, turn int, toolUses []types.ContentBlock) ([]types.Message, error) {
	a.emitTransition(turn, types.AgentLifecycleToolExecution, types.AgentTransitionToolExecutionStart)
	results := make([]types.Message, len(toolUses))
	total := len(toolUses)
	for i, tu := range toolUses {
		a.emit(types.AgentEvent{
			Type:           types.AgentEventToolQueued,
			Turn:           turn,
			ToolUseID:      tu.ToolUseID,
			ToolName:       tu.ToolName,
			ToolInput:      tu.Input,
			ToolQueueIndex: i + 1,
			ToolQueueTotal: total,
			Runtime:        a.RuntimeSnapshot(),
		})
	}

	// Partition into concurrent and sequential groups.
	var concurrentIdxs []int
	var sequentialIdxs []int

	toolCtx := types.ToolContext{
		WorkingDir: a.config.WorkingDir,
		Debug:      a.config.Debug,
	}

	mailbox := newToolEventMailbox(total)

	for i, tu := range toolUses {
		tool, ok := a.toolMap[tu.ToolName]
		if !ok {
			mailbox.Enqueue(i, a.executeSingleTool(ctx, turn, tu, toolCtx))
			continue
		}
		if !a.isToolAllowed(ctx, turn, tu, tool, toolCtx) {
			mailbox.Enqueue(i, toolExecutionEnvelope{
				message: types.NewToolResultMessage(tu.ToolUseID,
					fmt.Sprintf("error: permission denied for tool %q", tu.ToolName), true),
			})
			continue
		}
		if tool.IsConcurrencySafe(tu.Input) {
			concurrentIdxs = append(concurrentIdxs, i)
		} else {
			sequentialIdxs = append(sequentialIdxs, i)
		}
	}

	// Run concurrent tools in parallel.
	if len(concurrentIdxs) > 0 {
		var wg sync.WaitGroup

		for _, idx := range concurrentIdxs {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				tu := toolUses[i]
				mailbox.Enqueue(i, a.executeSingleTool(ctx, turn, tu, toolCtx))
			}(idx)
		}
		wg.Wait()
	}

	// Run sequential tools in order.
	for _, idx := range sequentialIdxs {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		tu := toolUses[idx]
		mailbox.Enqueue(idx, a.executeSingleTool(ctx, turn, tu, toolCtx))
	}

	envelopes, drainErr := mailbox.DrainInOrder()
	if drainErr != nil {
		return nil, drainErr
	}

	for i, env := range envelopes {
		for _, ev := range env.events {
			a.emit(ev)
		}
		results[i] = env.message
	}
	a.emitReplayCheckpoint(turn, "tools_complete")

	return results, nil
}

func (a *Agent) isToolAllowed(ctx context.Context, turn int, tu types.ContentBlock, tool types.Tool, toolCtx types.ToolContext) bool {
	decision := permissions.DecisionAsk
	if a.config.PermissionChecker != nil {
		decision = a.config.PermissionChecker.Check(tu.ToolName, tu.Input)
	} else {
		switch tool.CheckPermissions(tu.Input, toolCtx) {
		case types.PermissionAllowed:
			decision = permissions.DecisionAllow
		case types.PermissionDenied:
			decision = permissions.DecisionDeny
		case types.PermissionAsk:
			decision = permissions.DecisionAsk
		}
	}

	if decision == permissions.DecisionAsk {
		a.emit(types.AgentEvent{
			Type:               types.AgentEventPermissionAsk,
			Turn:               turn,
			ToolUseID:          tu.ToolUseID,
			ToolName:           tu.ToolName,
			ToolInput:          tu.Input,
			PermissionDecision: types.AgentPermissionAsk,
		})

		if a.config.ResolvePermission != nil {
			decision = a.config.ResolvePermission(ctx, PermissionRequest{
				Turn:     turn,
				ToolUse:  tu,
				Tool:     tool,
				ToolCtx:  toolCtx,
				Decision: decision,
			})
		} else {
			decision = permissions.DecisionDeny
		}
	}

	result := types.AgentPermissionDeny
	allowed := false
	if decision == permissions.DecisionAllow {
		allowed = true
		result = types.AgentPermissionAllow
	}

	a.emit(types.AgentEvent{
		Type:               types.AgentEventPermissionResult,
		Turn:               turn,
		ToolUseID:          tu.ToolUseID,
		ToolName:           tu.ToolName,
		ToolInput:          tu.Input,
		PermissionDecision: result,
	})

	return allowed
}

// executeSingleTool runs one tool and returns the result as a message.
func (a *Agent) executeSingleTool(ctx context.Context, turn int, tu types.ContentBlock, toolCtx types.ToolContext) toolExecutionEnvelope {
	tool, ok := a.toolMap[tu.ToolName]
	telemetryID := toolTelemetryID(turn, tu.ToolUseID, tu.ToolName)
	startedAtMS := time.Now().UnixMilli()

	startEvent := types.AgentEvent{
		Type:            types.AgentEventToolStart,
		Turn:            turn,
		ToolUseID:       tu.ToolUseID,
		ToolName:        tu.ToolName,
		ToolInput:       tu.Input,
		ToolTelemetryID: telemetryID,
		ToolStartedAtMS: startedAtMS,
	}
	startEvent.Runtime = a.runtimeToolStart(turn)

	if !ok {
		endedAtMS := time.Now().UnixMilli()
		endRuntime := a.runtimeToolEnd()
		return toolExecutionEnvelope{
			events: []types.AgentEvent{
				startEvent,
				{
					Type:            types.AgentEventToolEnd,
					Turn:            turn,
					ToolUseID:       tu.ToolUseID,
					ToolName:        tu.ToolName,
					ToolInput:       tu.Input,
					ToolError:       fmt.Sprintf("unknown tool %q", tu.ToolName),
					ToolTelemetryID: telemetryID,
					ToolStartedAtMS: startedAtMS,
					ToolEndedAtMS:   endedAtMS,
					ToolDurationMS:  endedAtMS - startedAtMS,
					Runtime:         endRuntime,
				},
			},
			message: types.NewToolResultMessage(tu.ToolUseID,
				fmt.Sprintf("error: unknown tool %q", tu.ToolName), true),
		}
	}

	if a.config.Hooks != nil {
		if err := a.config.Hooks.Fire(ctx, hooks.EventPreTool, map[string]string{
			"TOOL_NAME":  tu.ToolName,
			"TOOL_ID":    tu.ToolUseID,
			"TOOL_INPUT": string(tu.Input),
			"TURN":       strconv.Itoa(turn),
		}); err != nil {
			endedAtMS := time.Now().UnixMilli()
			endRuntime := a.runtimeToolEnd()
			return toolExecutionEnvelope{
				events: []types.AgentEvent{
					startEvent,
					{
						Type:            types.AgentEventToolEnd,
						Turn:            turn,
						ToolUseID:       tu.ToolUseID,
						ToolName:        tu.ToolName,
						ToolInput:       tu.Input,
						ToolError:       err.Error(),
						ToolTelemetryID: telemetryID,
						ToolStartedAtMS: startedAtMS,
						ToolEndedAtMS:   endedAtMS,
						ToolDurationMS:  endedAtMS - startedAtMS,
						Runtime:         endRuntime,
					},
				},
				message: types.NewToolResultMessage(tu.ToolUseID,
					fmt.Sprintf("error running pre-tool hook for %s: %v", tu.ToolName, err), true),
			}
		}
	}

	// Checkpoint before destructive tools.
	if a.config.EnableCheckpoints && tool.IsDestructive(tu.Input) {
		desc := fmt.Sprintf("before %s", tu.ToolName)
		if _, err := CreateCheckpoint(a.config.WorkingDir, desc); err != nil {
			if a.config.Debug {
				log.Printf("[agent] checkpoint failed: %v", err)
			}
		}
	}

	a.emitLifecycleEvent(turn, tu, "start")

	result, err := tool.Execute(ctx, tu.Input, toolCtx)
	attempt := 1
	for err != nil && shouldRetryToolError(ctx, err) && attempt < toolMaxAttempts {
		nextAttempt := attempt + 1
		a.emit(types.AgentEvent{
			Type:            types.AgentEventToolRetry,
			Turn:            turn,
			ToolUseID:       tu.ToolUseID,
			ToolName:        tu.ToolName,
			ToolInput:       tu.Input,
			ToolTelemetryID: telemetryID,
			ToolAttempt:     nextAttempt,
			ToolMaxAttempts: toolMaxAttempts,
			RetryKind:       "tool_execution",
			RetryAttempt:    nextAttempt,
			RetryMax:        toolMaxAttempts,
			Details:         err.Error(),
			Runtime:         a.RuntimeSnapshot(),
		})
		attempt = nextAttempt
		result, err = tool.Execute(ctx, tu.Input, toolCtx)
	}

	if err != nil {
		a.emitLifecycleEvent(turn, tu, "error")
		if a.config.Hooks != nil {
			_ = a.config.Hooks.Fire(ctx, hooks.EventPostTool, map[string]string{
				"TOOL_NAME":  tu.ToolName,
				"TOOL_ID":    tu.ToolUseID,
				"TOOL_INPUT": string(tu.Input),
				"TURN":       strconv.Itoa(turn),
				"IS_ERROR":   "true",
				"RESULT":     err.Error(),
			})
		}
		endedAtMS := time.Now().UnixMilli()
		endRuntime := a.runtimeToolEnd()
		return toolExecutionEnvelope{
			events: []types.AgentEvent{
				startEvent,
				{
					Type:            types.AgentEventToolEnd,
					Turn:            turn,
					ToolUseID:       tu.ToolUseID,
					ToolName:        tu.ToolName,
					ToolInput:       tu.Input,
					ToolError:       err.Error(),
					ToolTelemetryID: telemetryID,
					ToolStartedAtMS: startedAtMS,
					ToolEndedAtMS:   endedAtMS,
					ToolDurationMS:  endedAtMS - startedAtMS,
					ToolAttempt:     attempt,
					ToolMaxAttempts: toolMaxAttempts,
					Runtime:         endRuntime,
				},
			},
			message: types.NewToolResultMessage(tu.ToolUseID,
				fmt.Sprintf("error executing %s: %v", tu.ToolName, err), true),
		}
	}

	a.emitLifecycleEvent(turn, tu, "end")

	if a.config.Hooks != nil {
		if hookErr := a.config.Hooks.Fire(ctx, hooks.EventPostTool, map[string]string{
			"TOOL_NAME":  tu.ToolName,
			"TOOL_ID":    tu.ToolUseID,
			"TOOL_INPUT": string(tu.Input),
			"TURN":       strconv.Itoa(turn),
			"IS_ERROR":   strconv.FormatBool(result.IsError),
			"RESULT":     result.Content,
		}); hookErr != nil {
			endedAtMS := time.Now().UnixMilli()
			endRuntime := a.runtimeToolEnd()
			return toolExecutionEnvelope{
				events: []types.AgentEvent{
					startEvent,
					{
						Type:            types.AgentEventToolEnd,
						Turn:            turn,
						ToolUseID:       tu.ToolUseID,
						ToolName:        tu.ToolName,
						ToolInput:       tu.Input,
						ToolError:       hookErr.Error(),
						ToolTelemetryID: telemetryID,
						ToolStartedAtMS: startedAtMS,
						ToolEndedAtMS:   endedAtMS,
						ToolDurationMS:  endedAtMS - startedAtMS,
						ToolAttempt:     attempt,
						ToolMaxAttempts: toolMaxAttempts,
						Runtime:         endRuntime,
					},
				},
				message: types.NewToolResultMessage(tu.ToolUseID,
					fmt.Sprintf("error running post-tool hook for %s: %v", tu.ToolName, hookErr), true),
			}
		}
	}

	a.applyRuntimeLifecycleFromToolResult(tu.ToolName, tu.Input, result)

	endedAtMS := time.Now().UnixMilli()
	endRuntime := a.runtimeToolEnd()
	return toolExecutionEnvelope{
		events: []types.AgentEvent{
			startEvent,
			{
				Type:            types.AgentEventToolEnd,
				Turn:            turn,
				ToolUseID:       tu.ToolUseID,
				ToolName:        tu.ToolName,
				ToolInput:       tu.Input,
				ToolTelemetryID: telemetryID,
				ToolStartedAtMS: startedAtMS,
				ToolEndedAtMS:   endedAtMS,
				ToolDurationMS:  endedAtMS - startedAtMS,
				ToolAttempt:     attempt,
				ToolMaxAttempts: toolMaxAttempts,
				Runtime:         endRuntime,
			},
		},
		message: types.NewToolResultMessage(tu.ToolUseID, result.Content, result.IsError),
	}
}

func (a *Agent) emitLifecycleEvent(turn int, tu types.ContentBlock, action string) {
	kind, ok := lifecycleKindFromToolName(tu.ToolName)
	if !ok {
		return
	}
	a.emit(types.AgentEvent{
		Type:          types.AgentEventToolLifecycle,
		Turn:          turn,
		ToolUseID:     tu.ToolUseID,
		ToolName:      tu.ToolName,
		ToolInput:     tu.Input,
		ToolLifecycle: action,
		Runtime:       a.RuntimeSnapshot(),
	})
	teamName, taskID := lifecycleIDsFromToolInput(tu.ToolName, tu.Input)
	if kind == "team" {
		a.emit(types.AgentEvent{
			Type:       types.AgentEventTeamLifecycle,
			Turn:       turn,
			ToolUseID:  tu.ToolUseID,
			ToolName:   tu.ToolName,
			TeamName:   teamName,
			TeamStatus: action,
			Runtime:    a.RuntimeSnapshot(),
		})
	}
	if taskID != "" {
		a.emit(types.AgentEvent{
			Type:       types.AgentEventTaskLifecycle,
			Turn:       turn,
			ToolUseID:  tu.ToolUseID,
			ToolName:   tu.ToolName,
			TaskID:     taskID,
			TaskStatus: action,
			Runtime:    a.RuntimeSnapshot(),
		})
	}
	a.emit(types.AgentEvent{
		Type: types.AgentEventAssistantChunk,
		Turn: turn,
		AssistantChunk: fmt.Sprintf(
			"<lifecycle kind=%q action=%q tool=%q tool_use_id=%q>",
			kind,
			action,
			tu.ToolName,
			tu.ToolUseID,
		),
	})
}

func lifecycleKindFromToolName(name string) (string, bool) {
	switch name {
	case "Agent":
		return "subagent", true
	case "team_create", "team_update", "team_delete", "team_status", "team_list":
		return "team", true
	case "send_message":
		return "message", true
	default:
		return "", false
	}
}

func (a *Agent) emit(ev types.AgentEvent) {
	now := time.Now().UnixMilli()
	a.eventMu.Lock()
	a.eventSeq++
	ev.Sequence = a.eventSeq
	ev.TimestampMS = now
	ev.ID = fmt.Sprintf("evt-%d", ev.Sequence)

	a.eventHistory = append(a.eventHistory, ev)
	if a.eventHistoryCap <= 0 {
		a.eventHistoryCap = 256
	}
	if over := len(a.eventHistory) - a.eventHistoryCap; over > 0 {
		a.eventHistory = a.eventHistory[over:]
	}
	a.eventMu.Unlock()

	if a.eventCB != nil {
		a.eventCB(ev)
	}
}

// ReplayCursor returns the current replay window for emitted events.
func (a *Agent) ReplayCursor() types.AgentReplayCursor {
	a.eventMu.RLock()
	defer a.eventMu.RUnlock()
	if len(a.eventHistory) == 0 {
		return types.AgentReplayCursor{}
	}
	return types.AgentReplayCursor{
		Start: a.eventHistory[0].Sequence,
		End:   a.eventHistory[len(a.eventHistory)-1].Sequence,
		Size:  len(a.eventHistory),
	}
}

// ReplayEvents returns events in [fromSeq, toSeq] inclusive.
func (a *Agent) ReplayEvents(fromSeq, toSeq uint64) []types.AgentEvent {
	a.eventMu.RLock()
	defer a.eventMu.RUnlock()
	if toSeq > 0 && fromSeq > toSeq {
		fromSeq, toSeq = toSeq, fromSeq
	}
	out := make([]types.AgentEvent, 0)
	for _, ev := range a.eventHistory {
		if fromSeq > 0 && ev.Sequence < fromSeq {
			continue
		}
		if toSeq > 0 && ev.Sequence > toSeq {
			continue
		}
		out = append(out, ev)
	}
	return out
}

func (a *Agent) emitTransition(turn int, to types.AgentLifecycleState, reason types.AgentTransitionReason) {
	a.eventMu.Lock()
	from := a.lifecycleState
	a.lifecycleState = to
	a.eventMu.Unlock()
	transition := types.AgentTransition{
		From:   from,
		To:     to,
		Reason: reason,
	}
	validationDetails := ""
	if err := types.ValidateLifecycleTransition(transition); err != nil {
		validationDetails = err.Error()
	}
	a.emit(types.AgentEvent{
		Type:       types.AgentEventStateTransition,
		Turn:       turn,
		State:      to,
		Transition: transition,
		Replay:     a.ReplayCursor(),
		Details:    validationDetails,
	})
}

func (a *Agent) emitTurnPhase(turn int, phase, prev types.AgentTurnPhase, transition string, canceled bool, recoveryBranch string) {
	a.runtimeSetPhase(turn, phase)
	state := types.AgentTurnPhaseState{
		Turn:           turn,
		TurnIndex:      turn,
		Phase:          phase,
		PreviousPhase:  prev,
		Transition:     transition,
		Canceled:       canceled,
		RecoveryBranch: recoveryBranch,
	}
	validationDetails := ""
	if err := types.ValidateTurnPhaseState(state); err != nil {
		validationDetails = err.Error()
	}
	a.emit(types.AgentEvent{
		Type:      types.AgentEventTurnPhase,
		Turn:      turn,
		TurnPhase: state,
		Runtime:   a.RuntimeSnapshot(),
		Details:   validationDetails,
	})
}

func (a *Agent) emitStop(reason types.AgentStopReason, providerReason types.StopReason, details string, turn int) {
	a.emitTurnPhase(turn, types.AgentTurnPhaseTurnEnd, a.RuntimeSnapshot().Phase, "turn_end", reason == types.AgentStopCanceled, "")
	a.runtimeSetTurn(turn)
	a.runtimeSetStop(reason)
	a.emitTransition(turn, types.AgentLifecycleStopped, types.AgentTransitionStopped)
	a.emit(types.AgentEvent{
		Type:               types.AgentEventStop,
		Turn:               turn,
		StopReason:         reason,
		ProviderStopReason: providerReason,
		Details:            details,
		Runtime:            a.RuntimeSnapshot(),
	})
	a.emit(types.AgentEvent{
		Type:       types.AgentEventTurnEnd,
		Turn:       turn,
		StopReason: reason,
		Details:    details,
		Runtime:    a.RuntimeSnapshot(),
	})
	a.emitReplayCheckpoint(turn, "turn_end")

	if a.config.Hooks != nil {
		stopReasonClass := hooks.ClassifyStopReason(string(reason))
		terminalStop := hooks.IsTerminalStopReason(string(reason))
		mappedProviderStop := mapProviderStopReason(providerReason)
		_ = a.config.Hooks.Fire(context.Background(), hooks.EventTurnEnd, map[string]string{
			"TURN":                   strconv.Itoa(turn),
			"STOP_REASON":            string(reason),
			"STOP_REASON_CLASS":      string(stopReasonClass),
			"STOP_REASON_TERMINAL":   strconv.FormatBool(terminalStop),
			"PROVIDER_STOP_REASON":   string(providerReason),
			"MAPPED_PROVIDER_REASON": string(mappedProviderStop),
			"DETAILS":                details,
		})
		_ = a.config.Hooks.Fire(context.Background(), hooks.EventStop, map[string]string{
			"TURN":                   strconv.Itoa(turn),
			"STOP_REASON":            string(reason),
			"STOP_REASON_CLASS":      string(stopReasonClass),
			"STOP_REASON_TERMINAL":   strconv.FormatBool(terminalStop),
			"PROVIDER_STOP_REASON":   string(providerReason),
			"MAPPED_PROVIDER_REASON": string(mappedProviderStop),
			"DETAILS":                details,
		})
	}
}

func (a *Agent) emitReplayCheckpoint(turn int, label string) {
	a.emit(types.AgentEvent{
		Type:        types.AgentEventReplayCheckpoint,
		Turn:        turn,
		ReplayLabel: label,
		Replay:      a.ReplayCursor(),
	})
}

func toolTelemetryID(turn int, toolUseID, toolName string) string {
	id := toolUseID
	if id == "" {
		id = "unknown"
	}
	return fmt.Sprintf("turn-%d:%s:%s", turn, toolName, id)
}

func mapProviderStopReason(reason types.StopReason) types.AgentStopReason {
	switch reason {
	case types.StopEndTurn:
		return types.AgentStopEndTurn
	case types.StopMaxTokens:
		return types.AgentStopMaxTokens
	case types.StopContextLimit:
		return types.AgentStopContextLimit
	case types.StopStopSequence:
		return types.AgentStopStopSeq
	case types.StopToolUse:
		return types.AgentStopToolUseMalformed
	case types.StopContentFilter:
		return types.AgentStopContentFilter
	case types.StopRefusal:
		return types.AgentStopRefusal
	case types.StopCanceled:
		return types.AgentStopCanceled
	case types.StopError:
		return types.AgentStopProviderError
	default:
		return types.AgentStopUnknown
	}
}

func (a *Agent) runtimeSetTurn(turn int) {
	a.runtimeMu.Lock()
	a.runtimeTurnIdx = turn
	if turn > a.runtimeTurns {
		a.runtimeTurns = turn
	}
	a.runtimeMu.Unlock()
}

func (a *Agent) runtimeSetPhase(turn int, phase types.AgentTurnPhase) {
	a.runtimeMu.Lock()
	if turn > 0 {
		a.runtimeTurnIdx = turn
		if turn > a.runtimeTurns {
			a.runtimeTurns = turn
		}
	}
	a.runtimePhase = phase
	a.runtimeMu.Unlock()
}

func (a *Agent) runtimeSetStop(reason types.AgentStopReason) {
	a.runtimeMu.Lock()
	a.runtimeLastStop = reason
	a.runtimeMu.Unlock()
}

func (a *Agent) runtimeToolStart(turn int) types.AgentRuntimeSnapshot {
	a.runtimeMu.Lock()
	a.runtimeInflight++
	if turn > a.runtimeTurns {
		a.runtimeTurns = turn
	}
	out := a.runtimeSnapshotLocked()
	a.runtimeMu.Unlock()
	return out
}

func (a *Agent) runtimeToolEnd() types.AgentRuntimeSnapshot {
	a.runtimeMu.Lock()
	if a.runtimeInflight > 0 {
		a.runtimeInflight--
	}
	out := a.runtimeSnapshotLocked()
	a.runtimeMu.Unlock()
	return out
}

func (a *Agent) runtimeSnapshotLocked() types.AgentRuntimeSnapshot {
	out := types.AgentRuntimeSnapshot{
		Turns:          a.runtimeTurns,
		TurnIndex:      a.runtimeTurnIdx,
		Phase:          a.runtimePhase,
		ToolInflight:   a.runtimeInflight,
		LastStopReason: a.runtimeLastStop,
	}
	for _, status := range a.runtimeTasks {
		out.TasksTotal++
		switch status {
		case "running":
			out.TasksRunning++
		case "completed":
			out.TasksCompleted++
		}
	}
	for _, status := range a.runtimeTeams {
		out.TeamsTotal++
		if status == "active" {
			out.TeamsActive++
		}
	}
	return out
}

func (a *Agent) applyRuntimeLifecycleFromToolResult(toolName string, input json.RawMessage, result types.ToolResult) {
	if result.IsError {
		return
	}
	switch toolName {
	case "task_create", "task_update", "task_stop":
		a.applyTaskRuntimeLifecycle(toolName, input, result.Content)
	case "team_create", "team_update", "team_delete", "team_status", "team_list":
		a.applyTeamRuntimeLifecycle(toolName, result.Content)
	}
}

func normalizeRuntimeTaskStatus(status string) string {
	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case "running", "completed", "failed", "canceled":
		return status
	default:
		return "running"
	}
}

func normalizeRuntimeTeamStatus(status string) string {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "canceled" {
		return "canceled"
	}
	return "active"
}

func (a *Agent) applyTaskRuntimeLifecycle(toolName string, input json.RawMessage, raw string) {
	var in struct {
		TaskID string `json:"task_id"`
		Status string `json:"status"`
	}
	_ = json.Unmarshal(input, &in)

	var out struct {
		Task *struct {
			ID     string `json:"task_id"`
			Status string `json:"status"`
		} `json:"task"`
	}
	_ = json.Unmarshal([]byte(raw), &out)

	id := strings.TrimSpace(in.TaskID)
	status := strings.TrimSpace(in.Status)
	if out.Task != nil {
		if strings.TrimSpace(out.Task.ID) != "" {
			id = strings.TrimSpace(out.Task.ID)
		}
		if strings.TrimSpace(out.Task.Status) != "" {
			status = strings.TrimSpace(out.Task.Status)
		}
	}
	if toolName == "task_stop" {
		status = "canceled"
	}
	if toolName == "task_create" && strings.TrimSpace(status) == "" {
		status = "running"
	}
	if id == "" {
		return
	}

	a.runtimeMu.Lock()
	a.runtimeTasks[id] = normalizeRuntimeTaskStatus(status)
	a.runtimeMu.Unlock()
	a.emit(types.AgentEvent{
		Type:       types.AgentEventTaskLifecycle,
		ToolName:   toolName,
		TaskID:     id,
		TaskStatus: normalizeRuntimeTaskStatus(status),
		Runtime:    a.RuntimeSnapshot(),
	})
}

func (a *Agent) applyTeamRuntimeLifecycle(toolName, raw string) {
	if toolName == "team_list" {
		var list struct {
			Teams []struct {
				Name   string `json:"name"`
				Status string `json:"status"`
			} `json:"teams"`
		}
		if err := json.Unmarshal([]byte(raw), &list); err != nil {
			return
		}
		type lifecycleEvent struct {
			name   string
			status string
		}
		events := make([]lifecycleEvent, 0, len(list.Teams))
		a.runtimeMu.Lock()
		for _, team := range list.Teams {
			name := strings.TrimSpace(team.Name)
			if name == "" {
				continue
			}
			status := normalizeRuntimeTeamStatus(team.Status)
			a.runtimeTeams[name] = status
			events = append(events, lifecycleEvent{name: name, status: status})
		}
		a.runtimeMu.Unlock()
		for _, ev := range events {
			a.emit(types.AgentEvent{Type: types.AgentEventTeamLifecycle, ToolName: toolName, TeamName: ev.name, TeamStatus: ev.status, Runtime: a.RuntimeSnapshot()})
		}
		return
	}

	var withTeam struct {
		Team *struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"team"`
	}
	if err := json.Unmarshal([]byte(raw), &withTeam); err == nil && withTeam.Team != nil {
		name := strings.TrimSpace(withTeam.Team.Name)
		if name != "" {
			status := normalizeRuntimeTeamStatus(withTeam.Team.Status)
			a.runtimeMu.Lock()
			a.runtimeTeams[name] = status
			a.runtimeMu.Unlock()
			a.emit(types.AgentEvent{Type: types.AgentEventTeamLifecycle, ToolName: toolName, TeamName: name, TeamStatus: status, Runtime: a.RuntimeSnapshot()})
			return
		}
	}

	if toolName == "team_delete" {
		var del struct {
			TeamName string `json:"team_name"`
			Status   string `json:"status"`
			Deleted  bool   `json:"deleted"`
		}
		if err := json.Unmarshal([]byte(raw), &del); err != nil {
			return
		}
		if !del.Deleted {
			return
		}
		name := strings.TrimSpace(del.TeamName)
		if name == "" {
			return
		}
		status := strings.TrimSpace(del.Status)
		if status == "" {
			status = "canceled"
		}
		a.runtimeMu.Lock()
		normalized := normalizeRuntimeTeamStatus(status)
		a.runtimeTeams[name] = normalized
		a.runtimeMu.Unlock()
		a.emit(types.AgentEvent{Type: types.AgentEventTeamLifecycle, ToolName: toolName, TeamName: name, TeamStatus: normalized, Runtime: a.RuntimeSnapshot()})
	}
}

func shouldRetryToolError(ctx context.Context, err error) bool {
	return defaultToolRetryPolicy().ShouldRetry(ctx, err)
}

func lifecycleIDsFromToolInput(toolName string, input json.RawMessage) (teamName string, taskID string) {
	var payload map[string]any
	if err := json.Unmarshal(input, &payload); err != nil {
		return "", ""
	}
	if toolName == "team_create" || toolName == "team_update" || toolName == "team_delete" || toolName == "team_status" {
		if v, ok := payload["team_name"].(string); ok {
			teamName = strings.TrimSpace(v)
		}
	}
	if toolName == "task_create" || toolName == "task_update" || toolName == "task_stop" {
		if v, ok := payload["task_id"].(string); ok {
			taskID = strings.TrimSpace(v)
		}
	}
	return teamName, taskID
}

func (a *Agent) appendMessage(msg types.Message) error {
	a.mu.Lock()
	a.messages = append(a.messages, msg)
	store := a.session
	a.mu.Unlock()

	if store != nil {
		if err := store.AppendMessage(msg); err != nil {
			return err
		}
	}

	return nil
}

func (a *Agent) appendMessages(msgs []types.Message) error {
	for _, msg := range msgs {
		if err := a.appendMessage(msg); err != nil {
			return err
		}
	}
	return nil
}

// accumulateUsage adds usage from a single event to the running totals.
func (a *Agent) accumulateUsage(u *types.Usage) {
	if u == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.usage.InputTokens += u.InputTokens
	a.usage.OutputTokens += u.OutputTokens
	a.usage.CacheHits += u.CacheHits
	a.totalCostUSD += float64(u.InputTokens)/1_000_000*a.costPerMInput + float64(u.OutputTokens)/1_000_000*a.costPerMOutput
	if math.IsInf(a.totalCostUSD, 0) || math.IsNaN(a.totalCostUSD) {
		a.totalCostUSD = 0
	}
}
