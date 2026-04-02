package permissions

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type matchMode int

const (
	matchModeGlob matchMode = iota
	matchModeRegex
	matchModeShell
)

func matchDSL(pattern, value string, defaultMode matchMode) (bool, error) {
	trimmed := strings.TrimSpace(pattern)
	if trimmed == "" {
		return false, fmt.Errorf("matcher pattern cannot be empty")
	}

	switch {
	case strings.HasPrefix(trimmed, "regex:"):
		body := strings.TrimSpace(strings.TrimPrefix(trimmed, "regex:"))
		if body == "" {
			return false, fmt.Errorf("regex matcher cannot be empty")
		}
		return matchRegex(body, value)
	case strings.HasPrefix(trimmed, "glob:"):
		body := strings.TrimSpace(strings.TrimPrefix(trimmed, "glob:"))
		if body == "" {
			return false, fmt.Errorf("glob matcher cannot be empty")
		}
		return matchGlob(body, value)
	default:
		if defaultMode == matchModeShell {
			return matchShellRule(trimmed, value)
		}
		if defaultMode == matchModeRegex {
			return matchRegex(trimmed, value)
		}
		return matchGlob(trimmed, value)
	}
}

func validateMatcher(pattern string, defaultMode matchMode) error {
	_, err := matchDSL(pattern, "", defaultMode)
	return err
}

func matchGlob(pattern, value string) (bool, error) {
	if strings.Contains(pattern, "**") {
		return matchDoubleStarGlob(pattern, value)
	}
	matched, err := filepath.Match(pattern, value)
	if err != nil {
		return false, fmt.Errorf("invalid glob pattern %q: %w", pattern, err)
	}
	return matched, nil
}

func matchDoubleStarGlob(pattern, value string) (bool, error) {
	normalizedPattern := filepath.ToSlash(pattern)
	normalizedValue := filepath.ToSlash(value)

	regexBody := globPatternToRegex(normalizedPattern)
	re, err := regexp.Compile("^" + regexBody + "$")
	if err != nil {
		return false, fmt.Errorf("invalid glob pattern %q: %w", pattern, err)
	}
	return re.MatchString(normalizedValue), nil
}

func globPatternToRegex(pattern string) string {
	var b strings.Builder
	for i := 0; i < len(pattern); i++ {
		ch := pattern[i]
		if ch == '*' {
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				b.WriteString(".*")
				i++
				continue
			}
			b.WriteString("[^/]*")
			continue
		}
		if ch == '?' {
			b.WriteString("[^/]")
			continue
		}
		b.WriteString(regexp.QuoteMeta(string(ch)))
	}
	return b.String()
}

func matchRegex(pattern, value string) (bool, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, fmt.Errorf("invalid regex pattern %q: %w", pattern, err)
	}
	return re.MatchString(value), nil
}

func matchShellRule(pattern, command string) (bool, error) {
	trimmed := strings.TrimSpace(pattern)
	if trimmed == "" {
		return false, fmt.Errorf("shell matcher cannot be empty")
	}

	if prefix, ok := shellPrefix(trimmed); ok {
		if command == prefix {
			return true, nil
		}
		return strings.HasPrefix(command, prefix+" "), nil
	}

	if hasUnescapedWildcard(trimmed) {
		return matchShellWildcard(trimmed, command)
	}

	if looksLikeRegex(trimmed) {
		return matchRegex(trimmed, command)
	}

	return trimmed == command, nil
}

func shellPrefix(pattern string) (string, bool) {
	if strings.HasSuffix(pattern, ":*") && len(pattern) > 2 {
		return pattern[:len(pattern)-2], true
	}
	return "", false
}

func hasUnescapedWildcard(pattern string) bool {
	if strings.HasSuffix(pattern, ":*") {
		return false
	}
	for i := 0; i < len(pattern); i++ {
		if pattern[i] != '*' {
			continue
		}
		backslashes := 0
		for j := i - 1; j >= 0 && pattern[j] == '\\'; j-- {
			backslashes++
		}
		if backslashes%2 == 0 {
			return true
		}
	}
	return false
}

func matchShellWildcard(pattern, command string) (bool, error) {
	var b strings.Builder
	b.Grow(len(pattern) + 8)

	unescapedStars := 0
	for i := 0; i < len(pattern); i++ {
		ch := pattern[i]
		if ch == '\\' && i+1 < len(pattern) {
			next := pattern[i+1]
			if next == '*' || next == '\\' {
				b.WriteString(regexp.QuoteMeta(string(next)))
				i++
				continue
			}
		}
		if ch == '*' {
			backslashes := 0
			for j := i - 1; j >= 0 && pattern[j] == '\\'; j-- {
				backslashes++
			}
			if backslashes%2 == 0 {
				b.WriteString(".*")
				unescapedStars++
				continue
			}
		}
		b.WriteString(regexp.QuoteMeta(string(ch)))
	}

	regexBody := b.String()
	if unescapedStars == 1 && strings.HasSuffix(pattern, " *") && strings.HasSuffix(regexBody, " .*") {
		regexBody = strings.TrimSuffix(regexBody, " .*") + "( .*)?"
	}

	re, err := regexp.Compile("(?s)^" + regexBody + "$")
	if err != nil {
		return false, fmt.Errorf("invalid shell wildcard pattern %q: %w", pattern, err)
	}
	return re.MatchString(command), nil
}

func looksLikeRegex(pattern string) bool {
	return strings.ContainsAny(pattern, "^$[](){}+?|\\")
}
