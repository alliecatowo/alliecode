package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/alliecatowo/alliecode/internal/types"
)

// EstimateTokens gives a rough token count for a slice of messages.
// Uses the heuristic that 1 token ~= 4 characters.
func EstimateTokens(messages []types.Message) int {
	totalChars := 0
	for _, m := range messages {
		for _, block := range m.Content {
			switch block.Type {
			case types.ContentText, types.ContentThinking:
				totalChars += len(block.Text)
			case types.ContentToolUse:
				totalChars += len(block.ToolName) + len(block.Input)
			case types.ContentToolResult:
				totalChars += len(block.Content)
			}
		}
		// Small overhead per message for role, structure, etc.
		totalChars += 16
	}
	return totalChars / 4
}

// ShouldCompact returns true when the estimated token count of the conversation
// exceeds 80% of the context window, indicating compaction is advisable.
func ShouldCompact(messages []types.Message, contextWindow int) bool {
	if contextWindow <= 0 || len(messages) < 4 {
		return false
	}
	estimated := EstimateTokens(messages)
	threshold := int(float64(contextWindow) * 0.80)
	return estimated > threshold
}

// CompactMessages summarizes older messages to free context space.
//
// Strategy:
//  1. Keep the first message (usually the initial user prompt — important context).
//  2. Keep the last N messages intact (recent context is critical).
//  3. Summarize the middle chunk via the provider.
//  4. Return [first_message, summary_message, ...recent_messages].
func CompactMessages(
	ctx context.Context,
	messages []types.Message,
	provider types.Provider,
	model string,
) ([]types.Message, *types.CompactionBoundary, error) {
	if len(messages) < 6 {
		// Too few messages to compact meaningfully.
		return messages, nil, nil
	}

	// Keep first message and last 4 messages; summarize the middle.
	const keepRecent = 4
	first := messages[0]
	recent := messages[len(messages)-keepRecent:]
	middle := messages[1 : len(messages)-keepRecent]

	if len(middle) == 0 {
		return messages, nil, nil
	}

	// Build a summary request.
	summaryPrompt := buildSummaryPrompt(middle)

	req := types.ChatRequest{
		Messages: []types.Message{
			types.NewTextMessage(types.RoleUser, summaryPrompt),
		},
		System:    "You are a conversation summarizer. Produce a concise summary of the conversation, preserving all key decisions, file paths, code changes, tool results, and important context. Be thorough but compact.",
		Model:     model,
		MaxTokens: 4096,
	}

	resp, err := provider.ChatSync(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("compaction summary request failed: %w", err)
	}

	summaryText := resp.Message.GetText()
	if summaryText == "" {
		return nil, nil, fmt.Errorf("compaction produced empty summary")
	}

	replacement := []types.Message{
		types.NewTextMessage(types.RoleUser,
			fmt.Sprintf("[Conversation compacted - summary of %d earlier messages]\n\n%s", len(middle), summaryText)),
		// Need an assistant ack so the conversation alternation is valid.
		types.NewTextMessage(types.RoleAssistant,
			"Understood. I have the context from the summarized conversation. Continuing."),
	}

	// Build the compacted conversation.
	compacted := make([]types.Message, 0, 2+len(recent))
	compacted = append(compacted, first)
	compacted = append(compacted, replacement...)
	compacted = append(compacted, recent...)

	boundary := &types.CompactionBoundary{
		StartIndex:  1,
		EndIndex:    len(messages) - keepRecent,
		Replacement: replacement,
	}

	return compacted, boundary, nil
}

// buildSummaryPrompt formats the middle messages into a prompt for summarization.
func buildSummaryPrompt(messages []types.Message) string {
	var sb strings.Builder
	sb.WriteString("Summarize this conversation excerpt. Preserve all key information including file paths, code changes, decisions, and tool results.\n\n---\n\n")

	for _, m := range messages {
		sb.WriteString(fmt.Sprintf("[%s]\n", m.Role))
		for _, block := range m.Content {
			switch block.Type {
			case types.ContentText:
				sb.WriteString(block.Text)
				sb.WriteByte('\n')
			case types.ContentToolUse:
				sb.WriteString(fmt.Sprintf("Tool call: %s (id: %s)\nInput: %s\n", block.ToolName, block.ToolUseID, string(block.Input)))
			case types.ContentToolResult:
				content := block.Content
				// Truncate very long tool results in the summary prompt.
				if len(content) > 2000 {
					content = content[:2000] + "\n... [truncated]"
				}
				label := "Result"
				if block.IsError {
					label = "Error"
				}
				sb.WriteString(fmt.Sprintf("Tool %s (for: %s): %s\n", label, block.ForToolUseID, content))
			case types.ContentThinking:
				// Skip thinking blocks in summary prompt to save space.
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
