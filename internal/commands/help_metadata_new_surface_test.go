package commands

import "testing"

func TestSuggestionsIncludeNewCommandMetadata(t *testing.T) {
	r := DefaultRegistry()
	cases := []struct {
		query string
		name  string
	}{
		{query: "bridge", name: "bridge"},
		{query: "autofix", name: "autofix-pr"},
		{query: "backfill", name: "backfill-sessions"},
		{query: "ultraplan", name: "ultraplan"},
	}
	for _, tc := range cases {
		items := r.Suggestions(tc.query)
		if len(items) == 0 {
			t.Fatalf("expected suggestions for %q", tc.query)
		}
		found := false
		for _, item := range items {
			if item.Name == tc.name {
				if item.Category == "" || item.HelpHint == "" {
					t.Fatalf("missing metadata fields for %s: %+v", tc.name, item)
				}
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected %q suggestion in query %q", tc.name, tc.query)
		}
	}
}

func TestSuggestionsIncludeSkillsOperationalHints(t *testing.T) {
	r := DefaultRegistry()
	items := r.Suggestions("skills")
	if len(items) == 0 {
		t.Fatalf("expected suggestions for skills")
	}
	for _, item := range items {
		if item.Name != "skills" {
			continue
		}
		if item.ArgumentHint == "" || item.HelpHint == "" || len(item.Diagnostics) == 0 {
			t.Fatalf("expected rich skills metadata, got %+v", item)
		}
		if len(item.Examples) == 0 {
			t.Fatalf("expected skills examples in metadata")
		}
		return
	}
	t.Fatalf("expected skills suggestion entry")
}
