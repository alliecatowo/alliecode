package commands

import "strings"

func defaultLoginProvider(current string) string {
	provider := strings.TrimSpace(current)
	if provider == "" {
		return "anthropic"
	}
	return provider
}

func normalizeAccountName(raw string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(raw), " "))
}
