package tui

import (
	"strings"
	"testing"
)

func TestStageHSearchTextKeepsPlainAssistantText(t *testing.T) {
	row := timelineEntry{kind: timelineAssistant, text: "Status ready and stable."}
	got := timelineSearchText(row)
	if !strings.Contains(got, "Status ready and stable.") {
		t.Fatalf("expected plain assistant text in search corpus, got %q", got)
	}
}
