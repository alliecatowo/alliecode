package commands

import "testing"

func TestSuggestionMetadataRichnessFieldsPresent(t *testing.T) {
	r := DefaultRegistry()
	items := r.Suggestions("provider")
	if len(items) == 0 {
		t.Fatalf("expected provider suggestion")
	}
	if items[0].Group == "" || items[0].HelpHint == "" {
		t.Fatalf("expected group/help metadata: %+v", items[0])
	}
	if len(items[0].Shortcuts) == 0 {
		t.Fatalf("expected shortcut metadata")
	}
}
