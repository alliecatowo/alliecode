package state

import (
	"path/filepath"
	"sort"
	"strings"
)

func normalizeEquals(input string) string {
	return strings.ToLower(strings.TrimSpace(input))
}

func matchesEquals(value, expected string) bool {
	if expected == "" {
		return true
	}
	return normalizeEquals(value) == expected
}

func normalizeContains(input string) string {
	return strings.ToLower(strings.TrimSpace(input))
}

func matchesContains(haystack, needle string) bool {
	if needle == "" {
		return true
	}
	return strings.Contains(strings.ToLower(strings.TrimSpace(haystack)), needle)
}

func clampLimit(limit int, size int) int {
	if limit <= 0 || limit > size {
		return size
	}
	return limit
}

func paginateBounds(size, offset, limit int) (start, end int) {
	if size <= 0 {
		return 0, 0
	}
	if offset < 0 {
		offset = 0
	}
	if offset > size {
		offset = size
	}
	start = offset
	max := clampLimit(limit, size-start)
	end = start + max
	if end > size {
		end = size
	}
	return start, end
}

func sortByStringStable(values []string) {
	sort.SliceStable(values, func(i, j int) bool {
		return values[i] < values[j]
	})
}

func cleanPath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}
	return filepath.Clean(trimmed)
}
