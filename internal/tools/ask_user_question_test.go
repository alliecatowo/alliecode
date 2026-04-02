package tools

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestAskUserQuestionToolNeedsUserInput(t *testing.T) {
	in := `{"questions":[{"id":"q1","question":"Pick one?","options":[{"id":"a","label":"A"},{"id":"b","label":"B"}]}]}`
	res, err := (&AskUserQuestionTool{}).Execute(context.Background(), []byte(in), types.ToolContext{})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %s", res.Content)
	}

	var out map[string]any
	if err := json.Unmarshal([]byte(res.Content), &out); err != nil {
		t.Fatalf("invalid json output: %v", err)
	}
	if out["status"] != "needs_user_input" {
		t.Fatalf("unexpected status: %v", out["status"])
	}
	summary, _ := out["summary"].(map[string]any)
	if summary["question_count"] != float64(1) || summary["answered_count"] != float64(0) {
		t.Fatalf("unexpected summary: %v", summary)
	}
	if summary["pending_count"] != float64(1) || summary["all_questions_answered"] != false {
		t.Fatalf("unexpected pending summary: %v", summary)
	}
}

func TestAskUserQuestionToolAnswered(t *testing.T) {
	in := `{"questions":[{"id":"q1","question":"Pick one?","options":[{"id":"a","label":"A","preview":"<div>A</div>"},{"id":"b","label":"B"}]}],"answers":[{"question_id":"q1","option_ids":[" a "],"notes":"go with a"}],"metadata":{"source":"planner"}}`
	res, err := (&AskUserQuestionTool{}).Execute(context.Background(), []byte(in), types.ToolContext{})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %s", res.Content)
	}

	var out map[string]any
	if err := json.Unmarshal([]byte(res.Content), &out); err != nil {
		t.Fatalf("invalid json output: %v", err)
	}
	if out["status"] != "answered" {
		t.Fatalf("unexpected status: %v", out["status"])
	}
	if out["metadata"] == nil {
		t.Fatalf("expected metadata passthrough")
	}
	resolved, _ := out["resolved_answers"].([]any)
	if len(resolved) != 1 {
		t.Fatalf("expected resolved answer, got %d", len(resolved))
	}
	answers, _ := out["answers"].([]any)
	if len(answers) != 1 {
		t.Fatalf("expected one answer")
	}
	answer0, _ := answers[0].(map[string]any)
	optionIDs, _ := answer0["option_ids"].([]any)
	if len(optionIDs) != 1 || optionIDs[0] != "a" {
		t.Fatalf("expected trimmed option id in answer payload, got %v", optionIDs)
	}
	summary, _ := out["summary"].(map[string]any)
	if summary["answered_count"] != float64(1) {
		t.Fatalf("unexpected summary: %v", summary)
	}
	if summary["pending_count"] != float64(0) || summary["all_questions_answered"] != true {
		t.Fatalf("unexpected completion summary: %v", summary)
	}

	const expectedPrefix = `{"status":"answered","questions":[{"id":"q1","question":"Pick one?","options":[{"id":"a","label":"A","preview":"\u003cdiv\u003eA\u003c/div\u003e"},{"id":"b","label":"B"}]}],"answers":[{"question_id":"q1","option_ids":["a"],"notes":"go with a"}],"resolved_answers":[{"question_id":"q1","option_ids":["a"],"options":[{"id":"a","label":"A","preview":"\u003cdiv\u003eA\u003c/div\u003e"}],"notes":"go with a"}],"summary":{"question_count":1,"answered_count":1,"pending_count":0,"answered_question_ids":["q1"],"all_questions_answered":true},"metadata":{"source":"planner"}}`
	if res.Content != expectedPrefix {
		t.Fatalf("expected deterministic output payload, got: %s", res.Content)
	}
}

func TestAskUserQuestionToolPartialStatusAndSummary(t *testing.T) {
	in := `{"questions":[{"id":"q1","question":"Pick one?","options":[{"id":"a","label":"A"},{"id":"b","label":"B"}]},{"id":"q2","question":"Pick one?","options":[{"id":"x","label":"X"},{"id":"y","label":"Y"}]}],"answers":[{"question_id":"q2","option_ids":["x"]}]}`
	res, err := (&AskUserQuestionTool{}).Execute(context.Background(), []byte(in), types.ToolContext{})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %s", res.Content)
	}

	var out map[string]any
	if err := json.Unmarshal([]byte(res.Content), &out); err != nil {
		t.Fatalf("invalid json output: %v", err)
	}
	if out["status"] != "partial" {
		t.Fatalf("unexpected status: %v", out["status"])
	}
	summary, _ := out["summary"].(map[string]any)
	if summary["pending_count"] != float64(1) || summary["answered_count"] != float64(1) {
		t.Fatalf("unexpected summary counts: %v", summary)
	}
	if summary["all_questions_answered"] != false {
		t.Fatalf("expected all_questions_answered=false: %v", summary)
	}
	pending, _ := summary["pending_question_ids"].([]any)
	if !reflect.DeepEqual(pending, []any{"q1"}) {
		t.Fatalf("unexpected pending question ids: %v", pending)
	}
	answered, _ := summary["answered_question_ids"].([]any)
	if !reflect.DeepEqual(answered, []any{"q2"}) {
		t.Fatalf("unexpected answered question ids: %v", answered)
	}
}

func TestAskUserQuestionToolValidatesAnswerOptionIDs(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{
			name: "duplicate option id in answer",
			in:   `{"questions":[{"id":"q1","question":"Pick one?","options":[{"id":"a","label":"A"},{"id":"b","label":"B"}]}],"answers":[{"question_id":"q1","option_ids":["a","a"]}]}`,
		},
		{
			name: "empty option id in answer",
			in:   `{"questions":[{"id":"q1","question":"Pick one?","options":[{"id":"a","label":"A"},{"id":"b","label":"B"}]}],"answers":[{"question_id":"q1","option_ids":[""]}]}`,
		},
		{
			name: "duplicate answer for question",
			in:   `{"questions":[{"id":"q1","question":"Pick one?","options":[{"id":"a","label":"A"},{"id":"b","label":"B"}]}],"answers":[{"question_id":"q1","option_ids":["a"]},{"question_id":"q1","option_ids":["b"]}]}`,
		},
		{
			name: "missing option label",
			in:   `{"questions":[{"id":"q1","question":"Pick one?","options":[{"id":"a","label":""},{"id":"b","label":"B"}]}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := (&AskUserQuestionTool{}).Execute(context.Background(), []byte(tt.in), types.ToolContext{})
			if err != nil {
				t.Fatalf("Execute error: %v", err)
			}
			if !res.IsError {
				t.Fatalf("expected validation error, got success: %s", res.Content)
			}
		})
	}
}
