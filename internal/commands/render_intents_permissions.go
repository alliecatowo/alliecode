package commands

import (
	"fmt"

	"github.com/alliecatowo/alliecode/internal/types"
)

func permissionsModeIntents(mode string) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Permissions", "Current permission mode.", field("Mode", mode))}
}

func permissionsModesIntents(modes []string) []types.RenderIntent {
	rows := make([]types.RenderTableRow, 0, len(modes))
	items := make([]types.RenderChecklistItem, 0, len(modes))
	for _, mode := range modes {
		rows = append(rows, tableRow(mode))
		items = append(items, checklistItem(mode, true, "available mode"))
	}
	return []types.RenderIntent{
		tableIntent("Permission modes", "Available permission modes.", []string{"Mode"}, rows...),
		checklistIntent("Permission modes checklist", "Checklist compatibility view for available modes.", items...),
	}
}

func permissionsSummaryIntents(state *RuntimeState, groupedDenials []string) []types.RenderIntent {
	return []types.RenderIntent{
		summaryCardIntent("Permissions summary", "Active mode, rules, and denial counts.",
			field("Mode", modeString(state.PermissionMode)),
			field("Rules", fmt.Sprintf("%d", len(state.PermissionRules))),
			field("Denials", fmt.Sprintf("%d", len(state.PermissionDenials))),
			field("Groups", fmt.Sprintf("%d", len(groupedDenials))),
		),
		actionHintsIntent("Actions", hint("Show rules", "/permissions rules"), hint("Show denials", "/permissions denials")),
	}
}

func permissionRulesIntents(rules []string) []types.RenderIntent {
	items := make([]types.RenderChecklistItem, 0, len(rules))
	for _, rule := range rules {
		items = append(items, checklistItem(rule, true, "active rule"))
	}
	return []types.RenderIntent{checklistIntent("Permission rules", "Ordered active rules.", items...)}
}

func permissionDenialsIntents(denials []PermissionDenial, groups []string) []types.RenderIntent {
	groupRows := make([]types.RenderGroup, 0, len(groups)+1)
	if len(groups) > 0 {
		groupRows = append(groupRows, group("Groups", "Grouped by denial class.", groups...))
	}
	items := make([]string, 0, len(denials))
	for _, denial := range denials {
		items = append(items, fmt.Sprintf("%s (%s)", denial.Command, denial.Reason))
	}
	groupRows = append(groupRows, group("Commands", "Denied commands.", items...))
	return []types.RenderIntent{groupedListIntent("Permission denials", "Recently denied permission requests.", groupRows...)}
}

func permissionRetryIntents(commands []string) []types.RenderIntent {
	items := make([]types.RenderChecklistItem, 0, len(commands))
	for _, command := range commands {
		items = append(items, checklistItem(command, false, "retry candidate"))
	}
	return []types.RenderIntent{checklistIntent("Retry denied commands", "Commands eligible for retry.", items...)}
}
