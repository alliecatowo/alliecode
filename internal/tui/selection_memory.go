package tui

import (
	"slices"
	"strings"
)

func recallClosestSelection(memory map[string]string, query string) string {
	if len(memory) == 0 {
		return ""
	}
	normQuery := strings.TrimSpace(strings.ToLower(query))
	if remembered := strings.TrimSpace(memory[normQuery]); remembered != "" {
		return remembered
	}
	bestKey := ""
	bestValue := ""
	bestScore := -1
	keys := make([]string, 0, len(memory))
	for key := range memory {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		value := memory[key]
		key = strings.TrimSpace(strings.ToLower(key))
		value = strings.TrimSpace(strings.ToLower(value))
		if key == "" || value == "" {
			continue
		}
		score := selectionQueryAffinity(normQuery, key)
		if score < 0 {
			continue
		}
		if score > bestScore || (score == bestScore && len(key) > len(bestKey)) {
			bestKey = key
			bestValue = value
			bestScore = score
		}
	}
	return bestValue
}

func selectionQueryAffinity(query, candidate string) int {
	query = strings.TrimSpace(strings.ToLower(query))
	candidate = strings.TrimSpace(strings.ToLower(candidate))
	if candidate == "" {
		return -1
	}
	if query == "" {
		if candidate == "" {
			return 0
		}
		return 1
	}
	if query == candidate {
		return 1000
	}
	if strings.HasPrefix(query, candidate) || strings.HasPrefix(candidate, query) {
		return 800 + min(len(query), len(candidate))
	}
	queryTokens := strings.Fields(query)
	candidateTokens := strings.Fields(candidate)
	shared := 0
	for _, qt := range queryTokens {
		for _, ct := range candidateTokens {
			if qt == ct {
				shared++
				break
			}
		}
	}
	if shared > 0 {
		return 500 + shared*20
	}
	if isSubsequence(query, candidate) || isSubsequence(candidate, query) {
		return 200 + min(len(query), len(candidate))
	}
	return -1
}
