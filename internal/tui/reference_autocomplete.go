package tui

import (
	"strings"

	"github.com/alliecatowo/alliecode/internal/references"
)

type referenceToken struct {
	Start int
	End   int
	Query string
}

func currentReferenceToken(input string) (referenceToken, bool) {
	if input == "" {
		return referenceToken{}, false
	}
	cursor := len(input)
	i := cursor - 1
	for i >= 0 && isReferenceTokenChar(input[i]) {
		i--
	}
	start := i + 1
	if start <= 0 || start > cursor {
		return referenceToken{}, false
	}
	if input[start-1] != '@' {
		return referenceToken{}, false
	}
	if start-1 > 0 && isReferenceWordLike(input[start-2]) {
		return referenceToken{}, false
	}

	query := input[start:cursor]
	if strings.Contains(query, "#") {
		return referenceToken{}, false
	}
	return referenceToken{Start: start - 1, End: cursor, Query: query}, true
}

func buildReferenceInsertion(input string, token referenceToken, suggestion references.Suggestion) string {
	if token.Start < 0 || token.End < token.Start || token.End > len(input) {
		return input
	}
	insert := "@" + strings.TrimSpace(suggestion.Path)
	if insert == "@" {
		return input
	}
	next := input[:token.Start] + insert + input[token.End:]
	insertEnd := token.Start + len(insert)
	if insertEnd >= len(next) {
		return next + " "
	}
	if !isInlineSpace(next[insertEnd]) {
		return next[:insertEnd] + " " + next[insertEnd:]
	}
	return next
}

func isInlineSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

func isReferenceWordLike(b byte) bool {
	return (b >= 'a' && b <= 'z') ||
		(b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9') ||
		b == '_' || b == '.' || b == '-'
}

func isReferenceTokenChar(b byte) bool {
	if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') {
		return true
	}
	switch b {
	case '/', '\\', '.', '_', '-', '~', ':':
		return true
	default:
		return false
	}
}
