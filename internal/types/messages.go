// Package types defines AllieCode's universal message format.
// This is AllieCode's own schema — providers translate to/from their native formats.
package types

import "encoding/json"

// Role represents the sender of a message.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

// ContentBlockType identifies the kind of content in a message.
type ContentBlockType string

const (
	ContentText       ContentBlockType = "text"
	ContentToolUse    ContentBlockType = "tool_use"
	ContentToolResult ContentBlockType = "tool_result"
	ContentImage      ContentBlockType = "image"
	ContentThinking   ContentBlockType = "thinking"
)

// ContentBlock is a single piece of content within a message.
type ContentBlock struct {
	Type ContentBlockType `json:"type"`

	// Text content (when Type == ContentText or ContentThinking)
	Text string `json:"text,omitempty"`

	// Tool use (when Type == ContentToolUse)
	ToolUseID string          `json:"tool_use_id,omitempty"`
	ToolName  string          `json:"tool_name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`

	// Tool result (when Type == ContentToolResult)
	ForToolUseID string `json:"for_tool_use_id,omitempty"`
	Content      string `json:"content,omitempty"`
	IsError      bool   `json:"is_error,omitempty"`

	// Image (when Type == ContentImage)
	MediaType string `json:"media_type,omitempty"`
	ImageData []byte `json:"image_data,omitempty"`
	ImageURL  string `json:"image_url,omitempty"`
}

// Message is the universal message type used throughout AllieCode.
type Message struct {
	Role    Role           `json:"role"`
	Content []ContentBlock `json:"content"`
}

// CompactionBoundary describes a replay-safe transcript compaction operation.
// During replay, messages in [StartIndex:EndIndex] are replaced with Replacement.
type CompactionBoundary struct {
	StartIndex  int       `json:"start_index"`
	EndIndex    int       `json:"end_index"`
	Replacement []Message `json:"replacement"`
}

// NewTextMessage creates a simple text message.
func NewTextMessage(role Role, text string) Message {
	return Message{
		Role: role,
		Content: []ContentBlock{
			{Type: ContentText, Text: text},
		},
	}
}

// NewToolUseMessage creates an assistant message with a tool use block.
func NewToolUseMessage(toolUseID, toolName string, input json.RawMessage) Message {
	return Message{
		Role: RoleAssistant,
		Content: []ContentBlock{
			{
				Type:      ContentToolUse,
				ToolUseID: toolUseID,
				ToolName:  toolName,
				Input:     input,
			},
		},
	}
}

// NewToolResultMessage creates a user message with a tool result block.
func NewToolResultMessage(forToolUseID, content string, isError bool) Message {
	return Message{
		Role: RoleUser,
		Content: []ContentBlock{
			{
				Type:         ContentToolResult,
				ForToolUseID: forToolUseID,
				Content:      content,
				IsError:      isError,
			},
		},
	}
}

// GetText extracts concatenated text from all text content blocks.
func (m Message) GetText() string {
	var text string
	for _, block := range m.Content {
		if block.Type == ContentText {
			if text != "" {
				text += "\n"
			}
			text += block.Text
		}
	}
	return text
}

// GetToolUses extracts all tool use blocks from the message.
func (m Message) GetToolUses() []ContentBlock {
	var uses []ContentBlock
	for _, block := range m.Content {
		if block.Type == ContentToolUse {
			uses = append(uses, block)
		}
	}
	return uses
}

// StreamEventType identifies the kind of streaming event.
type StreamEventType string

const (
	StreamStart         StreamEventType = "stream_start"
	StreamRequestStart  StreamEventType = "request_start"
	StreamContentDelta  StreamEventType = "content_delta"
	StreamContentDone   StreamEventType = "content_done"
	StreamToolUseStart  StreamEventType = "tool_use_start"
	StreamToolUseDelta  StreamEventType = "tool_use_delta"
	StreamToolUseDone   StreamEventType = "tool_use_done"
	StreamToolBoundary  StreamEventType = "tool_boundary"
	StreamThinkingDelta StreamEventType = "thinking_delta"
	StreamUsageDelta    StreamEventType = "usage_delta"
	StreamUsageTotal    StreamEventType = "usage_total"
	StreamRetry         StreamEventType = "retry"
	StreamMessageDone   StreamEventType = "message_done"
	StreamError         StreamEventType = "error"
)

// StreamBoundaryType indicates a semantic edge in stream control-flow.
type StreamBoundaryType string

const (
	StreamBoundaryToolBegin StreamBoundaryType = "tool_begin"
	StreamBoundaryToolEnd   StreamBoundaryType = "tool_end"
	StreamBoundaryTurnEnd   StreamBoundaryType = "turn_end"
)

// StreamRetryKind indicates why a request retried.
type StreamRetryKind string

const (
	StreamRetryProviderError StreamRetryKind = "provider_error"
	StreamRetryPromptTooLong StreamRetryKind = "prompt_too_long"
	StreamRetryMaxTokens     StreamRetryKind = "max_tokens"
	StreamRetryToolError     StreamRetryKind = "tool_error"
)

// StreamEvent represents a single event in a streaming response.
type StreamEvent struct {
	Type StreamEventType `json:"type"`

	Provider string `json:"provider,omitempty"`
	Model    string `json:"model,omitempty"`

	Boundary     StreamBoundaryType `json:"boundary,omitempty"`
	RetryKind    StreamRetryKind    `json:"retry_kind,omitempty"`
	RetryAttempt int                `json:"retry_attempt,omitempty"`
	RetryMax     int                `json:"retry_max,omitempty"`

	// For content deltas
	Delta string `json:"delta,omitempty"`

	// For tool use events
	ToolUseID string          `json:"tool_use_id,omitempty"`
	ToolName  string          `json:"tool_name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`

	// For message_done
	Message *Message `json:"message,omitempty"`

	// For errors
	Error error `json:"-"`

	// Usage tracking
	Usage *Usage `json:"usage,omitempty"`

	// UsageCumulative indicates Usage is a cumulative snapshot for the request,
	// not an incremental delta for this single event.
	UsageCumulative bool `json:"usage_cumulative,omitempty"`

	// StopReason carries provider stop metadata when available in stream events.
	StopReason StopReason `json:"stop_reason,omitempty"`
}

// Usage tracks token consumption.
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	CacheHits    int `json:"cache_hits,omitempty"`
}

// StopReason indicates why the model stopped generating.
type StopReason string

const (
	StopEndTurn       StopReason = "end_turn"
	StopMaxTokens     StopReason = "max_tokens"
	StopContextLimit  StopReason = "context_window_exceeded"
	StopToolUse       StopReason = "tool_use"
	StopStopSequence  StopReason = "stop_sequence"
	StopContentFilter StopReason = "content_filter"
	StopRefusal       StopReason = "refusal"
	StopCanceled      StopReason = "canceled"
	StopError         StopReason = "error"
	StopUnknown       StopReason = "unknown"
)

// ChatResponse is the complete response from a model.
type ChatResponse struct {
	Message    Message    `json:"message"`
	Usage      Usage      `json:"usage"`
	StopReason StopReason `json:"stop_reason"`
	Model      string     `json:"model"`
}
