package commands

import (
	"errors"
	"strings"
)

var (
	ErrNotSlashCommand = errors.New("input is not a slash command")
	ErrEmptyCommand    = errors.New("slash command is empty")
	ErrUnclosedQuote   = errors.New("slash command contains an unclosed quote")
)

// Invocation is a parsed slash command.
type Invocation struct {
	Raw     string
	Name    string
	Args    []string
	IsSlash bool
}

// Parse parses a slash command line such as "/model gpt-4o".
func Parse(input string) (Invocation, error) {
	raw := strings.TrimSpace(input)
	if !strings.HasPrefix(raw, "/") {
		return Invocation{}, ErrNotSlashCommand
	}

	line := strings.TrimSpace(strings.TrimPrefix(raw, "/"))
	if line == "" {
		return Invocation{}, ErrEmptyCommand
	}

	tokens, err := tokenize(line)
	if err != nil {
		return Invocation{}, err
	}
	if len(tokens) == 0 {
		return Invocation{}, ErrEmptyCommand
	}

	return Invocation{
		Raw:     raw,
		Name:    strings.ToLower(tokens[0]),
		Args:    tokens[1:],
		IsSlash: true,
	}, nil
}

func tokenize(input string) ([]string, error) {
	var tokens []string
	for i := 0; i < len(input); {
		for i < len(input) && (input[i] == ' ' || input[i] == '\t') {
			i++
		}
		if i >= len(input) {
			break
		}

		if input[i] == '"' || input[i] == '\'' {
			quote := input[i]
			i++
			var b strings.Builder
			for i < len(input) {
				if input[i] == '\\' && i+1 < len(input) {
					i++
					b.WriteByte(input[i])
					i++
					continue
				}
				if input[i] == quote {
					i++
					tokens = append(tokens, b.String())
					goto next
				}
				b.WriteByte(input[i])
				i++
			}
			return nil, ErrUnclosedQuote
		}

		{
			var b strings.Builder
			for i < len(input) && input[i] != ' ' && input[i] != '\t' {
				if input[i] == '\\' && i+1 < len(input) {
					i++
					b.WriteByte(input[i])
					i++
					continue
				}
				b.WriteByte(input[i])
				i++
			}
			tokens = append(tokens, b.String())
		}

	next:
	}

	return tokens, nil
}
