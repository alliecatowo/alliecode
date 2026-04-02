package types

import "strings"

// IsRoleMessage reports whether the message uses the given role.
func IsRoleMessage(msg Message, role Role) bool {
	return msg.Role == role
}

// IsUserMessage reports whether the message is user-authored.
func IsUserMessage(msg Message) bool {
	return msg.Role == RoleUser
}

// IsAssistantMessage reports whether the message is assistant-authored.
func IsAssistantMessage(msg Message) bool {
	return msg.Role == RoleAssistant
}

// IsSystemMessage reports whether the message is system-authored.
func IsSystemMessage(msg Message) bool {
	return msg.Role == RoleSystem
}

// IsContentBlockType reports whether a content block has the requested type.
func IsContentBlockType(block ContentBlock, blockType ContentBlockType) bool {
	return block.Type == blockType
}

// IsToolUseBlock reports whether a content block is a valid tool use block.
func IsToolUseBlock(block ContentBlock) bool {
	return block.Type == ContentToolUse && strings.TrimSpace(block.ToolUseID) != "" && strings.TrimSpace(block.ToolName) != ""
}

// IsToolResultBlock reports whether a content block is a valid tool result block.
func IsToolResultBlock(block ContentBlock) bool {
	return block.Type == ContentToolResult && strings.TrimSpace(block.ForToolUseID) != ""
}

// IsThinkingBlock reports whether a block carries model-thinking content.
func IsThinkingBlock(block ContentBlock) bool {
	return block.Type == ContentThinking && strings.TrimSpace(block.Text) != ""
}

// MessageHasContentType reports whether any content block matches blockType.
func MessageHasContentType(msg Message, blockType ContentBlockType) bool {
	for _, block := range msg.Content {
		if block.Type == blockType {
			return true
		}
	}
	return false
}

// MessageHasToolUse reports whether the message contains at least one tool use.
func MessageHasToolUse(msg Message) bool {
	for _, block := range msg.Content {
		if IsToolUseBlock(block) {
			return true
		}
	}
	return false
}

// MessageHasToolResult reports whether the message contains at least one tool result.
func MessageHasToolResult(msg Message) bool {
	for _, block := range msg.Content {
		if IsToolResultBlock(block) {
			return true
		}
	}
	return false
}

// MessageHasThinking reports whether the message contains model-thinking content.
func MessageHasThinking(msg Message) bool {
	for _, block := range msg.Content {
		if IsThinkingBlock(block) {
			return true
		}
	}
	return false
}
