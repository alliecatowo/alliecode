package session

import "testing"

func TestQuerySummaryDistinctFields(t *testing.T) {
	items := []Transcript{
		{SessionID: "s1", ProjectPath: "/repo/a", Provider: "openai", Model: "gpt", MessageCount: 1, EventCount: 2},
		{SessionID: "s2", ProjectPath: "/repo/b", Provider: "anthropic", Model: "claude", MessageCount: 2, EventCount: 4},
	}
	s := summarizeTranscripts(items)
	if s.DistinctProjects != 2 || s.DistinctModels != 2 || s.DistinctSessions != 2 {
		t.Fatalf("unexpected summary: %+v", s)
	}
}
