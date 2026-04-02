package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ProjectOnboardingState struct {
	HasCompleted           bool   `json:"has_completed"`
	SeenCount              int    `json:"seen_count"`
	LastShownAtUnix        int64  `json:"last_shown_at_unix,omitempty"`
	CompletedAtUnix        int64  `json:"completed_at_unix,omitempty"`
	Skipped                bool   `json:"skipped,omitempty"`
	LastStepKey            string `json:"last_step_key,omitempty"`
	LastStepCompleted      bool   `json:"last_step_completed,omitempty"`
	LastWorkspaceIsEmpty   bool   `json:"last_workspace_is_empty,omitempty"`
	LastHasClaudeMd        bool   `json:"last_has_claude_md,omitempty"`
	LastEvaluationAtUnix   int64  `json:"last_evaluation_at_unix,omitempty"`
	LastShownSessionID     string `json:"last_shown_session_id,omitempty"`
	LastCompletedSessionID string `json:"last_completed_session_id,omitempty"`
}

func ReadProjectOnboardingState(path string) (ProjectOnboardingState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ProjectOnboardingState{}, nil
		}
		return ProjectOnboardingState{}, fmt.Errorf("read project onboarding state: %w", err)
	}
	if len(data) == 0 {
		return ProjectOnboardingState{}, nil
	}
	var state ProjectOnboardingState
	if err := json.Unmarshal(data, &state); err != nil {
		return ProjectOnboardingState{}, fmt.Errorf("decode project onboarding state: %w", err)
	}
	if state.SeenCount < 0 {
		state.SeenCount = 0
	}
	state.LastStepKey = strings.TrimSpace(state.LastStepKey)
	state.LastShownSessionID = strings.TrimSpace(state.LastShownSessionID)
	state.LastCompletedSessionID = strings.TrimSpace(state.LastCompletedSessionID)
	if state.LastShownAtUnix < 0 {
		state.LastShownAtUnix = 0
	}
	if state.CompletedAtUnix < 0 {
		state.CompletedAtUnix = 0
	}
	if state.LastEvaluationAtUnix < 0 {
		state.LastEvaluationAtUnix = 0
	}
	if state.CompletedAtUnix > 0 {
		state.HasCompleted = true
	}
	return state, nil
}

func WriteProjectOnboardingState(path string, state ProjectOnboardingState) error {
	if state.SeenCount < 0 {
		state.SeenCount = 0
	}
	state.LastStepKey = strings.TrimSpace(state.LastStepKey)
	state.LastShownSessionID = strings.TrimSpace(state.LastShownSessionID)
	state.LastCompletedSessionID = strings.TrimSpace(state.LastCompletedSessionID)
	if state.LastShownAtUnix < 0 {
		state.LastShownAtUnix = 0
	}
	if state.CompletedAtUnix < 0 {
		state.CompletedAtUnix = 0
	}
	if state.LastEvaluationAtUnix < 0 {
		state.LastEvaluationAtUnix = 0
	}
	if state.CompletedAtUnix > 0 {
		state.HasCompleted = true
	}
	if err := os.MkdirAll(filepath.Dir(path), stateDirPerm); err != nil {
		return fmt.Errorf("create project onboarding directory: %w", err)
	}
	payload, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode project onboarding state: %w", err)
	}
	payload = append(payload, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, payload, stateFilePerm); err != nil {
		return fmt.Errorf("write project onboarding tmp file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("replace project onboarding state file: %w", err)
	}
	return nil
}

func IncrementOnboardingSeenCount(path string, sessionID string, shownAtUnix int64) (ProjectOnboardingState, error) {
	st, err := ReadProjectOnboardingState(path)
	if err != nil {
		return ProjectOnboardingState{}, err
	}
	st.SeenCount++
	if st.SeenCount < 0 {
		st.SeenCount = 0
	}
	if shownAtUnix > 0 {
		st.LastShownAtUnix = shownAtUnix
	}
	st.LastShownSessionID = strings.TrimSpace(sessionID)
	if err := WriteProjectOnboardingState(path, st); err != nil {
		return ProjectOnboardingState{}, err
	}
	return st, nil
}

func MarkOnboardingCompleted(path string, sessionID string, completedAtUnix int64) (ProjectOnboardingState, error) {
	st, err := ReadProjectOnboardingState(path)
	if err != nil {
		return ProjectOnboardingState{}, err
	}
	st.HasCompleted = true
	st.Skipped = false
	if completedAtUnix > 0 {
		st.CompletedAtUnix = completedAtUnix
		st.LastEvaluationAtUnix = completedAtUnix
	}
	st.LastCompletedSessionID = strings.TrimSpace(sessionID)
	if err := WriteProjectOnboardingState(path, st); err != nil {
		return ProjectOnboardingState{}, err
	}
	return st, nil
}

func UpdateOnboardingEvaluation(path string, state ProjectOnboardingState) (ProjectOnboardingState, error) {
	current, err := ReadProjectOnboardingState(path)
	if err != nil {
		return ProjectOnboardingState{}, err
	}

	if strings.TrimSpace(state.LastStepKey) != "" {
		current.LastStepKey = strings.TrimSpace(state.LastStepKey)
	}
	current.LastStepCompleted = state.LastStepCompleted
	current.LastWorkspaceIsEmpty = state.LastWorkspaceIsEmpty
	current.LastHasClaudeMd = state.LastHasClaudeMd
	if state.LastEvaluationAtUnix > 0 {
		current.LastEvaluationAtUnix = state.LastEvaluationAtUnix
	}
	if state.Skipped {
		current.Skipped = true
	}

	if err := WriteProjectOnboardingState(path, current); err != nil {
		return ProjectOnboardingState{}, err
	}
	return current, nil
}
