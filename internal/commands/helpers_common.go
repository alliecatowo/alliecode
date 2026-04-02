package commands

import "strings"

func joinTrimmed(parts []string) string {
	return strings.TrimSpace(strings.Join(parts, " "))
}

func equalFoldTrimmed(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}

func hasAnyPrefixFold(value string, prefixes ...string) bool {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	for _, prefix := range prefixes {
		if strings.HasPrefix(trimmed, strings.ToLower(strings.TrimSpace(prefix))) {
			return true
		}
	}
	return false
}
