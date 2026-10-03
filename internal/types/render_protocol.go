package types

import (
	"regexp"
	"strings"
)

var contractHeaderPattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]{2,}$`)

// LooksLikeStructuredContract reports whether text appears to be a legacy
// command/tool protocol contract payload instead of free-form assistant prose.
func LooksLikeStructuredContract(text string) bool {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return false
	}
	lines := strings.Split(trimmed, "\n")
	nonEmpty := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			nonEmpty = append(nonEmpty, line)
		}
	}
	if len(nonEmpty) < 2 {
		return false
	}
	if !contractHeaderPattern.MatchString(nonEmpty[0]) {
		return false
	}
	for _, line := range nonEmpty[1:] {
		if strings.Contains(line, "=") {
			return true
		}
	}
	return false
}
