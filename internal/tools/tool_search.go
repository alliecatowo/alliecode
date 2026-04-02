package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/alliecatowo/alliecode/internal/types"
)

// ToolSearchTool searches registered tool metadata.
type ToolSearchTool struct{}

type toolSearchInput struct {
	Query      string `json:"query"`
	MaxResults int    `json:"max_results"`
}

type toolSearchOutput struct {
	Query              string          `json:"query"`
	Matches            []types.ToolDef `json:"matches"`
	TotalDeferredTools int             `json:"total_deferred_tools"`
}

func (t *ToolSearchTool) Name() string { return "tool_search" }

func (t *ToolSearchTool) Description() string {
	return "Searches registered tool names and descriptions."
}

func (t *ToolSearchTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"query": {
				Type:        "string",
				Description: "Keyword query or select:<tool-name>",
			},
			"max_results": {
				Type:        "integer",
				Description: "Maximum number of matches to return.",
			},
		},
		Required: []string{"query"},
	}
}

func (t *ToolSearchTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	var in toolSearchInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Invalid input: %v", err), IsError: true}, nil
	}
	q := strings.TrimSpace(in.Query)
	if q == "" {
		return types.ToolResult{Content: "query is required", IsError: true}, nil
	}
	max := in.MaxResults
	if max <= 0 {
		max = 5
	}

	defs := registeredToolDefsSnapshot()
	matches := findToolMatches(defs, q, max)
	out := toolSearchOutput{
		Query:              q,
		Matches:            matches,
		TotalDeferredTools: len(defs),
	}
	b, err := json.Marshal(out)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}
	return types.ToolResult{Content: string(b)}, nil
}

func findToolMatches(defs []types.ToolDef, query string, max int) []types.ToolDef {
	selectPrefix := "select:"
	if strings.HasPrefix(strings.ToLower(query), selectPrefix) {
		name := strings.TrimSpace(query[len(selectPrefix):])
		for _, d := range defs {
			if strings.EqualFold(d.Name, name) {
				return []types.ToolDef{d}
			}
		}
		return []types.ToolDef{}
	}

	terms := strings.Fields(strings.ToLower(query))
	type scoredDef struct {
		def           types.ToolDef
		score         int
		matchedTerms  int
		exactName     bool
		prefixName    bool
		descriptionID int
	}
	scored := make([]scoredDef, 0, len(defs))
	for _, d := range defs {
		name := strings.ToLower(d.Name)
		desc := strings.ToLower(d.Description)
		score := 0
		matched := 0
		for _, term := range terms {
			termMatched := false
			if strings.Contains(name, term) {
				score += 3
				termMatched = true
			}
			if strings.Contains(desc, term) {
				score++
				termMatched = true
			}
			if termMatched {
				matched++
			}
		}
		exactName := name == strings.ToLower(strings.TrimSpace(query))
		prefixName := strings.HasPrefix(name, strings.ToLower(strings.TrimSpace(query)))
		if exactName {
			score += 50
		} else if prefixName {
			score += 10
		}
		if score > 0 {
			scored = append(scored, scoredDef{
				def:           d,
				score:         score,
				matchedTerms:  matched,
				exactName:     exactName,
				prefixName:    prefixName,
				descriptionID: len(d.Description) + len(strconv.Itoa(len(d.InputSchema.Properties))),
			})
		}
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		if scored[i].matchedTerms != scored[j].matchedTerms {
			return scored[i].matchedTerms > scored[j].matchedTerms
		}
		if scored[i].exactName != scored[j].exactName {
			return scored[i].exactName
		}
		if scored[i].prefixName != scored[j].prefixName {
			return scored[i].prefixName
		}
		if scored[i].descriptionID != scored[j].descriptionID {
			return scored[i].descriptionID > scored[j].descriptionID
		}
		return scored[i].def.Name < scored[j].def.Name
	})

	if len(scored) > max {
		scored = scored[:max]
	}
	res := make([]types.ToolDef, 0, len(scored))
	for _, item := range scored {
		res = append(res, item.def)
	}
	return res
}

func (t *ToolSearchTool) IsReadOnly(input types.ToolInput) bool { return true }

func (t *ToolSearchTool) IsDestructive(input types.ToolInput) bool { return false }

func (t *ToolSearchTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *ToolSearchTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return types.PermissionAllowed
}
