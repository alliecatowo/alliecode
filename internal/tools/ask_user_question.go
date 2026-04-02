package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/alliecatowo/alliecode/internal/types"
)

// AskUserQuestionTool carries structured user-choice prompts and answers.
type AskUserQuestionTool struct{}

type askQuestionOption struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Preview     string `json:"preview,omitempty"`
}

type askQuestion struct {
	ID          string              `json:"id"`
	Question    string              `json:"question"`
	Header      string              `json:"header,omitempty"`
	Options     []askQuestionOption `json:"options"`
	MultiSelect bool                `json:"multi_select,omitempty"`
}

type askQuestionAnswer struct {
	QuestionID string   `json:"question_id"`
	OptionIDs  []string `json:"option_ids"`
	Notes      string   `json:"notes,omitempty"`
}

type askQuestionResolvedOption struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Preview string `json:"preview,omitempty"`
}

type askQuestionResolvedAnswer struct {
	QuestionID string                      `json:"question_id"`
	OptionIDs  []string                    `json:"option_ids"`
	Options    []askQuestionResolvedOption `json:"options"`
	Notes      string                      `json:"notes,omitempty"`
}

type askUserQuestionInput struct {
	Questions []askQuestion       `json:"questions"`
	Answers   []askQuestionAnswer `json:"answers,omitempty"`
	Metadata  map[string]any      `json:"metadata,omitempty"`
}

type askUserQuestionSummary struct {
	QuestionCount        int      `json:"question_count"`
	AnsweredCount        int      `json:"answered_count"`
	PendingCount         int      `json:"pending_count"`
	PendingQuestionIDs   []string `json:"pending_question_ids,omitempty"`
	AnsweredQuestionIDs  []string `json:"answered_question_ids,omitempty"`
	AllQuestionsAnswered bool     `json:"all_questions_answered"`
}

type askUserQuestionResult struct {
	Status          string                      `json:"status"`
	Questions       []askQuestion               `json:"questions"`
	Answers         []askQuestionAnswer         `json:"answers,omitempty"`
	ResolvedAnswers []askQuestionResolvedAnswer `json:"resolved_answers,omitempty"`
	Summary         askUserQuestionSummary      `json:"summary"`
	Metadata        map[string]any              `json:"metadata,omitempty"`
}

func (t *AskUserQuestionTool) Name() string { return "AskUserQuestion" }

func (t *AskUserQuestionTool) Description() string {
	return "Asks structured multiple-choice questions and returns a machine-readable answer payload."
}

func (t *AskUserQuestionTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"questions": {
				Type:        "array",
				Description: "Questions to present to the user.",
				Items: &types.PropertySchema{
					Type: "object",
				},
			},
			"answers": {
				Type:        "array",
				Description: "Optional user answers keyed by question id.",
				Items: &types.PropertySchema{
					Type: "object",
				},
			},
			"metadata": {
				Type:        "object",
				Description: "Optional metadata map for callers.",
			},
		},
		Required: []string{"questions"},
	}
}

func (t *AskUserQuestionTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	var in askUserQuestionInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Invalid input: %v", err), IsError: true}, nil
	}
	if len(in.Questions) == 0 {
		return types.ToolResult{Content: "questions must not be empty", IsError: true}, nil
	}

	questionByID := make(map[string]askQuestion, len(in.Questions))
	for i, q := range in.Questions {
		if strings.TrimSpace(q.ID) == "" {
			return types.ToolResult{Content: fmt.Sprintf("questions[%d].id is required", i), IsError: true}, nil
		}
		q.ID = strings.TrimSpace(q.ID)
		if _, ok := questionByID[q.ID]; ok {
			return types.ToolResult{Content: fmt.Sprintf("duplicate question id %q", q.ID), IsError: true}, nil
		}
		if strings.TrimSpace(q.Question) == "" {
			return types.ToolResult{Content: fmt.Sprintf("questions[%d].question is required", i), IsError: true}, nil
		}
		q.Question = strings.TrimSpace(q.Question)
		q.Header = strings.TrimSpace(q.Header)
		if len(q.Options) < 2 {
			return types.ToolResult{Content: fmt.Sprintf("questions[%d].options must have at least 2 options", i), IsError: true}, nil
		}
		seen := make(map[string]struct{}, len(q.Options))
		for j, opt := range q.Options {
			if strings.TrimSpace(opt.ID) == "" {
				return types.ToolResult{Content: fmt.Sprintf("questions[%d].options[%d].id is required", i, j), IsError: true}, nil
			}
			opt.ID = strings.TrimSpace(opt.ID)
			if strings.TrimSpace(opt.Label) == "" {
				return types.ToolResult{Content: fmt.Sprintf("questions[%d].options[%d].label is required", i, j), IsError: true}, nil
			}
			opt.Label = strings.TrimSpace(opt.Label)
			opt.Description = strings.TrimSpace(opt.Description)
			opt.Preview = strings.TrimSpace(opt.Preview)
			if _, ok := seen[opt.ID]; ok {
				return types.ToolResult{Content: fmt.Sprintf("questions[%d] has duplicate option id %q", i, opt.ID), IsError: true}, nil
			}
			seen[opt.ID] = struct{}{}
			q.Options[j] = opt
		}
		questionByID[q.ID] = q
		in.Questions[i] = q
	}

	seenQuestionAnswers := make(map[string]struct{}, len(in.Answers))
	resolved := make([]askQuestionResolvedAnswer, 0, len(in.Answers))
	for i, ans := range in.Answers {
		if strings.TrimSpace(ans.QuestionID) == "" {
			return types.ToolResult{Content: fmt.Sprintf("answers[%d].question_id is required", i), IsError: true}, nil
		}
		ans.QuestionID = strings.TrimSpace(ans.QuestionID)
		if _, dup := seenQuestionAnswers[ans.QuestionID]; dup {
			return types.ToolResult{Content: fmt.Sprintf("duplicate answer for question_id %q", ans.QuestionID), IsError: true}, nil
		}
		seenQuestionAnswers[ans.QuestionID] = struct{}{}

		q, ok := questionByID[ans.QuestionID]
		if !ok {
			return types.ToolResult{Content: fmt.Sprintf("answers[%d] references unknown question_id %q", i, ans.QuestionID), IsError: true}, nil
		}
		if len(ans.OptionIDs) == 0 {
			return types.ToolResult{Content: fmt.Sprintf("answers[%d].option_ids must not be empty", i), IsError: true}, nil
		}
		if !q.MultiSelect && len(ans.OptionIDs) > 1 {
			return types.ToolResult{Content: fmt.Sprintf("answers[%d] has multiple options for single-select question %q", i, q.ID), IsError: true}, nil
		}
		allowed := make(map[string]struct{}, len(q.Options))
		optionsByID := make(map[string]askQuestionOption, len(q.Options))
		for _, opt := range q.Options {
			allowed[opt.ID] = struct{}{}
			optionsByID[opt.ID] = opt
		}
		seenOpt := make(map[string]struct{}, len(ans.OptionIDs))
		normalizedIDs := make([]string, 0, len(ans.OptionIDs))
		for _, optID := range ans.OptionIDs {
			if strings.TrimSpace(optID) == "" {
				return types.ToolResult{Content: fmt.Sprintf("answers[%d].option_ids must not contain empty values", i), IsError: true}, nil
			}
			optID = strings.TrimSpace(optID)
			if _, dup := seenOpt[optID]; dup {
				return types.ToolResult{Content: fmt.Sprintf("answers[%d] contains duplicate option id %q", i, optID), IsError: true}, nil
			}
			seenOpt[optID] = struct{}{}
			if _, ok := allowed[optID]; !ok {
				return types.ToolResult{Content: fmt.Sprintf("answers[%d] contains unknown option id %q", i, optID), IsError: true}, nil
			}
			normalizedIDs = append(normalizedIDs, optID)
		}
		ans.OptionIDs = normalizedIDs

		selected := make([]askQuestionResolvedOption, 0, len(ans.OptionIDs))
		for _, optID := range ans.OptionIDs {
			opt := optionsByID[optID]
			selected = append(selected, askQuestionResolvedOption{ID: opt.ID, Label: opt.Label, Preview: opt.Preview})
		}
		sort.SliceStable(selected, func(i, j int) bool { return selected[i].ID < selected[j].ID })
		resolved = append(resolved, askQuestionResolvedAnswer{
			QuestionID: ans.QuestionID,
			OptionIDs:  ans.OptionIDs,
			Options:    selected,
			Notes:      strings.TrimSpace(ans.Notes),
		})
		in.Answers[i] = ans
	}

	answeredQuestionIDs := make([]string, 0, len(in.Answers))
	for _, ans := range in.Answers {
		answeredQuestionIDs = append(answeredQuestionIDs, ans.QuestionID)
	}
	sort.Strings(answeredQuestionIDs)

	pendingQuestionIDs := make([]string, 0, len(in.Questions))
	for _, q := range in.Questions {
		if _, ok := seenQuestionAnswers[q.ID]; !ok {
			pendingQuestionIDs = append(pendingQuestionIDs, q.ID)
		}
	}
	sort.Strings(pendingQuestionIDs)

	result := askUserQuestionResult{
		Questions: in.Questions,
		Summary: askUserQuestionSummary{
			QuestionCount:        len(in.Questions),
			AnsweredCount:        len(in.Answers),
			PendingCount:         len(pendingQuestionIDs),
			PendingQuestionIDs:   pendingQuestionIDs,
			AnsweredQuestionIDs:  answeredQuestionIDs,
			AllQuestionsAnswered: len(pendingQuestionIDs) == 0,
		},
	}
	if len(in.Metadata) > 0 {
		result.Metadata = in.Metadata
	}

	if len(in.Answers) == 0 {
		result.Status = "needs_user_input"
	} else {
		if len(pendingQuestionIDs) > 0 {
			result.Status = "partial"
		} else {
			result.Status = "answered"
		}
		result.Answers = in.Answers
		result.ResolvedAnswers = resolved
	}

	b, err := json.Marshal(result)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}
	return types.ToolResult{Content: string(b)}, nil
}

func (t *AskUserQuestionTool) IsReadOnly(input types.ToolInput) bool { return true }

func (t *AskUserQuestionTool) IsDestructive(input types.ToolInput) bool { return false }

func (t *AskUserQuestionTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *AskUserQuestionTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return types.PermissionAsk
}
