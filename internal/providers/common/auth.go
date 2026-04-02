package common

import "os"

// ResolveAuthValue returns provider configuration using env > config > fallback precedence.
func ResolveAuthValue(envVar, configured, fallback string) string {
	if v := os.Getenv(envVar); v != "" {
		return v
	}
	if configured != "" {
		return configured
	}
	return fallback
}
