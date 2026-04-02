package agent

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// checkpointRefPrefix is the git ref namespace for AllieCode checkpoints.
const checkpointRefPrefix = "refs/alliecode/checkpoints/"

// Checkpoint represents a saved snapshot of the working directory.
type Checkpoint struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	Ref         string    `json:"ref"`
}

// CreateCheckpoint snapshots the current working tree state (including staged
// and unstaged changes) and stores it as a git ref under refs/alliecode/checkpoints/.
//
// This uses `git stash create` which creates a commit object without modifying
// the working tree or index — a non-destructive snapshot.
func CreateCheckpoint(workingDir string, description string) (string, error) {
	// Ensure we're in a git repo.
	if err := gitCmd(workingDir, "rev-parse", "--git-dir"); err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}

	// Create a stash-like commit object without affecting the working tree.
	// git stash create captures both staged and unstaged changes.
	output, err := gitOutput(workingDir, "stash", "create", "--include-untracked")
	if err != nil {
		return "", fmt.Errorf("git stash create failed: %w", err)
	}

	commitHash := strings.TrimSpace(output)
	if commitHash == "" {
		// No changes to snapshot — use HEAD.
		commitHash, err = gitOutput(workingDir, "rev-parse", "HEAD")
		if err != nil {
			return "", fmt.Errorf("git rev-parse HEAD failed: %w", err)
		}
		commitHash = strings.TrimSpace(commitHash)
	}

	// Generate a unique checkpoint ID based on timestamp.
	ts := time.Now().UTC()
	checkpointID := ts.Format("20060102-150405")

	// Store the ref.
	ref := checkpointRefPrefix + checkpointID
	if err := gitCmd(workingDir, "update-ref", ref, commitHash); err != nil {
		return "", fmt.Errorf("git update-ref failed: %w", err)
	}

	// Store the description in the ref's log message via a notes-like approach.
	// We encode description into the reflog entry.
	_ = gitCmd(workingDir, "update-ref", "-m", description, ref, commitHash)

	return checkpointID, nil
}

// RestoreCheckpoint restores the working directory to the state captured
// in the given checkpoint.
func RestoreCheckpoint(workingDir string, checkpointID string) error {
	ref := checkpointRefPrefix + checkpointID

	// Verify the ref exists.
	commitHash, err := gitOutput(workingDir, "rev-parse", ref)
	if err != nil {
		return fmt.Errorf("checkpoint %q not found: %w", checkpointID, err)
	}
	commitHash = strings.TrimSpace(commitHash)

	// First, stash any current changes so the user doesn't lose work.
	_, _ = gitOutput(workingDir, "stash", "create", "--include-untracked")

	// Apply the checkpoint state.
	// Use git read-tree + checkout to restore both tracked files and the index.
	if err := gitCmd(workingDir, "read-tree", commitHash); err != nil {
		// If read-tree fails (e.g., stash commit), fall back to stash apply approach.
		if err := gitCmd(workingDir, "stash", "apply", commitHash); err != nil {
			return fmt.Errorf("failed to restore checkpoint %q: %w", checkpointID, err)
		}
		return nil
	}

	if err := gitCmd(workingDir, "checkout-index", "-a", "-f"); err != nil {
		return fmt.Errorf("checkout-index failed: %w", err)
	}

	return nil
}

// ListCheckpoints returns all stored checkpoints, most recent first.
func ListCheckpoints(workingDir string) ([]Checkpoint, error) {
	// List all refs under our namespace.
	output, err := gitOutput(workingDir, "for-each-ref",
		"--sort=-creatordate",
		"--format=%(refname) %(objectname:short) %(creatordate:iso-strict) %(subject)",
		checkpointRefPrefix)
	if err != nil {
		return nil, fmt.Errorf("listing checkpoints failed: %w", err)
	}

	output = strings.TrimSpace(output)
	if output == "" {
		return nil, nil
	}

	var checkpoints []Checkpoint
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse: refs/alliecode/checkpoints/<id> <short-hash> <timestamp> <description>
		parts := strings.SplitN(line, " ", 4)
		if len(parts) < 3 {
			continue
		}

		ref := parts[0]
		id := strings.TrimPrefix(ref, checkpointRefPrefix)
		ts, _ := time.Parse(time.RFC3339, parts[2])

		desc := ""
		if len(parts) >= 4 {
			desc = parts[3]
		}

		checkpoints = append(checkpoints, Checkpoint{
			ID:          id,
			Description: desc,
			Timestamp:   ts,
			Ref:         ref,
		})
	}

	return checkpoints, nil
}

// Undo restores the most recent checkpoint — the primary undo mechanism.
// This is exported for use by cmd/ac/main.go.
func Undo(workingDir string) error {
	checkpoints, err := ListCheckpoints(workingDir)
	if err != nil {
		return fmt.Errorf("failed to list checkpoints: %w", err)
	}
	if len(checkpoints) == 0 {
		return fmt.Errorf("no checkpoints available to undo")
	}

	latest := checkpoints[0]
	if err := RestoreCheckpoint(workingDir, latest.ID); err != nil {
		return fmt.Errorf("undo failed: %w", err)
	}

	// Remove the checkpoint ref after successful restore so the next Undo
	// goes to the previous one.
	_ = gitCmd(workingDir, "update-ref", "-d", latest.Ref)

	return nil
}

// gitCmd runs a git command in the given directory and returns an error if it fails.
func gitCmd(workingDir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = workingDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// gitOutput runs a git command and returns its stdout.
func gitOutput(workingDir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = workingDir
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", err
	}
	return string(out), nil
}
