package history

import (
	"testing"
	"time"
)

func TestBuildIndexFiltersAndSorts(t *testing.T) {
	events := []Event{
		{Type: EventCustom, SessionID: "a", Project: "/repo/a", Display: "build", Provider: "openai", Model: "gpt", Timestamp: time.Unix(1, 0).UTC()},
		{Type: EventCustom, SessionID: "b", Project: "/repo/b", Display: "deploy", Provider: "anthropic", Model: "claude", Timestamp: time.Unix(2, 0).UTC()},
	}
	items := BuildIndex(events, IndexQuery{ProjectContains: "/repo/b", DisplayContains: "dep", Limit: 1})
	if len(items) != 1 || items[0].SessionID != "b" {
		t.Fatalf("unexpected index result: %+v", items)
	}
}
