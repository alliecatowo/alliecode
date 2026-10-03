package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStageJIntentFamilyPanelTimelineMatrixNoLegacyContract(t *testing.T) {
	type matrixCase struct {
		name   string
		intent types.RenderIntent
		tokens []string
		legacy string
	}

	cases := []matrixCase{
		{
			name:   "summary_card",
			intent: types.RenderIntent{Kind: types.RenderIntentSummaryCard, Title: "Status", Fields: []types.RenderField{{Label: "Provider", Value: "openai"}}},
			tokens: []string{"Status", "Provider: openai"},
			legacy: "STATUS_REPORT",
		},
		{
			name:   "grouped_list",
			intent: types.RenderIntent{Kind: types.RenderIntentGroupedList, Title: "Tasks", Groups: []types.RenderGroup{{Title: "Open", Items: []string{"lane5"}}}},
			tokens: []string{"Tasks", "Open", "- lane5"},
			legacy: "TASKS_LIST",
		},
		{
			name:   "checklist",
			intent: types.RenderIntent{Kind: types.RenderIntentChecklist, Title: "Startup chain", Items: []types.RenderChecklistItem{{Label: "login", Done: true, Detail: "openai"}}},
			tokens: []string{"Startup chain", "[x] login - openai"},
			legacy: "CHECKLIST_REPORT",
		},
		{
			name:   "table",
			intent: types.RenderIntent{Kind: types.RenderIntentTable, Title: "Model matrix", Columns: []string{"Provider", "Model"}, Rows: []types.RenderTableRow{{Cells: []string{"openai", "gpt-4o"}}}},
			tokens: []string{"Model matrix", "Provider | Model", "openai"},
			legacy: "MODEL_LIST_ALL",
		},
		{
			name:   "detail_rows",
			intent: types.RenderIntent{Kind: types.RenderIntentDetailRows, Title: "Runtime", DetailRows: []types.RenderDetailRow{{Label: "Provider", Value: "anthropic", Status: "ready", Detail: "resolved at send"}}},
			tokens: []string{"Runtime", "Provider: anthropic [ready] - resolved at send"},
			legacy: "RUNTIME_STATUS",
		},
		{
			name:   "option_list",
			intent: types.RenderIntent{Kind: types.RenderIntentOptionList, Title: "Providers", Options: []types.RenderOption{{Label: "openai", Selected: true, Status: "ready", Detail: "primary", Hint: "/provider set openai"}}},
			tokens: []string{"Providers", "* openai [ready]: primary -> /provider set openai"},
			legacy: "PROVIDER_OPTIONS",
		},
		{
			name:   "action_list",
			intent: types.RenderIntent{Kind: types.RenderIntentActionList, Title: "Follow-ups", Actions: []types.RenderAction{{Label: "Run status", Command: "/status", Status: "ready", Detail: "verify runtime"}}},
			tokens: []string{"Follow-ups", "- Run status: /status [ready] - verify runtime"},
			legacy: "ACTION_HINTS",
		},
		{
			name:   "diagnostics",
			intent: types.RenderIntent{Kind: types.RenderIntentDiagnostics, Title: "Doctor", Diagnostics: []types.RenderDiagnostic{{Label: "Provider auth", Status: "warn", Detail: "relogin required"}}, Hints: []types.RenderActionHint{{Label: "Fix", Command: "/login provider openai"}}},
			tokens: []string{"Doctor", "Provider auth [warn]: relogin required", "- Fix: /login provider openai"},
			legacy: "DOCTOR_REPORT",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			app := readySizedApp(t, 180, 30)
			intents := []types.RenderIntent{tc.intent}
			app.commandPanel.activate(commands.InteractivePanel{Title: "panel", Items: []commands.InteractivePanelItem{{Label: tc.name, Status: "item", PreviewIntents: intents, Preview: []string{tc.legacy}}}}, "")

			panel := stripANSIForTest(app.renderCommandPanel())
			timeline, _, _, _, _ := renderTimeline([]timelineEntry{{kind: timelineAssistant, turn: 1, text: tc.legacy, intents: intents}}, "", 120)
			timeline = stripANSIForTest(timeline)

			for _, token := range tc.tokens {
				if !strings.Contains(panel, token) {
					t.Fatalf("panel missing %q\n%s", token, panel)
				}
				if !strings.Contains(timeline, token) {
					t.Fatalf("timeline missing %q\n%s", token, timeline)
				}
			}
			if strings.Contains(panel, tc.legacy) {
				t.Fatalf("panel unexpectedly contains legacy token %q\n%s", tc.legacy, panel)
			}
			if strings.Contains(timeline, tc.legacy) {
				t.Fatalf("timeline unexpectedly contains legacy token %q\n%s", tc.legacy, timeline)
			}
		})
	}
}
