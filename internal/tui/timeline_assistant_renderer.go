package tui

import "strings"

type assistantRenderBoundary struct {
	body        string
	usedIntents bool
	usedRawText bool
}

func renderTimelineAssistantBody(row timelineEntry) assistantRenderBoundary {
	if rendered := strings.TrimSpace(renderIntentsPlain(row.intents)); rendered != "" {
		return assistantRenderBoundary{body: rendered, usedIntents: true}
	}
	body := strings.TrimSpace(row.text)
	if body == "" {
		return assistantRenderBoundary{}
	}
	if strings.Contains(body, "Deterministic runtime snapshot for the current session.") {
		body = "Status\n" + body
	}
	if strings.HasPrefix(body, "Provider  | Model") {
		body = "All models\n" + body
	}
	return assistantRenderBoundary{body: body, usedRawText: true}
}
