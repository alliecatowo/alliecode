package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProjectOnboardingStateReadWrite(t *testing.T) {
	paths, err := ResolvePaths(t.TempDir(), t.TempDir())
	if err != nil {
		t.Fatalf("ResolvePaths() error = %v", err)
	}

	state := ProjectOnboardingState{HasCompleted: true, SeenCount: 3}
	if err := WriteProjectOnboardingState(paths.ProjectOnboardingStateFile, state); err != nil {
		t.Fatalf("WriteProjectOnboardingState() error = %v", err)
	}

	got, err := ReadProjectOnboardingState(paths.ProjectOnboardingStateFile)
	if err != nil {
		t.Fatalf("ReadProjectOnboardingState() error = %v", err)
	}
	if got != state {
		t.Fatalf("state mismatch: got %+v want %+v", got, state)
	}
}

func TestProjectOnboardingStateMissingIsDefault(t *testing.T) {
	got, err := ReadProjectOnboardingState(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("ReadProjectOnboardingState() error = %v", err)
	}
	if got != (ProjectOnboardingState{}) {
		t.Fatalf("default state mismatch: got %+v", got)
	}
}

func TestProjectOnboardingStateRejectsInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "project-onboarding-state.json")
	if err := os.WriteFile(path, []byte("{"), stateFilePerm); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	_, err := ReadProjectOnboardingState(path)
	if err == nil || !strings.Contains(err.Error(), "decode project onboarding state") {
		t.Fatalf("expected decode error, got %v", err)
	}
}

func TestProjectOnboardingStateClampsNegativeSeenCount(t *testing.T) {
	path := filepath.Join(t.TempDir(), "project-onboarding-state.json")
	if err := WriteProjectOnboardingState(path, ProjectOnboardingState{SeenCount: -10}); err != nil {
		t.Fatalf("WriteProjectOnboardingState() error = %v", err)
	}
	got, err := ReadProjectOnboardingState(path)
	if err != nil {
		t.Fatalf("ReadProjectOnboardingState() error = %v", err)
	}
	if got.SeenCount != 0 {
		t.Fatalf("SeenCount = %d, want 0", got.SeenCount)
	}
}

func TestProjectOnboardingStateExtendedFieldsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "project-onboarding-state.json")
	now := time.Now().UTC().Unix()
	state := ProjectOnboardingState{
		HasCompleted:           true,
		SeenCount:              4,
		LastShownAtUnix:        now - 5,
		CompletedAtUnix:        now,
		LastStepKey:            "claudemd",
		LastStepCompleted:      true,
		LastWorkspaceIsEmpty:   false,
		LastHasClaudeMd:        true,
		LastEvaluationAtUnix:   now,
		LastShownSessionID:     "s1",
		LastCompletedSessionID: "s1",
	}
	if err := WriteProjectOnboardingState(path, state); err != nil {
		t.Fatalf("WriteProjectOnboardingState() error = %v", err)
	}
	got, err := ReadProjectOnboardingState(path)
	if err != nil {
		t.Fatalf("ReadProjectOnboardingState() error = %v", err)
	}
	if got != state {
		t.Fatalf("state mismatch: got %+v want %+v", got, state)
	}
}

func TestIncrementAndCompleteOnboardingHelpers(t *testing.T) {
	path := filepath.Join(t.TempDir(), "project-onboarding-state.json")
	now := time.Now().UTC().Unix()
	state, err := IncrementOnboardingSeenCount(path, "s1", now)
	if err != nil {
		t.Fatalf("IncrementOnboardingSeenCount() error = %v", err)
	}
	if state.SeenCount != 1 || state.LastShownSessionID != "s1" {
		t.Fatalf("unexpected state after increment: %+v", state)
	}
	state, err = MarkOnboardingCompleted(path, "s1", now+1)
	if err != nil {
		t.Fatalf("MarkOnboardingCompleted() error = %v", err)
	}
	if !state.HasCompleted || state.LastCompletedSessionID != "s1" {
		t.Fatalf("unexpected state after completion: %+v", state)
	}
}
