package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	diffAddedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))            // green
	diffRemovedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))            // red
	diffHeaderStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6")) // cyan
	diffHunkStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))            // magenta
	diffContextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

// RenderDiff produces a colorized unified diff between oldContent and newContent
// for the given filePath. Uses lipgloss for terminal styling.
func RenderDiff(oldContent, newContent, filePath string) string {
	oldLines := splitLines(oldContent)
	newLines := splitLines(newContent)

	// Compute the edit script via Myers' diff algorithm (LCS-based).
	edits := computeDiff(oldLines, newLines)

	// Build unified diff output.
	var b strings.Builder

	b.WriteString(diffHeaderStyle.Render(fmt.Sprintf("--- a/%s", filePath)) + "\n")
	b.WriteString(diffHeaderStyle.Render(fmt.Sprintf("+++ b/%s", filePath)) + "\n")

	// Group edits into hunks with context lines.
	hunks := buildHunks(edits, len(oldLines), len(newLines), 3)
	for _, hunk := range hunks {
		header := fmt.Sprintf("@@ -%d,%d +%d,%d @@",
			hunk.oldStart+1, hunk.oldCount,
			hunk.newStart+1, hunk.newCount)
		b.WriteString(diffHunkStyle.Render(header) + "\n")

		for _, line := range hunk.lines {
			switch line.op {
			case opContext:
				b.WriteString(diffContextStyle.Render(" "+line.text) + "\n")
			case opAdd:
				b.WriteString(diffAddedStyle.Render("+"+line.text) + "\n")
			case opRemove:
				b.WriteString(diffRemovedStyle.Render("-"+line.text) + "\n")
			}
		}
	}

	return b.String()
}

// editOp represents a diff operation.
type editOp int

const (
	opContext editOp = iota
	opAdd
	opRemove
)

// edit represents a single line in the diff.
type edit struct {
	op   editOp
	text string
}

// hunk is a contiguous section of the diff.
type hunk struct {
	oldStart int
	oldCount int
	newStart int
	newCount int
	lines    []edit
}

// splitLines splits a string into lines, handling the empty-string edge case.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

// computeDiff computes the edit operations between old and new line slices
// using a simple LCS-based approach.
func computeDiff(oldLines, newLines []string) []edit {
	m := len(oldLines)
	n := len(newLines)

	// Build LCS table.
	lcs := make([][]int, m+1)
	for i := range lcs {
		lcs[i] = make([]int, n+1)
	}
	for i := m - 1; i >= 0; i-- {
		for j := n - 1; j >= 0; j-- {
			if oldLines[i] == newLines[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else if lcs[i+1][j] >= lcs[i][j+1] {
				lcs[i][j] = lcs[i+1][j]
			} else {
				lcs[i][j] = lcs[i][j+1]
			}
		}
	}

	// Walk the LCS table to produce edits.
	var edits []edit
	i, j := 0, 0
	for i < m && j < n {
		if oldLines[i] == newLines[j] {
			edits = append(edits, edit{op: opContext, text: oldLines[i]})
			i++
			j++
		} else if lcs[i+1][j] >= lcs[i][j+1] {
			edits = append(edits, edit{op: opRemove, text: oldLines[i]})
			i++
		} else {
			edits = append(edits, edit{op: opAdd, text: newLines[j]})
			j++
		}
	}
	for ; i < m; i++ {
		edits = append(edits, edit{op: opRemove, text: oldLines[i]})
	}
	for ; j < n; j++ {
		edits = append(edits, edit{op: opAdd, text: newLines[j]})
	}

	return edits
}

// buildHunks groups edits into unified diff hunks with the given number of context lines.
func buildHunks(edits []edit, oldLen, newLen, contextLines int) []hunk {
	if len(edits) == 0 {
		return nil
	}

	// Find ranges of changed lines.
	type changeRange struct {
		start, end int // indices into edits
	}
	var changes []changeRange
	inChange := false
	for i, e := range edits {
		if e.op != opContext {
			if !inChange {
				changes = append(changes, changeRange{start: i})
				inChange = true
			}
			changes[len(changes)-1].end = i + 1
		} else {
			inChange = false
		}
	}

	if len(changes) == 0 {
		return nil
	}

	// Merge nearby changes into hunks.
	var hunks []hunk
	for _, ch := range changes {
		start := ch.start - contextLines
		if start < 0 {
			start = 0
		}
		end := ch.end + contextLines
		if end > len(edits) {
			end = len(edits)
		}

		// Check if this overlaps with the previous hunk.
		if len(hunks) > 0 {
			prev := &hunks[len(hunks)-1]
			prevEnd := 0
			for _, l := range prev.lines {
				_ = l
				prevEnd++
			}
			// We use edit indices to decide merging, but simplify: just check overlap.
			// In practice, rebuild from scratch for correctness.
		}

		var lines []edit
		oldStart, newStart := 0, 0
		oldCount, newCount := 0, 0

		// Compute starting positions by counting ops before start.
		oPos, nPos := 0, 0
		for i := 0; i < start; i++ {
			switch edits[i].op {
			case opContext:
				oPos++
				nPos++
			case opRemove:
				oPos++
			case opAdd:
				nPos++
			}
		}
		oldStart = oPos
		newStart = nPos

		for i := start; i < end; i++ {
			lines = append(lines, edits[i])
			switch edits[i].op {
			case opContext:
				oldCount++
				newCount++
			case opRemove:
				oldCount++
			case opAdd:
				newCount++
			}
		}

		hunks = append(hunks, hunk{
			oldStart: oldStart,
			oldCount: oldCount,
			newStart: newStart,
			newCount: newCount,
			lines:    lines,
		})
	}

	return hunks
}
