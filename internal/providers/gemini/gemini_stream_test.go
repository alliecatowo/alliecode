package gemini

import (
	"io"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStreamResponseParsesTextToolAndUsage(t *testing.T) {
	provider := &Provider{}
	body := strings.Join([]string{
		"data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"hello\"}]},\"finishReason\":\"STOP\"}],\"usageMetadata\":{\"promptTokenCount\":4,\"candidatesTokenCount\":2}}",
		"data: {\"candidates\":[{\"content\":{\"parts\":[{\"functionCall\":{\"name\":\"fake\",\"args\":{\"x\":1}}}]},\"finishReason\":\"STOP\"}],\"usageMetadata\":{\"promptTokenCount\":5,\"candidatesTokenCount\":3}}",
	}, "\n") + "\n"

	ch := make(chan types.StreamEvent, 16)
	go provider.streamResponse(io.NopCloser(strings.NewReader(body)), ch)

	var sawText, sawToolDone, sawUsage, sawMessageDone bool
	for ev := range ch {
		if ev.Type == types.StreamContentDelta && ev.Delta == "hello" {
			sawText = true
		}
		if ev.Type == types.StreamToolUseDone && ev.ToolName == "fake" {
			sawToolDone = true
		}
		if ev.Usage != nil {
			sawUsage = true
			if !ev.UsageCumulative {
				t.Fatalf("usage should be cumulative")
			}
		}
		if ev.Type == types.StreamMessageDone {
			sawMessageDone = true
			if ev.StopReason != types.StopEndTurn {
				t.Fatalf("stop reason = %q, want %q", ev.StopReason, types.StopEndTurn)
			}
		}
	}

	if !sawText {
		t.Fatalf("expected text delta")
	}
	if !sawToolDone {
		t.Fatalf("expected tool done event")
	}
	if !sawUsage {
		t.Fatalf("expected usage event")
	}
	if !sawMessageDone {
		t.Fatalf("expected message done event")
	}
}

func TestBuildRequestMapsToolsAndSystem(t *testing.T) {
	provider := &Provider{}
	req := types.ChatRequest{
		System: "sys",
		Messages: []types.Message{
			types.NewTextMessage(types.RoleUser, "hello"),
		},
		Tools: []types.ToolDef{
			{
				Name:        "fake",
				Description: "desc",
				InputSchema: types.ToolSchema{
					Type: "object",
					Properties: map[string]types.PropertySchema{
						"x": {Type: "number"},
					},
					Required: []string{"x"},
				},
			},
		},
	}

	apiReq := provider.buildRequest(req)
	if apiReq.SystemInstruction == nil || len(apiReq.SystemInstruction.Parts) == 0 || apiReq.SystemInstruction.Parts[0].Text != "sys" {
		t.Fatalf("expected system instruction to be set")
	}
	if len(apiReq.Tools) != 1 || len(apiReq.Tools[0].FunctionDeclarations) != 1 {
		t.Fatalf("expected one function declaration")
	}
	decl := apiReq.Tools[0].FunctionDeclarations[0]
	if decl.Name != "fake" {
		t.Fatalf("tool name = %q, want fake", decl.Name)
	}
	params, ok := decl.Parameters["properties"].(map[string]any)
	if !ok || len(params) == 0 {
		t.Fatalf("expected tool schema properties")
	}
}
